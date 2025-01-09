// global/v.go 文件定义了应用程序的全局变量和配置。
// 它包含了数据库连接、配置信息和日志记录器等重要的全局对象。

package global

import (
	"database/sql"
	"email/config"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	Config    *config.Config
	SqlDB     *sql.DB
	PsqlDB    *gorm.DB
	Log       *logrus.Logger
	EmailType = []string{EmailTypeInbox, EmailTypeSent, EmailTypeDeleted, EmailTypeTrash, EmailTypeDraft}
)

const (
	EmailTypeInbox   = "inbox"
	EmailTypeSent    = "sent"
	EmailTypeTrash   = "trash"
	EmailTypeDraft   = "draft"
	EmailTypeDeleted = "deleted"
)

type EmailFlag string

const (
	FlagSeen    EmailFlag = "S" // 已读
	FlagReplied EmailFlag = "R" // 已回复
	FlagFlagged EmailFlag = "F" // 已标记/星标
	FlagDraft   EmailFlag = "D" // 草稿
	FlagTrashed EmailFlag = "T" // 已删除
)

const (
	ChangeReadStatus = "change_read_status"
	MoveEmail        = "move_email"
	AppendEmail      = "append_email"
)
