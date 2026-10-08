package common

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// GenerateVerificationCode 生成随机数字串
func (c *CommonUtils) GenerateVerificationCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", int(n.Int64())), nil
}

// CheckPasswordFormat 检查password字符串的组成
func (c *CommonUtils) CheckPasswordFormat(password string) bool {
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

// ToStringTime int64转为string类型 返回标准时间格式
func (c *CommonUtils) ToStringTime(ts int64) string {
	var t time.Time
	// 判断是13位毫秒 还是10位秒
	if ts > 1e12 {
		t = time.UnixMilli(ts)
	} else {
		t = time.Unix(ts, 0)
	}

	return t.Format("2006-01-02 15:04:05")
}
