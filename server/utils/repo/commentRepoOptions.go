package repo

import (
	"Noviterra/server/app/model"
	"Noviterra/server/utils/common"
	"context"
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
	userLikedArtcles map[string][]string,
	articleLikeCount map[string]int64,
	userCollectArticles map[string][]string,
	articleCollect map[string]int64,
	articleViewCount map[string]int64,
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
			var dbLikeArticles []string
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
			var toLikeRemove []string
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
			var dbCollectArticles []string
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
			var toRemoveCollect []string
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
