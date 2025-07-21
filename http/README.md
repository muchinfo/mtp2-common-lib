# http 包说明

## 功能简介

- 标准 HTTP 请求工具，支持 context、自定义 client、logger
- 支持多种 body 类型（map、form、string、[]byte）
- 常用签名算法（MD5、HMAC-SHA256、RSA）
- 支持 RSA 加解密

## 主要接口

### HttpCall

```go
func HttpCall(method HttpMethod, urlStr string, body any, header map[string]string, opt *HttpCallOpt) (respBody []byte, status int, respHeader http.Header, err error)
```

- 支持 GET/POST/PUT/DELETE 等
- body 支持 map[string]any、map[string]string、url.Values、string、[]byte
- opt 可自定义 context、http.Client、logger

### 签名相关

- `SignWithMapAndMD5(m map[string]any, key string, ignoreKeys ...string) (sign string, err error)`
- `SignWithMapStringMD5(m map[string]string, key string, ignoreKeys ...string) (sign string, err error)`
- `SignWithMapHMACSHA256(m map[string]any, key string, ignoreKeys ...string) (sign string, err error)`
- `SignWithMapStringHMACSHA256(m map[string]string, key string, ignoreKeys ...string) (sign string, err error)`
- `SignWithRSA(data string, privateKeyPEM string) (sign string, err error)`
- `VerifySignature(data string, signatureBase64 string, publicKeyPEM string) (ok bool, err error)`
- `VerifySignWithMap`/`VerifySignWithMapString`/`VerifySignWithMapHMACSHA256`/`VerifySignWithMapStringHMACSHA256`

### 加解密

- `EncryptWithRSA(data []byte, publicKey string) (encryptedValue string, err error)`
- `DecryptRSAByPublicKey(data string, publicKeyPEM string) (decrypted []byte, err error)`

## 用法示例

详见 example/httpcall_example.go

```go
// GET 请求
resp, status, header, err := http.HttpCall("GET", "https://httpbin.org/get", nil, nil, nil)

// POST JSON
data := map[string]any{"foo": "bar"}
resp, status, _, err := http.HttpCall("POST", "https://httpbin.org/post", data, nil, nil)
```

## 单元测试

详见 example/httpcall_test.go
