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
func AIRouterInit(r *gin.RouterGroup) {
	var AIController controller.DomainController
	domain := r.Group("/ai")
	{
		// * 对域名下的 account 进行操作
		{
			domain.GET("/reply", middleware.AIReplyMiddleware(), AIController.GetDomainEmailList)
		}
	}
}
