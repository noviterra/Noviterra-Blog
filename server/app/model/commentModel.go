package model

import "time"

// ArticleStat 文章互动统计（点赞 / 收藏 / 访问量）
type ArticleStat struct {
	ArticleID    string    `gorm:"column:article_id;primaryKey;autoIncrement:false;size:64;comment:文章ID"`
	ViewCount    int64     `gorm:"column:view_count;not null;default:0;comment:访问量"`
	LikeCount    int64     `gorm:"column:like_count;not null;default:0;comment:点赞量"`
	CollectCount int64     `gorm:"column:collect_count;not null;default:0;comment:收藏量"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime;comment:数据更新时间"`
}

func (ArticleStat) TableName() string {
	return "article_stats"
}

// UserLike 用户点赞收藏文章记录
type UserLike struct {
	ID            int64     `gorm:"type:bigint;primaryKey;autoIncrement;comment:主键"`
	ArticleID     string    `gorm:"type:varchar(36);not null;uniqueIndex:uk_user_article,priority:2;index:idx_article_like,priority:1;index:idx_article_collect,priority:1;comment:文章ID"`
	UserID        string    `gorm:"type:varchar(36);not null;uniqueIndex:uk_user_article,priority:1;index:idx_user_created,priority:1;comment:用户ID"`
	LikeStatus    int8      `gorm:"type:tinyint;not null;default:0;index:idx_article_like,priority:2;comment:点赞状态 1已赞 0未赞"`
	CollectStatus int8      `gorm:"type:tinyint;not null;default:0;index:idx_article_collect,priority:2;comment:收藏状态 1已藏 0未藏"`
	CreatedAt     time.Time `gorm:"autoCreateTime;index:idx_user_created,priority:2;comment:创建时间"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime;comment:最后操作时间"`
}

func (UserLike) TableName() string { return "user_like" }
