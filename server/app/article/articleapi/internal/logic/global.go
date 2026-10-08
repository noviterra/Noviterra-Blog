package logic

import "github.com/go-playground/validator/v10"

var Validate = validator.New()

type class map[string]string

var TypeName = class{
	"ArticleID":   "文章ID",
	"Page":        "页数",
	"PageSize":    "数量",
	"Title":       "文章标题",
	"Cover":       "封面",
	"Tags":        "标签",
	"Information": "简介",
	"Query":       "检索关键词",
}
