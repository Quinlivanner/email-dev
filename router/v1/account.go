package v1

import (
	"email/controller"

	"github.com/gin-gonic/gin"
)

// v1 下 关于account的Api初始化
func AccountRouterInit(r *gin.RouterGroup) {
	AccountController := controller.AccountController{}

	account := r.Group("/account")
	{
		update := account.Group("/update")
		{
			update.POST("/password", AccountController.UpdateAccountPassword)
		}

		//獲取郵箱信息
	}
}
