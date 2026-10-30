package model

import (
	"time"

	"gorm.io/gorm"
)

// FriendLink 友链记录表
type FriendLink struct {
	ID          int64          `gorm:"type:bigint;primaryKey;autoIncrement;comment:主键" json:"id"`
	Link        string         `gorm:"type:varchar(255);not null;comment:链接地址" json:"link"`
	LinkName    string         `gorm:"type:varchar(100);not null;comment:链接名称" json:"linkName"`
	Information string         `gorm:"type:varchar(500);comment:站点简介" json:"information"`
	Applicant   string         `gorm:"type:varchar(100);not null;index:idx_applicant;comment:申请人" json:"applicant"`
	Email       string         `gorm:"type:varchar(100);not null;comment:申请者联系邮箱" json:"email"`
	Status      int            `gorm:"type:bigint;not null;default:0;comment:友链当前状态 0申请中 1申请成功 2取消或已被删除"`
	CreatedAt   time.Time      `gorm:"comment:创建时间" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"comment:更新时间" json:"updatedAt"`
	IsCanceled  bool           `gorm:"type:tinyint(1);not null;default:0;index:idx_is_canceled;comment:是否已取消" json:"isCanceled"`
	Canceler    string         `gorm:"type:varchar(100);comment:取消人（美式拼写；英式为 Canceller）" json:"canceler"`
	DeletedAt   gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (FriendLink) TableName() string {
	return "friend_links"
}
