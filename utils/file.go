package utils

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// CutFileByNameAndPath 根据文件名和路径使用包含关系匹配文件，并将文件剪切到指定路径下
func CutFileByPath(srcPath string, dstPath string) error {
	// 获取源文件信息
	fileInfo, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("source file error: %v", err)
	}

	// 确保源文件存在且不是目录
	if fileInfo.IsDir() {
		return fmt.Errorf("source path is a directory: %s", srcPath)
	}

	// 构建目标文件路径
	dstFile := filepath.Join(dstPath, fileInfo.Name())

	// 确保目标目录存在
	if err := os.MkdirAll(dstPath, os.ModePerm); err != nil {
		return fmt.Errorf("create destination directory failed: %v", err)
	}

	// 移动文件
	if err := os.Rename(srcPath, dstFile); err != nil {
		return fmt.Errorf("move file failed: %v", err)
	}

	return nil
}

func GenerateRandomFilePrefix() string {
	// 設置隨機種子
	rand.Seed(time.Now().UnixNano())

	// 定義字符集（大小寫字母 + 數字）
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 7

	// 生成隨機字符串
	randomString := make([]byte, length)
	for i := range randomString {
		randomString[i] = charset[rand.Intn(len(charset))]
	}

	return string(randomString)
}
