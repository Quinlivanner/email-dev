package email

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// EmailStorageService 处理邮件存储相关的操作
type EmailStorageService struct {
	BaseDir string
}

// NewEmailStorageService 创建一个新的 EmailStorageService 实例
func NewEmailStorageService(baseDir string) *EmailStorageService {
	return &EmailStorageService{
		BaseDir: baseDir,
	}
}

// SaveEmail 保存邮件内容到指定文件夹
func (s *EmailStorageService) SaveEmail(userID, mailboxName string, emailContent []byte) error {
	// 构建保存路径
	savePath := filepath.Join(s.BaseDir, userID, mailboxName, "new")

	// 确保目录存在
	if err := os.MkdirAll(savePath, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 生成唯一的文件名（这里使用时间戳作为示例）
	fileName := fmt.Sprintf("%d.eml", time.Now().UnixNano())
	filePath := filepath.Join(savePath, fileName)

	// 写入文件
	if err := ioutil.WriteFile(filePath, emailContent, 0644); err != nil {
		return fmt.Errorf("保存邮件失败: %w", err)
	}

	return nil
}
