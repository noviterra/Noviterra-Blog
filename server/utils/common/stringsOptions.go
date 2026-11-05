package common

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
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

// ToString int64转为string类型
func (c *CommonUtils) ToString(num int64) string {
	return strconv.FormatInt(num, 10)
}

// ToInt64 string类型转为int64类型
func (c *CommonUtils) ToInt64(data string) (int64, error) {
	return strconv.ParseInt(data, 10, 64)
}

// ToSet 集合操作
// 暂时放在这
func (c *CommonUtils) ToSet(slice []int64) map[int64]bool {
	set := make(map[int64]bool, len(slice))
	for _, s := range slice {
		set[s] = true
	}
	return set
}
