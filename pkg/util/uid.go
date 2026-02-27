package util

import (
	"fmt"
	"github.com/google/uuid"
	"math/rand/v2"
	"strings"
	"time"
)

// GenerateShortID 生成自定义的16位ID
func GenerateShortID() string {
	// 生成UUID并取前16个字符
	fullUUID := uuid.New().String()
	shortID := strings.ReplaceAll(fullUUID, "-", "")[:16]
	return shortID
}

// GetRandom 生成随机字符串
// length: 随机数长度，date: 可选的时间参数，如果提供会追加时间戳
func GetRandom(length *int, date *time.Time) string {
	// 生成随机数并取上限值
	ceil := rand.Float64() * 100000000000000
	ceilStr := fmt.Sprintf("%.0f", ceil)
	// 确定截取长度
	strLength := 4
	if length != nil {
		strLength = *length
	}
	// 截取指定长度
	var substring string
	if len(ceilStr) > strLength {
		substring = ceilStr[:strLength]
	} else {
		substring = ceilStr
	}
	// 如果提供了时间参数，追加时间戳
	if date != nil {
		substring += fmt.Sprintf("%d", date.UnixMilli())
	}
	return substring
}
