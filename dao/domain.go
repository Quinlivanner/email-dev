package dao

import (
	"email/global"
	"email/models"
	"fmt"
)

// 新增域名
func AddDomain(domainName, adminEmail string) error {
	d := models.Domain{
		DomainName: domainName,
		AdminEmail: adminEmail,
	}
	result := global.PsqlDB.Create(&d)
	if result.Error != nil {
		global.Log.Error(fmt.Sprintf("插入 [ %s ] 域名失败: ", domainName), result.Error)
		return result.Error
	}
	return nil
}

// 查询指定域名信息
func GetDomainDetailsByName(domainName string) (*models.Domain, error) {
	var domain models.Domain
	result := global.PsqlDB.Where("domain_name = ?", domainName).First(&domain)
	if result.Error != nil {
		global.Log.Error(fmt.Sprintf("查询域名 [ %s ] 失败: ", domainName), result.Error)
		return nil, result.Error
	}
	return &domain, nil
}
