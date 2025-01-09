package controller

import (
	"email/controller/response"
	"email/dao"
	"email/models"
	"email/service"
	"fmt"
	"strconv"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type EmailController struct{}

// GetEmailList 获取收件箱邮件列表  -> Inbox
func (EmailController) GetEmailList(c *gin.Context, emailType string) {
	// 获取并验证 token
	reqAccount, err := getAccountDataByParseAccessToken(c)
	if err != nil {
		response.FailedReq(c, response.InvalidAccessTokenCode)
		return
	}
	page := c.Query("page")
	if page == "" {
		page = "1"
	}
	//进入service层开始具体处理
	eList, e := service.GetEmailListProcess(reqAccount.EmailAddress, reqAccount.UserID, emailType, page)
	if e != nil {
		response.FailedReq(c, response.GetEmailListFailedCode, e.Error())
		return
	}
	response.SuccessReq(c, eList)
}

// MoveEmail 移动邮件到对应分类
func (EmailController) MoveEmail(c *gin.Context) {
	// 获取并验证 token
	reqAccount, err := getAccountDataByParseAccessToken(c)
	if err != nil {
		response.FailedReq(c, response.InvalidAccessTokenCode)
		return
	}
	// 从上下文中获取 moveEmailReq
	moveEmailReqInterface, exists := c.Get("moveEmailReq")
	if !exists {
		response.FailedReq(c, response.InvalidParametersCode)
		return
	}
	moveEmailReq, ok := moveEmailReqInterface.(models.MoveEmailRequest)
	if !ok {
		response.FailedReq(c, response.InvalidParametersCode)
		return
	}
	// 进入service层开始具体处理
	s, err := service.MoveEmailProcess(reqAccount.UserID, moveEmailReq)
	if err != nil && !s {
		if err == gorm.ErrRecordNotFound {
			response.FailedReq(c, response.EmailNotFoundCode)
			return
		}
		response.FailedReq(c, response.MoveEmailFailedCode, err.Error())
		return
	}
	response.SuccessReq(c, nil)
}

// SendNewEmail 发送新邮件
func (EmailController) SendNewEmail(c *gin.Context) {
	// 获取并验证 token
	reqAccount, err := getAccountDataByParseAccessToken(c)
	if err != nil {
		response.FailedReq(c, response.InvalidAccessTokenCode)
		return
	}
	// 从上下文中获取 sendNewEmailReq
	sendNewEmailReqInterface, exists := c.Get("sendNewEmailReq")
	if !exists {
		response.FailedReq(c, response.MissingParametersCode)
		return
	}
	sendNewEmailReq, ok := sendNewEmailReqInterface.(models.SendNewEmailRequest)
	if !ok {
		response.FailedReq(c, response.InvalidParametersCode)
		return
	}
	// 进入service层开始具体处理
	eid, err := service.SendNewEmailProcess(reqAccount.UserID, reqAccount.EmailAddress, sendNewEmailReq)
	if err != nil {
		response.FailedReq(c, response.SendEmailFailedCode, err.Error())
		return
	}
	response.SuccessReq(c, gin.H{"email_id": eid})
}

// WebSendNewEmail 网页端发送邮件
func (EmailController) WebSendNewEmail(c *gin.Context) {
	// 获取并验证 token
	reqAccount, err := getAccountDataByParseAccessToken(c)
	if err != nil {
		response.FailedReq(c, response.InvalidAccessTokenCode)
		return
	}
	// 从上下文中获取 webSendEmailReq
	webSendEmailReqInterface, exists := c.Get("webSendEmailReq")
	if !exists {
		response.FailedReq(c, response.MissingParametersCode)
		return
	}
	webSendEmailReq, ok := webSendEmailReqInterface.(models.WebSendEmailRequest)
	if !ok {
		response.FailedReq(c, response.InvalidParametersCode)
		return
	}
	// 进入service层开始具体处理
	eid, err := service.WebSendNewEmailProcess(reqAccount.UserID, reqAccount.EmailAddress, &webSendEmailReq)
	if err != nil {
		response.FailedReq(c, response.SendEmailFailedCode, err.Error())
		return
	}
	response.SuccessReq(c, gin.H{"email_id": eid})
}

// GetEmailDetails 获取邮件详情
func (EmailController) GetEmailDetails(c *gin.Context) {
	// 获取并验证 token
	reqAccount, err := getAccountDataByParseAccessToken(c)
	if err != nil {
		response.FailedReq(c, response.InvalidAccessTokenCode)
		return
	}
	//提取url中的emailID参数
	emailID := c.Query("email_id")
	if emailID == "" {
		response.FailedReq(c, response.IncorrectEmailIDParameterCode)
		return
	}
	//将emailID转为int
	emailIDInt, err := strconv.Atoi(emailID)
	if err != nil || emailIDInt <= 0 {
		response.FailedReq(c, response.IncorrectEmailIDParameterCode)
		return
	}
	//进入service层开始具体处理
	eDetail, err := service.GetEmailDetailsProcess(reqAccount.UserID, emailIDInt)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.FailedReq(c, response.EmailNotFoundCode)
			return
		}
		response.FailedReq(c, response.GetEmailDetailsFailedCode, err.Error())
		return
	}
	response.SuccessReq(c, eDetail)
}

// ReplyEmail 回复邮件
func (EmailController) ReplyEmail(c *gin.Context) {
	// 获取并验证 token
	reqAccount, err := getAccountDataByParseAccessToken(c)
	if err != nil {
		response.FailedReq(c, response.InvalidAccessTokenCode)
		return
	}
	// 从上下文中获取 sendNewEmailReq
	replyEmailReqInterface, exists := c.Get("replyEmailReq")
	if !exists {
		response.FailedReq(c, response.MissingParametersCode)
		return
	}
	replyEmailReq, ok := replyEmailReqInterface.(models.ReplyEmailRequest)
	if !ok {
		response.FailedReq(c, response.InvalidParametersCode, "Invalid parameters.")
		return
	}
	// 进入service层开始具体处理
	eid, err := service.ReplyEmailProcess(reqAccount.UserID, reqAccount.EmailAddress, reqAccount.UserName, replyEmailReq)
	if err != nil {
		response.FailedReq(c, response.ReplyEmailFailedCode, err.Error())
		return
	}
	response.SuccessReq(c, gin.H{
		"email_id": eid,
	})
}

// GetLatestInboxEmailList 获取最新收件箱邮件列表  -> Inbox
func (EmailController) GetLatestInboxEmailList(c *gin.Context) {
	// 获取并验证 token
	reqAccount, err := getAccountDataByParseAccessToken(c)
	if err != nil {
		response.FailedReq(c, response.InvalidAccessTokenCode)
		return
	}
	fmt.Println("c.Query(\"eid\") => ", c.Query("eid"))
	//提取url中的last_email_id参数
	lastEmailID := c.Query("eid")
	if lastEmailID == "" {
		lastEmailID = "1"
	}

	//将last_email_id转为int
	lastEmailIDInt, err := strconv.Atoi(lastEmailID)
	if err != nil || lastEmailIDInt <= 0 {
		response.FailedReq(c, response.IncorrectEmailIDParameterCode)
		return
	}
	//进入service层开始具体处理
	es, err := service.GetLatestInboxEmailListProcess(reqAccount.UserID, reqAccount.EmailAddress, lastEmailIDInt)
	if err != nil {
		response.FailedReq(c, response.GetEmailListFailedCode, err.Error())
		return
	}
	response.SuccessReq(c, es)
}

/*网页端API---------------------------------------------------------------------------------------------------------------*/
// GetEmailList 使用last_email_id获取收件箱邮件列表  -> Inbox
func (EmailController) GetEmailListByEmailId(c *gin.Context, emailType string) {
	// 获取并验证 token
	reqAccount, err := getAccountDataByParseAccessToken(c)
	if err != nil {
		response.FailedReq(c, response.InvalidAccessTokenCode)
		return
	}
	//提取url中的last_email_id参数
	id := c.Query("last_email_id")
	if id == "" {
		response.FailedReq(c, response.IncorrectEmailIDParameterCode)
		return
	}
	//将last_email_id转为int
	lastEmailIDInt, err := strconv.Atoi(id)
	if err != nil || lastEmailIDInt < 0 {
		response.FailedReq(c, response.IncorrectEmailIDParameterCode)
		return
	}
	//进入service层开始具体处理
	es, err := service.GetEmailListByEmailIDProcess(reqAccount.UserID, reqAccount.EmailAddress, emailType, lastEmailIDInt)
	if err != nil {
		response.FailedReq(c, response.GetInboxEmailListFailedCode, err.Error())
		return
	}
	response.SuccessReq(c, es)
}

// MakeEmailReadByEmailID 使用email_id将邮件阅读状态标记为true
func (EmailController) MakeEmailReadByEmailID(c *gin.Context) {
	// 获取并验证 token
	reqAccount, err := getAccountDataByParseAccessToken(c)
	if err != nil {
		response.FailedReq(c, response.InvalidAccessTokenCode)
		return
	}
	// 从URL参数中获取邮件ID
	emailID := c.Query("email_id")
	if emailID == "" {
		response.FailedReq(c, response.IncorrectEmailIDParameterCode)
		return
	}
	// 将邮件ID转换为整数
	emailIDInt, err := strconv.Atoi(emailID)
	if err != nil || emailIDInt <= 0 {
		response.FailedReq(c, response.IncorrectEmailIDParameterCode)
		return
	}
	details, err := dao.GetEmailDetailFullFileds(emailIDInt, reqAccount.UserID)
	if err != nil {
		return
	}
	// 进入service层开始具体处理
	err = service.MakeEmailReadByEmailIDProcess(details)
	if err != nil {
		response.FailedReq(c, response.MarkEmailAsReadFailedCode, err.Error())
		return
	}
	response.SuccessReq(c, nil)
}

// MakeEmailUnreadByEmailID 使用email_id将邮件阅读状态标记为false
func (EmailController) MakeEmailUnreadByEmailID(c *gin.Context) {
	// 获取并验证 token
	reqAccount, err := getAccountDataByParseAccessToken(c)
	if err != nil {
		response.FailedReq(c, response.InvalidAccessTokenCode)
		return
	}
	// 从URL参数中获取邮件ID
	emailID := c.Query("email_id")
	if emailID == "" {
		response.FailedReq(c, response.IncorrectEmailIDParameterCode)
		return
	}
	// 将邮件ID转换为整数
	emailIDInt, err := strconv.Atoi(emailID)
	if err != nil || emailIDInt <= 0 {
		response.FailedReq(c, response.IncorrectEmailIDParameterCode)
		return
	}
	details, err := dao.GetEmailDetailFullFileds(emailIDInt, reqAccount.UserID)
	if err != nil {
		return
	}
	// 进入service层开始具体处理
	err = service.MakeEmailUnReadByEmailIDProcess(details)
	if err != nil {
		response.FailedReq(c, response.MarkEmailAsUnReadFailedCode, err.Error())
		return
	}
	response.SuccessReq(c, nil)
}
