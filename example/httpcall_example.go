package main

import (
	"fmt"
	mhttp "mtp2-common-lib/http"
	"net/http"
	"net/url"
	"time"
)

func RunHttpExample() {
	resp, status, _, err := mhttp.HttpCall("GET", "https://httpbin.org/get", nil, nil, nil)
	fmt.Println("GET:", status, err, string(resp))

	// 示例2：POST JSON
	data := map[string]any{"foo": "bar", "num": 123}
	resp, status, _, err = mhttp.HttpCall("POST", "https://httpbin.org/post", data, map[string]string{"X-Test": "1"}, nil)
	fmt.Println("POST JSON:", status, err, string(resp))

	// 示例3：POST 表单
	form := url.Values{"a": {"1"}, "b": {"2"}}
	resp, status, _, err = mhttp.HttpCall("POST", "https://httpbin.org/post", form, nil, nil)
	fmt.Println("POST form:", status, err, string(resp))
	fmt.Println("POST form:", status, err, string(resp))

	// 示例4：自定义超时
	opt := &mhttp.HttpCallOpt{Client: &http.Client{Timeout: 2 * time.Second}}
	resp, status, _, err = mhttp.HttpCall("GET", "https://httpbin.org/delay/1", nil, nil, opt)
	fmt.Println("GET with timeout:", status, err, string(resp))
}
