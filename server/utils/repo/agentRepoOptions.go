package repo

import (
	"Noviterra/server/app/model"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const defaultAgentModelID = 1

type AgentRepository struct {
	db *gorm.DB
}

func NewAgentRepository(repoConfig RepoConfig, field string) *ArticleRepository {
	return &ArticleRepository{
		db: NewDB(repoConfig, field),
	}
}

// GetModelConfig 得到模型配置
func (ar *ArticleRepository) GetModelConfig(ctx context.Context) (*model.AgentModelConfig, error) {
	cfg := new(model.AgentModelConfig)
	if err := ar.db.WithContext(ctx).
		Model(&model.AgentModelConfig{}).
		Where("id = ?", defaultAgentModelID).
		First(cfg).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("站点助手尚未配置模型")
		}
		return nil, fmt.Errorf("读取站点助手模型配置失败：%w", err)
	}
	return cfg, nil
}

// SaveModelConfig 保存模型配置
func (ar *ArticleRepository) SaveModelConfig(ctx context.Context,
	cfg *model.AgentModelConfig) error {
	if cfg == nil {
		return errors.New("模型配置为空")
	}

	cfg.ID = defaultAgentModelID

	if err := ar.db.WithContext(ctx).
		Model(&model.AgentModelConfig{}).
		Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(cfg).
		Error; err != nil {
		return fmt.Errorf("保存站点助手模型配置失败：%w", err)
	}
	return nil
}
