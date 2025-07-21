package http

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
)

// EncryptWithRSA 支持 PKCS1/PKCS8 公钥格式的加密
// publicKey: base64/PEM内容（不含头尾）
func EncryptWithRSA(data []byte, publicKey string) (encryptedValue string, err error) {
	publicKeyPEM := fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----", publicKey)
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		err = fmt.Errorf("解析公钥 PEM 块失败")
		return
	}
	var pub *rsa.PublicKey
	// 先尝试 PKCS8
	ifc, e := x509.ParsePKIXPublicKey(block.Bytes)
	if e == nil {
		var ok bool
		pub, ok = ifc.(*rsa.PublicKey)
		if !ok {
			err = fmt.Errorf("不是有效的 RSA 公钥")
			return
		}
	} else {
		// 再尝试 PKCS1
		pkcs1, e2 := x509.ParsePKCS1PublicKey(block.Bytes)
		if e2 != nil {
			err = fmt.Errorf("解析公钥失败: %v, %v", e, e2)
			return
		}
		pub = pkcs1
	}
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, pub, data)
	if err != nil {
		err = fmt.Errorf("加密数据失败: %v", err)
		return
	}
	encryptedValue = base64.StdEncoding.EncodeToString(encrypted)
	return
}

// DecryptRSAByPublicKey 使用公钥解密数据
func DecryptRSAByPublicKey(data string, publicKeyPEM string) (decrypted []byte, err error) {
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

	// 解码Base64数据
	encryptedValue, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		err = fmt.Errorf("Base64 数据解码失败: %v", err)
		return
	}

	c := new(big.Int).SetBytes(encryptedValue)
	m := new(big.Int).Exp(c, big.NewInt(int64(rsaPublicKey.E)), rsaPublicKey.N)
	decrypted = m.Bytes()

	// PKCS#1 padding，需要手动处理
	if len(decrypted) < 11 {
		err = fmt.Errorf("解密数据过短")
		return
	}

	return
}
