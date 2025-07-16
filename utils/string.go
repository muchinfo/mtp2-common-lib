package utils

import (
	"fmt"
	"sort"
	"strings"
)

// MapToSortedQueryString 按 map key 排序后输出字符串，类似 key1=value1&key2=value2
func MapToSortedQueryString(m map[string]interface{}, ignored ...string) string {
	// 提取并排序键
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	ignoredMap := make(map[string]bool)
	for _, item := range ignored {
		ignoredMap[item] = true
	}

	// 构建键值对切片
	var pairs []string
	for _, k := range keys {
		if _, ig := ignoredMap[k]; !ig {
			v := m[k]
			// 将值转换为字符串（支持任意类型）
			strV := ""
			switch v := v.(type) {
			case float64:
				// 如果是 float64 类型，保留两位小数
				strV = fmt.Sprintf("%.2f", v)
			case string:
				if v == "" {
					continue
				}
				strV = v
			default:
				strV = fmt.Sprintf("%v", v)
			}
			pairs = append(pairs, fmt.Sprintf("%s=%s", k, strV))
		}
	}

	// 拼接为最终字符串
	return strings.Join(pairs, "&")
}
