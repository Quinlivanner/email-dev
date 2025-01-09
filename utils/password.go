// utils/password.go 包含与密码处理相关的实用函数。
// 该文件提供了密码哈希、验证和强度检查等功能，用于确保用户密码的安全性。
// 主要功能包括使用 bcrypt 对密码进行哈希处理、验证哈希密码的正确性，
// 以及检查密码是否符合特定的强度要求（如长度、大小写字母、数字和特殊字符的使用）。

package utils

import (
	"crypto/rand"
	"errors"
	"math/big"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	lowerChars   = "abcdefghjkmnpqrstuvwxyz"
	upperChars   = "ABCDEFGHJKMNPQRSTUVWXYZ"
	numberChars  = "123456789"
	specialChars = "!@#"
	passwordLen  = 9
)

// 使用 crypto/rand 生成随机字符
func secureRandomChar(charset string) string {
	newInt := big.NewInt(int64(len(charset)))
	n, err := rand.Int(rand.Reader, newInt)
	if err != nil {
		panic(err)
	}
	return string(charset[n.Int64()])
}

// 打乱字符串
func shuffle(s []string) {
	for i := len(s) - 1; i > 0; i-- {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		j := int(n.Int64())
		s[i], s[j] = s[j], s[i]
	}
}

// 生成安全的随机密码
func GenerateSecurePassword() string {
	password := []string{
		secureRandomChar(lowerChars),
		secureRandomChar(upperChars),
		secureRandomChar(specialChars),
		secureRandomChar(numberChars),
	}

	allChars := lowerChars + upperChars + numberChars + specialChars

	for i := len(password); i < passwordLen; i++ {
		password = append(password, secureRandomChar(allChars))
	}

	shuffle(password)
	return strings.Join(password, "")
}

// HashPassword 使用 bcrypt 对密码进行哈希
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CheckPasswordHash 比较明文密码和哈希密码，验证密码是否正确
func CheckPasswordHash(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// ValidatePassword 验证密码是否符合强度要求
func ValidatePassword(password string) error {
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	if len(password) >= 8 && len(password) <= 17 {
		hasMinLen = true
	}

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasMinLen {
		return errors.New("密码长度至少为 8 个字符")
	}
	if !hasUpper {
		return errors.New("密码需要包含至少一个大写字母")
	}
	if !hasLower {
		return errors.New("密码需要包含至少一个小写字母")
	}
	if !hasNumber {
		return errors.New("密码需要包含至少一个数字")
	}
	if !hasSpecial {
		return errors.New("密码需要包含至少一个特殊字符")
	}

	return nil
}

// ValidateEmailAddress 验证邮箱地址是否符合格式
func ValidateEmailAddress(email string) error {
	// 使用正则表达式验证邮箱格式
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(pattern)

	if !regex.MatchString(email) {
		return errors.New("Invalid email address.")
	}
	// 检查邮箱长度
	if len(email) > 254 {
		return errors.New("Invalid email address.")
	}
	// 检查本地部分长度
	parts := strings.Split(email, "@")
	if len(parts[0]) > 64 {
		return errors.New("Invalid email address.")
	}
	return nil
}
