package common

import (
	"errors"
	"regexp"
	"unicode/utf8"
)

func (c *CommonUtils) ValidateAge(age int) error {
	// 限制年龄在0~100之间
	if age <= 0 || age >= 100 {
		return errors.New("年龄设置有误！")
	}

	return nil
}

var specialChar = regexp.MustCompile(`^[\p{Han}a-zA-Z·\-\s\.]+$`)

func (c *CommonUtils) ValidateUserName(name string) error {
	// 姓名不能有特殊字符
	if !specialChar.MatchString(name) {
		return errors.New("用户名中出现了特殊字符！")
	}

	// 限制2~20字符内 一个中文占两个字符
	length := utf8.RuneCountInString(name)
	if length > 20 || length < 2 {
		return errors.New("用户名长度不符合！")
	}

	return nil
}

func (c *CommonUtils) ValidateGender(gender int) error {
	if gender != 0 && gender != 1 && gender != 2 { // 分别对应未知 男 女
		return errors.New("用户性别设置异常！")
	}

	return nil
}

func (c *CommonUtils) ValidateSelfIntroduction(selfIntroduction string) error {
	// 限制长度 100
	length := utf8.RuneCountInString(selfIntroduction)
	if length < 0 || length > 100 {
		return errors.New("个人简介过长！")
	}

	return nil
}
