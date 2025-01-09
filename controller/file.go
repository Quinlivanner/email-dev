package controller

import (
	"email/controller/response"
	"email/dao"
	"email/models"
	"email/service/aws"
	"email/service/shortlink"
	"email/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type FileManageController struct{}

// RedirectToOriginal 重定向到原始 S3 链接
func (FileManageController) DownloadFile(c *gin.Context) {
	code := c.Param("code")
	//判断code是否存在
	if code == "" {
		response.FailedReq(c, response.MissingParametersCode, "code is required")
		return
	}
	//根据code获取附件信息
	att, err := dao.GetAttachmentByCode(code)
	if err != nil {
		response.FailedReq(c, response.DatabaseQueryFailedCode, err.Error())
		return
	}
	if att == nil {
		response.FailedReq(c, response.GetEmailAttachmentsFailedCode, "attachment not found")
		return
	}

	//根据附件信息获取文件
	url, err := aws.GeneratePresignedURL(att.S3StoragePath)
	if err != nil {
		response.FailedReq(c, response.GetEmailAttachmentsFailedCode, err.Error())
		return
	}
	response.SuccessReq(c, gin.H{
		"url": url,
	})
	return

}

// 上传文件Controller
func (FileManageController) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.FailedReq(c, response.GetEmailAttachmentsFailedCode, err.Error())
		return
	}

	// 获取文件类型和大小
	contentType := file.Header.Get("Content-Type")
	fileSize := file.Size
	// 检查文件大小是否超过33MB
	if fileSize > 33*1024*1024 {
		response.FailedReq(c, response.GetEmailAttachmentsFailedCode, "文件大小不能超过33MB")
		return
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		response.FailedReq(c, response.GetEmailAttachmentsFailedCode, err.Error())
		return
	}
	defer src.Close()

	// 读取文件内容
	buffer := make([]byte, fileSize)
	_, err = src.Read(buffer)
	if err != nil {
		response.FailedReq(c, response.GetEmailAttachmentsFailedCode, err.Error())
		return
	}

	attachmentHash := utils.ComputeContentHash(buffer)
	exist, err := dao.GetAttachmentByHash(attachmentHash)

	if err != nil {
		response.FailedReq(c, response.DatabaseQueryFailedCode, err.Error())
	}
	if exist != nil {
		response.SuccessReq(c, gin.H{
			"filename": file.Filename,
			"size":     fileSize,
			"type":     contentType,
			"code":     exist.ShortUrlCode,
		})
		return
	}
	// 生成S3对象键
	objectKey := aws.GenerateS3ObjectKey(file.Filename)

	// 保存到S3
	err = aws.SaveAttachmentToS3(objectKey, buffer, contentType)
	if err != nil {
		response.FailedReq(c, response.UploadAttachmentToS3FailedCode, err.Error())
		return
	}

	// 生成预签名 URL
	presignedURL, err := aws.GeneratePresignedURL(objectKey)
	if err != nil {
		response.FailedReq(c, response.UploadAttachmentToS3FailedCode, err.Error())
		return
	}

	// 构造 Attachment 结构体并追加
	att := models.Attachment{
		FileHash:       attachmentHash,
		FileName:       file.Filename,
		FileType:       contentType,
		FileSize:       fileSize,
		S3FromEmailKey: "WEB_SEND",
		ShortUrlCode:   shortlink.CreateShortLinkCode(objectKey),
		DownloadURL:    presignedURL,
		S3StoragePath:  objectKey,
		ExpireTime:     time.Now().AddDate(0, 0, 6),
	}

	err = dao.AddAttachmentToDBBatchPostgres([]models.Attachment{att})
	if err != nil {
		response.FailedReq(c, response.UploadAttachmentToS3FailedCode, err.Error())
		return
	}
	//返回的时候，移除att部分参数
	response.SuccessReq(c, gin.H{
		"filename": file.Filename,
		"size":     fileSize,
		"type":     contentType,
		"code":     att.ShortUrlCode,
	})
}
