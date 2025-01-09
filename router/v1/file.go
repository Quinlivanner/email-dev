package v1

import (
	"email/controller"

	"github.com/gin-gonic/gin"
)

func FileManageRouterInit(r *gin.RouterGroup) {
	var FileManageController controller.FileManageController
	file := r.Group("/file")
	Upload := file.Group("/upload")
	Download := file.Group("/download")

	Upload.POST("/", FileManageController.UploadFile)
	Download.GET("/:code", FileManageController.DownloadFile)
}
