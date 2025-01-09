package service

import (
	"email/controller/response"
	"email/dao"
	"email/global"
	"email/models"
	"email/service/aws"
	"email/service/dovecot"
	"email/utils"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jhillyerd/enmime"
	"gorm.io/gorm"
)

type apiErrorRes response.ApiErrorRes

// GetEmailListProcess 获取邮件列表
func GetEmailListProcess(emailAddress string, userID uint, emailType string, page string) (*models.EmailList, error) {
	//将page转为int
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt <= 0 {
		return nil, errors.New(response.IncorrectPageParameterCode.ErrMessage)
	}
	// 调用数据访问层获取收件箱邮件列表
	emails, total, err := dao.GetEmailList(userID, pageInt, global.Config.API.EmailCountPerPage, emailType)
	if err != nil {
		//获取邮件失败，返回实际错误
		return nil, err
	}
	return &models.EmailList{
		EmailAddress: emailAddress,
		Emails:       emails,
		EmailType:    emailType,
		Total:        total,
		Page:         pageInt,
	}, nil
}

// MoveEmailProcess 处理移动邮件的请求
func MoveEmailProcess(userID uint, moveEmailReq models.MoveEmailRequest) (bool, error) {
	// 调用数据访问层移动邮件
	go func(models.MoveEmailRequest) {
		sourceMailbox := ""
		targetMailBox := ""
		emailDetail, err := dao.GetEmailDetailFullFileds(moveEmailReq.EmailID, userID)
		if err != nil {
			return
		}
		switch moveEmailReq.SourceType {
		case "inbox":
			sourceMailbox = "Inbox"
		case "trash":
			sourceMailbox = "Junk"
		case "deleted":
			sourceMailbox = "Trash"
		}

		switch moveEmailReq.TargetType {
		case "inbox":
			targetMailBox = "Inbox"
		case "trash":
			targetMailBox = "Junk"
		case "deleted":
			targetMailBox = "Trash"
		}

		options := models.ImapOperatorOptions{
			UserName:      emailDetail.EmailAddress,
			MoveType:      global.MoveEmail,
			SourceMailBox: sourceMailbox,
			TargetMailBox: targetMailBox,
			MessageID:     emailDetail.EmailMessageID,
		}

		dovecot.ImapOperatorEmail(options)

	}(moveEmailReq)

	err := dao.MoveEmail(moveEmailReq.EmailID, userID, moveEmailReq.SourceType, moveEmailReq.TargetType)
	if err != nil {
		return false, err
	}

	return true, nil
}

// GetEmailDetailsProcess 获取邮件详情
func GetEmailDetailsProcess(userID uint, emailId int) (*models.EmailDetailsResponse, error) {
	// 调用数据访问层获取邮件详情
	email, err := dao.GetEmailDetails(emailId, userID)
	if err != nil {
		return nil, err
	}
	return email, nil
}

// SendNewEmailProcess 处理发送新邮件的请求
func SendNewEmailProcess(userID uint, emailAddress string, sendNewEmailReq models.SendNewEmailRequest) (int, error) {
	// 根据userid获取用户信息进行比对from email address
	account, err := dao.GetAccountByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, errors.New(response.AccountNotFoundCode.ErrMessage)
		}
		return 0, err
	}
	//数据库中邮件地址是否存在
	if account.EmailAddress == "" {
		return 0, errors.New(response.AccountNotFoundCode.ErrMessage)
	}
	//无法发送邮件给自己
	if account.EmailAddress == sendNewEmailReq.To {
		return 0, errors.New("Cannot send email to yourself.")
	}
	//token邮件地址和数据库中是否匹配
	if account.EmailAddress != emailAddress {
		return 0, errors.New(response.AccountNotFoundCode.ErrMessage)
	}
	//调用AWS SES 服务
	err = aws.SendNewEmailByAwsSes(&sendNewEmailReq, account.EmailAddress, account.UserName)
	if err != nil {
		return 0, err
	}
	//调用Dao层保存邮件
	newEmail := models.EmailDetails{
		DomainID:       account.DomainID,
		DomainName:     account.DomainName,
		EmailAccountID: account.ID,
		EmailAddress:   account.EmailAddress,
		RecipientEmail: sendNewEmailReq.To,
		S3Key:          "N/A",
		SenderName:     account.UserName,
		SenderEmail:    account.EmailAddress,
		Subject:        sendNewEmailReq.Subject,
		BodyText:       sendNewEmailReq.TextBody,
		BodyHTML:       sendNewEmailReq.HtmlBody,
		IsRead:         true,
		EmailType:      "sent",
		ReceivedAt:     time.Now(),
		LastUpdateAt:   time.Now(),
	}
	eid, err := dao.AddNewEmailToDB(&newEmail, account, nil)
	if err != nil {
		return 0, err
	}
	// 返回成功响应
	return int(eid), nil
}

// WebSendNewEmailProcess 网页端发送邮件
func WebSendNewEmailProcess(userID uint, emailAddress string, webSendEmailReq *models.WebSendEmailRequest) (int, error) {
	// 根据userid获取用户信息进行比对from email address
	account, err := dao.GetAccountByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, errors.New(response.AccountNotFoundCode.ErrMessage)
		}
		return 0, err
	}
	//数据库中邮件地址是否存在
	if account.EmailAddress == "" {
		return 0, errors.New(response.AccountNotFoundCode.ErrMessage)
	}
	//token邮件地址和数据库中是否匹配
	if account.EmailAddress != emailAddress {
		return 0, errors.New(response.AccountNotFoundCode.ErrMessage)
	}
	////调用AWS SES 服务
	rawMessage, msgId, atts, err := aws.SendEmailByAwsSmtp(emailAddress, webSendEmailReq)
	if err != nil {
		return 0, err
	}
	//调用Dao层保存邮件
	newEmail := models.EmailDetails{
		DomainID:       account.DomainID,
		DomainName:     account.DomainName,
		EmailAccountID: account.ID,
		EmailAddress:   account.EmailAddress,
		RecipientEmail: strings.Join(webSendEmailReq.To, ","),
		Cc:             strings.Join(webSendEmailReq.Cc, ","),
		Bcc:            strings.Join(webSendEmailReq.Bcc, ","),
		S3Key:          "N/A",
		SenderName:     account.UserName,
		SenderEmail:    account.EmailAddress,
		Subject:        webSendEmailReq.Subject,
		BodyText:       webSendEmailReq.TextBody,
		BodyHTML:       webSendEmailReq.HtmlBody,
		IsRead:         true,
		EmailType:      global.EmailTypeSent,
		ReceivedAt:     time.Now(),
		LastUpdateAt:   time.Now(),
		EmailMessageID: msgId,
		FileName:       "N/A",
		EmailHash:      utils.ComputeContentHash([]byte(rawMessage)),
	}
	eid, err := dao.AddNewEmailToDB(&newEmail, account, atts)
	if err != nil {
		return 0, err
	}
	// 返回成功响应
	fmt.Println(rawMessage)

	if global.Config.System.Env {

		options := models.ImapOperatorOptions{
			UserName:        emailAddress,
			MoveType:        global.AppendEmail,
			TargetMailBox:   "Sent",
			MessageID:       msgId,
			EmailRawMessage: []byte(rawMessage),
		}

		dovecot.ImapOperatorEmail(options)
	}

	return int(eid), nil
}

// ReplyEmailProcess 处理回复邮件的请求
func ReplyEmailProcess(userID uint, emailAddress, userName string, replyEmailReq models.ReplyEmailRequest) (int, error) {

	// 根据userid和email id 获取详细的邮件信息
	emailDetails, err := dao.GetEmailDetailFullFileds(replyEmailReq.EmailID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, errors.New(response.EmailNotFoundCode.ErrMessage)
		}
		return 0, err
	}
	//判断收件人
	if emailDetails.SenderEmail != replyEmailReq.To {
		return 0, errors.New(response.ReplyEmailToAddressNotMatch.ErrMessage)
	}
	//无法发送邮件给自己
	if emailDetails.EmailAddress == replyEmailReq.To {
		return 0, errors.New(response.RecipientAndSenderSameCode.ErrMessage)
	}
	//邮件地址和数据库中是否匹配
	if emailDetails.EmailAddress != emailAddress {
		return 0, errors.New(response.AccountNotFoundCode.ErrMessage)
	}
	//调用AWS SES 服务
	err = aws.ReplyEmailByAwsSes(&replyEmailReq, emailAddress, userName)
	if err != nil {
		return 0, err
	}

	emailContent, messageId := utils.GenerateEmailRawMessage(&models.EmailContent{
		From:     fmt.Sprintf("%s <%s>", userName, emailAddress),
		To:       replyEmailReq.To,
		Subject:  replyEmailReq.Subject,
		TextBody: replyEmailReq.TextBody,
		HtmlBody: replyEmailReq.HtmlBody,
	})

	if global.Config.System.Env {
		options := models.ImapOperatorOptions{
			UserName:        emailAddress,
			MoveType:        global.AppendEmail,
			TargetMailBox:   "Sent",
			MessageID:       messageId,
			EmailRawMessage: []byte(emailContent),
		}

		err = dovecot.ImapOperatorEmail(options)
		if err != nil {
			return 0, err
		}
	}

	//调用Dao层保存邮件
	newEmail := models.EmailDetails{
		DomainID:       emailDetails.DomainID,
		FileName:       "",
		EmailMessageID: messageId,
		DomainName:     emailDetails.DomainName,
		EmailAccountID: emailDetails.EmailAccountID,
		EmailAddress:   emailDetails.EmailAddress,
		RecipientEmail: emailDetails.SenderEmail,
		S3Key:          "N/A",
		SenderName:     userName,
		SenderEmail:    emailDetails.EmailAddress,
		Subject:        replyEmailReq.Subject,
		ReplyEmailID:   emailDetails.ID,
		BodyText:       replyEmailReq.TextBody,
		BodyHTML:       strings.ReplaceAll(replyEmailReq.HtmlBody, "\n", ""),
		IsRead:         true,
		EmailType:      "sent",
		ReceivedAt:     time.Now(),
		LastUpdateAt:   time.Now(),
	}
	eid, err := dao.AddNewEmailToDB(&newEmail, &models.EmailAccount{ID: emailDetails.EmailAccountID, EmailAddress: emailDetails.EmailAddress, DomainID: emailDetails.DomainID}, nil)
	if err != nil {
		return 0, err
	}
	return int(eid), nil
}

// GetLatestInboxEmailListProcess 获取最新邮件列表
func GetLatestInboxEmailListProcess(userID uint, emailAddress string, lastEmailID int) (*models.EmailList, error) {

	// 调用数据访问层获取收件箱邮件列表
	emails, total, err := dao.GetLatestInboxEmailList(userID, lastEmailID)
	if err != nil {
		//获取邮件失败，返回实际错误
		return nil, err
	}
	return &models.EmailList{
			EmailAddress: emailAddress,
			Emails:       emails,
			EmailType:    "inbox",
			Total:        total,
			Page:         1},
		nil
}

// ----------------------------------------------------------------------------------------------------------------------
// GetEmailListProcess 获取邮件列表
func GetEmailListByEmailIDProcess(userID uint, emailAddress string, emailType string, emailID int) (*models.EmailList, error) {
	// 调用数据访问层获取收件箱邮件列表
	emails, total, err := dao.GetEmailListByID(userID, emailID, emailType)
	if err != nil {
		return nil, err
	}
	return &models.EmailList{
		EmailAddress: emailAddress,
		Emails:       emails,
		EmailType:    emailType,
		Total:        total,
	}, nil
}

// MakeEmailReadByEmailIDProcess 将指定邮件标记为已读
func MakeEmailReadByEmailIDProcess(emailDetail *models.EmailDetails) error {
	// 调用数据访问层将邮件标记为已读
	sourceMailBox := ""
	switch emailDetail.EmailType {
	case "inbox":
		sourceMailBox = "Inbox"
	case "trash":
		sourceMailBox = "Junk"
	case "deleted":
		sourceMailBox = "Trash"
	case "sent":
		sourceMailBox = "Sent"

	}
	if global.Config.System.Env {

		options := models.ImapOperatorOptions{
			UserName:      emailDetail.EmailAddress,
			MoveType:      global.ChangeReadStatus,
			SourceMailBox: sourceMailBox,
			MessageID:     emailDetail.EmailMessageID,
			ReadStatus:    true,
		}
		err := dovecot.ImapOperatorEmail(options)
		if err != nil {
			return err
		}
	}

	err := dao.MarkEmailAsRead(emailDetail.EmailAccountID, int(emailDetail.ID))
	if err != nil {
		return err
	}
	return nil
}

// MakeEmailReadByEmailIDProcess 将指定邮件标记为未读
func MakeEmailUnReadByEmailIDProcess(emailDetail *models.EmailDetails) error {
	// 调用数据访问层将邮件标记为已读
	sourceMailBox := ""
	switch emailDetail.EmailType {
	case "inbox":
		sourceMailBox = "Inbox"
	case "trash":
		sourceMailBox = "Junk"
	case "deleted":
		sourceMailBox = "Trash"
	case "sent":
		sourceMailBox = "Sent"

	}

	if global.Config.System.Env {
		options := models.ImapOperatorOptions{
			UserName:      emailDetail.EmailAddress,
			MoveType:      global.ChangeReadStatus,
			SourceMailBox: sourceMailBox,
			MessageID:     emailDetail.EmailMessageID,
			ReadStatus:    false,
		}
		err := dovecot.ImapOperatorEmail(options)
		if err != nil {
			return err
		}
	}
	// 调用数据访问层将邮件标记为已读
	err := dao.MarkEmailAsUnRead(emailDetail.EmailAccountID, int(emailDetail.ID))
	if err != nil {
		return err
	}
	return nil

}

// ----------------------------------------------------------------------------------------------------------------------
// SaveThirtyPartySendEmailProcess 保存第三方发送的邮件
func SaveThirtyPartySendEmailProcess(env *enmime.Envelope, from string, to, cc, bcc []string) error {
	ad, err := dao.IsAccountExist(from, strings.Split(from, "@")[1])
	if err != nil {
		return err
	}
	receivedTime, err := utils.ParseTime(env.GetHeader("Date"))
	if err != nil {
		return fmt.Errorf("解析邮件时间失败: %v", err)
	}
	s3Key, err := utils.GenerateSendID()
	if err != nil {
		return err
	}
	client, err := aws.CreateS3Client()
	if err != nil {
		return err
	}
	atts := aws.EmailAttachmentProcessor(client, global.Config.AWS.S3Bucket, s3Key, env.Attachments)
	// 创建邮件详情对象
	emailDetails := &models.EmailDetails{
		DomainName:     strings.Split(from, "@")[1],
		EmailAddress:   from,
		RecipientEmail: strings.Join(to, ","),
		S3Key:          s3Key,
		SenderName:     strings.Split(from, "@")[0],
		SenderEmail:    from,
		EmailMessageID: env.GetHeader("Message-Id"),
		Cc:             strings.Join(cc, ","),
		Bcc:            strings.Join(bcc, ","),
		Subject:        env.GetHeader("Subject"),
		BodyText:       env.Text,
		BodyHTML:       env.HTML,
		IsRead:         true,
		EmailType:      global.EmailTypeSent,
		ReceivedAt:     receivedTime,
	}

	// 保存邮件到数据库
	_, err = dao.AddNewEmailToDB(emailDetails, ad, atts)
	if err != nil {
		return fmt.Errorf("保存邮件到数据库失败: %v", err)
	}
	return nil
}
