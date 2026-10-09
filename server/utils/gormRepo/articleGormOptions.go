package gormrepo

import (
	"Noviterra/server/app/model"
	"context"

	"gorm.io/gorm"
)

type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(repoConfig RepoConfig) *ArticleRepository {
	return &ArticleRepository{
		db: NewDB(repoConfig),
	}
}

// CreateArticle 插入文章
func (ar *ArticleRepository) CreateArticle(ctx context.Context, article *model.ArticleInfo) error {
	return ar.db.WithContext(ctx).Create(article).Error
}
