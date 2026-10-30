package model

import (
	"time"

	"gorm.io/gorm"
)

type Notices struct {
	ID        int64          `gorm:"type:bigint;primaryKey;autoIncrement;comment:主键" json:"id"`
	Title     string         `gorm:"type:varchar(128);not null;comment:公告标题" json:"title"`
	Message   string         `gorm:"type:text;not null;comment:公告正文" json:"message"`
	Publisher string         `gorm:"type:varchar(100);comment:发布人" json:"publisher"`
	CreatedAt time.Time      `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"comment:更新时间" json:"updated_at"`
	IsDeleted bool           `gorm:"default:false;index;comment:是否删除" json:"is_deleted"`
	Deleter   string         `gorm:"type:varchar(100);comment:删除人" json:"deleter"`
	DeletedAt gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (Notices) TableName() string {
	return "notices"
}

// NoticeLookDTO 返回Notice可展示的字段
type NoticeLookDTO struct {
	Title     string    `gorm:"type:varchar(128);not null;comment:公告标题" json:"title"`
	Message   string    `gorm:"type:text;not null;comment:公告正文" json:"message"`
	Publisher string    `gorm:"type:varchar(100);comment:发布人" json:"publisher"`
	CreatedAt time.Time `gorm:"comment:创建时间" json:"created_at"`
}
