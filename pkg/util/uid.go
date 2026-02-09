package util

import (
	"github.com/google/uuid"
	"strings"
)

// GenerateShortID 生成自定义的16位ID
func GenerateShortID() string {
	// 生成UUID并取前16个字符
	fullUUID := uuid.New().String()
	shortID := strings.ReplaceAll(fullUUID, "-", "")[:16]
	return shortID
}
