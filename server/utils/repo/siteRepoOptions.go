package repo

import (
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
	"context"

	"gorm.io/gorm"
)

type SiteRepository struct {
	db *gorm.DB
}

func NewSiteRepository(repoConfig RepoConfig, field string) *SiteRepository {
	return &SiteRepository{
		db: NewDB(repoConfig, field),
	}
}

// ApplyFriendLink 添加友链申请 申请的
func (sr *SiteRepository) ApplyFriendLink(ctx context.Context, linkInfo model.FriendLink) error {
	// 此方法为申请的函数 防止外部传入初始状态失败 此处强制初始化状态
	linkInfo.Status = 0
	linkInfo.IsCanceled = false
	linkInfo.Canceler = ""
	linkInfo.DeletedAt = gorm.DeletedAt{}

	// 使用事务操作 筛滤掉无效请求
	return sr.db.WithContext(ctx).
		Transaction(func(tx *gorm.DB) error {
			var count int64
			// 同一申请人不能有未处理的申请
			if err := tx.Model(&model.FriendLink{}).
				Where("applicant = ? AND is_canceled = ?", linkInfo.Applicant, false).
				Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return errcode.NewRepoError("您已有待处理的友链申请，请勿重复提交！")
			}

			// 同一链接不能重复申请
			if err := tx.Model(&model.FriendLink{}).
				Where("link = ? AND status IN ? AND is_canceled = ?",
					linkInfo.Link, []int{0, 1}, false).
				Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return errcode.NewRepoError("该链接已存在申请或已通过审核！")
			}

			// 插入
			return tx.Create(&linkInfo).Error
		})
}
