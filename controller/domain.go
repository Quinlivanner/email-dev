package controller

import (
	"email/controller/response"
	"email/models"
	"email/service"
	"github.com/gin-gonic/gin"
)

type DomainController struct{}

// GetDomainEmailList 获取域名邮箱列表
func (DomainController) GetDomainEmailList(c *gin.Context) {
	// 获取并验证 token
	reqAccount, err := getAccountDataByParseAccessToken(c)
	if err != nil {
		response.FailedReq(c, response.InvalidAccessTokenCode)
		return
	}
	//进入service层开始具体处理
	es, e := service.GetDomainEmailListProcess(reqAccount.EmailAddress)
	if e != nil {
		response.FailedReq(c, response.GetDomainEmailListFailedCode, e.Error())
		return
	}
	response.SuccessReq(c, es)
	return
}

// AddDomainEmailAccount 添加邮件账户
func (DomainController) AddDomainEmailAccount(c *gin.Context) {
	// 获取并验证 token
	reqAccount, err := getAccountDataByParseAccessToken(c)
	if err != nil {
		response.FailedReq(c, response.InvalidAccessTokenCode)
		return
	}
	// 从上下文中获取 addAccountReq
	addAccountReqInterface, exists := c.Get("addAccountReq")
	if !exists {
		// 如果请求信息不存在，返回错误响应
		response.FailedReq(c, response.MissingCriticalLoginParametersCode)
	}
	// 将接口类型断言为 AddDomainEmailAccount 类型
	addAccountReq, ok := addAccountReqInterface.(models.AddDomainEmailAccount)
	if !ok {
		// 如果类型断言失败，返回错误响应
		response.FailedReq(c, response.MissingCriticalLoginParametersCode)
	}
	// 调用 service 层处理添加邮件账户的逻辑
	acd, e := service.AddDomainEmailAccountProcess(reqAccount.EmailAddress, addAccountReq)
	if e != nil {
		response.FailedReq(c, response.CreateEmailAccountFailedCode, e.Error())
		return
	}
	response.SuccessReq(c, acd)
}
