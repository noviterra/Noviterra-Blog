package logic

import "github.com/go-playground/validator/v10"

var Validate = validator.New()

type class map[string]string

var TypeName = class{
	// 用户基础信息
	"ID":               "用户ID",
	"Role":             "用户角色",
	"Username":         "用户名",
	"Email":            "邮箱",
	"Avatar":           "头像",
	"Age":              "年龄",
	"Gender":           "性别",
	"SelfIntroduction": "个人简介",

	// 密码相关
	"OldPassword":      "旧密码",
	"NewPassword":      "新密码",
	"VerificationCode": "验证码",

	// 分页相关
	"Page": "页数",
	"Size": "页面展示用户数量",
}
