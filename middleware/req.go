package middleware

import (
	"email/controller/response"
	"email/global"
	"email/models"
	"email/utils"

	"github.com/gin-gonic/gin"
)

// 验证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		// TODO: validate token
		c.Next()
	}
}

// 登录中间件
func LoginMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginReq models.LoginRequest
		if err := c.ShouldBindJSON(&loginReq); err != nil {
			response.FailedReq(c, response.ParseDataFailedCode)
			c.Abort()
			return
		}
		c.Set("loginRequest", loginReq)
		c.Next()
	}

}

// 添加邮箱账户中间件
func AddDomainEmailAccountMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var addAccountReq models.AddDomainEmailAccount
		if err := c.ShouldBindJSON(&addAccountReq); err != nil {
			response.FailedReq(c, response.ParseDataFailedCode)
			c.Abort()
			return
		}
		if err := utils.ValidatePassword(addAccountReq.Password); err != nil {
			response.FailedReq(c, response.IncorrectPasswordFormatCode)
			c.Abort()
			return
		}
		c.Set("addAccountReq", addAccountReq)
		c.Next()
	}
}

// 移动邮件中间件
func MoveEmailMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var moveEmailReq models.MoveEmailRequest
		if err := c.ShouldBindJSON(&moveEmailReq); err != nil {
			response.FailedReq(c, response.ParseDataFailedCode)
			c.Abort()
			return
		}

		// 验证源类别和目标类别是否相同
		if moveEmailReq.SourceType == moveEmailReq.TargetType {
			response.FailedReq(c, response.InvalidParametersCode, "Source and target types cannot be the same.")
			c.Abort()
			return
		}

		// 判断源类别和目标类别是否在global.EmailType范围内
		if !utils.Contains(global.EmailType, moveEmailReq.SourceType) || !utils.Contains(global.EmailType, moveEmailReq.TargetType) {
			response.FailedReq(c, response.InvalidParametersCode, "Invalid source or target type.")
			c.Abort()
			return
		}

		// 验证邮件ID是否有效
		if moveEmailReq.EmailID <= 0 {
			response.FailedReq(c, response.IncorrectEmailIDParameterCode)
			c.Abort()
			return
		}
		c.Set("moveEmailReq", moveEmailReq)
		c.Next()
	}
}

// 发送邮件中间件
func SendEmailMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var sendEmailReq models.SendNewEmailRequest
		if err := c.ShouldBindJSON(&sendEmailReq); err != nil {
			response.FailedReq(c, response.ParseDataFailedCode)
			c.Abort()
			return
		}
		// 验证邮箱地址是否有效
		if err := utils.ValidateEmailAddress(sendEmailReq.To); err != nil {
			response.FailedReq(c, response.IncorrectEmailAddressCode, err.Error())
			c.Abort()
			return
		}
		// 验证邮件主题是否为空
		if sendEmailReq.Subject == "" || sendEmailReq.TextBody == "" {
			response.FailedReq(c, response.InvalidParametersCode, "The subject or body cannot be empty.")
			c.Abort()
			return
		}
		c.Set("sendNewEmailReq", sendEmailReq)
		c.Next()
	}
}

// 网页端发送邮件中间件
func WebSendEmailMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var webSendEmailReq models.WebSendEmailRequest
		if err := c.ShouldBindJSON(&webSendEmailReq); err != nil {
			response.FailedReq(c, response.ParseDataFailedCode)
			c.Abort()
			return
		}
		// 验证邮箱地址是否有效
		for _, email := range webSendEmailReq.To {
			if err := utils.ValidateEmailAddress(email); err != nil {
				response.FailedReq(c, response.IncorrectEmailAddressCode, err.Error())
				c.Abort()
				return
			}
		}
		c.Set("webSendEmailReq", webSendEmailReq)
		c.Next()
	}
}

// 回复邮件中间件
func ReplyEmailMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var replyEmailReq models.ReplyEmailRequest
		if err := c.ShouldBindJSON(&replyEmailReq); err != nil {
			response.FailedReq(c, response.ParseDataFailedCode)
			c.Abort()
			return
		}
		// 验证邮箱地址是否有效
		if err := utils.ValidateEmailAddress(replyEmailReq.To); err != nil {
			response.FailedReq(c, response.IncorrectEmailAddressCode, err.Error())
			c.Abort()
			return
		}
		// 验证邮件主题和内容是否为空
		if replyEmailReq.Subject == "" || (replyEmailReq.TextBody == "" && replyEmailReq.HtmlBody == "") {
			response.FailedReq(c, response.InvalidParametersCode, "The subject or body cannot be empty.")
			c.Abort()
			return
		}
		c.Set("replyEmailReq", replyEmailReq)
		c.Next()
	}
}

// AI 回复中间件
func AIReplyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var aiReplyEmailReq models.AIReplyEmailRequest
		if err := c.ShouldBindJSON(&aiReplyEmailReq); err != nil {
			response.FailedReq(c, response.ParseDataFailedCode, "Submit data parsing failed.")
			c.Abort()
			return
		}
		// 验证
		// 验证邮件主题和内容是否为空
		if aiReplyEmailReq.Subject == "" || (aiReplyEmailReq.TextBody == "" && aiReplyEmailReq.HtmlBody == "") {
			response.FailedReq(c, response.InvalidParametersCode, "The subject or body cannot be empty.")
			c.Abort()
			return
		}
		c.Set("aiReplyEmailReq", aiReplyEmailReq)
		c.Next()
	}
}
