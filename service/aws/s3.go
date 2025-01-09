package aws

import (
	"bytes"
	"context"
	"email/dao"
	"email/global"
	"email/models"
	"email/service/dovecot"
	"email/service/shortlink"
	"email/utils"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/jhillyerd/enmime"
)

// 创建新的S3客户端
func CreateS3Client() (*s3.Client, error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(global.Config.AWS.ConfigRegion),
		config.WithSharedConfigProfile(global.Config.AWS.ConfigProfile),
	)
	if err != nil {
		return nil, fmt.Errorf("无法加载 AWS 配置: %w", err)
	}
	S3Client := s3.NewFromConfig(cfg)
	return S3Client, nil
}

// 获取s3中邮件的数据
func S3EmailDataProcessor(key string, recipients *models.Recipients) error {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(global.Config.AWS.ConfigRegion),
		config.WithSharedConfigProfile(global.Config.AWS.ConfigProfile),
	)
	if err != nil {
		global.Log.Errorf("无法加载 AWS 配置: %v", err)
		return err
	}
	client := s3.NewFromConfig(cfg)
	bucket := global.Config.AWS.S3Bucket
	input := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}
	// 获取详细数据
	result, err := client.GetObject(ctx, input)
	if err != nil {
		global.Log.Errorf("无法获取对象: %v", err)
		return err
	}
	defer result.Body.Close()
	// 读取数据
	body, err := io.ReadAll(result.Body)
	if err != nil {
		global.Log.Errorf("无法读取对象内容: %v", err)
		return err
	}
	// 转换数据
	emailStr := string(body)
	// 解析数据
	env, err := enmime.ReadEnvelope(strings.NewReader(emailStr))
	if err != nil {
		global.Log.Errorf("无法解析邮件内容: %v", err)
		return err
	}

	if env.GetHeader("Message-ID") == "" {
		newMessageID := fmt.Sprintf("<%s@%s>", uuid.New().String(), "mountex.net")
		err = env.AddHeader("Message-ID", newMessageID)
		if err != nil {
			return err
		}
		dateIndex := strings.Index(emailStr, "Date:")
		if dateIndex != -1 {
			dateEnd := strings.Index(emailStr[dateIndex:], "\n")
			if dateEnd != -1 {
				insertPos := dateIndex + dateEnd + 1
				body = []byte(emailStr[:insertPos] + fmt.Sprintf("Message-ID: %s\n", newMessageID) + emailStr[insertPos:])
			} else {
				return errors.New("can't find the end of date header")
			}
		} else {
			return errors.New("can't find the start of date header")
		}
	}
	//获取邮件内容hash
	hashConent := GetEmailFileHash(env)
	//判断邮件地址是否存在
	//解析收件人

	for _, address := range recipients.Recipients {
		accountData, err := dao.IsAccountExist(address, strings.Split(address, "@")[1])
		if err != nil {
			global.Log.Warnf("账户 [ %s ] 不存在，将从收件人列表中移除", address)
			continue
		}
		emailExist, err := dao.IsEmailExistByHash(accountData.ID, hashConent)
		if err != nil {
			return err
		}
		if emailExist {
			global.Log.Warnf("[%s]邮件已存在，将从收件人列表中移除", address)
			continue
		}
		err = SaveEmailToServerWithDB(accountData, key, bucket, &body, &hashConent, env, address, recipients)

		if err != nil {
			return err
		}
	}

	return nil
}

// 保存邮件到服务器与数据库
func SaveEmailToServerWithDB(accountData *models.EmailAccount, key, bucket string, emailRawMessage *[]byte, hashContent *string, env *enmime.Envelope, deliver string, recipients *models.Recipients) error {
	// 记录日志
	global.Log.Info(fmt.Sprintf("查询账户 [ %s ] 成功，即将开始保存邮件至邮件服务器！", deliver))

	//imap通信保存至指定文件夹
	if global.Config.System.Env {

		options := models.ImapOperatorOptions{
			UserName:        accountData.EmailAddress,
			MoveType:        global.AppendEmail,
			TargetMailBox:   "Inbox",
			MessageID:       env.GetHeader("Message-Id"),
			ReadStatus:      false,
			EmailRawMessage: *emailRawMessage,
		}

		err := dovecot.ImapOperatorEmail(options)
		if err != nil {
			return errors.New(err.Error() + "保存邮件至邮件服务器失败")
		}
	}

	//解析时间
	p, err := utils.ParseTime(decodeText(env.GetHeader("Date")))
	if err != nil {
		global.Log.Errorf("解析时间失败: %s", err)
		return err
	}
	//解析流程结束
	global.Log.Info("解析完毕，开始处理邮件数据...")

	from := utils.ParseFromEmailAddress(env.GetHeader("From"))

	r := models.EmailDetails{
		EmailMessageID: env.GetHeader("Message-Id"),
		FileName:       "",
		DomainName:     strings.Split(deliver, "@")[1],
		EmailHash:      *hashContent,
		EmailAddress:   deliver,
		S3Key:          key,
		RecipientEmail: strings.Join(recipients.To, ","),
		SenderName:     from.DisplayName,
		SenderEmail:    from.Address,
		Cc:             strings.Join(recipients.Cc, ","),
		Bcc:            "",
		Subject:        decodeText(env.GetHeader("Subject")),
		BodyText:       env.Text,
		BodyHTML:       env.HTML,
		IsRead:         false,
		EmailType:      global.EmailTypeInbox,
		ReceivedAt:     p,
	}
	//判断是否有附件
	var atts []models.Attachment
	if len(env.Attachments) > 0 || len(env.Inlines) > 0 {
		client, err := CreateS3Client()
		if err != nil {
			return err
		}
		newparts := append(env.Attachments, env.Inlines...)
		atts = EmailAttachmentProcessor(client, bucket, key, newparts)
		global.Log.Info("附件处理完毕，开始保存邮件数据...")
	}
	_, err = dao.AddNewEmailToDB(&r, accountData, atts)
	if err != nil {
		return err
	}
	return nil
}

// 处理附件
//func EmailAttachmentProcessor(client *s3.Client, bucket, key string, parts []*enmime.Part) []models.Attachment {
//	// 计算所有附件的总大小
//	attachments := []models.Attachment{}
//	var totalAttachmentSize int64
//	for _, part := range parts {
//		totalAttachmentSize += int64(len(part.Content))
//	}
//	maxSize := utils.GetMaxAttachmentSize()
//	if totalAttachmentSize > maxSize {
//		global.Log.Errorf("所有附件的总大小为 %d 字节，超过了 %d 字节的限制，附件将不被解析。\n", totalAttachmentSize, maxSize)
//		return attachments
//	}
//
//	for _, part := range parts {
//		// 解码文件名
//		decodedFileName := decodeText(part.FileName)
//		if decodedFileName == "" {
//			global.Log.Warn("附件文件名为空，跳过该附件")
//			continue
//		}
//
//		attachmentSize := int64(len(part.Content))
//		contentType := part.ContentType
//
//		global.Log.Infof("文件名: %s, 大小: %d 字节, MIME类型: %s", decodedFileName, attachmentSize, contentType)
//
//		// 生成附件的哈希值
//		attachmentHash := ComputeAttachmentHash(part.Content)
//		global.Log.Infof("附件哈希: %s", attachmentHash)
//
//		//去数据库比对hash值，如果存在则不再上传
//		attachment, err := dao.GetAttachmentByHash(attachmentHash)
//		if err != nil {
//			global.Log.Errorf("查询附件失败: %v", err)
//			return attachments
//		}
//		if attachment != nil {
//			attachments = append(attachments, *attachment)
//			global.Log.Warnf("附件 %s 已存在，跳过该附件", decodedFileName)
//			continue
//		}
//		// 生成 S3 对象键
//		sanitizedKey := strings.TrimPrefix(key, "email/")
//		objectKey := fmt.Sprintf("attachment/%s/%s", sanitizedKey, decodedFileName)
//
//		// 将附件保存到 S3
//		err = SaveAttachmentToS3(client, bucket, objectKey, part.Content, contentType)
//		if err != nil {
//			global.Log.Errorf("无法保存附件 %s 到 S3: %v", decodedFileName, err)
//			return attachments
//		}
//
//		global.Log.Infof("附件 %s 已保存到 S3，对象键：%s", decodedFileName, objectKey)
//		global.Log.Info("生成短链接Code = > " + shortlink.CreateShortLinkCode(objectKey))
//		// 构造 Attachment 结构体并追加
//		att := models.Attachment{
//			FileHash:       attachmentHash,
//			FileName:       decodedFileName,
//			FileType:       contentType,
//			FileSize:       attachmentSize,
//			S3FromEmailKey: key,
//			ShortUrlCode:   shortlink.CreateShortLinkCode(objectKey),
//			S3StoragePath:  objectKey,
//			ExpireTime:     time.Now().AddDate(0, 0, 6),
//		}
//		attachments = append(attachments, att)
//	}
//
//	err := dao.AddAttachmentToDBBatchPostgres(attachments)
//	if err != nil {
//		global.Log.Errorf("附件信息保存至数据库失败: %v", err)
//		return attachments
//	}
//	return attachments
//}

// 处理附件
func EmailAttachmentProcessor(client *s3.Client, bucket, key string, parts []*enmime.Part) []models.Attachment {
	// 计算所有附件的总大小
	attachments := []models.Attachment{}
	var totalAttachmentSize int64
	for _, part := range parts {
		totalAttachmentSize += int64(len(part.Content))
	}
	maxSize := utils.GetMaxAttachmentSize()
	if totalAttachmentSize > maxSize {
		global.Log.Errorf("所有附件的总大小为 %d 字节，超过了 %d 字节的限制，附件将不被解析。\n", totalAttachmentSize, maxSize)
		return attachments
	}

	for _, part := range parts {
		// 解码文件名
		decodedFileName := decodeText(part.FileName)
		if decodedFileName == "" {
			global.Log.Warn("附件文件名为空，跳过该附件")
			continue
		}

		attachmentSize := int64(len(part.Content))
		contentType := part.ContentType

		global.Log.Infof("文件名: %s, 大小: %d 字节, MIME类型: %s", decodedFileName, attachmentSize, contentType)

		// 生成附件的哈希值
		attachmentHash := utils.ComputeContentHash(part.Content)
		global.Log.Infof("附件哈希: %s", attachmentHash)
		//去数据库比对hash值，如果存在则不再上传
		attachment, err := dao.GetAttachmentByHash(attachmentHash)
		if err != nil {
			global.Log.Errorf("查询附件失败: %v", err)
			return attachments
		}
		if attachment != nil {
			attachments = append(attachments, *attachment)
			global.Log.Warnf("附件 %s 已存在，跳过该附件", decodedFileName)
			continue
		}
		// 生成 S3 对象键
		sanitizedKey := strings.TrimPrefix(key, "email/")
		objectKey := fmt.Sprintf("attachment/%s/%s", sanitizedKey, decodedFileName)

		// 将附件保存到 S3
		err = SaveAttachmentToS3(objectKey, part.Content, contentType)
		if err != nil {
			global.Log.Errorf("无法保存附件 %s 到 S3: %v", decodedFileName, err)
			return attachments
		}

		// 生成预签名 URL
		presignedURL, err := GeneratePresignedURL(objectKey)
		if err != nil {
			global.Log.Errorf("生成预签名 URL 失败: %v", err)
			return attachments
		}
		global.Log.Infof("附件 %s 已保存到 S3，对象键：%s", decodedFileName, objectKey)
		global.Log.Info("生成短链接Code = > " + shortlink.CreateShortLinkCode(objectKey))
		// 构造 Attachment 结构体并追加
		att := models.Attachment{
			FileHash:       attachmentHash,
			FileName:       decodedFileName,
			FileType:       contentType,
			FileSize:       attachmentSize,
			S3FromEmailKey: key,
			ShortUrlCode:   shortlink.CreateShortLinkCode(objectKey),
			DownloadURL:    presignedURL,
			S3StoragePath:  objectKey,
			ExpireTime:     time.Now().Add(time.Hour * time.Duration(global.Config.AWS.FileExpireTime)),
		}
		attachments = append(attachments, att)
	}

	err := dao.AddAttachmentToDBBatchPostgres(attachments)
	if err != nil {
		global.Log.Errorf("附件信息保存至数据库失败: %v", err)
		return attachments
	}
	return attachments
}

// 生成s3对象键
func GenerateS3ObjectKey(fileName string) string {
	// 获取文件扩展名
	ext := filepath.Ext(fileName)
	// 获取文件名（不含扩展名）
	baseName := fileName[:len(fileName)-len(ext)]
	// 生成时间戳
	timestamp := time.Now().Format("20060102")
	// 生成唯一标识符（可选）
	// uuid := generateUUID()

	// 组合生成对象键，例如：attachments/filename_20241007_123456.ext
	objectKey := fmt.Sprintf("web_attachments/%s/%s_%s%s", timestamp, baseName, utils.GenerateRandomFilePrefix(), ext)
	return objectKey
}

// 从s3下载附件
func DownloadAttachmentFromS3(objectKey string) (string, error) {
	ctx := context.TODO()
	client, err := CreateS3Client()
	if err != nil {
		return "", err
	}
	bucket := global.Config.AWS.S3Bucket

	ext := filepath.Ext(objectKey)

	// 创建临时文件
	tempFile, err := os.CreateTemp("", "temp_attachment-*"+ext)
	if err != nil {
		return "", fmt.Errorf("create temp file failed: %w", err)
	}
	defer tempFile.Close()

	// 从S3下载对象
	result, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return "", fmt.Errorf("download attachment failed: %w", err)
	}
	defer result.Body.Close()

	// 将内容写入临时文件
	_, err = io.Copy(tempFile, result.Body)
	if err != nil {
		return "", fmt.Errorf("write temp file failed: %w", err)
	}

	return tempFile.Name(), nil
}

// 保存附件到S3
func SaveAttachmentToS3(objectKey string, content []byte, contentType string) error {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(global.Config.AWS.ConfigRegion),
		config.WithSharedConfigProfile(global.Config.AWS.ConfigProfile),
	)
	if err != nil {
		global.Log.Errorf("load aws config failed: %v", err)
		return err
	}
	client := s3.NewFromConfig(cfg)
	bucket := global.Config.AWS.S3Bucket

	input := &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(content),
		ContentType: aws.String(contentType),
	}

	_, err = client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("upload to s3 failed: %w", err)
	}
	return nil
}

// 预签名URL
func GeneratePresignedURL(key string) (string, error) {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(global.Config.AWS.ConfigRegion),
		config.WithSharedConfigProfile(global.Config.AWS.ConfigProfile),
	)
	if err != nil {
		global.Log.Errorf("load aws config failed: %v", err)
		return "", err
	}
	client := s3.NewFromConfig(cfg)
	bucket := global.Config.AWS.S3Bucket

	presignClient := s3.NewPresignClient(client)
	presignParams := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
		//配置下载属性
	}

	ctx = context.Background()
	presignedGetObjectOutput, err := presignClient.PresignGetObject(ctx, presignParams, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(global.Config.AWS.FileExpireTime) * time.Hour

	})
	if err != nil {
		return "", err
	}

	return presignedGetObjectOutput.URL, nil
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

func GetEmailFileHash(env *enmime.Envelope) string {
	hashConent := fmt.Sprintf("Subject:%s\n"+
		"From:%s\n"+
		"To:%s\n"+
		"Date:%s\n"+
		"Content-Type:%s\n"+
		"Text:%s\n"+
		"HTML:%s",
		decodeText(env.GetHeader("Subject")),
		decodeText(env.GetHeader("From")),
		decodeText(env.GetHeader("To")),
		decodeText(env.GetHeader("Date")),
		env.GetHeader("Content-Type"),
		env.Text,
		env.HTML,
	)
	return utils.ComputeContentHash([]byte(hashConent))
}
