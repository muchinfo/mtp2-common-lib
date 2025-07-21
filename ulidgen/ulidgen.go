package ulidgen

import (
	crypto_rand "crypto/rand"
	math_rand "math/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

// GenerateULID 生成一个 ULID 字符串，长度为 26 位（标准 ULID 长度）。
func GenerateULID() (string, error) {
	entropy := ulid.Monotonic(crypto_rand.Reader, 0)
	id, err := ulid.New(ulid.Timestamp(time.Now()), entropy)
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

// GenerateShortULID 生成一个自定义长度的 ULID 字符串，最短 18 位，最长 26 位。
// 若 length < 18，则自动调整为 18；若 length > 26，则自动调整为 26。
func GenerateShortULID(length int) (string, error) {
	if length < 18 {
		length = 18
	}
	if length > 26 {
		length = 26
	}
	full, err := GenerateULID()
	if err != nil {
		return "", err
	}
	return full[:length], nil
}

// GenerateULIDWithPrefix 生成带前缀的 ULID，自动截断保证总长度不超过 maxLen。
// 若 prefix+ulid 超过 maxLen，则 ulid 部分会被截断。
func GenerateULIDWithPrefix(prefix string, maxLen int) (string, error) {
	ulidStr, err := GenerateULID()
	if err != nil {
		return "", err
	}
	remain := maxLen - len(prefix)
	if remain < 6 {
		remain = 6 // 保证 ULID 至少有 6 位
	}
	if remain > 26 {
		remain = 26
	}
	return prefix + ulidStr[:remain], nil
}

// GenerateULIDWithRandSource 允许自定义随机源，便于测试。
func GenerateULIDWithRandSource(t time.Time, r *math_rand.Rand) string {
	entropy := ulid.Monotonic(r, 0)
	id := ulid.MustNew(ulid.Timestamp(t), entropy)
	return id.String()
}
