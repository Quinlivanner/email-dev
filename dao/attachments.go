package dao

import (
	"email/global"
	"email/models"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AddAttachmentToDBBatchPostgres 批量添加附件，已存在的附件会被跳过
func AddAttachmentToDBBatchPostgres(atts []models.Attachment) error {
	if len(atts) == 0 {
		return nil // 无需操作
	}

	// 使用事务执行数据库操作
	err := global.PsqlDB.Transaction(func(tx *gorm.DB) error {
		// 使用 ON CONFLICT DO NOTHING 语法跳过已存在的记录
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&atts).Error; err != nil {
			global.Log.Error("批量插入附件失败: ", err)
			return err // 返回错误触发事务回滚
		}
		return nil // 返回 nil 提交事务
	})

	if err != nil {
		return fmt.Errorf("批量添加附件事务失败: %w", err)
	}
	global.Log.Infof("批量添加附件成功，处理了 %d 个附件", len(atts))
	return nil
}

// GetAttachmentByHash 通过哈希值查询附件
func GetAttachmentByHash(hash string) (*models.Attachment, error) {
	var att models.Attachment
	err := global.PsqlDB.Where("file_hash = ?", hash).First(&att).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 未找到记录，返回 nil, nil
		}
		global.Log.Error("查询附件失败: ", err)
		return nil, fmt.Errorf("查询附件失败: %w", err)
	}
	return &att, nil
}

// GetAttachmentByCode 通过code查询附件
func GetAttachmentByCode(code string) (*models.Attachment, error) {
	var att models.Attachment
	err := global.PsqlDB.Where("short_url_code = ?", code).First(&att).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 未找到记录，返回 nil, nil
		}
		global.Log.Error("查询附件失败: ", err)
		return nil, fmt.Errorf("search attachment failed: %w", err)
	}
	return &att, nil
}

// GetAttachmentsDataByCodes 通过[]code查询附件
func GetAttachmentsDataByCodes(frAtts []models.FrontendAttachment) ([]models.WebSendMailAttGenerate, []models.Attachment, error) {
	codes := make([]string, len(frAtts))
	for i, att := range frAtts {
		codes[i] = att.Code
	}
	var dbAtts []models.Attachment
	err := global.PsqlDB.Where("short_url_code IN ?", codes).Find(&dbAtts).Error
	if err != nil {
		global.Log.Error("查询附件失败: ", err)
		return nil, nil, fmt.Errorf("search attachments failed: %w", err)
	}

	codeBindKey := make(map[string]string)
	for _, dbAtt := range dbAtts {
		codeBindKey[dbAtt.ShortUrlCode] = dbAtt.S3StoragePath

	}

	// 用前端传来的文件名作为 key
	result := make([]models.WebSendMailAttGenerate, len(frAtts))
	for i, frAtt := range frAtts {
		if key, exists := codeBindKey[frAtt.Code]; exists {
			result[i] = models.WebSendMailAttGenerate{
				FileCode: frAtt.Code,
				FileName: frAtt.Filename,
				FileKey:  key,
			}
		}
	}

	return result, dbAtts, nil
}
