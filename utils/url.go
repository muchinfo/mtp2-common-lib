package utils

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

func MapToURLValues(m map[string]any) url.Values {
	values := url.Values{}
	for k, v := range m {
		if v == nil {
			continue
		}
		values.Set(k, fmt.Sprintf("%v", v))
	}
	return values
}

// 将结构体转换为 url.Values
func StructToURLValues(s interface{}) (values url.Values, err error) {
	values = url.Values{}

	v := reflect.ValueOf(s)
	// 检查是否为结构体
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		err = fmt.Errorf("expected a struct, but got %s", v.Kind())
		return
	}

	// 遍历结构体的字段
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("form")
		tagParts := strings.Split(tag, ",")
		formKey := tagParts[0]

		// 跳过没有 form 标签的字段
		if formKey == "" {
			continue
		}

		// 处理omitempty标志
		omitEmpty := len(tagParts) > 1 && tagParts[1] == "omitempty"
		fieldValue := v.Field(i)

		// 检查字段是否为空值
		if omitEmpty && isEmptyValue(fieldValue) {
			continue
		}

		// 将字段值转换为字符串并加入 url.Values
		values.Set(formKey, fmt.Sprintf("%v", fieldValue.Interface()))
	}

	return
}

// 检查字段是否为零值
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Slice, reflect.Map, reflect.Array, reflect.Chan, reflect.Ptr:
		return v.IsNil()
	default:
		return false
	}
}
