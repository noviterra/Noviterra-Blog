package model

import "time"

type ArticleInfo struct {
	ArticleID   string   `gorm:"primaryKey;size:256;comment:文章ID（雪花ID）"`
	PublisherID string   `gorm:"size:64;index;not null;comment:发布者用户ID"`
	Title       string   `gorm:"size:128;not null;comment:文章标题"`
	Content     string   `gorm:"type:longtext;comment:文章正文"`
	Cover       string   `gorm:"size:255;comment:封面图URL"`
	Information string   `gorm:"size:256;comment:文章摘要"`
	Tags        []string `gorm:"type:json;comment:标签列表，JSON存储"`

	IsDeleted bool       `gorm:"default:false;index;not null;comment:是否删除"`
	DeleterID string     `gorm:"size:64;default:'';comment:删除者用户ID"`
	DeletedAt *time.Time `gorm:"index;comment:删除时间"`

	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (ArticleInfo) TableName() string {
	return "article"
}
