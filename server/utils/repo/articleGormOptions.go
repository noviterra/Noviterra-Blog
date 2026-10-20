package repo

import (
	"Noviterra/server/app/model"
	"context"
	"errors"
	"fmt"
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
func (ar *ArticleRepository) GetArticleDetail(ctx context.Context, articleID string) (*model.ArticleInfo, error) {
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
	return articleInfo, nil
}

// GetPublisherIDByArticleID 根据文章ID返回该文章的发布者ID
func (ar *ArticleRepository) GetPublisherIDByArticleID(ctx context.Context, articleID string) (string, error) {
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
func (ar *ArticleRepository) DeleteArticle(ctx context.Context, articleID string, deleterID string) (bool, error) {
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
func (ar *ArticleRepository) UpdateArticleInfo(ctx context.Context, articleID string, article *model.UpdateArticle) (bool, error) {
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
func (ar *ArticleRepository) UpdateArticleCover(ctx context.Context, articleID, cover string) (bool, error) {
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
