package model

import (
	"time"

	"gorm.io/gorm"
)

// ArticleStat 文章互动统计（点赞 / 收藏 / 访问量）
type ArticleStat struct {
	ID           int64     `gorm:"type:bigint;primaryKey;autoIncrement;comment:主键"`
	ArticleID    int64     `gorm:"column:article_id;autoIncrement:false;size:64;comment:文章ID"`
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
	ArticleID     int64     `gorm:"type:varchar(36);not null;uniqueIndex:uk_user_article,priority:2;index:idx_article_like,priority:1;index:idx_article_collect,priority:1;comment:文章ID"`
	UserID        string    `gorm:"type:varchar(36);not null;uniqueIndex:uk_user_article,priority:1;index:idx_user_created,priority:1;comment:用户ID"`
	LikeStatus    int8      `gorm:"type:tinyint;not null;default:0;index:idx_article_like,priority:2;comment:点赞状态 1已赞 0未赞"`
	CollectStatus int8      `gorm:"type:tinyint;not null;default:0;index:idx_article_collect,priority:2;comment:收藏状态 1已藏 0未藏"`
	CreatedAt     time.Time `gorm:"autoCreateTime;index:idx_user_created,priority:2;comment:创建时间"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime;comment:最后操作时间"`
}

func (UserLike) TableName() string { return "user_like" }

// ArticleComment 对应 article_comments 表
type ArticleComment struct {
	CommentID int64          `gorm:"column:comment_id;primaryKey;autoIncrement;type:bigint unsigned;comment:评论的唯一ID标识"`
	Address   int64          `gorm:"column:address;type:varchar(100);not null;index:idx_address_root,priority:1;comment:评论的地址某篇文章or站点"`
	RootID    int64          `gorm:"column:root_id;type:bigint unsigned;default:null;index:idx_address_root,priority:2;index:idx_root_created,priority:1;comment:根节点即首评ID"`
	ParentID  int64          `gorm:"column:parent_id;type:bigint unsigned;default:null;index:idx_parent_created,priority:1;comment:父节点即被回复的评论ID"`
	Depth     int32          `gorm:"column:depth;type:tinyint unsigned;not null;default:0;comment:多层评论深度"`
	SpeakerID string         `gorm:"column:speaker_id;type:varchar(100);not null;index:idx_speaker,priority:1;comment:评论者ID"`
	Content   string         `gorm:"column:content;type:varchar(200);not null;comment:评论内容"`
	Status    int32          `gorm:"column:status;type:tinyint;not null;default:1;comment:1已发布0已删除"`
	CreatedAt time.Time      `gorm:"column:created_at;type:datetime(3);not null;autoCreateTime;index:idx_address_root,priority:3;index:idx_root_created,priority:2;index:idx_parent_created,priority:2;index:idx_speaker,priority:2;comment:发布时间"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间"`
}

func (ArticleComment) TableName() string {
	return "article_comments"
}

// SyncEvent 异步同步事件 DTO
// Redis List消息队列操作 后续使用Kafka消息队列
type SyncEvent struct {
	ArticleID int64
	UserID    string
	Action    string // "like"、"unlike" 或 "collect、"uncollect"
}

// CommentNode DTO 多层评论中单个评论属性
type CommentNode struct {
	CommentID string    // 文章唯一ID
	ParentID  string    // 父节点评论ID  为空表示为顶层评论
	SpeakerID string    // 回复者ID
	Content   string    // 评论内容
	CreatedAt time.Time // 评论时间
}

// TopCommentNode 顶层评论属性
type TopCommentNode struct {
	CommentID string
	SpeakerID string
	Content   string
	Count     int64
	CreatedAt time.Time
}
