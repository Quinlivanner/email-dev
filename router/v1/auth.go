package v1

import (
	"email/controller"
	"email/middleware"

	"github.com/gin-gonic/gin"
)

func AuthRouterInit(r *gin.RouterGroup) {
	var AuthController controller.AuthController

	// 公共路由
	r.POST("/auth/login", middleware.LoginMiddleware(), AuthController.Login)

	// 受保护的路由
	auth := r.Group("/auth")
	auth.Use(middleware.AuthMiddleware())
	{
		//auth.POST("/register", nil)
		// auth.POST("/refresh", nil)
		// auth.GET("/logout", nil)
	}
}
