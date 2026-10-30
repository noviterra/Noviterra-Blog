package logic

import (
	"github.com/go-playground/validator/v10"
)

type class map[string]string

var TypeName = class{
	"Link":            "有效友链接",
	"LinkLogo":        "网页Logo",
	"LinkName":        "网页名称",
	"SelfInformation": "网页介绍",
	"Email":           "联系邮箱",
}

var Validate = validator.New()
