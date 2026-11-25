package model

import "time"

// ==================================================Models===========================================================

// AgentModelConfig 站点助手模型配置 全局单例 ID固定为1
// APIKey不落明文 加密后存储
type AgentModelConfig struct {
	ID             int64  `gorm:"primaryKey;autoIncrement:false;comment:配置ID（固定1）"`
	Provider       string `gorm:"size:32;not null;comment:服务商标识 openai/deepseek/siliconflow/ollama..."`
	BaseURL        string `gorm:"size:255;not null;comment:OpenAI兼容接口地址（含/v1）"`
	Model          string `gorm:"size:128;not null;comment:对话模型"`
	EmbeddingModel string `gorm:"size:128;not null;default:'';comment:向量模型，为空表示不启用向量检索"`
	APIKeyCipher   string `gorm:"type:text;not null;comment:APIKey密文（AES-256-GCM，Base64URL）"`
	UpdatedBy      string `gorm:"size:64;not null;default:'';comment:最后修改人"`

	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (AgentModelConfig) Name() string {
	return "agent_model_config"
}

// AgentKnowledgeChunk 站点知识库切片
// 来源：已发布的文章、传入的文档等
type AgentKnowledgeChunk struct {
	ID         int64  `gorm:"primaryKey;autoIncrement;comment:切片ID"`
	SourceType string `gorm:"size:32;not null;uniqueIndex:uk_agent_chunk,priority:1;comment:来源类型：article"`
	SourceID   int64  `gorm:"not null;uniqueIndex:uk_agent_chunk,priority:2;comment:来源ID：article_id"`
	ChunkIndex int    `gorm:"not null;uniqueIndex:uk_agent_chunk,priority:3;comment:切片序号（从0开始）"`

	Title     string `gorm:"size:255;not null;default:'';comment:来源标题（冗余，检索展示用）"`
	Content   string `gorm:"type:text;not null;comment:切片正文"`
	CharCount int    `gorm:"not null;default:0;comment:切片字符数"`

	// SourceHash 来源内容指纹：标题+正文的 sha256，用于增量同步时跳过未变更的来源
	SourceHash string `gorm:"size:64;not null;default:'';index;comment:来源内容指纹"`
	// SourceUpdate 来源 updated_at（unix秒），用于水位增量同步
	SourceUpdate int64 `gorm:"not null;default:0;index;comment:来源更新时间"`

	Embedding  []float32 `gorm:"type:json;serializer:json;comment:向量（JSON数组）"`
	EmbedModel string    `gorm:"size:128;not null;default:'';comment:生成该向量的模型"`

	CreatedAt time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}

func (AgentKnowledgeChunk) TableName() string {
	return "agent_knowledge_chunk"
}

// ==================================================DTO===========================================================

// ArticleIndexSource 索引源的投影，避免把整表加载进内存
type ArticleIndexSource struct {
	ArticleID   int64
	Title       string
	Content     string
	Information string
	Tags        string // 原始 JSON 文本，由索引器自行解析
	IsDeleted   bool
	UpdatedAt   int64
}
