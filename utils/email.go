package utils

import (
	"bufio"
	"bytes"
	"email/global"
	"email/models"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jhillyerd/enmime"
	"gorm.io/datatypes"
	"io"
	"log"
	"math/rand"
	"mime"
	"net/mail"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// EmailStatus 表示邮件的完整状态
type EmailStatus struct {
	IsRead    bool
	IsReplied bool
	IsFlagged bool
	IsDraft   bool
	IsTrashed bool
}

// -----------------------------第三方状态同步 有关工具函数 ----------------------------------

// IsEmailFlagSet 检查特定标记是否已设置
func IsEmailFlagSet(filepath string, flag global.EmailFlag) (bool, error) {
	flags, err := GetEmailFlags(filepath)
	if err != nil {
		return false, err
	}
	return flags[flag], nil
}

// MarkAsRead 标记邮件为已读
func MarkAsRead(filepath string, userID, emailID uint) error {
	return ChangeEmailFlag(filepath, global.FlagSeen, true, userID, emailID)
}

// MarkAsUnread 标记邮件为未读
func MarkAsUnread(filepath string, userID, emailID uint) error {
	return ChangeEmailFlag(filepath, global.FlagSeen, false, userID, emailID)
}

//// MarkAsReplied 标记邮件为已回复
//func MarkAsReplied(filepath string) error {
//	return ChangeEmailFlag(filepath, global.FlagReplied, true)
//}
//
//// MarkAsFlagged 标记邮件为已标记/星标
//func MarkAsFlagged(filepath string) error {
//	return ChangeEmailFlag(filepath, global.FlagFlagged, true)
//}

// ClearAllFlags 移除所有标记
func ClearAllFlags(filepath string) error {
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return fmt.Errorf("file not found: %v", err)
	}

	dir := path.Dir(filepath)
	oldName := fileInfo.Name()
	flagIndex := strings.LastIndex(oldName, ":2,")

	if flagIndex == -1 {
		return nil
	}

	// 移除所有标记
	newName := oldName[:flagIndex] + ":2,"
	newPath := path.Join(dir, newName)

	if err := os.Rename(filepath, newPath); err != nil {
		return fmt.Errorf("clear flags failed: %v", err)
	}

	return nil
}

//// RemoveFlag 移除特定标记
//func RemoveFlag(filepath string, flag global.EmailFlag) error {
//	return ChangeEmailFlag(filepath, flag, false)
//}

// ChangeEmailFlag 更改邮件标记状态
func ChangeEmailFlag(filepath string, flag global.EmailFlag, addFlag bool, userId, emailID uint) error {
	// 验证文件是否存在
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return fmt.Errorf("file not found: %v", err)
	}

	// 获取文件路径信息
	dir := path.Dir(filepath)
	oldName := fileInfo.Name()

	// 验证是否是有效的邮件文件
	if !strings.Contains(oldName, ",S=") {
		return fmt.Errorf("not a valid email file: %s", oldName)
	}

	// 解析
	flagIndex := strings.LastIndex(oldName, ":2,")
	if flagIndex == -1 {
		newPath := filepath + ":2,"
		if addFlag {
			newPath += string(flag)
		}
		return os.Rename(filepath, newPath)
	}

	// 分离基标记
	baseName := oldName[:flagIndex]
	flags := make(map[string]bool)

	// 获取现有标记
	if len(oldName) > flagIndex+3 {
		existingFlags := oldName[flagIndex+3:]
		for _, f := range existingFlags {
			flags[string(f)] = true
		}
	}

	// 更新标记
	if addFlag {
		flags[string(flag)] = true
	} else {
		delete(flags, string(flag))
	}

	// 构建新的标记字符串
	var newFlags []string
	for f := range flags {
		newFlags = append(newFlags, f)
	}
	sort.Strings(newFlags)

	// 构建新文件名
	newName := fmt.Sprintf("%s:2,%s", baseName, strings.Join(newFlags, ""))
	newPath := path.Join(dir, newName)

	// 检查是否需要重命名
	if newPath == filepath {
		return nil // 无需更改
	}

	// 重命名
	if err := os.Rename(filepath, newPath); err != nil {
		return fmt.Errorf("rename file failed: %v", err)
	}
	err = UpdateEmailFileNameByEmailID(userId, emailID, newName)
	if err != nil {
		return err
	}
	return nil
}

// GetEmailStatus 获取邮件的当前状态
func GetEmailStatus(filepath string) (*EmailStatus, error) {
	flags, err := GetEmailFlags(filepath)
	if err != nil {
		return nil, err
	}

	return &EmailStatus{
		IsRead:    flags[global.FlagSeen],
		IsReplied: flags[global.FlagReplied],
		IsFlagged: flags[global.FlagFlagged],
		IsDraft:   flags[global.FlagDraft],
		IsTrashed: flags[global.FlagTrashed],
	}, nil
}

// GetEmailFlags 获取邮件的所有标记
func GetEmailFlags(filepath string) (map[global.EmailFlag]bool, error) {
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %v", err)
	}

	flags := make(map[global.EmailFlag]bool)
	filename := fileInfo.Name()

	flagIndex := strings.LastIndex(filename, ":2,")
	if flagIndex != -1 && len(filename) > flagIndex+3 {
		for _, flag := range filename[flagIndex+3:] {
			flags[global.EmailFlag(string(flag))] = true
		}
	}

	return flags, nil
}

// SaveEmailToMaildir 保存邮件到 Maildir 格式
func SaveEmailToMaildir(content *models.EmailContent) (string, string, error) {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Join(content.MailDir, "cur"), 0755); err != nil {
		return "", "", fmt.Errorf("create directory failed: %v", err)
	}

	//生成唯一的 Message-ID 和 boundary
	domain := strings.Split(content.From, "@")[1]
	messageID := fmt.Sprintf("<%s@%s>", uuid.New().String(), strings.ReplaceAll(domain, ">", ""))
	boundary := fmt.Sprintf("------------%s", uuid.New().String()[:16]) // 更标准的 boundary 格式

	// 3. 构建邮件内容
	emailContent := fmt.Sprintf(`From: %s
To: %s
Subject: %s
Date: %s
Message-ID: %s
MIME-Version: 1.0
Content-Type: multipart/alternative;
 boundary="%s"
Content-Transfer-Encoding: 7bit

--%s
Content-Type: text/plain; charset=UTF-8
Content-Transfer-Encoding: base64

%s

--%s
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: base64

%s

--%s--

`, // 注意最后添加一个空行
		content.From,
		content.To,
		mime.QEncoding.Encode("utf-8", strings.TrimSpace(content.Subject)), // 去除可能的空格
		time.Now().Format(time.RFC1123Z),
		messageID,
		boundary,
		boundary,
		encodeBody(content.TextBody),
		boundary,
		encodeBody(content.HtmlBody),
		boundary,
	) // 4. 生成文件名
	now := time.Now()
	filename := fmt.Sprintf("%d.M%dP%d.%s,S=%d:2,S",
		now.Unix(),
		now.UnixMicro()%1000000,
		os.Getpid(), // 添加进程 ID 增加唯一性
		global.Config.Dovecot.Host,
		len(emailContent),
	)

	// 5. 写入文件
	curPath := filepath.Join(content.MailDir, "cur", filename)
	if err := os.WriteFile(curPath, []byte(emailContent), 0644); err != nil {
		return "", "", fmt.Errorf("write to cur failed: %v", err)
	}

	// 6. 设置所有者
	if err := os.Chown(curPath, 5000, 5000); err != nil {
		return "", "", fmt.Errorf("change file ownership failed: %v", err)
	}

	return filename, messageID, nil
}

func GenerateEmailRawMessage(content *models.EmailContent) (string, string) {

	//生成唯一的 Message-ID 和 boundary
	domain := strings.Split(content.From, "@")[1]
	messageID := fmt.Sprintf("<%s@%s>", uuid.New().String(), strings.ReplaceAll(domain, ">", ""))
	boundary := fmt.Sprintf("------------%s", uuid.New().String()[:16]) // 更标准的 boundary 格式

	// 3. 构建邮件内容
	emailContent := fmt.Sprintf(`From: %s
To: %s
Subject: %s
Date: %s
Message-ID: %s
MIME-Version: 1.0
Content-Type: multipart/alternative;
 boundary="%s"
Content-Transfer-Encoding: 7bit

--%s
Content-Type: text/plain; charset=UTF-8
Content-Transfer-Encoding: base64

%s

--%s
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: base64

%s

--%s--

`,
		content.From,
		content.To,
		mime.QEncoding.Encode("utf-8", strings.TrimSpace(content.Subject)), // 去除可能的空格
		time.Now().Format(time.RFC1123Z),
		messageID,
		boundary,
		boundary,
		encodeBody(content.TextBody),
		boundary,
		encodeBody(content.HtmlBody),
		boundary,
	)

	return emailContent, messageID
}

// encodeBody base64 编码内容
func encodeBody(content string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(content))
	var lines []string
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		lines = append(lines, encoded[i:end])
	}
	return strings.Join(lines, "\n")
}

// -----------------------------service 有关工具函数 ----------------------------------
// 解析发件人邮件地址
func ParseSenderAddress(sender string) (*mail.Address, error) {
	from := decodeText(sender)
	fromAddress, err := mail.ParseAddress(from)
	if err != nil {
		return nil, fmt.Errorf("解析发件人地址失败: %w", err)
	}
	return fromAddress, nil
}

// 解析收件人邮件地址
func ParseRecipientAddress(recipient string) ([]*mail.Address, error) {
	to := decodeText(recipient)
	toAddress, err := mail.ParseAddressList(to)
	if err != nil {
		return nil, fmt.Errorf("解析收件人地址失败: %w", err)
	}
	return toAddress, nil
}

// 解码
func decodeText(headerValue string) string {
	// 创建一个新的 WordDecoder
	decoder := new(mime.WordDecoder)
	// 尝试解码邮件头
	decoded, err := decoder.DecodeHeader(headerValue)
	if err != nil {
		// 解码失败，返回原始值
		return headerValue
	}
	return decoded
}

// 获取邮件附件最大大小
func GetMaxAttachmentSize() int64 {
	if global.Config != nil && global.Config.AWS.MaxFileSize != 0 {
		return 1024 * 1024 * global.Config.AWS.MaxFileSize
	}
	// 返回默认值
	return 30 * 1024 * 20 // 默认 20
}

// 判断邮件类型是否在指定范围内
func Contains(slice []string, item string) bool {
	for _, value := range slice {
		if value == item {
			return true
		}
	}
	return false
}

// 辅助函数：验证邮箱地址格式
func IsValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(pattern)
	return regex.MatchString(email)
}

// 生成aws s3 key
func GenerateSendID() (string, error) {
	randomLength := 36
	// 生成随机字节
	randomBytes := make([]byte, randomLength*5/8+1)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	// 使用 base32 编码，但不包含填充字符
	encoding := base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding(base32.NoPadding)
	randomString := strings.ToLower(encoding.EncodeToString(randomBytes))

	// 截取所需长度并添加 "send" 前缀
	result := "send" + randomString[:randomLength]

	return result, nil
}

// 只读取文件头部来获取 Message-ID
func GetEmailMessageIDFromFile(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 只读取前 4KB，通常邮件头部不会超过这个大小
	headerData := make([]byte, 32*1024)
	// 使用更大的缓冲区
	n, err := file.Read(headerData)
	if err != nil && err != io.EOF {
		return "", err
	}

	// 查找头部结束位置
	headerEnd := bytes.Index(headerData[:n], []byte("\r\n\r\n"))
	if headerEnd == -1 {
		// 如果没找到空行，可能需要更大的缓冲区
		log.Printf("警告: 文件 %s 的头部可能超过 %d KB", filepath, 32*1024/1024)
		headerEnd = n
	}

	// 只解析头部
	reader := bytes.NewReader(headerData[:headerEnd])
	env, err := enmime.ReadEnvelope(reader)
	if err != nil {
		// 如果解析失败，尝试使用备用方法
		return getMessageIDByScanning(file)
	}

	return env.GetHeader("Message-ID"), nil
}

// 备用方法：通过扫描文件来获取 Message-ID
func getMessageIDByScanning(file *os.File) (string, error) {
	// 确保文件指针回到开始
	file.Seek(0, 0)

	scanner := bufio.NewScanner(file)
	// 增加 scanner 的缓冲区
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // 最大允许 1MB 的行

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Message-ID:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Message-ID:")), nil
		}
		// 如果遇到空行，说明头部结束
		if line == "" {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", nil
}

// 根据EemailID 更新filename
func UpdateEmailFileNameByEmailID(accountID, emailID uint, fileName string) error {
	tableName := fmt.Sprintf("user_%d_emails", accountID)
	// 更新邮件的��件名
	if err := global.PsqlDB.Table(tableName).
		Where("id = ?", emailID).
		Update("file_name", fileName).Error; err != nil {
		return err
	}
	return nil

}

func ParseSliceJson(v interface{}) datatypes.JSON {
	data, err := json.Marshal(v)
	if err != nil {
		return []byte("[]")
	}
	return data
}

// 解析邮件地址列表，提取纯邮件地址
func ParseAddressList(addresses []string) []string {
	if len(addresses) == 0 {
		return nil
	}

	result := make([]string, 0, len(addresses))
	for _, addr := range addresses {
		// 处理可能的多个地址（以逗号分隔）
		parsed, err := mail.ParseAddressList(addr)
		if err != nil {
			// 如果解析失败，尝试作为单个地址解析
			if parsedAddr, err := mail.ParseAddress(addr); err == nil {
				result = append(result, parsedAddr.Address)
			} else {
				// 记录错误日志
				log.Printf("邮件地址格式无效: %s, 错误: %v", addr, err)
				// 不保留无效地址，跳过处理
				continue
			}
			continue
		}

		// 添加解析成功的地址
		for _, parsedAddr := range parsed {
			result = append(result, parsedAddr.Address)
		}
	}

	return result
}

func ParseFromEmailAddress(e string) models.EmailAddress {
	resAdd := models.EmailAddress{}
	if strings.Contains(e, "<") && strings.Contains(e, ">") {
		as := strings.Split(e, "<")
		resAdd.DisplayName = as[0]
		resAdd.Address = strings.ReplaceAll(as[1], ">", "")
		return resAdd
	}
	resAdd.Address = e
	resAdd.DisplayName = strings.Split(e, "@")[0]
	return resAdd
}
