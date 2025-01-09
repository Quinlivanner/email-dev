package controller

import (
	"email/controller/response"
	"email/dao"
	"email/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AccountController struct{}

// GetAccountPassword 获取账号密码
func (AccountController) GetAccountPassword(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "GetAccountPassword",
		"status":  200,
		"result":  nil,
		"data":    nil,
	})
}

func (AccountController) UpdateAccountPassword(c *gin.Context) {
	// 获取并验证 token
	reqAccount, err := getAccountDataByParseAccessToken(c)
	if err != nil {
		response.FailedReq(c, response.InvalidAccessTokenCode)
		return
	}
	var req models.UpdateAccountPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailedReq(c, response.ParseDataFailedCode, err.Error())
		return
	}
	if req.NewPassword != req.ConfirmPassword {
		response.FailedReq(c, response.InvalidParametersCode, "new password and confirm password do not match")
		return
	}
	err = dao.UpdateAccountPassword(reqAccount.EmailAddress, req.CurrentPassword, req.NewPassword)
	if err != nil {
		response.FailedReq(c, response.UpdatePasswordFailedCode, err.Error())
		return
	}
	response.SuccessReq(c, nil)
}
