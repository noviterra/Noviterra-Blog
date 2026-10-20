package common

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (c *CommonUtils) SaveImageFile(header *multipart.FileHeader, file multipart.File,
	saveDir, fileName string) (string, error) {
	// 文件校验
	if header == nil || file == nil {
		return "", errors.New("未上传有效文件！")
	}
	if header.Size <= 0 {
		return "", errors.New("未上传有效文件！")
	}
	if header.Size > int64(maxSize) {
		return "", errors.New("上传文件超过了5MB！")
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if _, ok := allowedExt[ext]; !ok {
		return "", errors.New("上传文件仅支持jpg/jpeg/png/webp！")
	}

	// 将封面保存到static/articlecover中 保存名cover+articleID
	sd := filepath.Join("./static", saveDir)
	if err := os.MkdirAll(sd, 0755); err != nil {
		return "", errors.New("文件上传失败！")
	}

	fn := fmt.Sprintf("cover%stime%d%s", fileName, time.Now().Unix(), ext)
	savePath := filepath.Join(sd, fn)

	// 每次写入前删除旧文件
	matches, _ := filepath.Glob(filepath.Join(sd, fmt.Sprintf("cover%stime*.*", fileName)))
	for _, m := range matches {
		_ = os.Remove(m)
	}
	dst, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("保存文件失败！%w", err)
	}

	// 写入失败时删除残余file
	if _, err := io.Copy(dst, file); err != nil {
		_ = dst.Close()
		_ = os.Remove(savePath)
		return "", fmt.Errorf("文件保存失败！%w", err)
	}
	_ = dst.Close()

	return fn, nil
}
