package validate

import (
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"
)

type IdentityValidate struct {
}

func NewIdentityValidate() *IdentityValidate {
	return &IdentityValidate{}
}

const (
	MaxUserIDLength = 10
)

var (
	errUserIDEmpty                 = errors.New("用户ID输入为空！")
	errUSerIDValidLength           = errors.New("用户ID长度不合法！")
	errUserPasswordValidLength     = errors.New("密码长度不符合要求！")
	errUserPasswordValid           = errors.New("密码格式不符合要求！")
	errUserPasswordIncludeSpace    = errors.New("密码内不能含有空格！")
	errVerificationCodeEmpty       = errors.New("验证码输入为空！")
	errVerificationCodeValid       = errors.New("验证码格式错误！")
	errVerificationCodeValidLength = errors.New("验证码长度不合法！")
	errEmailEmpty                  = errors.New("邮箱输入为空！")
	errEmailIncludeSpace           = errors.New("邮箱输入含有空格！")
	errEmailValidLength            = errors.New("邮箱长度不合法！")
	errEmailValid                  = errors.New("邮箱格式不合法！")
	errDoublePasswordNotSame       = errors.New("两次输入的密码不一致！")
)

func (iv *IdentityValidate) ValidateUserID(userID string) (string, error) {
	uid := strings.TrimSpace(userID)
	if uid == "" {
		return "", errUserIDEmpty
	}
	if utf8.RuneCountInString(uid) > MaxUserIDLength {
		return "", errUSerIDValidLength
	}
	return uid, nil
}

// ValidatePasswordFormat 检查password字符串的组成
func (iv *IdentityValidate) ValidatePasswordFormat(password string) (string, error) {
	p := strings.TrimSpace(password)
	if p != password {
		return "", errUserPasswordIncludeSpace
	}
	if len(p) < 6 || len(p) > 20 {
		return "", errUserPasswordValidLength
	}
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, c := range p {
		switch {
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= '0' && c <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	if !(hasLower && hasUpper && hasDigit && hasSpecial) {
		return "", errUserPasswordValid
	}

	return p, nil
}

// ValidateVerificationCode 校验验证码
func (iv *IdentityValidate) ValidateVerificationCode(verificationCode string) (string, error) {
	vc := strings.TrimSpace(verificationCode)
	if vc == "" {
		return "", errVerificationCodeEmpty
	}

	if len(vc) != 6 {
		return "", errVerificationCodeValidLength
	}

	var other = false
	for _, c := range vc {
		switch {
		case c >= '0' && c <= '9':
			continue
		default:
			other = true
		}
	}
	if other {
		return "", errVerificationCodeValid
	}

	return vc, nil
}

// ValidateEmail 校验邮箱
func (iv *IdentityValidate) ValidateEmail(email string) (string, error) {
	if email == "" {
		return "", errEmailEmpty
	}
	if strings.TrimSpace(email) != email {
		return "", errEmailIncludeSpace
	}
	if len(email) > 254 {
		return "", errEmailValidLength
	}

	// mail库 校验email格式
	address, err := mail.ParseAddress(email)
	if err != nil {
		return "", errEmailValid
	}
	if address.Address != email {
		return "", errEmailValid
	}
	return email, nil
}

// ValidateDoublePassword 修改密码中校验两个输入的密码
func (iv *IdentityValidate) ValidateDoublePassword(newPassword,
	oldPassword string) (string, string, error) {
	npwd := strings.TrimSpace(newPassword)
	opwd := strings.TrimSpace(oldPassword)
	if npwd == "" || opwd == "" {
		return "", "", errUserPasswordValidLength
	}
	if npwd != opwd {
		return "", "", errDoublePasswordNotSame
	}
	npwd, err := iv.ValidatePasswordFormat(npwd)
	if err != nil {
		return "", "", err
	}
	opwd, err = iv.ValidatePasswordFormat(opwd)
	if err != nil {
		return "", "", err
	}

	return npwd, opwd, nil
}
