package repo

import (
	"Noviterra/server/app/model"
	"context"
	"time"

	"gorm.io/gorm"
)

type NoticeRepository struct {
	db *gorm.DB
}

func NewNoticeRepository(repoConfig RepoConfig, field string) *NoticeRepository {
	return &NoticeRepository{
		db: NewDB(repoConfig, field),
	}
}

func (nr *NoticeRepository) PublishNoticeInfo(ctx context.Context, notice model.Notices) error {
	// 部分字段初始化
	notice.DeletedAt = gorm.DeletedAt{}
	notice.Deleter = ""
	notice.IsDeleted = false

	// 插入操作
	if err := nr.db.WithContext(ctx).
		Model(&model.Notices{}).
		Create(&notice).Error; err != nil {
		return err
	}

	return nil
}

func (nr *NoticeRepository) GetNoticeInfo(ctx context.Context) ([]model.NoticeLookDTO, error) {
	var dtos []model.NoticeLookDTO

	senvenDaysAgo := time.Now().AddDate(0, 0, -7) // 七天内发布的公告

	if err := nr.db.WithContext(ctx).
		Model(&model.Notices{}).
		Where("created_at >= ? AND is_deleted = ?",
			senvenDaysAgo, false).
		Order("created_at DESC").
		Find(&dtos).
		Error; err != nil {
		return nil, err
	}

	return dtos, nil
}
