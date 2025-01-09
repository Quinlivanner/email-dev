package v1

import (
	"email/controller"
	"email/middleware"

	"github.com/gin-gonic/gin"
)

/*
* v1 下关于domain的Api初始化
! API 风格待定 > RESTful Api or 普通
TODO: 与前端对接采用何种形式
*/

// @title Domain Operation Api
// @version v1.0
// @description 关于域名下的功能操作
// @host localhost:8080
// @BasePath /api/v1
func EmailRouterInit(r *gin.RouterGroup) {
	var EmailController controller.EmailController
	email := r.Group("/email")
	{
		// * 对域名下的 account 进行操作
		{
			// * 获取所有邮件
			email.GET("/inbox", func(c *gin.Context) { EmailController.GetEmailList(c, "inbox") })
			// * 获取所有已发送邮件
			email.GET("/sent", func(c *gin.Context) { EmailController.GetEmailList(c, "sent") })
			// * 获取所有垃圾邮件
			email.GET("/trash", func(c *gin.Context) { EmailController.GetEmailList(c, "trash") })
			// * 获取所有已删除邮件
			email.GET("/deleted", func(c *gin.Context) { EmailController.GetEmailList(c, "deleted") })
			// * 移动邮件到指定分类

			//获取邮件详情
			email.GET("/details", EmailController.GetEmailDetails)

			// * 获取邮件列表
			email.GET("/inbox/uid", func(c *gin.Context) { EmailController.GetEmailListByEmailId(c, "inbox") })
			// * 获取已发送邮件列表
			email.GET("/sent/uid", func(c *gin.Context) { EmailController.GetEmailListByEmailId(c, "sent") })
			// * 获取垃圾邮件列表
			email.GET("/trash/uid", func(c *gin.Context) { EmailController.GetEmailListByEmailId(c, "trash") })
			// * 获取已删除邮件列表
			email.GET("/deleted/uid", func(c *gin.Context) { EmailController.GetEmailListByEmailId(c, "deleted") })
			// * 移动邮件到指定分类
			email.POST("/move", middleware.MoveEmailMiddleware(), EmailController.MoveEmail)
			// * 标记邮件为已读
			email.GET("/read/uid", EmailController.MakeEmailReadByEmailID)
			// * 标记邮件为未读
			email.GET("/unread/uid", EmailController.MakeEmailUnreadByEmailID)

			// * 发送邮件
			email.POST("/new/send", middleware.SendEmailMiddleware(), EmailController.SendNewEmail)
			// * 回复邮件
			email.POST("/new/reply", middleware.ReplyEmailMiddleware(), EmailController.ReplyEmail)
			// * 网页端发送邮件
			email.POST("/web/send", middleware.WebSendEmailMiddleware(), EmailController.WebSendNewEmail)

			// * 获取最新的收件箱邮件
			email.GET("/latest/inbox", EmailController.GetLatestInboxEmailList)
		}
	}

	/*网页端API*/
	/*网页端API*/
	//---------------------------------------------------------------------------------------------------------------------------

	//---------------------------------------------------------------------------------------------------------------------------

}
