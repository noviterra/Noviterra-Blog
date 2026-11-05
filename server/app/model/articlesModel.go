package model

import (
	"time"
)

// ==================================================Models===========================================================

// ArticleInfo 文章属性
type ArticleInfo struct {
	ArticleID   int64    `gorm:"primaryKey;autoIncrement:false;comment:文章ID（雪花ID）"`
	PublisherID string   `gorm:"size:64;index;not null;comment:发布者用户ID"`
	Title       string   `gorm:"size:128;not null;comment:文章标题"`
	Content     string   `gorm:"type:longtext;comment:文章正文"`
	Cover       string   `gorm:"size:255;comment:封面图;default:'defaultCover.jpg'"`
	Information string   `gorm:"size:256;comment:文章摘要"`
	Tags        []string `gorm:"type:json;serializer:json;not null;comment:标签列表"`

	IsDeleted bool       `gorm:"default:false;index;not null;comment:是否删除"`
	DeleterID string     `gorm:"size:64;default:'';comment:删除者用户ID"`
	DeletedAt *time.Time `gorm:"index;comment:删除时间"`

	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (ArticleInfo) TableName() string {
	return "article"
}

// ==================================================DTO===========================================================

// ArticleList articlerepository中使用 仅作为映射列表信息的载体 (dto)
type ArticleList struct {
	ArticleID   int64
	Title       string
	Cover       string
	Information string
	Tags        []string
	CreatedAt   time.Time
}

// Article DTO articlecache中使用 存储前端中展示的内容
type Article struct {
	ArticleID   int64
	PublisherID string
	Title       string
	Cover       string
	Information string
	Content     string
	Tags        []string
	CreatedAt   int64
	UpdatedAt   int64
}

// UpdateArticle DTO 存储需要更新的文章字段（不含封面）
type UpdateArticle struct {
	Title       string
	Information string
	Content     string
	Tags        []string
}

type ArticleArchive struct {
	Title  string
	Action string
	Time   int64
}
