package aws

import (
	"email/dao"
	"email/global"
	"email/models"
	"fmt"
	"strings"

	"github.com/google/uuid"
	mail "github.com/xhit/go-simple-mail/v2"
)

func SendEmailByAwsSmtp(senderEmail string, webSendEmailReq *models.WebSendEmailRequest) (string, string, []models.Attachment, error) {
	server := mail.NewSMTPClient()
	server.Host = global.Config.AWS.SmtpHost
	server.Port = global.Config.AWS.SmtpPort
	server.Username = global.Config.AWS.SmtpUsername
	server.Password = global.Config.AWS.SmtpPassword
	server.Encryption = mail.EncryptionSTARTTLS

	smtpClient, err := server.Connect()
	if err != nil {
		return "", "", nil, err
	}

	msgId := fmt.Sprintf("<%s@%s>", uuid.New().String(), strings.Split(senderEmail, "@")[1])
	email := mail.NewMSG()
	email.AddHeader("Message-ID", msgId)

	//email.SetFrom(senderEmail).SetReturnPath(senderEmail).AddTo(webSendEmailReq.To...).AddCc(webSendEmailReq.Cc...).AddBcc(webSendEmailReq.Bcc...)
	email.SetFrom(senderEmail).AddTo(webSendEmailReq.To...).AddCc(webSendEmailReq.Cc...).AddBcc(webSendEmailReq.Bcc...)

	email.SetSubject(webSendEmailReq.Subject)

	var attachments []models.Attachment

	if len(webSendEmailReq.Attachments) > 0 {
		fileList, atts, err := dao.GetAttachmentsDataByCodes(webSendEmailReq.Attachments)
		if err != nil {
			return "", "", nil, err
		}
		attachments = atts
		for _, file := range fileList {
			path, err := DownloadAttachmentFromS3(file.FileKey)
			if err != nil {
				return "", "", nil, err
			}
			email.Attach(&mail.File{FilePath: path, Name: file.FileName})
		}
	}

	email.SetBody(mail.TextPlain, webSendEmailReq.TextBody)
	email.AddAlternative(mail.TextHTML, webSendEmailReq.HtmlBody)

	if email.Error != nil {
		return "", "", nil, email.Error
	}

	if err = email.Send(smtpClient); err != nil {
		return "", "", nil, err
	}

	return email.GetMessage(), msgId, attachments, nil
}
