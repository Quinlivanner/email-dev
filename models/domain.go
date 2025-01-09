package models

import (
	"email/global"
	"time"
)

// Domain 表示 domains 表的 GORM 模型
type Domain struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`               // 主键，自增
	DomainName string    `gorm:"type:varchar(255);uniqueIndex;not null"` // 域名名称，唯一且不为空
	AdminEmail string    `gorm:"type:varchar(255);not null"`             // 管理员邮箱，不为空
	CreatedAt  time.Time `gorm:"autoCreateTime"`                         // 创建时间，自动填充
}

func (Domain) TableName() string {
	return global.Config.DatabseTableNames.Domains
}
