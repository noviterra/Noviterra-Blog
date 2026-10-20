package model

import (
	"time"
)

// ==================================================GORM Models===========================================================

type ArticleInfo struct {
	ArticleID   string   `gorm:"primaryKey;size:32;comment:文章ID（雪花ID）"`
	PublisherID string   `gorm:"size:64;index;not null;comment:发布者用户ID"`
	Title       string   `gorm:"size:128;not null;comment:文章标题"`
	Content     string   `gorm:"type:longtext;comment:文章正文"`
	Cover       string   `gorm:"size:255;comment:封面图;default:'defaultCover.jpg'"`
	Information string   `gorm:"size:256;comment:文章摘要"`
	Tags        []string `gorm:"type:json;serializer:json;comment:标签列表"`

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
	ArticleID   string    `gorm:"column:article_id"`
	Title       string    `gorm:"column:title"`
	Cover       string    `gorm:"column:cover"`
	Information string    `gorm:"column:information"`
	Tags        []string  `gorm:"column:tags;serializer:json"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

// Article DTO articlecache中使用 存储前端中展示的内容
type Article struct {
	ArticleID   string   `gorm:"column:article_id"`
	PublisherID string   `gorm:"column:publisher_id"`
	Title       string   `gorm:"column:title"`
	Cover       string   `gorm:"column:cover"`
	Information string   `gorm:"column:information"`
	Content     string   `gorm:"column:content"`
	Tags        []string `gorm:"column:tags;serializer:json"`
	// 时间戳设置为int64类型便于转换
	CreatedAt int64 `gorm:"column:created_at"`
	UpdatedAt int64 `gorm:"column:updated_at"`
}

// UpdateArticle DTO 存储需要更新的文章字段（不含封面）
type UpdateArticle struct {
	Title       string   `gorm:"column:title"`
	Information string   `gorm:"column:information"`
	Content     string   `gorm:"column:content"`
	Tags        []string `gorm:"column:tags;serializer:json"`
}
