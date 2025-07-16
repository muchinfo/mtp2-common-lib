package utils

import "time"

// GetTime 获取目标时间
//
// timeStr 时间字串
// parsedTime 获取目标时间
//
// err 错误
func GetTime(timeStr string) (parsedTime time.Time, err error) {
	format := "2006-01-02 15:04:05"
	parsedTime, err = time.Parse(format, timeStr)

	return
}

// GetTimeString 获取目标时间字串
//
// srcTime 目标时间
//
// timeStr 时间字串
func GetTimeString(srcTime time.Time) (timeStr string) {
	format := "2006-01-02 15:04:05"
	timeStr = srcTime.Format(format)

	return
}
