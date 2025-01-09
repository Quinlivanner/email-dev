package utils

import (
	"fmt"
	"time"

	"github.com/araddon/dateparse"
)

// 解析时间为time.Time
func ParseTime(dateStr string) (time.Time, error) {
	// 解析时间字符串为 time.Time
	resTime := time.Time{}
	parsedDate, err := dateparse.ParseAny(dateStr)
	if err != nil {
		fmt.Printf("无法解析日期字符串: %v\n", err)
		// 解析失败时返回当前时间
		resTime = time.Now()
	}
	resTime = parsedDate
	// 获取时区
	localLocation := time.Now().Location()
	// 转换为当地时间
	localTime := resTime.In(localLocation)
	return localTime, nil
}
