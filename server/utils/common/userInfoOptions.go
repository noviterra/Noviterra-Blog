package common

import (
	"errors"
	"regexp"
	"unicode/utf8"
)

type ValidateUserInfoInput struct {
	Age              int
	Name             string
	Gender           int
	SelfIntroduction string
}

type ValidateUserInfoOutput struct {
	Age              int
	Name             string
	Gender           int
	SelfIntroduction string
}

func (c *CommonUtils) ValidateUserInfo(in *ValidateUserInfoInput) (*ValidateUserInfoOutput, error) {
	age, err := c.ValidateAge(in.Age)
	if err != nil {
		return nil, err
	}
	name, err := c.ValidateUserName(in.Name)
	if err != nil {
		return nil, err
	}
	gender, err := c.ValidateGender(in.Gender)
	if err != nil {
		return nil, err
	}
	selfIntroduction, err := c.ValidateSelfIntroduction(in.SelfIntroduction)
	if err != nil {
		return nil, err
	}
	return &ValidateUserInfoOutput{
		Age:              age,
		Name:             name,
		Gender:           gender,
		SelfIntroduction: selfIntroduction,
	}, nil
}

func (c *CommonUtils) ValidateAge(age int) (int, error) {
	// 限制年龄在0~100之间
	if age <= 0 || age >= 100 {
		return -1, errors.New("年龄设置有误！")
	}

	return age, nil
}

var specialChar = regexp.MustCompile(`^[\p{Han}a-zA-Z·\-\s\.]+$`)

func (c *CommonUtils) ValidateUserName(name string) (string, error) {
	// 姓名不能有特殊字符
	if !specialChar.MatchString(name) {
		return "", errors.New("用户名中出现了特殊字符！")
	}

	// 限制2~20字符内 一个中文占两个字符
	length := utf8.RuneCountInString(name)
	if length > 20 || length < 2 {
		return "", errors.New("用户名长度不符合！")
	}

	return name, nil
}

func (c *CommonUtils) ValidateGender(gender int) (int, error) {
	if gender != 0 && gender != 1 && gender != 2 { // 分别对应未知 男 女
		return 0, errors.New("用户性别设置异常！")
	}

	return gender, nil
}

func (c *CommonUtils) ValidateSelfIntroduction(selfIntroduction string) (string, error) {
	// 限制长度 100
	length := utf8.RuneCountInString(selfIntroduction)
	if length < 0 || length > 100 {
		return "", errors.New("个人简介过长！")
	}

	return selfIntroduction, nil
}
