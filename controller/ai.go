package controller

import (
	"email/controller/response"
	"github.com/gin-gonic/gin"
)

type AIController struct{}

// AIReplyEmail
func (AIController) AIReplyEmail(c *gin.Context) {
	// 获取并验证 token
	_, err := getAccountDataByParseAccessToken(c)
	if err != nil {
		response.FailedReq(c, response.InvalidAccessTokenCode, "Invalid access token.")
		return
	}
	// 进入service层开始具体处理
	//service.MakeEmailReadByEmailIDProcess(c, reqAccount)
}
