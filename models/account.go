package models

import (
	"email/global"
	"time"
)

// EmailAccount 表示 email_accounts 表的 GORM 模型
type EmailAccount struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"`                                                  // 主键，自增
	DomainID     uint      `gorm:"foreignKey:DomainID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;not null"` // 关联的域名 ID，不为空
	DomainName   string    `gorm:"type:varchar(255);not null"`                                                // 域名名称，不为空
	EmailAddress string    `gorm:"type:varchar(255);uniqueIndex;not null"`                                    // 邮箱地址，唯一且不为空
	PasswordHash string    `gorm:"type:varchar(255);not null"`                                                // 哈希后的密码，不为空
	JwtTokenHash string    `gorm:"type:text"`                                                                 // JWT令牌
	UserName     string    `gorm:"type:varchar(255);not null"`                                                // 用户名，不为空
	Status       string    `gorm:"type:varchar(50);not null;default:'active'"`                                // 账号状态，默认 'active'
	StorageUsed  int64     `gorm:"type:bigint;not null;default:0" json:"storage_used"`                        // 新增字段
	CreatedAt    time.Time `gorm:"autoCreateTime;not null"`
}

func (EmailAccount) TableName() string {
	return global.Config.DatabseTableNames.EmailAccounts
}

// SafeEmailAccount 表示 email_accounts 表回传前端模型
type SafeEmailAccount struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"`                                                  // 主键，自增
	DomainID     uint      `gorm:"foreignKey:DomainID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;not null"` // 关联的域名 ID，不为空
	DomainName   string    `gorm:"type:varchar(255);not null"`                                                // 域名名称，不为空
	EmailAddress string    `gorm:"type:varchar(255);uniqueIndex;not null"`                                    // 邮箱地址，唯一且不为空
	UserName     string    `gorm:"type:varchar(255);not null"`                                                // 用户名，不为空
	Status       string    `gorm:"type:varchar(50);not null;default:'active'"`                                // 账号状态，默认 'active'
	StorageUsed  int64     `gorm:"type:bigint;not null;default:0" json:"storage_used"`                        // 新增字段
	CreatedAt    time.Time `gorm:"autoCreateTime;not null"`                                                   // 创建时间，自动填充，不为空

}

// 在 EmailAccount 结构体中添加一个方法来创建 SafeEmailAccount
func (ea *EmailAccount) ToSafeEmailAccount() SafeEmailAccount {
	return SafeEmailAccount{
		ID:           ea.ID,
		DomainID:     ea.DomainID,
		DomainName:   ea.DomainName,
		EmailAddress: ea.EmailAddress,
		UserName:     ea.UserName,
		Status:       ea.Status,
		StorageUsed:  ea.StorageUsed,
		CreatedAt:    ea.CreatedAt,
	}
}
