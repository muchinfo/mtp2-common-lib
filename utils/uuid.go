package utils

import (
	"github.com/google/uuid"
)

// GetUUID 获取 UUID
func GetUUID() string {
	return uuid.New().String()
}
