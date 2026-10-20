package logic

import "github.com/go-playground/validator/v10"

var Validate = validator.New()

type class map[string]string

var TypeName = class{
	"ID":               "用户ID",
	"Email":            "邮箱",
	"Type":             "邮箱作用",
	"Password":         "密码",
	"VerificationCode": "验证码",
	"NewPassword":      "新密码",
	"VerifPassword":    "确认新密码",
	"RefreshToken":     "身份凭证",
	"OldPassword":      "旧密码",
}
