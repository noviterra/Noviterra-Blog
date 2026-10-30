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
