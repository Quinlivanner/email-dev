package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// 计算hash值
func ComputeContentHash(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}
