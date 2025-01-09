package shortlink

import (
	"crypto/sha256"
	"email/global"
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// CreateShortLink 创建短链接Code
func CreateShortLinkCode(s3StorgePath string) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// 获取当前时间戳（纳秒级）
	timestamp := time.Now().UnixNano()
	// 对 S3 路径进行哈希处理
	hash := sha256.Sum256([]byte(s3StorgePath))
	s3Hash := base64.RawURLEncoding.EncodeToString(hash[:])[:8]
	// 生成随机字符串
	randomStr := make([]byte, 4)
	for i := range randomStr {
		randomStr[i] = charset[rand.Intn(len(charset))]
	}
	// 组合所有元素
	code := fmt.Sprintf("%d%s%s", timestamp, s3Hash, string(randomStr))
	// 确保长度为17位
	if len(code) > global.Config.API.ShortUrlCodeLength {
		code = code[:global.Config.API.ShortUrlCodeLength]
	} else {
		code = code + strings.Repeat("0", global.Config.API.ShortUrlCodeLength-len(code))
	}
	return code
}
