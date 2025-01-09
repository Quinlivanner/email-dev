package service

import (
	"email/models"
	"github.com/gin-gonic/gin"
)

// AIReplyEmailProcess
func AIReplyEmailProcess(c *gin.Context, reqAccount *models.CustomJwtClaims) {
	// 从上下文中获取 AIReplyEmailRequest
	//aiReplyEmailReqInterface, exists := c.Get("AIReplyEmailRequest")
	//if !exists {
	//	response.FailedReq(c, http.StatusBadRequest, response.InvalidParametersCode, "Invalid parameters.")
	//	return
	//}
	//aiReplyEmailReq, ok := aiReplyEmailReqInterface.(models.ReplyEmailRequest)
	//if !ok {
	//	response.FailedReq(c, http.StatusBadRequest, response.InvalidParametersCode, "Invalid parameters.")
	//	return
	//}

}
