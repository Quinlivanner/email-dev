package models

import (
	"email/global"
	"fmt"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Attachments 表示 attachments 表的 GORM 模型
// Attachment 既可用于 GORM 也可作为普通结构体
type Attachment struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	FileHash       string    `gorm:"type:char(64);uniqueIndex" json:"file_hash"`
	FileName       string    `gorm:"type:varchar(255)" json:"file_name"`
	FileType       string    `gorm:"type:varchar(255)" json:"file_type"`
	FileSize       int64     `json:"file_size"`
	S3FromEmailKey string    `gorm:"type:varchar(512)" json:"s3_from_email_key"`
	ShortUrlCode   string    `gorm:"type:varchar(255);default:'N/A'" json:"short_url_code"`
	DownloadURL    string    `gorm:"type:varchar(1024)" json:"download_url"`
	S3StoragePath  string    `gorm:"type:varchar(512);default:'N/A'" json:"s3_storage_path"`
	ExpireTime     time.Time `json:"expire_time"`
}

func (Attachment) TableName() string {
	return global.Config.DatabseTableNames.Attachments
}

// EmailDetails 表示用户 emails 表的 GORM 模型
type EmailDetails struct {
	ID             uint           `gorm:"primaryKey;" json:"id"`
	EmailMessageID string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email_message_id,omitempty"`
	FileName       string         `gorm:"type:varchar(255);index;not null" json:"file_name,omitempty"`                                                // 添加索引和非空约束
	DomainID       uint           `gorm:"foreignKey:DomainID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;not null;index" json:"domain_id,omitempty"` // 添加索引
	DomainName     string         `gorm:"type:varchar(255);not null" json:"domain_name,omitempty"`                                                    // 添加类型
	EmailHash      string         `gorm:"type:char(64);uniqueIndex" json:"email_hash,omitempty"`
	EmailAccountID uint           `gorm:"foreignKey:EmailAccountID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;not null;index" json:"email_account_id"` // 添加索引
	EmailAddress   string         `gorm:"type:varchar(255);not null" json:"email_address"`                                                               // 添加非空约束
	RecipientEmail string         `gorm:"type:varchar(255);not null" json:"recipient_email"`
	S3Key          string         `gorm:"type:varchar(255);not null" json:"s3_key,omitempty"`
	SenderName     string         `gorm:"type:varchar(255)" json:"sender_name,omitempty"`
	SenderEmail    string         `gorm:"type:varchar(255);not null" json:"sender_email"`
	Cc             string         `gorm:"type:varchar(255)" json:"cc,omitempty"`
	Bcc            string         `gorm:"type:varchar(255)" json:"bcc,omitempty"`
	ReplyEmailID   uint           `gorm:"index" json:"reply_email_id,omitempty"` // 添加索引
	Subject        string         `gorm:"type:varchar(255);not null" json:"subject,omitempty"`
	BodyText       string         `gorm:"type:text" json:"body_text,omitempty"`
	BodyHTML       string         `gorm:"type:text" json:"body_html,omitempty"`
	IsRead         bool           `gorm:"not null;default:false" json:"is_read"`
	EmailType      string         `gorm:"type:varchar(255);not null" json:"email_type,omitempty"` // 移除错误的 default:false
	ReceivedAt     time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"received_at"`  // 修改时间默认值
	AttachmentInfo datatypes.JSON `gorm:"type:jsonb;default:'[]';not null" json:"attachment_info,omitempty"`
	CreatedAt      time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at,omitempty"`                    // 修改时间默认值
	LastUpdateAt   time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"last_update_at,omitempty"` // 添加自动更新
}

func (e *EmailDetails) BeforeCreate(tx *gorm.DB) (err error) {
	tx.Statement.Table = fmt.Sprintf("user_%d_emails", e.EmailAccountID)
	return
}

// func (e *EmailDetails) TableName() string {
// 	return fmt.Sprintf("user_%d_emails", e.EmailAccountID)
// }

// SentEmail 表示 sent_emails 表的 GORM 模型
type SentEmail struct {
	ID             uint           `gorm:"primaryKey;autoIncrement"`                                                         // 主键，自增
	DomainID       uint           `gorm:"foreignKey:DomainID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;not null"`        // 关联的域名 ID，不为空
	DomainName     string         `gorm:"not null"`                                                                         // 关联的域名对象
	EmailAccountID uint           `gorm:"foreignKe y:EmailAccountID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;not null"` // 关联的邮箱账号 ID，不为空
	EmailAddress   string         `gorm:"not null"`                                                                         // 关联的邮箱账号对象
	RecipientEmail string         `gorm:"type:varchar(255);not null"`                                                       // 收件人邮箱，不为空
	Subject        string         `gorm:"type:varchar(255)"`                                                                // 邮件主题，可为空
	BodyText       string         `gorm:"type:text"`                                                                        // 邮件正文（纯文本），可为空
	BodyHTML       string         `gorm:"type:text"`                                                                        // 邮件正文（HTML），可为空
	SentTimestamp  time.Time      `gorm:"not null"`                                                                         // 邮件发送时间，不为空
	AttachmentInfo datatypes.JSON `gorm:"type:jsonb;default:'[]';not null"`                                                 // 附件信息，JSON 格式，不为空
	CreatedAt      time.Time      `gorm:"autoCreateTime"`                                                                   // 创建时间，自动填充
}

func (SentEmail) TableName() string {
	return global.Config.DatabseTableNames.SentEmails
}
