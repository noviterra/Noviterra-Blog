package common

import (
	"encoding/binary"
	"errors"
	"strconv"
	"strings"
)

type TypesConversion struct {
}

func NewTypesConversion() *TypesConversion {
	return &TypesConversion{}
}

// Int64ConversionBytes int64 -> []byte
func (tc *TypesConversion) Int64ConversionBytes(data int64) []byte {
	// 创建一个长度为8的字节切片 int64固定占用8字节
	buf := make([]byte, 8)

	// 使用大端序将int64写入切片
	// 将int64强制转为uint64 保证底层二进制位不发生变化
	// 对正数的雪花ID无影响 对负数来说直接转uint64可以保留其补码形式的二进制位
	binary.BigEndian.PutUint64(buf, uint64(data))

	return buf
}

// Int64ConversionString int64 -> string
func (tc *TypesConversion) Int64ConversionString(data int64) string {
	return strconv.FormatInt(data, 10)
}

// StringConversionInt64 string -> int64
func (tc *TypesConversion) StringConversionInt64(data string) (int64, error) {
	d := strings.TrimSpace(data)
	if d == "" {
		return 0, errors.New("字符串为空，无法抓转换！")
	}
	return strconv.ParseInt(d, 10, 64)
}
