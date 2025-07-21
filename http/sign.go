package http

import (
	"crypto"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"mtp2-common-lib/utils"
)

// SignWithRSA 使用PKCS8格式的私钥对数据进行SHA256WithRSA签名
// data: 待签名字符串
// privateKeyPEM: base64/PEM内容（不含头尾）
func SignWithRSA(data string, privateKeyPEM string) (sign string, err error) {
	privateKeyPEM = fmt.Sprintf("-----BEGIN PRIVATE KEY-----\n%s\n-----END PRIVATE KEY-----", privateKeyPEM)

	// 将私钥字符串转换为PEM块
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		err = errors.New("解析私钥 PEM 块失败")
		return
	}

	// 解析PKCS8格式的私钥
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		err = fmt.Errorf("解析私钥失败: %v", err)
		return
	}

	// 将私钥转换为rsa.PrivateKey类型
	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		err = fmt.Errorf("不是有效的 RSA 私钥")
		return
	}

	// 计算SHA256哈希值
	hashed := sha256.Sum256([]byte(data))

	// 使用私钥对哈希值进行签名
	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, crypto.SHA256, hashed[:])
	if err != nil {
		err = fmt.Errorf("签名数据失败: %v", err)
		return
	}

	// 将签名转换为Base64编码字符串
	sign = base64.StdEncoding.EncodeToString(signature)
	return
}

// SignWithMapAndMD5 针对 map[string]any 排序后生成 MD5 签名
func SignWithMapAndMD5(m map[string]any, key string, ignoreKeys ...string) (sign string, err error) {
	s := utils.MapToSortedQueryString(m, ignoreKeys...)
	s += "&key=" + key
	md := md5.New()
	md.Write([]byte(s))
	sign = hex.EncodeToString(md.Sum(nil))
	return
}

// SignWithMapStringMD5 针对 map[string]string 排序后生成 MD5 签名
func SignWithMapStringMD5(m map[string]string, key string, ignoreKeys ...string) (sign string, err error) {
	s := utils.MapStringToSortedQueryString(m, ignoreKeys...)
	s += "&key=" + key
	md := md5.New()
	md.Write([]byte(s))
	sign = hex.EncodeToString(md.Sum(nil))
	return
}

// SignWithMapHMACSHA256 针对 map[string]any 排序后生成 HMAC-SHA256 签名
func SignWithMapHMACSHA256(m map[string]any, key string, ignoreKeys ...string) (sign string, err error) {
	s := utils.MapToSortedQueryString(m, ignoreKeys...)
	s += "&key=" + key
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(s))
	sign = hex.EncodeToString(h.Sum(nil))
	return
}

// SignWithMapStringHMACSHA256 针对 map[string]string 排序后生成 HMAC-SHA256 签名
func SignWithMapStringHMACSHA256(m map[string]string, key string, ignoreKeys ...string) (sign string, err error) {
	s := utils.MapStringToSortedQueryString(m, ignoreKeys...)
	s += "&key=" + key
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(s))
	sign = hex.EncodeToString(h.Sum(nil))
	return
}

// VerifySignWithRSA 使用公钥验证数据的签名
// data: 原始字符串
// signatureBase64: base64签名
// publicKeyPEM: base64/PEM内容（不含头尾）
func VerifySignWithRSA(data string, signatureBase64 string, publicKeyPEM string) (ok bool, err error) {
	publicKeyPEM = fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----", publicKeyPEM)

	// 将公钥字符串转换为PEM块
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		err = fmt.Errorf("解析公钥 PEM 块失败")
		return
	}

	// 解析PKCS8格式的公钥
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		err = fmt.Errorf("解析公钥失败: %v", err)
		return
	}

	// 将公钥转换为rsa.PublicKey类型
	rsaPublicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		err = fmt.Errorf("不是有效的 RSA 公钥")
		return
	}

	// 计算SHA256哈希值
	hashed := sha256.Sum256([]byte(data))

	// 解码Base64签名
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		err = fmt.Errorf("Base64 签名解码失败: %v", err)
		return
	}

	// 使用公钥验证签名
	err = rsa.VerifyPKCS1v15(rsaPublicKey, crypto.SHA256, hashed[:], signature)
	if err != nil {
		err = fmt.Errorf("签名验证失败: %v", err)
		return
	}

	ok = true
	return
}

// VerifySignWithMap 验证 map[string]any 的 MD5 签名
func VerifySignWithMap(m map[string]any, key string, signature string, ignoreKeys ...string) (ok bool, err error) {
	expectedSignature, err := SignWithMapAndMD5(m, key, ignoreKeys...)
	if err != nil {
		return
	}
	ok = (expectedSignature == signature)
	return
}

// VerifySignWithMapString 验证 map[string]string 的 MD5 签名
func VerifySignWithMapString(m map[string]string, key string, signature string, ignoreKeys ...string) (ok bool, err error) {
	expectedSignature, err := SignWithMapStringMD5(m, key, ignoreKeys...)
	if err != nil {
		return
	}
	ok = (expectedSignature == signature)
	return
}

// VerifySignWithMapHMACSHA256 验证 map[string]any 的 HMAC-SHA256 签名
func VerifySignWithMapHMACSHA256(m map[string]any, key string, signature string, ignoreKeys ...string) (ok bool, err error) {
	expectedSignature, err := SignWithMapHMACSHA256(m, key, ignoreKeys...)
	if err != nil {
		return
	}
	ok = (expectedSignature == signature)
	return
}

// VerifySignWithMapStringHMACSHA256 验证 map[string]string 的 HMAC-SHA256 签名
func VerifySignWithMapStringHMACSHA256(m map[string]string, key string, signature string, ignoreKeys ...string) (ok bool, err error) {
	expectedSignature, err := SignWithMapStringHMACSHA256(m, key, ignoreKeys...)
	if err != nil {
		return
	}
	ok = (expectedSignature == signature)
	return
}
