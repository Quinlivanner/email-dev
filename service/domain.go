package service

import (
	"email/controller/response"
	"email/dao"
	"email/models"
	"email/utils"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// GetDomainEmailListProcess 获取域名邮箱列表
func GetDomainEmailListProcess(emailAddress string) (*models.DomainAccountList, error) {
	// 检查用户是否为域名管理员
	domainDetails, err := IsDomainAdmin(emailAddress)
	if err != nil {
		return nil, err
	}
	//身份确认后，开始查询账号列表
	es, err := dao.GetAccountsByDomainID(domainDetails.ID)
	if err != nil {
		// 如果查询失败，返回错误响应
		return nil, errors.New(fmt.Sprintf("%s [ Reason: %s ]", response.GetDomainEmailListFailedCode.ErrMessage, err.Error()))
	}
	// 如果查询成功，返回账户列表
	return &models.DomainAccountList{Accounts: es, Total: len(es)}, nil

}

// AddDomainEmailAccountProcess 处理添加域名邮箱账户的请求
func AddDomainEmailAccountProcess(emailAddress string, addAccountReq models.AddDomainEmailAccount) (*models.NewAccountDetails, error) {
	// 检查用户是否为域名管理员
	domainDetails, err := IsDomainAdmin(emailAddress)
	if err != nil {
		// 如果用户不是域名管理员，返回权限错误
		return nil, err
	}
	// 生成随机密码
	pwd := utils.GenerateSecurePassword()
	// 对密码进行哈希处理
	hashPwd, err := utils.HashPassword(pwd)
	if err != nil {
		// 如果密码哈希失败，返回错误响应
		return nil, errors.New(fmt.Sprintf("%s [ Reason: %s ]", response.CreateEmailAccountFailedCode.ErrMessage, err.Error()))
	}
	// 创建新的邮箱账户
	err = dao.AddAccount(domainDetails.DomainName, addAccountReq.UserName, strings.Split(addAccountReq.EmailAddress, "@")[0], hashPwd)
	if err != nil {
		// 如果账户已存在，返回错误响应
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.New(response.AccountExistCode.ErrMessage)
		}
		// 如果创建账户失败，返回错误响应

		return nil, errors.New(fmt.Sprintf("%s [ Reason: %s ]", response.CreateEmailAccountFailedCode.ErrMessage, err.Error()))
	}
	// 创建成功，返回新账户的详细信息
	return &models.NewAccountDetails{
		EmailAddress: strings.Split(addAccountReq.EmailAddress, "@")[0] + "@" + addAccountReq.DomainName,
		Password:     pwd,
		UserName:     addAccountReq.UserName,
	}, nil
}

// IsDomainAdmin 判断是否是域名邮箱管理员并获取域名数据详情
func IsDomainAdmin(emailAddress string) (*models.Domain, error) {
	// 获取域名
	domain := strings.Split(emailAddress, "@")[1]
	// 检查是否为管理员
	domainDetails, err := dao.GetDomainDetailsByName(domain)
	if err != nil {
		return nil, err
	}
	if domainDetails.AdminEmail != emailAddress {
		return nil, errors.New("You do not have permission to access this domain.")
	}
	return domainDetails, nil
}
