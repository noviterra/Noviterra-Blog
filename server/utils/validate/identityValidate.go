package validate

import (
	"errors"
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
	ErrUserIDEmpty   = errors.New("用户ID输入为空！")
	ErrUSerIDTooLong = errors.New("用户ID长度超出限制！")
)

func (iv *IdentityValidate) ValidateUserID(userID string) (string, error) {
	uid := strings.TrimSpace(userID)
	if uid == "" {
		return "", ErrUserIDEmpty
	}
	if utf8.RuneCountInString(uid) > MaxUserIDLength {
		return "", ErrUSerIDTooLong
	}
	return uid, nil
}

// CheckPasswordFormat 检查password字符串的组成
func (iv *IdentityValidate) CheckPasswordFormat(password string) bool {
	if len(password) < 6 || len(password) > 20 {
		return false
	}
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, c := range password {
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

	return hasLower && hasUpper && hasDigit && hasSpecial
}
