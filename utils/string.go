package utils

import (
	"fmt"
	"sort"
	"strings"
)

// MapToSortedQueryString 按 map[string]any key 排序后输出字符串，类似 key1=value1&key2=value2
func MapToSortedQueryString(m map[string]any, ignored ...string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	ignoredMap := make(map[string]bool)
	for _, item := range ignored {
		ignoredMap[item] = true
	}

	var pairs []string
	for _, k := range keys {
		if _, ig := ignoredMap[k]; !ig {
			v := m[k]
			strV := ""
			switch v := v.(type) {
			case float64:
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
	return strings.Join(pairs, "&")
}

// MapStringToSortedQueryString 按 map[string]string key 排序后输出字符串，类似 key1=value1&key2=value2
func MapStringToSortedQueryString(m map[string]string, ignored ...string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	ignoredMap := make(map[string]bool)
	for _, item := range ignored {
		ignoredMap[item] = true
	}

	var pairs []string
	for _, k := range keys {
		if _, ig := ignoredMap[k]; !ig {
			v := m[k]
			if v == "" {
				continue
			}
			pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
		}
	}
	return strings.Join(pairs, "&")
}
