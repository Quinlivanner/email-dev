package dao

import (
	"email/global"
	"email/models"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// 新邮件入库 - 包括收件发件 [待完善]
func AddNewEmailToDB(r *models.EmailDetails, d *models.EmailAccount, atts []models.Attachment) (uint, error) {
	global.Log.Info(fmt.Sprintf("开始插入 [ %s ] 邮件", r.S3Key))
	// 使用事务执行数据库操作
	var emailID uint
	err := global.PsqlDB.Transaction(func(gd *gorm.DB) error {
		// 将结构体序列化为 JSON 字节
		// #转换为 datatypes.JSON 类型
		if atts != nil {
			jsonBytes, err := json.Marshal(atts)
			if err != nil {
				global.Log.Error("JSON 1序列化错误:", err)
				return err
			}
			r.AttachmentInfo = jsonBytes
		}
		// 设置外键
		r.EmailAccountID = d.ID
		r.DomainID = d.DomainID
		// 插入新邮件记录
		if err := gd.Create(&r).Error; err != nil {
			global.Log.Error(fmt.Sprintf("插入 [ %s ] 邮件失败: ", r.S3Key), err)
			return err // 返回错误会触发事务回滚
		}
		emailID = r.ID // 获取插入后的ID
		global.Log.Info(fmt.Sprintf("插入 [ %s ] 邮件成功", r.S3Key))
		return nil // 返回 nil 提交事务
	})
	if err != nil {
		// 事务回滚时的错误处理
		return 0, err
	}
	return emailID, nil
}

// 获取收件箱邮件列表
func GetEmailList(accountID uint, page, pageSize int, emailType string) ([]models.EmailDetailsResponse, int64, error) {
	var emails []models.EmailDetailsResponse
	var total int64
	tableName := fmt.Sprintf("user_%d_emails", accountID)

	// 计算偏移量
	offset := (page - 1) * pageSize
	// 查询总数
	if err := global.PsqlDB.Table(tableName).
		Where("email_type = ? ", emailType).
		Count(&total).Error; err != nil {
		global.Log.Error("获取收件箱邮件总数失败:", err)
		return nil, 0, err
	}

	if total == 0 {
		return []models.EmailDetailsResponse{}, 0, nil
	}

	// 检查页码是否有效
	totalPages := int(total+int64(pageSize)-1) / pageSize
	if page > totalPages {
		return []models.EmailDetailsResponse{}, total, errors.New("Requested page number out of range")
	}
	// 查询邮件列表，只选择指定的字段
	if err := global.PsqlDB.Table(tableName).
		Select("id, email_account_id, email_address,recipient_email, sender_name, sender_email, subject, body_text, body_html, is_read, received_at").
		Where("email_type = ? ", emailType).
		Order("received_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&emails).Error; err != nil {
		global.Log.Error("获取收件箱邮件列表失败:", err)
		return nil, 0, err
	}
	return emails, total, nil
}

// 获取邮件详情
func GetEmailDetails(emailID int, accountID uint) (*models.EmailDetailsResponse, error) {
	var email models.EmailDetailsResponse
	tableName := fmt.Sprintf("user_%d_emails", accountID)
	result := global.PsqlDB.Table(tableName).
		Select("id, recipient_email,sender_name, sender_email, subject, body_text, body_html, received_at").
		Where("id = ?", emailID).
		First(&email)
	if result.Error != nil {
		return nil, result.Error
	}
	//邮件的is_read字段如果是false的，则更新为true
	if !email.IsRead {
		go global.PsqlDB.Table(tableName).Model(&email).Update("is_read", true)
	}
	return &email, nil
}

// 获取邮件详情 - [ 所有字段  ]
func GetEmailDetailFullFileds(emailID int, accountID uint) (*models.EmailDetails, error) {
	var email models.EmailDetails
	tableName := fmt.Sprintf("user_%d_emails", accountID)
	result := global.PsqlDB.Table(tableName).
		Where("id = ?", emailID).
		First(&email)
	if result.Error != nil {
		global.Log.Error("获取邮件详情失败:", result.Error)
		return nil, result.Error
	}
	//邮件的is_read字段如果是false的，则更新为true
	return &email, nil
}

// MoveEmail 移动邮件到指定分类
func MoveEmail(emailID int, accountID uint, sourceType, targetType string) error {
	//拼接表名
	tableName := fmt.Sprintf("user_%d_emails", accountID)
	var email models.EmailDetails
	// 查询指定邮件
	result := global.PsqlDB.Table(tableName).
		Select("email_type").
		Where("id = ?", emailID).
		First(&email)
	//判断错误
	if result.Error != nil {
		return result.Error
	}
	//判断类别是否一致
	if email.EmailType != sourceType {
		return fmt.Errorf("The current type of the email does not match the source type")
	}
	//更新邮件类别
	result = global.PsqlDB.Table(tableName).
		Where("id = ?", emailID).
		Update("email_type", targetType)
	//判断错误
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// MoveEmail 移动邮件到指定分类
func MoveEmailDirect(emailMessageID string, accountID uint, targetType string) error {
	//拼接表名
	tableName := fmt.Sprintf("user_%d_emails", accountID)
	//更新邮件类别
	result := global.PsqlDB.Table(tableName).
		Where("email_message_id = ?", emailMessageID).
		Update("email_type", targetType)
	//判断错误
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// 获取收件箱最新邮件列表
func GetLatestInboxEmailList(accountID uint, lastEmailID int) ([]models.EmailDetailsResponse, int64, error) {
	var emails []models.EmailDetailsResponse
	var total int64
	tableName := fmt.Sprintf("user_%d_emails", accountID)

	// 查询总数
	if err := global.PsqlDB.Table(tableName).
		Where("email_type = ? AND id > ? ", "inbox", lastEmailID).
		Count(&total).Error; err != nil {
		global.Log.Error("获取收件箱邮件总数失败:", err)
		return nil, 0, err
	}

	if total == 0 {
		return []models.EmailDetailsResponse{}, 0, nil
	}
	// 查询邮件列表，只选择指定的字段
	limit := 20
	if total < 20 {
		limit = int(total)
	}
	if err := global.PsqlDB.Table(tableName).
		Select("id, email_account_id, email_address, sender_name, sender_email, subject, body_text, body_html, is_read, received_at").
		Where("email_type = ? AND id > ?", "inbox", lastEmailID).
		Order("received_at DESC").
		Limit(limit).
		Find(&emails).Error; err != nil {
		global.Log.Error("获取收件箱邮件列表失败:", err)
		return nil, 0, err
	}
	return emails, total, nil
}

// IsEmailExistByHash 根据账户ID和邮件hash判断邮件是否存在
func IsEmailExistByHash(accountID uint, emailHash string) (bool, error) {
	// 构建邮件表名
	tableName := fmt.Sprintf("user_%d_emails", accountID)

	// 查询计数
	var count int64
	err := global.PsqlDB.Table(tableName).Where("email_hash = ?", emailHash).Count(&count).Error
	if err != nil {
		global.Log.Error(fmt.Sprintf("查询邮件hash [%s] 对应的邮件失败: ", emailHash), err)
		return false, err
	}

	return count > 0, nil
}

// ------------------------------------------------------------------------------------------------------------------------
// 获取收件箱邮件列表
func GetEmailListByID(accountID uint, emailID int, emailType string) ([]models.EmailDetailsResponse, int64, error) {
	var emails []models.EmailDetailsResponse
	var total int64
	//判断是否是初始加载
	tableName := fmt.Sprintf("user_%d_emails", accountID)
	// 查询总数
	if err := global.PsqlDB.Table(tableName).
		Where("email_type = ? ", emailType).
		Count(&total).Error; err != nil {
		global.Log.Error("获取收件箱邮件总数失败:", err)
		return nil, 0, err
	}

	if total == 0 {
		return []models.EmailDetailsResponse{}, 0, nil
	}

	//此api每次默认返回10条，不足30条则返回全部
	// 查询邮件列表，只选择指定的字段
	query := global.PsqlDB.Table(tableName).
		Select("id, email_account_id, email_address, recipient_email, sender_name, sender_email, cc,subject, body_text, body_html, is_read, attachment_info,received_at").
		Where("email_type = ?", emailType).
		Order("received_at DESC").
		Limit(10)

	if emailID == 0 {
		query = query.Where("id > ?", emailID)
	} else {
		query = query.Where("id < ?", emailID)
	}

	if err := query.Find(&emails).Error; err != nil {
		global.Log.Error("获取邮件列表失败:", err)
		return nil, 0, err
	}

	return emails, total, nil
}

// MarkEmailAsRead 将指定邮件标记为已读
func MarkEmailAsRead(accountID uint, emailID int) error {
	tableName := fmt.Sprintf("user_%d_emails", accountID)

	// 更新邮件的已读状态
	if err := global.PsqlDB.Table(tableName).
		Where("id = ?", emailID).
		Update("is_read", true).Error; err != nil {
		return err
	}

	return nil
}

// MarkEmailAsUnRead 将指定邮件标记为已读
func MarkEmailAsUnRead(accountID uint, emailID int) error {
	tableName := fmt.Sprintf("user_%d_emails", accountID)

	// 更新邮件的已读状态
	if err := global.PsqlDB.Table(tableName).
		Where("id = ?", emailID).
		Update("is_read", false).Error; err != nil {
		return err
	}

	return nil
}

// 根据messageID删除邮件
func DeleteEmailByMessageID(accountID uint, emailMessageID string) error {
	tableName := fmt.Sprintf("user_%d_emails", accountID)
	// 删除邮件
	if err := global.PsqlDB.Table(tableName).
		Where("email_message_id = ?", emailMessageID).
		Delete(models.EmailDetails{}).Error; err != nil {
		return err
	}
	return nil
}

// 根据messageID更新邮件filename
func UpdateEmailFileNameByMessageID(accountID uint, emailMessageID, fileName string) error {
	tableName := fmt.Sprintf("user_%d_emails", accountID)
	// 更新邮件的��件名
	if err := global.PsqlDB.Table(tableName).
		Where("email_message_id = ?", emailMessageID).
		Update("file_name", fileName).Error; err != nil {
		return err
	}
	return nil

}

// 根据messageid更新邮件阅读状态
func UpdateEmailReadStatusByMessageID(accountID uint, emailMessageID string, readStatus bool) error {
	tableName := fmt.Sprintf("user_%d_emails", accountID)
	// 更新邮件状态
	if err := global.PsqlDB.Table(tableName).
		Where("email_message_id = ?", emailMessageID).
		Update("is_read", readStatus).Error; err != nil {
		return err
	}
	return nil
}

// 获取Email file_name
func GetEmailFileName(emailID int, accountID uint) (string, error) {
	var email models.EmailDetails
	tableName := fmt.Sprintf("user_%d_emails", accountID)
	result := global.PsqlDB.Table(tableName).
		Where("id = ?", emailID).
		First(&email)
	if result.Error != nil {
		global.Log.Error("获取邮件详情失败:", result.Error)
		return "", result.Error
	}
	//邮件的is_read字段如果是false的，则更新为true
	return email.FileName, nil
}
