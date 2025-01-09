/*
dao/account.go 文件包含与账户管理相关的数据访问对象（DAO）函数。
该文件提供了账户的增删改查等操作，用于与数据库进行交互。
主要功能包括：
- 添加新账户
- 验证账户信息
- 更新账户详情
- 删除账户
- 查询账户列表
*/

package dao

import (
	"email/global"
	"email/models"
	"email/utils"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// 新增账户
// AddAccount 原子性地添加新账户到数据库并创建对应的邮件表
func AddAccount(domainName, userName, accountPrefix, password string) error {
	global.Log.Info(fmt.Sprintf("开始添加账户 [ %s ]", userName))
	// 事务操作
	err := global.PsqlDB.Transaction(func(tx *gorm.DB) error {
		var d models.Domain

		// 查询域名是否存在
		global.Log.Info(fmt.Sprintf("查询域名 [ %s ]", domainName))
		if err := tx.Where("domain_name = ?", domainName).First(&d).Error; err != nil {
			global.Log.Error(fmt.Sprintf("查询 [ %s ] 域名失败: ", domainName), err)
			return err // 返回错误会触发事务回滚
		}

		//查询账户是否存在
		account, err := IsAccountExist(accountPrefix+"@"+domainName, domainName)
		if err != nil && err != gorm.ErrRecordNotFound {
			global.Log.Error(fmt.Sprintf("查询账户 [ %s@%s ] 失败: ", accountPrefix, domainName), err)
			return err // 返回错误会触发事务回滚
		}
		if account != nil {
			global.Log.Info(fmt.Sprintf("账户 [ %s@%s ] 已存在，未进行插入", accountPrefix, domainName))
			return gorm.ErrDuplicatedKey
		}

		// 哈希密码
		hashPassword, err := utils.HashPassword(password)
		if err != nil {
			global.Log.Error("密码哈希失败: ", err)
			return err // 返回错误会触发事务回滚
		}

		// 创建 EmailAccount 实例
		a := models.EmailAccount{
			DomainID:     d.ID,
			DomainName:   domainName,
			EmailAddress: accountPrefix + "@" + domainName,
			PasswordHash: hashPassword,
			UserName:     userName,
		}

		// 插入新账户记录
		global.Log.Info(fmt.Sprintf("插入账户 [ %s ]", a.EmailAddress))
		if err := tx.Create(&a).Error; err != nil {
			global.Log.Error(fmt.Sprintf("插入 [ %s ] 账户失败: ", userName), err)
			return err // 返回错误会触发事务回滚
		}

		global.Log.Info(fmt.Sprintf("插入 [ %s ] 账户成功", userName))
		// 创建账户对应的邮件表
		tableName := fmt.Sprintf("user_%d_emails", a.ID)
		sql := fmt.Sprintf(`
				CREATE TABLE IF NOT EXISTS %s (
					id               SERIAL PRIMARY KEY,
					email_message_id TEXT NOT NULL UNIQUE,  -- UNIQUE 约束会自动创建唯一索引
					file_name        VARCHAR(255) NOT NULL,
					domain_id        INT          NOT NULL,
					domain_name      VARCHAR(255) NOT NULL,
					email_hash       VARCHAR(255) NOT NULL,
					email_account_id INT          NOT NULL,
					email_address    VARCHAR(255) NOT NULL,
					recipient_email  TEXT		 NOT NULL,
					s3_key          VARCHAR(255) NOT NULL,
					sender_name      TEXT,
					sender_email     TEXT NOT NULL,
					cc              TEXT,
					bcc             TEXT,
					reply_email_id  INT,
					subject         TEXT,
					body_text       TEXT,
					body_html       TEXT,
					is_read         BOOLEAN      DEFAULT FALSE NOT NULL,
					email_type      VARCHAR(255) NOT NULL,
					received_at     TIMESTAMP    DEFAULT CURRENT_TIMESTAMP NOT NULL,
					attachment_info JSONB        DEFAULT '{}'::JSONB NOT NULL,
					created_at      TIMESTAMP    DEFAULT CURRENT_TIMESTAMP NOT NULL,
					last_update_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP NOT NULL,
				
					-- 外键约束
					FOREIGN KEY (domain_id)
						REFERENCES domains (id)
						ON DELETE CASCADE
						ON UPDATE CASCADE,
				
					FOREIGN KEY (email_account_id)
						REFERENCES email_accounts (id)
						ON DELETE CASCADE
						ON UPDATE CASCADE
				)
		`, tableName)

		if err := tx.Exec(sql).Error; err != nil {
			global.Log.Error(fmt.Sprintf("为账户 [ID: %d] 创建邮件表失败: ", a.ID), err)
			return err // 返回错误会触发事务回滚
		}

		global.Log.Info(fmt.Sprintf("为账户 [ID: %d] 成功创建邮件表 [%s]", a.ID, tableName))
		return nil // 返回 nil 提交事务
	})

	if err != nil {
		// 事务回滚时的错误处理
		return err
	}

	return nil
}

// 查询账户状态
// IsAccountExist 根据邮箱地址和域名查询账户是否存在，如果存在返回models.EmailAccount
func IsAccountExist(emailAddress, domain string) (*models.EmailAccount, error) {
	// 记录查询的日志
	// 查询数据库，检查邮箱地址和域名是否存
	var account models.EmailAccount
	//err := global.PsqlDB.Where("email_address = ? AND domain_name = ?", emailAddress, domain).First(&account).Error
	err := global.PsqlDB.Where(&models.EmailAccount{EmailAddress: emailAddress, DomainName: domain}).First(&account).Error
	// 如果查询过程中发生错误
	if err != nil {
		return nil, err
	}
	// 记录查询结果日志
	return &account, nil
}

// IsEmailExistByS3Key 根据S3Key查询邮件是否存在
func IsEmailExistByS3Key(accountID uint, s3Key string) (bool, error) {
	// 构建邮件表名
	tableName := fmt.Sprintf("user_%d_emails", accountID)
	// 查询计数
	var count int64
	err := global.PsqlDB.Table(tableName).Where("s3_key = ?", s3Key).Count(&count).Error
	if err != nil {
		global.Log.Error(fmt.Sprintf("查询S3Key [%s] 对应的邮件失败: ", s3Key), err)
		return false, err
	}
	return count > 0, nil
}

// 验证账户和密码是否匹配
func ValidateAccount(email, password string) (*models.EmailAccount, error) {
	var account models.EmailAccount
	result := global.PsqlDB.Where("email_address = ?", email).First(&account)
	if result.Error != nil {
		return nil, result.Error
	}
	// 验证密码
	s := utils.CheckPasswordHash(password, account.PasswordHash)
	if s {
		return &account, nil
	}
	return nil, errors.New("incorrect password")
}

// UpdateAccountJwtToken 更新账户的JWT令牌哈希
func UpdateAccountJwtToken(accountID uint, emailAddress string, jwtToken string) error {
	// 对JWT令牌进行哈希处理
	hashedToken := utils.HashJWTToken(jwtToken)
	// 更新数据库中的JWT令牌哈希
	result := global.PsqlDB.Model(&models.EmailAccount{}).
		Where("id = ? AND email_address = ?", accountID, emailAddress).
		Update("jwt_token_hash", hashedToken)
	if result.Error != nil {
		global.Log.Error("更新JWT令牌哈希失败:", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("未找到指定的账户")
	}
	global.Log.Info(fmt.Sprintf("成功更新账户 [ID: %d] 的JWT令牌哈希", accountID))
	return nil
}

// GetAccountsByDomainID 根据域名ID获取所有邮箱账户
func GetAccountsByDomainID(domainID uint) ([]models.SafeEmailAccount, error) {
	var accounts []models.EmailAccount
	result := global.PsqlDB.Where("domain_id = ?", domainID).Find(&accounts)
	if result.Error != nil {
		global.Log.Error(fmt.Sprintf("获取域名ID [%d] 的邮箱账户失败:", domainID), result.Error)
		return nil, result.Error
	}
	safeAccounts := make([]models.SafeEmailAccount, len(accounts))
	for i, account := range accounts {
		safeAccounts[i] = account.ToSafeEmailAccount()
	}
	global.Log.Info(fmt.Sprintf("成功获取域名ID [%d] 的邮箱账户，共 %d 个", domainID, len(accounts)))
	return safeAccounts, nil
}

// GetAccountByID 根据账户ID获取账户信息
func GetAccountByID(accountID uint) (*models.EmailAccount, error) {
	var account models.EmailAccount
	result := global.PsqlDB.Where("id = ?", accountID).First(&account)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			global.Log.Error(fmt.Sprintf("获取ID为 [%d] 的账户信息失败:", accountID), result.Error)
			return nil, result.Error
		}
	}
	return &account, nil
}

// 根据emailAddress获取accountid
func GetAccountIDByEmailAddress(emailAddress string) (uint, error) {
	var account models.EmailAccount
	result := global.PsqlDB.Where("email_address = ?", emailAddress).First(&account)
	if result.Error != nil {
		global.Log.Error(fmt.Sprintf("查询邮箱地址为 [%s] 的账户时发生错误:", emailAddress), result.Error)
		return 0, result.Error
	}
	return account.ID, nil
}

// 修改账户密码,先用email和密码去匹配，如果匹配成功就修改密码，否则返回错误
func UpdateAccountPassword(email, password, newPassword string) error {
	account, err := ValidateAccount(email, password)
	if err != nil {
		return err
	}
	if account == nil {
		return errors.New("account not found")
	}
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return global.PsqlDB.Model(&models.EmailAccount{}).Where("id = ?", account.ID).Update("password_hash", hashedPassword).Error
}
