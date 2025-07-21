package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"go.uber.org/zap"
)

// HttpCallOpt 用于自定义 http.Client、context、logger
type HttpCallOpt struct {
	Client *http.Client
	Ctx    context.Context
	Logger *zap.Logger
}

// HttpCall 支持 context、自定义 client，返回 body、状态码、header
// body 支持 map[string]any (json), url.Values (form), string, []byte
// opt 可为 nil，默认用全局 logger
func HttpCall(method HttpMethod, urlStr string, body any, header map[string]string, opt *HttpCallOpt) (respBody []byte, status int, respHeader http.Header, err error) {
	var reqBody io.Reader
	var contentType string
	var logger *zap.Logger
	if opt != nil && opt.Logger != nil {
		logger = opt.Logger
	}
	if body != nil {
		switch v := body.(type) {
		case map[string]any:
			b, e := json.Marshal(v)
			if e != nil {
				if logger != nil {
					logger.Error("序列化 JSON 请求体失败", zap.Error(e))
				}
				err = e
				return
			}
			reqBody = bytes.NewReader(b)
			contentType = "application/json"
		case url.Values:
			reqBody = strings.NewReader(v.Encode())
			contentType = "application/x-www-form-urlencoded"
		case map[string]string:
			var buf bytes.Buffer
			w := multipart.NewWriter(&buf)
			for key, val := range v {
				_ = w.WriteField(key, val)
			}
			w.Close()
			reqBody = &buf
			contentType = w.FormDataContentType()
		case string:
			reqBody = strings.NewReader(v)
		case []byte:
			reqBody = bytes.NewReader(v)
		default:
			b, e := json.Marshal(v)
			if e != nil {
				if logger != nil {
					logger.Error("序列化 JSON 请求体失败", zap.Error(e))
				}
				err = e
				return
			}
			reqBody = bytes.NewReader(b)
			contentType = "application/json"
		}
	}
	ctx := context.Background()
	if opt != nil && opt.Ctx != nil {
		ctx = opt.Ctx
	}
	req, e := http.NewRequestWithContext(ctx, string(method), urlStr, reqBody)
	if e != nil {
		if logger != nil {
			logger.Error("创建 HTTP 请求失败", zap.Error(e))
		}
		err = e
		return
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	if contentType != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", contentType)
	}
	client := http.DefaultClient
	if opt != nil && opt.Client != nil {
		client = opt.Client
	}
	if logger != nil {
		logger.Info("HTTP 请求", zap.String("方法", string(method)), zap.String("URL", urlStr), zap.Any("请求头", header))
	}
	response, e := client.Do(req)
	if e != nil {
		if logger != nil {
			logger.Error("HTTP 执行请求失败", zap.Error(e))
		}
		err = e
		return
	}
	defer response.Body.Close()
	respBody, e = io.ReadAll(response.Body)
	if e != nil {
		if logger != nil {
			logger.Error("读取响应体失败", zap.Error(e))
		}
		err = e
		return
	}
	if logger != nil {
		logger.Info("HTTP 响应", zap.Int("状态码", response.StatusCode), zap.String("响应体", string(respBody)))
	}
	status = response.StatusCode
	respHeader = response.Header
	return
}
