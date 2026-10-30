package logic

import "github.com/go-playground/validator/v10"

var Validate = validator.New()

type class map[string]string

var TypeName = class{
	"Email":   "邮箱",
	"Type":    "邮箱作用",
	"Title":   "公告标题",
	"Message": "公告正文",
}
