package repo

import (
	"Noviterra/server/app/model"
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"
)

type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(repoConfig RepoConfig, field string) *ArticleRepository {
	return &ArticleRepository{
		db: NewDB(repoConfig, field),
	}
}

// PublishArticle 发布文章推动的事务开展 文章表的插入与文章其下的属性（点赞、收藏、访问量）
func (ar *ArticleRepository) PublishArticle(ctx context.Context, article *model.ArticleInfo,
	stat *model.ArticleStat) error {
	// 使用事务 保证两个操作共同失败或成功
	return ar.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(article).Error; err != nil {
			return fmt.Errorf("article表插入失败，id：%s，err：%s", article.ArticleID, err.Error())
		}
		if err := tx.Create(stat).Error; err != nil {
			return fmt.Errorf("stat表插入失败，id：%s，err：%s", stat.ArticleID, err.Error())
		}
		return nil
	})
}

func (ar *ArticleRepository) GetArticleList(ctx context.Context, page, pageSize int) ([]model.ArticleList, int64, error) {
	// 统计总数
	var (
		total int64
		list  []model.ArticleList
	)
	if err := ar.db.WithContext(ctx).
		Model(&model.ArticleInfo{}).
		Where("is_deleted = ?", false).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计文章总数失败！ERROR：%s", err.Error())
	}
	if total == 0 {
		return list, total, nil
	}

	// 分页查询 映射到ArticleList中
	if err := ar.db.WithContext(ctx).
		Model(&model.ArticleInfo{}).
		Select("article_id", "title", "cover", "information", "tags", "created_at").
		Where("is_deleted = ?", false).
		Order("created_at DESC, article_id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("文章查询失败！ERROR：%s", err.Error())
	}
	return list, total, nil
}

// GetArticleDetail 获取文章的详细信息
func (ar *ArticleRepository) GetArticleDetail(ctx context.Context, articleID int64) (*model.Article, error) {
	articleInfo := new(model.ArticleInfo)
	if err := ar.db.WithContext(ctx).
		Model(&model.ArticleInfo{}).
		Where("article_id = ? AND is_deleted = false", articleID).
		First(articleInfo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &model.Article{
		ArticleID:   articleInfo.ArticleID,
		PublisherID: articleInfo.PublisherID,
		Title:       articleInfo.Title,
		Cover:       articleInfo.Cover,
		Information: articleInfo.Information,
		Content:     articleInfo.Content,
		Tags:        articleInfo.Tags,
		CreatedAt:   articleInfo.CreatedAt.Unix(),
		UpdatedAt:   articleInfo.UpdatedAt.Unix(),
	}, nil
}

// GetPublisherIDByArticleID 根据文章ID返回该文章的发布者ID
func (ar *ArticleRepository) GetPublisherIDByArticleID(ctx context.Context, articleID int64) (string, error) {
	var publisherID string
	if err := ar.db.WithContext(ctx).
		Model(&model.ArticleInfo{}).
		Select("publisher_id").
		Where("article_id = ? AND is_deleted = ?", articleID, 0).
		Limit(1).
		Scan(&publisherID).
		Error; err != nil {
		return "", err
	}
	return publisherID, nil
}

// DeleteArticle 软删除 删除文章
func (ar *ArticleRepository) DeleteArticle(ctx context.Context, articleID int64, deleterID string) (bool, error) {
	result := ar.db.WithContext(ctx).
		Model(&model.ArticleInfo{}).
		Where("article_id = ? AND is_deleted = ?", articleID, 0).
		Updates(map[string]interface{}{
			"is_deleted": 1,
			"deleter_id": deleterID,
			"deleted_at": time.Now(),
		})
	if err := result.Error; err != nil {
		return false, err
	}
	if result.RowsAffected == 0 {
		return false, nil
	}
	return true, nil
}

// UpdateArticleInfo 更新文章（不含封面）
func (ar *ArticleRepository) UpdateArticleInfo(ctx context.Context, articleID int64, article *model.UpdateArticle) (bool, error) {
	result := ar.db.WithContext(ctx).
		Model(&model.ArticleInfo{}).
		Where("article_id = ? AND is_deleted = ?", articleID, 0).
		Updates(map[string]interface{}{
			"title":       article.Title,
			"information": article.Information,
			"content":     article.Content,
			"tags":        article.Tags,
		})
	if err := result.Error; err != nil {
		return false, err
	}
	if result.RowsAffected == 0 {
		return false, nil
	}

	return true, nil
}

// UpdateArticleCover 更新文章封面
func (ar *ArticleRepository) UpdateArticleCover(ctx context.Context, articleID int64, cover string) (bool, error) {
	result := ar.db.WithContext(ctx).
		Model(&model.ArticleInfo{}).
		Where("article_id = ? AND is_deleted = ?", articleID, 0).
		Update("cover", cover)
	if err := result.Error; err != nil {
		return false, err
	}
	if result.RowsAffected == 0 {
		return false, nil
	}
	return true, nil
}

// SearchArticle 查询文章 展示文章列表 模糊查询
func (ar *ArticleRepository) SearchArticle(ctx context.Context, like string, page, pageSize int) ([]model.ArticleList, int64, error) {
	base := func() *gorm.DB {
		return ar.db.WithContext(ctx).
			Where("(title LIKE ? ESCAPE '\\\\' OR information LIKE ? ESCAPE '\\\\' OR tags LIKE ? ESCAPE '\\\\')",
				like, like, like)
	}

	var total int64
	if err := base().
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计查询到的文章篇数失败！%w", err)
	}

	list := make([]model.ArticleList, 0, pageSize)
	if err := base().
		WithContext(ctx).
		Model(&model.ArticleInfo{}).
		Where("is_deleted = ?", 0).
		Order("article_id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("分页查询文章列表失败！%w", err)
	}
	return list, total, nil
}

// GetArchiveList 文章归档总结 UNION ALL + 数据库层排序
// todo: 获取未删除的文章 根据发布时间和更新时间进行排序 （同一篇文章的发布与更新如果不是同一时间 就判为两条归档数据）
// 此处为GORM实现 + 内置排序
func (ar *ArticleRepository) GetArchiveList(ctx context.Context) ([]model.ArticleArchive, error) {
	var publishList []model.ArticleArchive
	if err := ar.db.WithContext(ctx).
		Model(&model.ArticleInfo{}).
		Select("title, ? AS `option`, UNIX_TIMESTAMP(created_at) AS `created_at`", "publish").
		Where("is_deleted = ?", 0).
		Scan(&publishList).Error; err != nil {
		return nil, fmt.Errorf("查询文章发布操作失败！%w", err)
	}

	var updateList []model.ArticleArchive
	if err := ar.db.WithContext(ctx).
		Model(&model.ArticleInfo{}).
		Select("title, ? AS `option`, UNIX_TIMESTAMP(updated_at)", "update").
		Where("is_deleted = ?", 0).
		Where("updated_at <> created_at").
		Scan(&updateList).Error; err != nil {
		return nil, fmt.Errorf("查询文章更新操作失败！%w", err)
	}

	// 内置排序 不占用数据库内存
	list := make([]model.ArticleArchive, 0, len(publishList)+len(updateList))
	list = append(list, publishList...)
	list = append(list, updateList...)
	sort.Slice(list, func(i, j int) bool {
		return list[i].Time > list[j].Time
	})
	return list, nil
}

// CheckArticleExist 检查文章是否存在
func (ar *ArticleRepository) CheckArticleExist(ctx context.Context, articleID int64) (bool, error) {
	if articleID <= 0 {
		return false, nil
	}
	var count int64
	if err := ar.db.WithContext(ctx).
		Model(&model.ArticleInfo{}).
		Where("article_id = ?", articleID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetUserPublishArticleList 得到用户发布的文章
func (ar *ArticleRepository) GetUserPublishArticleList(ctx context.Context, page, pageSize int,
	userID string) ([]model.ArticleList, int64, error) {
	var (
		total  int64
		result []model.ArticleList
	)

	// 事务保证原子一致性
	err := ar.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 定义scope 复用查询条件
		scope := func(tx *gorm.DB) *gorm.DB {
			return tx.Where("publisher_id = ? AND is_deleted = ?", userID, false)
		}

		if err := tx.Model(&model.ArticleInfo{}).
			Scopes(scope).
			Count(&total).Error; err != nil {
			return fmt.Errorf("统计用户发布文章总数失败：%w", err)
		}
		if total == 0 {
			result = []model.ArticleList{}
			return nil
		}

		if err := tx.Model(&model.ArticleInfo{}).
			Select([]string{"article_id", "title", "cover", "information", "tags", "created_at"}).
			Scopes(scope).
			Order("created_at DESC, article_id DESC").
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Find(&result).Error; err != nil {
			return fmt.Errorf("用户发布文章查询失败：%w", err)
		}
		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

// GetArticleListByArticleID 根据文章ID列表批量得到多个文章列表数据
func (ar *ArticleRepository) GetArticleListByArticleID(ctx context.Context,
	articleIDs []int64, page, pageSize int) ([]model.ArticleList, int64, error) {
	if len(articleIDs) == 0 {
		return []model.ArticleList{}, 0, nil
	}

	var (
		total  int64
		result []model.ArticleList
	)

	err := ar.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		scope := func(tx *gorm.DB) *gorm.DB {
			return tx.Model(&model.ArticleInfo{}).
				Where("article_id IN ? AND is_deleted = ?", articleIDs, false)
		}
		if err := tx.Scopes(scope).
			Count(&total).Error; err != nil {
			return fmt.Errorf("查询文章列表数量失败：%w", err)
		}
		if total == 0 {
			result = []model.ArticleList{}
			return nil
		}

		if err := tx.Scopes(scope).
			Select([]string{"article_id", "title", "cover", "information", "tags", "created_at"}).
			Order("created_at DESC, article_id DESC").
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Find(&result).Error; err != nil {
			return fmt.Errorf("查询文章列表失败：%w", err)
		}

		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return result, total, nil
}
