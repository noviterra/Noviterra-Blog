package logic

import "github.com/go-playground/validator/v10"

var Validate = validator.New()

type class map[string]string

var TypeName = class{
	// CommentArticleRequest / CommentNoviterraRequest
	"Comment": "评论结构体信息",

	// ArticleComment（嵌套在 Comment 内）
	"Comment.Root":      "回复所在的首评论",
	"Comment.CommentID": "被回复的评论唯一标识",
	"Comment.SpeakerID": "评论者ID",
	"Comment.Content":   "评论内容",

	// LikeArticleRequest / CollectArticleRequest
	"ArticleID": "文章ID",
	"Action":    "操作类型",
}
