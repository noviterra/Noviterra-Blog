package repo

import (
	"Noviterra/server/app/model"
	"Noviterra/server/utils/common"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CommentRepository struct {
	db          *gorm.DB
	commonUtils *common.CommonUtils
}

func NewCommentRepository(repoConfig RepoConfig, field string,
	commonUtils *common.CommonUtils) *CommentRepository {
	return &CommentRepository{
		db:          NewDB(repoConfig, field),
		commonUtils: commonUtils,
	}
}

// BatchUpsertSyncStatus 批量ORM操作更新 点赞状态与收藏状态
// 事务进行article_stat与user_like表的操作
// 相对上面的两个方法（UpsertUserLikes与UpsertArticleStatus）这里事务保障原子一致性 对比两边数据进行最小的变更 只对变化的数据操作
func (cr *CommentRepository) BatchUpsertSyncStatus(ctx context.Context,
	userLikedArtcles map[string][]int64,
	articleLikeCount map[int64]int64,
	userCollectArticles map[string][]int64,
	articleCollect map[int64]int64,
	articleViewCount map[int64]int64,
) error {
	// 依赖于redis中的数据 传参为redis清洗之后的数据
	return cr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 同步用户点赞记录
		// 实现步骤：先获取MySQL中用户的点赞数据 与redis进行对比 筛选出其中一方没有的
		// redis中有MySQL中没有：插入数据
		// redis中没有MySQL中有：删除数据
		// redis与MySQL中都有/没有：不操作
		for userID, redisArticleIDs := range userLikedArtcles {
			// 查询MySQL当前的点赞记录
			var dbLikeArticles []int64
			if err := tx.Model(&model.UserLike{}).
				Where("user_id = ? AND like_status = 1", userID).
				Pluck("article_id", &dbLikeArticles).Error; err != nil {
				return fmt.Errorf("查询用户%s的点赞文章列表失败：%w", userID, err)
			}
			redisLikeSet := cr.commonUtils.ToSet(redisArticleIDs)
			dbLikeSet := cr.commonUtils.ToSet(dbLikeArticles)
			// 新增：redis中有 MySQL中没有
			var toAddLike []model.UserLike
			for aid := range redisLikeSet {
				if !dbLikeSet[aid] {
					toAddLike = append(toAddLike, model.UserLike{
						UserID:     userID,
						ArticleID:  aid,
						LikeStatus: 1,
					})
				}
			}
			// 需要取消的 DB有但redis中没有
			var toLikeRemove []int64
			for aid := range dbLikeSet {
				if !redisLikeSet[aid] {
					toLikeRemove = append(toLikeRemove, aid)
				}
			}
			// 批量新增
			if len(toAddLike) > 0 {
				if err := tx.Clauses(clause.OnConflict{
					Columns: []clause.Column{
						{Name: "user_id"},
						{Name: "article_id"},
					},
					DoUpdates: clause.AssignmentColumns([]string{"like_status", "updated_at"}),
				}).Create(&toAddLike).Error; err != nil {
					return fmt.Errorf("新增用户%s的点赞记录失败：%w", userID, err)
				}
			}
			// 批量取消
			if len(toLikeRemove) > 0 {
				if err := tx.Model(&model.UserLike{}).
					Where("user_id = ? AND article_id IN ?", userID, toLikeRemove).
					Update("like_status", 0).Error; err != nil {
					return fmt.Errorf("取消用户%s的点赞记录失败：%w", userID, err)
				}
			}
		}

		// 同步用户文章点赞数据 操作同理
		for userID, redisArticles := range userCollectArticles {
			var dbCollectArticles []int64
			if err := tx.Model(&model.UserLike{}).
				Where("user_id = ? AND collect_status = 1", userID).
				Pluck("article_id", &dbCollectArticles).Error; err != nil {
				return fmt.Errorf("查询用户%s收藏的文章列表失败：%w", userID, err)
			}
			redisCollectArticleSet := cr.commonUtils.ToSet(redisArticles)  // 获取redis读取的数据的全部article_id
			dbCollectArticleSet := cr.commonUtils.ToSet(dbCollectArticles) // 获取MySQL中存在收藏状态的全部article_id

			// 开始对比两边数据 以redis为主 redis无则移除 mysql无则增加 都有的数据不变
			var toAddCollect []model.UserLike
			for articleID := range redisCollectArticleSet {
				if !dbCollectArticleSet[articleID] {
					toAddCollect = append(toAddCollect, model.UserLike{
						UserID:        userID,
						ArticleID:     articleID,
						CollectStatus: 1,
					})
				}
			}
			var toRemoveCollect []int64
			for articleID := range dbCollectArticleSet {
				if !redisCollectArticleSet[articleID] {
					toRemoveCollect = append(toRemoveCollect, articleID)
				}
			}

			// 批量新增
			if len(toAddCollect) > 0 {
				if err := tx.Clauses(clause.OnConflict{
					Columns: []clause.Column{
						{Name: "user_id"},
						{Name: "article_id"},
					},
					DoUpdates: clause.AssignmentColumns([]string{"collect_status", "updated_at"}),
				}).Create(&toAddCollect).Error; err != nil {
					return fmt.Errorf("新增用户%s收藏记录失败：%w", userID, err)
				}
			}

			if len(toRemoveCollect) > 0 {
				if err := tx.Model(&model.UserLike{}).
					Where("user_id = ? AND article_id in ?", userID, toRemoveCollect).
					Update("collect_status", 0).Error; err != nil {
					return fmt.Errorf("移除用户%s收藏文章的状态失败：%w", userID, err)
				}
			}
		}

		// 同步文章点赞总数
		for articleID, count := range articleLikeCount {
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "article_id"},
				},
				DoUpdates: clause.AssignmentColumns([]string{"like_count"}),
			}).Create(&model.ArticleStat{
				ArticleID: articleID,
				LikeCount: count,
			}).Error; err != nil {
				return fmt.Errorf("文章%s点赞总数更新失败：%w", articleID, err)
			}
		}

		// 同步文章收藏总数
		for articleID, count := range articleCollect {
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "article_id"},
				},
				DoUpdates: clause.AssignmentColumns([]string{"collect_count"}),
			}).Create(&model.ArticleStat{
				ArticleID:    articleID,
				CollectCount: count,
			}).Error; err != nil {
				return fmt.Errorf("文章%s收藏总数更新失败：%w", articleID, err)
			}
		}

		// 同步文章浏览总数
		for articleID, count := range articleViewCount {
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "article_id"},
				},
				DoUpdates: clause.Assignments(map[string]interface{}{
					"view_count": gorm.Expr("view_count + ?", count),
				}),
			}).Create(&model.ArticleStat{
				ArticleID: articleID,
				ViewCount: count, // 记录不存在时使用它作为初始值
			}).Error; err != nil {
				return fmt.Errorf("文章%s浏览总数更新失败：%w", articleID, err)
			}
		}

		return nil
	})
}

// InsertComment 插入一条评论
func (cr *CommentRepository) InsertComment(ctx context.Context, comment *model.ArticleComment) (int64, error) {
	var rootID int64
	err := cr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if comment.ParentID <= 0 { // 父节点为空 表示该为顶层评论
			comment.ParentID = 0
			comment.RootID = 0
		} else {
			// 子回复 开始查询父评论
			var parent model.ArticleComment
			if err := tx.Where("comment_id = ? AND status = ?", comment.ParentID, 1).
				First(&parent).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("回复的评论不存在或已删除！")
				}
				return fmt.Errorf("查询父评论失败：%w", err)
			}
			comment.RootID = parent.RootID // root_id集成父节点的
			comment.Depth = parent.Depth + 1
			rootID = parent.RootID
		}

		// 插入操作
		if err := tx.Create(comment).Error; err != nil {
			return fmt.Errorf("评论插入失败：%w", err)
		}

		// 回填RootID
		if comment.RootID == 0 {
			rootID = comment.CommentID
			if err := tx.Model(comment).
				Update("root_id", comment.CommentID).Error; err != nil {
				return fmt.Errorf("回填RootID失败: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return rootID, nil
}

// FirstLoadTopComment 首次进入文章后 加载顶层评论
func (cr *CommentRepository) FirstLoadTopComment(ctx context.Context,
	address int64) ([]model.TopCommentNode, error) {
	var comments []model.TopCommentNode
	err := cr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.ArticleComment{}).
			Where("address = ? AND parent_id = ? AND status = ?",
				address, 0, 1).Order("created_at ASC").
			Find(&comments).Error; err != nil {
			return fmt.Errorf("评论加载失败：%w", err)
		}

		if len(comments) == 0 {
			return nil
		}
		// 统计每个顶层评论下有多少评论
		// 收集顶层评论ID
		topIDs := make([]int64, 0, len(comments)) // 预分配容量 直接make([]string, legth)导致长度翻倍
		for _, c := range comments {
			topIDs = append(topIDs, c.CommentID)
		}
		// 查出全部后代 后续继续比对parentID
		type child struct {
			CommentID int64
			ParentID  int64
		}
		var childs []child
		if err := tx.Model(&model.ArticleComment{}).
			Select("comment_id,parent_id").
			Where("address = ? AND status = ? AND parent_id <> ?",
				address, 1, 0).
			Find(&childs).Error; err != nil {
			return fmt.Errorf("统计子评论失败：%w", err)
		}

		// parent -> children
		children := make(map[int64][]int64, len(childs))
		for _, cd := range childs {
			children[cd.ParentID] = append(children[cd.ParentID], cd.CommentID)
		}

		// BFS 计算每个顶层评论的后代总数
		countMap := make(map[int64]int64, len(topIDs))
		for _, id := range topIDs {
			var cnt int64
			queue := []int64{id}
			for len(queue) > 0 {
				cur := queue[0]
				queue = queue[1:]
				for _, cd := range children[cur] {
					cnt++
					queue = append(queue, cd)
				}
			}
			countMap[id] = cnt
		}

		// 回填
		for i := range comments {
			comments[i].Count = countMap[comments[i].CommentID]
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(comments) == 0 {
		return []model.TopCommentNode{}, nil
	}

	return comments, nil
}

// OpenComment 展开评论 根据顶层评论ID得到全部评论
func (cr *CommentRepository) OpenComment(ctx context.Context,
	commentID int64) ([]model.CommentNode, error) {
	// 校验顶层评论的有效性
	var top model.TopCommentNode
	if err := cr.db.WithContext(ctx).
		Where("comment_id = ? AND status = ?", commentID, 1).
		First(&top).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("展开的评论不存在或已被删除！")
		}
		return nil, fmt.Errorf("校验顶层评论失败：%w", err)
	}

	var comments []model.CommentNode
	if err := cr.db.WithContext(ctx).
		Model(&model.ArticleComment{}).
		Where("root_id = ? AND status = ?", commentID, 1).
		Order("created_at ASC, comment_id ASC").
		Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("评论展开失败：%w", err)
	}
	if len(comments) == 0 {
		return []model.CommentNode{}, nil
	}

	return comments, nil
}
