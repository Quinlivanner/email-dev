package models

import (
	"time"

	"gorm.io/datatypes"

	"github.com/dgrijalva/jwt-go"
)

// ---------------------------------------------------------------------------
//type LoginResult struct {
//	Token     AccessToken
//	ErrorCode response.APIErrorCode
//	ErrorMsg  string
//	Error     error
//}

// --------------------------------------------------------------------------

// 登录请求结构体
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=30"`
}

// 添加域名邮箱账户结构体
type AddDomainEmailAccount struct {
	DomainName   string `json:"domain_name" binding:"required"`
	EmailAddress string `json:"email_address" binding:"required,email"`
	Password     string `json:"password"`
	UserName     string `json:"user_name"`
}

// --------------------------------------------------------------------------
// 结构体用于回传信息给前端

// jwt token
type CustomJwtClaims struct {
	EmailAddress string `json:"email_address"`
	UserName     string `json:"userName"`
	UserID       uint   `json:"userId"`
	jwt.StandardClaims
}

// access token
type AccessToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

// 包含账户信息和总数
type DomainAccountList struct {
	Accounts []SafeEmailAccount `json:"accounts"`
	Total    int                `json:"total"`
}

// 新账户
type NewAccountDetails struct {
	EmailAddress string `json:"email_address"`
	Password     string `json:"password"`
	UserName     string `json:"user_name"`
}

// inbox邮件列表
type EmailList struct {
	EmailAddress string                 `json:"email_address"`
	Emails       []EmailDetailsResponse `json:"emails"`
	EmailType    string                 `json:"email_type"`
	Total        int64                  `json:"total"`
	Page         int                    `json:"page,omitempty"`
}

// 邮件API返回结构体
type EmailDetailsResponse struct {
	ID             uint           `json:"id"`
	EmailAccountID uint           `json:"email_account_id"`
	EmailAddress   string         `json:"email_address"`
	RecipientEmail string         `json:"recipient_email"`
	SenderName     string         `json:"sender_name"`
	SenderEmail    string         `json:"sender_email"`
	Cc             string         `json:"cc"`
	Subject        string         `json:"subject"`
	BodyText       string         `json:"body_text"`
	BodyHTML       string         `json:"body_html"`
	IsRead         bool           `json:"is_read"`
	AttachmentInfo datatypes.JSON `json:"attachment_info"`
	ReceivedAt     time.Time      `json:"received_at"`
}

// 移动邮件结构体
type MoveEmailRequest struct {
	EmailID    int    `json:"email_id" binding:"required"`
	SourceType string `json:"source_type" binding:"required"`
	TargetType string `json:"target_type" binding:"required"`
}

type WebSendEmailRequest struct {
	To          []string             `json:"to"`
	Cc          []string             `json:"cc"`
	Bcc         []string             `json:"bcc"`
	Subject     string               `json:"subject"`
	TextBody    string               `json:"text_body"`
	HtmlBody    string               `json:"html_body"`
	Attachments []FrontendAttachment `json:"attachments"`
}

// 前端附件结构体
type FrontendAttachment struct {
	Code     string `json:"code"`
	Filename string `json:"filename"`
}

// Web send mail att generate
type WebSendMailAttGenerate struct {
	FileCode string
	FileName string
	FileKey  string
	FilePath string
}

type ResponseAttachmentData struct {
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	FileCode string `json:"short_url_code"`
}

// 发送新邮件结构体
type SendNewEmailRequest struct {
	To          string   `json:"to" binding:"required"`
	Cc          []string `json:"cc,omitempty"`
	Subject     string   `json:"subject" binding:"required"`
	TextBody    string   `json:"text_body" binding:"required"`
	HtmlBody    string   `json:"html_body" binding:"required"`
	Attachments []string `json:"attachments,omitempty"`
}

// 回复邮件结构体
type ReplyEmailRequest struct {
	EmailID     int      `json:"email_id" binding:"required"`
	To          string   `json:"to" binding:"required,email"`
	Subject     string   `json:"subject" binding:"required"`
	TextBody    string   `json:"text_body" binding:"required"`
	HtmlBody    string   `json:"html_body" binding:"required"`
	Attachments []string `json:"attachments,omitempty"`
	CC          []string `json:"cc,omitempty"`
	BCC         []string `json:"bcc,omitempty"`
}

// AI回复邮件结构体
type AIReplyEmailRequest struct {
	Subject     string   `json:"subject" binding:"required"`
	TextBody    string   `json:"text_body" binding:"required"`
	HtmlBody    string   `json:"html_body" binding:"required"`
	Attachments []string `json:"attachments,omitempty"`
}

type ImapOperatorOptions struct {
	UserName        string
	MoveType        string
	SourceMailBox   string
	TargetMailBox   string
	MessageID       string
	ReadStatus      bool
	EmailRawMessage []byte
}

// 修改密码请求结构体
type UpdateAccountPasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
}
