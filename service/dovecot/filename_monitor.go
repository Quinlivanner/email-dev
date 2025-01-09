package dovecot

import (
	"bufio"
	"email/dao"
	"email/global"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// 需要忽略的文件
var ignoreFiles = map[string]bool{
	"dovecot-uidlist":        true,
	"dovecot-uidlist.lock":   true,
	"dovecot-uidlist.tmp":    true,
	"dovecot.index.log":      true,
	"dovecot.index.cache":    true,
	"dovecot.list.index.log": true,
}

var categoryList = map[string]string{
	".Junk":   "trash",
	".Trash":  "deleted",
	".Sent":   "sent",
	".Drafts": "draft",
	".Inbox":  "inbox",
}

// 邮件ID正则表达式
var emailIDRegex = regexp.MustCompile(`(\d+\.\w+\.\w+)`)

// 文件操作记录
type FileOperation struct {
	Path string
	Time time.Time
}

var fileOps = make(map[string]*FileOperation)

// 检查是否是需要忽略的文件
func shouldIgnore(path string) bool {
	filename := filepath.Base(path)
	if ignoreFiles[filename] {
		return true
	}
	if strings.Contains(path, "/tmp/") {
		return true
	}
	return false
}

// 获取邮件ID和分类
func getEmailInfo(path string) (string, string, bool, string, string) {
	// 提取完整的文件名
	filename := filepath.Base(path)
	// 提取邮件ID
	matches := emailIDRegex.FindStringSubmatch(filename)
	if len(matches) < 2 {
		return "", "", false, "", path
	}
	emailID := matches[1]

	// 提取分类
	parts := strings.Split(path, "/Maildir/")
	if len(parts) < 2 {
		return emailID, "", false, filename, path
	}

	folderPart := strings.Split(parts[1], "/cur/")[0]

	// 如果是空字符串或者不包含点号，说明是收件箱
	if folderPart == "" || !strings.Contains(folderPart, ".") {
		return emailID, ".Inbox", false, filename, path
	}

	// 如果以.开头，是特殊文件夹
	if strings.HasPrefix(folderPart, ".") {
		// 检查是否是删除操作（在垃圾箱或垃圾邮件文件夹且标记为T）
		isDelete := (folderPart == ".Trash" || folderPart == ".Junk") &&
			(strings.HasSuffix(path, ":2,T") || strings.HasSuffix(path, ":2,ST"))
		return emailID, folderPart, isDelete, filename, path
	}

	return emailID, ".Inbox", false, filename, path
}

// 清理旧的操作记录
func cleanOldOperations() {
	now := time.Now()
	for file, op := range fileOps {
		if now.Sub(op.Time) > time.Minute {
			delete(fileOps, file)
		}
	}
}

// 检查是否是需要忽略的目录
func shouldIgnoreDir(path string) bool {
	return (strings.Contains(path, "/tmp") || strings.Contains(path, "/new"))
}

func WatchEmailDir(root string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// 首次添加所有现存目录
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && !shouldIgnoreDir(path) {
			err = watcher.Add(path)
			if err != nil {
				log.Printf("添加监控目录失败 %s: %v", path, err)
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	// 定期清理旧的操作记录
	go func() {
		for {
			time.Sleep(time.Minute)
			cleanOldOperations()
		}
	}()

	log.Printf("开始监控邮件目录: %s", root)
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// 如果是新建目录，添加到监控
			if event.Op&fsnotify.Create == fsnotify.Create {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					err = filepath.Walk(event.Name, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							return err
						}
						if info.IsDir() && !shouldIgnoreDir(path) {
							err = watcher.Add(path)
							if err != nil {
								log.Printf("添加新目录到监控失败 %s: %v", path, err)
							} else {
								log.Printf("添加新目录到监控: %s", path)
							}
						}
						return nil
					})
					if err != nil {
						log.Printf("遍历新目录失败: %v", err)
					}
					continue
				}
			}

			if shouldIgnore(event.Name) {
				continue
			}

			if !strings.Contains(event.Name, "/cur/") {
				continue
			}

			// 忽略包含 imap.mountex.net 的路径
			if strings.Contains(event.Name, "imap.mountex.net") {
				continue
			}

			matches := emailIDRegex.FindStringSubmatch(filepath.Base(event.Name))
			if len(matches) < 2 {
				continue
			}
			emailID := matches[1]

			switch {
			case event.Op&fsnotify.Create == fsnotify.Create:
				emailID, folder, isDelete, filename, path := getEmailInfo(event.Name)
				if emailID != "" {
					if isDelete {
						log.Printf("邮件删除: ID=%s, 分类=%s, 文件名=%s, 文件路径=%s", emailID, folder, filename, path)
					} else if folder != "" {
						go func(string, string) {
							global.Log.Info(fmt.Sprintf("准备更新 Email File Name"))
							hendelEmailNameChaned(path, folder, filename)
							log.Printf("邮件移动: ID=%s, 分类=%s, 文件名=%s, 文件路径=%s", emailID, folder, filename, path)
						}(path, folder)
					}
				}

			case event.Op&fsnotify.Remove == fsnotify.Remove:
				fileOps[emailID] = &FileOperation{
					Path: event.Name,
					Time: time.Now(),
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("监控错误: %v", err)
		}
	}
}

func handleEmailDeleted() {

}

func hendelEmailNameChaned(path, category, fileName string) {
	msgId, err := GetMessageIDFromFile(path)
	if err != nil {
		global.Log.Errorf("GetMessageIDFromFile Error:%s\n", err)
		return
	}
	global.Log.Info(fmt.Sprintf("成功获取到Email Msg Id => %s", msgId))
	if _, ok := categoryList[category]; !ok {
		global.Log.Errorf("Category:%s not found\n", category)
		return
	}
	global.Log.Info("成功获取到Email Category")

	emailAddress, err := ExtractEmailFromPath(path)
	if err != nil {
		global.Log.Errorf("EmailAddress Get Error:%s\n", err)
		return
	}
	global.Log.Info(fmt.Sprintf("成功获取到Email Address => %s", emailAddress))

	accountId, err := dao.GetAccountIDByEmailAddress(emailAddress)
	if err != nil {
		global.Log.Errorf("AccountId Get Error:%d\n", err)

		return
	}
	global.Log.Info(fmt.Sprintf("成功获取到AccountId => %d", accountId))
	err = dao.UpdateEmailFileNameByMessageID(accountId, msgId, fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

// 获取messageid
func GetMessageIDFromFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 创建一个缓冲读取器
	reader := bufio.NewReader(file)

	// 只读取到空行（头部和正文的分隔）
	var messageID string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		// 去除行尾的 \r\n
		line = strings.TrimRight(line, "\r\n")

		// 如果是空行，表示头部结束
		if line == "" {
			break
		}

		// 检查是否是 Message-ID 行
		if strings.HasPrefix(strings.ToLower(line), "message-id:") {
			messageID = strings.TrimSpace(strings.TrimPrefix(line, "Message-ID:"))
			messageID = strings.TrimSpace(strings.TrimPrefix(messageID, "message-id:"))
			break // 找到 Message-ID 后直接退出
		}
	}

	return messageID, nil
}

// 获取邮箱地址
func ExtractEmailFromPath(filepath string) (string, error) {
	// 使用 Maildir 作为关键字来分割路径
	parts := strings.Split(filepath, "Maildir")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid path: Maildir not found")
	}

	// 获取 Maildir 之前的路径部分
	beforeMaildir := strings.TrimSpace(parts[0])
	// 分割路径获取最后两个部分（用户名和域名）
	pathParts := strings.Split(strings.Trim(beforeMaildir, "/"), "/")
	if len(pathParts) < 2 {
		return "", fmt.Errorf("invalid path: cannot extract username and domain")
	}
	// 获取最后两个部分
	username := pathParts[len(pathParts)-1]
	domain := pathParts[len(pathParts)-2]

	// 拼接邮箱地址
	email := fmt.Sprintf("%s@%s", username, domain)

	return email, nil
}

func DovecotFileNameMonitor() {
	emailDir := "/email_save"
	err := WatchEmailDir(emailDir)
	if err != nil {
		log.Fatal(err)
	}
}

//func SaveNewEmail(username, mailbox, emailContent string) error {
//	// 创建临时文件
//	tmpFile, err := os.CreateTemp("", "email-*.eml")
//	if err != nil {
//		return fmt.Errorf("create temp file failed: %v", err)
//	}
//	defer os.Remove(tmpFile.Name()) // 清理临时文件
//
//	// 写入邮件内容
//	if _, err := tmpFile.WriteString(emailContent); err != nil {
//		return fmt.Errorf("write email content failed: %v", err)
//	}
//	tmpFile.Close()
//
//	// 使用 doveadm 保存邮件
//	cmd := exec.Command("doveadm", "save", "-u", username, "-m", mailbox, tmpFile.Name())
//	output, err := cmd.CombinedOutput()
//	if err != nil {
//		return fmt.Errorf("save email failed: %v, output: %s", err, string(output))
//	}
//
//	return nil
//}
