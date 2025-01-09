package controller

import (
	"email/models"
	"email/utils"
	"github.com/gin-gonic/gin"
	"strings"
)

// 获取token以及一个包含信息的结构体
func getAccountDataByParseAccessToken(c *gin.Context) (*models.CustomJwtClaims, error) {
	tokenStr := c.GetHeader("Authorization")
	token := strings.Replace(tokenStr, "Bearer ", "", 1)
	e, err := utils.ParseJwtToken(token)
	if err != nil {
		return nil, err
	}
	return e, nil
}
