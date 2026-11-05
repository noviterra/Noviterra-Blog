package logic

import "github.com/go-playground/validator/v10"

var Validate = validator.New()

type class map[string]string

var TypeName = class{
	"ArticleID":   "文章ID",
	"PublisherID": "发布者ID",
	"Page":        "页数",
	"PageSize":    "数量",
	"Title":       "文章标题",
	"Content":     "文章内容",
	"Cover":       "封面",
	"Tags":        "标签",
	"Information": "简介",
	"Query":       "检索关键词",
	"CreatedAt":   "创建时间",
	"UpdatedAt":   "更新时间",
	"Action":      "操作类型",
	"ActionTime":  "操作时间",
	"UserID":      "用户ID",
	"Fields":      "查询类型",
	"ArticleList": "归档文章列表",
}
