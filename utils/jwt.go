/*
utils/jwt.go 包含与 JWT（JSON Web Token）相关的实用函数。
这个文件提供了生成、验证和处理 JWT 的功能，用于用户身份验证和授权。
主要功能包括生成新的 JWT、验证现有的 JWT 的有效性，以及其他与 JWT 操作相关的辅助函数。
*/

package utils

import (
	"crypto/sha256"
	"email/global"
	"email/models"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// GenerateToken 生成JWT token
func GenerateJwtToken(account *models.EmailAccount, t time.Time) (string, error) {
	// 创建JWT声明
	claims := models.CustomJwtClaims{
		EmailAddress: account.EmailAddress,
		UserName:     account.UserName,
		UserID:       account.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: t.Unix(),
		},
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用密钥签名token
	tokenString, err := token.SignedString([]byte(global.Config.Jwt.SecretKey))
	if err != nil {
		return "", fmt.Errorf("Generate token failed: %v", err)
	}

	return tokenString, nil
}

// ParseJwtToken 解析JWT token
func ParseJwtToken(tokenString string) (*models.CustomJwtClaims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &models.CustomJwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		// 返回用于验证的密钥
		return []byte(global.Config.Jwt.SecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("Parse token failed: %v", err)
	}

	// 验证token并提取声明
	if claims, ok := token.Claims.(*models.CustomJwtClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("Invalid token.")
}

// ValidateToken 验证JWT token
func ValidateToken(tokenString string) (*jwt.Token, error) {
	fmt.Print("hellp")
	// 实现token验证逻辑
	return nil, nil
}

// ... 其他JWT相关函数

// HashJWTToken 对JWT令牌进行哈希处理
func HashJWTToken(token string) string {
	// 使用SHA-256算法对令牌进行哈希
	hasher := sha256.New()
	hasher.Write([]byte(token))
	hashedBytes := hasher.Sum(nil)
	// 将哈希结果转换为十六进制字符串
	hashedToken := hex.EncodeToString(hashedBytes)
	return hashedToken
}
