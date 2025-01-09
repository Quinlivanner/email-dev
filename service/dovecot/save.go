package dovecot

import (
	"crypto/sha256"
	"email/global"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func SaveEmailToMaildir(emailContent, domainName, userName string) (string, error) {
	// 构造 Maildir 路径
	newPath := filepath.Join("/email_save", domainName, userName, "Maildir", "cur")

	// 生成唯一的文件名
	timestamp := time.Now().Unix()
	pid := os.Getpid()
	size := len(emailContent)
	flags := "2," // 没有特殊标志

	fileName := fmt.Sprintf("%d.%d.%s,S=%d:%s", timestamp, pid, global.Config.Dovecot.Host, size, flags)
	filePath := filepath.Join(newPath, fileName)

	// 写入文件
	err := ioutil.WriteFile(filePath, []byte(emailContent), 0600)
	if err != nil {
		return fileName, fmt.Errorf("写入文件失败: %v", err)
	}

	// 更改文件所有者为 vmail
	err = os.Chown(filePath, 5000, 5000)
	if err != nil {
		return fileName, fmt.Errorf("更改文件所有者失败: %v", err)
	}

	fmt.Printf("邮件已保存到: %s\n", filePath)
	return fileName, nil
}

func DeleteEmailFromMaildir(domainName, userName, emailContent string) error {
	// 构造 Maildir 路径
	newPath := filepath.Join("/email_save", domainName, userName, "Maildir", "new")

	// 计算邮件内容的哈希值
	hash := sha256.Sum256([]byte(emailContent))
	emailHash := hex.EncodeToString(hash[:])

	// 遍历目录查找匹配的文件
	err := filepath.Walk(newPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			fileHash := sha256.Sum256(content)
			if hex.EncodeToString(fileHash[:]) == emailHash {
				// 找到匹配的文件，删除它
				return os.Remove(path)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("删除邮件文件失败: %v", err)
	}

	return nil
}

// IsEmailExistInMaildir 检查邮件是否已经存在于Maildir中
func IsEmailExistInMaildir(domainName, userName, emailContent string) (bool, error) {
	// 构造 Maildir 路径
	newPath := filepath.Join("/email_save", domainName, userName, "Maildir", "cur")
	// 计算邮件内容的哈希值
	hash := sha256.Sum256([]byte(emailContent))
	emailHash := hex.EncodeToString(hash[:])
	// 标记是否找到匹配的文件
	var found bool
	// 遍历目录查找匹配的文件
	err := filepath.Walk(newPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			fileHash := sha256.Sum256(content)
			if hex.EncodeToString(fileHash[:]) == emailHash {
				found = true
				return filepath.SkipDir // 找到后停止遍历
			}
		}
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("检查邮件文件是否存在时发生错误: %v", err)
	}
	return found, nil
}
