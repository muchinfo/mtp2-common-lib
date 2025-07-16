package utils

import (
	"crypto/md5"
	"encoding/hex"
	"os"
)

// GetMD5 获取byte对应MD5
func GetMD5(s []byte) string {
	m := md5.New()
	m.Write([]byte(s))
	return hex.EncodeToString(m.Sum(nil))
}

// 获取文件byte
func ReadFileMd5(sfile string) (string, error) {
	ssconfig, err := os.ReadFile(sfile)
	if err != nil {
		return "", err
	}
	return GetMD5(ssconfig), nil
}
