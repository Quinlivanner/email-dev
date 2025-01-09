package router

import (
	_ "email/docs"
	v1 "email/router/v1"
	"time"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type RouterEngine struct {
	*gin.Engine
}

/*
	在Go语言中，可以使用以下方式进行段落注释：
	! 这里有个小问题
	? .
	TODO: .
	* .
*/

// InitRouter
func InitRouter() {
	r := gin.Default()
	//r.Use(middleware.RateLimitMiddleware())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173", "https://mail.web.mail.mountex.net", "https://webmail.mountex.net"}, // 允许的前端地址
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	//router.Use()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 初始化 v1 版本的路由
	v1.V1RouterInit(r)

	r.Run(":8080")
}
