package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 统一错误响应
type ApiErrorRes struct {
	ErrorCodeInfo ErrorCodeInfo
	ErrMessage    string
}

type ErrorCodeInfo struct {
	//自定义错误代码
	Code int
	//应返回http状态码
	HttpStatus int
	//错误信息
	ErrMessage string
}

var (
	//成功 [成功]
	SuccessCode = ErrorCodeInfo{0, http.StatusOK, "Success"}
	//一般错误 [一般错误]
	GeneralErrorCode = ErrorCodeInfo{1, http.StatusInternalServerError, "General error occurred"}
	//获取账户信息失败 - [ 详情见报错 ]
	GetAccountDetailsFailedCode = ErrorCodeInfo{7701, http.StatusInternalServerError, "Failed to retrieve account details"}
	//无效的请求参数 [请求参数不正确或缺失]
	InvalidParametersCode = ErrorCodeInfo{7001, http.StatusBadRequest, "Invalid request parameters"}
	//未授权 [未携带Access Token 请求]
	UnauthorizedCode = ErrorCodeInfo{7002, http.StatusUnauthorized, "Unauthorized access"}
	//登录失败 [登录信息解析失败或格式不正确]
	LoginFailedCode = ErrorCodeInfo{7003, http.StatusBadRequest, "Login failed due to parsing error or incorrect format"}
	//账户已存在 [账户已存在]
	AccountExistCode = ErrorCodeInfo{7004, http.StatusBadRequest, "Account already exists"}
	//账户不存在 [账户不存在]
	AccountNotFoundCode = ErrorCodeInfo{7005, http.StatusBadRequest, "Account not found"}
	//邮箱地址错误 [邮箱地址格式不正确]
	IncorrectEmailAddressCode = ErrorCodeInfo{7006, http.StatusBadRequest, "Incorrect email address format"}
	//密码错误 [密码错误]
	IncorrectPasswordCode = ErrorCodeInfo{7007, http.StatusBadRequest, "Incorrect password"}
	//密码格式不合格 [ 8-17 位置，至少包含 1 个字母，1 个数字，1 个特殊字符 ]
	IncorrectPasswordFormatCode = ErrorCodeInfo{7008, http.StatusBadRequest, "Password format does not meet requirements,the password must be a combination of letters, numbers and special characters from 8-17 digits."}
	//关键登录参数丢失 [登录信息解析失败或格式不正确]
	MissingCriticalLoginParametersCode = ErrorCodeInfo{7009, http.StatusBadRequest, "Missing critical login parameters"}
	//生成JWT令牌失败 [建议用户再试一次]
	GenerateJwtTokenFailedCode = ErrorCodeInfo{7010, http.StatusInternalServerError, "Failed to generate JWT token"}
	//无效的access token [Jwt token 过期或有错]
	InvalidAccessTokenCode = ErrorCodeInfo{7011, http.StatusUnauthorized, "Invalid access token"}
	//域名无效 [域名不存在于数据库或已删除]
	InvalidDomainCode = ErrorCodeInfo{7012, http.StatusBadRequest, "Invalid domain"}
	//获取域名信息失败 [数据库查询失败]
	GetDomainDetailsFailedCode = ErrorCodeInfo{7013, http.StatusInternalServerError, "Failed to retrieve domain details"}
	//获取域名邮件列表失败 [数据库查询失败]
	GetDomainEmailListFailedCode = ErrorCodeInfo{7014, http.StatusInternalServerError, "Failed to retrieve domain email list"}
	//没有管理员权限操作域名邮箱 [域名管理员权限验证失败]
	NoPermissionToAccessDomainEmail = ErrorCodeInfo{7015, http.StatusForbidden, "No permission to access domain email"}
	//添加域名邮箱账户失败 [详情见报错]
	CreateEmailAccountFailedCode = ErrorCodeInfo{7016, http.StatusInternalServerError, "Failed to create email account"}
	//更新域名邮箱账户失败 [详情见报错]
	UpdateEmailAccountFailedCode = ErrorCodeInfo{7017, http.StatusInternalServerError, "Failed to update email account"}
	//删除域名邮箱账户失败 [详情见报错]
	DeleteEmailAccountFailedCode = ErrorCodeInfo{7018, http.StatusInternalServerError, "Failed to delete email account"}
	//获取收件箱邮件列表失败 [详情见报错]
	GetInboxEmailListFailedCode = ErrorCodeInfo{7019, http.StatusInternalServerError, "Failed to retrieve inbox email list"}
	//获取发件箱邮件列表失败 [详情见报错]
	GetSentEmailListFailedCode = ErrorCodeInfo{7020, http.StatusInternalServerError, "Failed to retrieve sent email list"}
	//获取草稿箱邮件列表失败 [详情见报错]
	GetDraftEmailListFailedCode = ErrorCodeInfo{7021, http.StatusInternalServerError, "Failed to retrieve draft email list"}
	//获取垃圾箱邮件列表失败 [详情见报错]
	GetTrashEmailListFailedCode = ErrorCodeInfo{7022, http.StatusInternalServerError, "Failed to retrieve trash email list"}
	//获取邮件详情失败 [详情见报错]
	GetEmailDetailsFailedCode = ErrorCodeInfo{7023, http.StatusInternalServerError, "Failed to retrieve email details"}
	//发送邮件失败 [详情见报错]
	SendEmailFailedCode = ErrorCodeInfo{7024, http.StatusInternalServerError, "Failed to send email"}
	//删除邮件失败 [详情见报错]
	DeleteEmailFailedCode = ErrorCodeInfo{7025, http.StatusInternalServerError, "Failed to delete email"}
	//移动邮件失败 [详情见报错]
	MoveEmailFailedCode = ErrorCodeInfo{7026, http.StatusInternalServerError, "Failed to move email"}
	//获取邮件附件失败 [详情见报错]
	GetEmailAttachmentsFailedCode = ErrorCodeInfo{7027, http.StatusInternalServerError, "Failed to retrieve email attachments"}
	//邮箱Page参数错误 [Page参数必须为Int]
	IncorrectPageParameterCode = ErrorCodeInfo{7028, http.StatusBadRequest, "Incorrect page parameter"}
	//邮件ID参数错误 [在请求邮件详情时，必须携带正确的邮件ID]
	IncorrectEmailIDParameterCode = ErrorCodeInfo{7029, http.StatusBadRequest, "Incorrect email ID parameter"}
	//邮件未找到 [在请求邮件详情或者移动邮件类别时，邮件ID不存在或格式不正确]
	EmailNotFoundCode = ErrorCodeInfo{7030, http.StatusNotFound, "Email not found"}
	//标记邮件为已读失败 [详情见报错]
	MarkEmailAsReadFailedCode = ErrorCodeInfo{7031, http.StatusInternalServerError, "Failed to mark email as read"}
	//回复邮件时，收件人和邮件发件人不匹配 [回复邮件时，收件人和邮件发件人不匹配]
	ReplyEmailToAddressNotMatch = ErrorCodeInfo{7032, http.StatusBadRequest, "Recipient and email sender do not match"}
	//提交数据解析失败 [提交数据解析失败]
	ParseDataFailedCode = ErrorCodeInfo{7033, http.StatusBadRequest, "Failed to parse submitted data"}
	//获取邮件列表失败 [详情见报错]
	GetEmailListFailedCode = ErrorCodeInfo{7034, http.StatusInternalServerError, "Failed to retrieve email list"}
	//缺少请求参数 [缺少请求参数]
	MissingParametersCode = ErrorCodeInfo{7035, http.StatusBadRequest, "Missing parameters"}
	//密码哈希失败 [密码哈希失败]
	HashPasswordFailedCode = ErrorCodeInfo{7036, http.StatusInternalServerError, "Failed to hash password"}
	//收件人与发件人相同 [收件人与发件人相同]
	RecipientAndSenderSameCode = ErrorCodeInfo{7037, http.StatusBadRequest, "Recipient and sender are the same"}
	//回复邮件失败 [详情见报错]
	ReplyEmailFailedCode = ErrorCodeInfo{7038, http.StatusInternalServerError, "Failed to reply email"}
	//标记邮件为未读失败 [详情见报错]
	MarkEmailAsUnReadFailedCode = ErrorCodeInfo{7039, http.StatusInternalServerError, "Failed to mark email as unread"}
	//数据库查询失败
	DatabaseQueryFailedCode = ErrorCodeInfo{7040, http.StatusInternalServerError, "Query failed"}
	//上传附件到s3失败 [详情见报错]
	UploadAttachmentToS3FailedCode = ErrorCodeInfo{7041, http.StatusInternalServerError, "Failed to upload attachment"}
	//修改密码失败 [详情见报错]
	UpdatePasswordFailedCode = ErrorCodeInfo{7042, http.StatusInternalServerError, "Failed to update password"}
)

// Response 定义统一的响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 发送成功响应
func SuccessReq(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    SuccessCode.Code,
		Message: "Success",
		Data:    data,
	})

}

// Fail 发送失败响应
func FailedReq(c *gin.Context, errCode ErrorCodeInfo, details ...string) {
	message := errCode.ErrMessage
	if len(details) > 0 && details[0] != "" {
		message = details[0]
	}
	c.JSON(errCode.HttpStatus, Response{
		Code:    errCode.Code,
		Message: message,
	})

}

// Unauthorized 发送未授权响应
func Unauthorized(c *gin.Context) {
	FailedReq(c, ErrorCodeInfo{Code: UnauthorizedCode.Code, HttpStatus: http.StatusUnauthorized}, "Unauthorised.")
}
