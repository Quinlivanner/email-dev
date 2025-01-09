package v1

import (
	"email/middleware"

	"github.com/gin-gonic/gin"
)

// v1 Api 初始化
func V1RouterInit(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	// 筛选认证的路由组
	publicRoutes := v1.Group("/")
	{
		AuthRouterInit(publicRoutes)
	}
	// 需要认证的路由组
	protectedRoutes := v1.Group("/")
	protectedRoutes.Use(middleware.AuthMiddleware())
	{
		FileManageRouterInit(protectedRoutes)
		DomainRouterInit(protectedRoutes)
		AccountRouterInit(protectedRoutes)
		EmailRouterInit(protectedRoutes)
		// 在这里添加其他需要认证的路由初始化
	}
}
