package dao

import "fmt"

func getEmailTableName(accountID uint) string {
	return fmt.Sprintf("user_%d_emails", accountID)
}
