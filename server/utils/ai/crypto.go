package ai

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

// cipherPrefix 版本前缀，便于以后更换算法或密钥时平滑迁移
const cipherPrefix = "v1:"

// EncryptSecret 使用 AES-256-GCM 加密敏感配置（如模型 APIKey）
// key 可以是任意长度字符串，内部用 sha256 派生出 32 字节密钥
func EncryptSecret(plain, key string) (string, error) {
	if plain == "" {
		return "", errors.New("待加密内容为空")
	}
	aead, err := newAEAD(key)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("生成随机数失败：%w", err)
	}

	// 密文 = nonce + ciphertext + tag
	sealed := aead.Seal(nonce, nonce, []byte(plain), nil)
	return cipherPrefix + base64.RawURLEncoding.EncodeToString(sealed), nil
}

// DecryptSecret 解密 EncryptSecret 的产物
func DecryptSecret(cipherText, key string) (string, error) {
	raw := strings.TrimSpace(cipherText)
	if raw == "" {
		return "", errors.New("密文为空")
	}
	raw = strings.TrimPrefix(raw, cipherPrefix)

	data, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return "", fmt.Errorf("密文格式错误：%w", err)
	}

	aead, err := newAEAD(key)
	if err != nil {
		return "", err
	}
	if len(data) < aead.NonceSize() {
		return "", errors.New("密文长度不足")
	}

	nonce, sealed := data[:aead.NonceSize()], data[aead.NonceSize():]
	plain, err := aead.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", errors.New("密文解密失败：密钥不匹配或密文被篡改")
	}
	return string(plain), nil
}

func newAEAD(key string) (cipher.AEAD, error) {
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("加密密钥未配置")
	}
	sum := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, fmt.Errorf("初始化加密算法失败：%w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("初始化加密算法失败：%w", err)
	}
	return aead, nil
}
