package controller

import (
	"email/controller/response"
	"email/models"
	"email/service"
	"github.com/gin-gonic/gin"
)

type AuthController struct{}

// 登录 Router
func (AuthController) Login(c *gin.Context) {
	// 从上下文中获取 loginRequest
	loginReqInterface, exists := c.Get("loginRequest")
	if !exists {
		response.FailedReq(c, response.MissingCriticalLoginParametersCode)
		return
	}
	loginReq, ok := loginReqInterface.(models.LoginRequest)
	if !ok {
		response.FailedReq(c, response.MissingCriticalLoginParametersCode)
		return
	}
	at, e := service.LoginProcess(loginReq)
	if e != nil {
		response.FailedReq(c, response.LoginFailedCode, e.Error())
		return
	}
	response.SuccessReq(c, at)
	return
}

// 退出登录
func (AuthController) Logout() {

}
