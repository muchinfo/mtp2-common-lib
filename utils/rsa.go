package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
)

// 签名算法
// 算法		RSA
// Key格式	PKCS8
// 签名算法	SHA256WithRSA
// 密钥长度	2048

// SignWithRSA 使用PKCS8格式的私钥对数据进行SHA256WithRSA签名
func SignWithRSA(data string, privateKeyPEM string) (sign string, err error) {
	privateKeyPEM = fmt.Sprintf("-----BEGIN PRIVATE KEY-----\n%s\n-----END PRIVATE KEY-----", privateKeyPEM)

	// 将私钥字符串转换为PEM块
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		err = errors.New("failed to decode PEM block containing the private key")
		return
	}

	// 解析PKCS8格式的私钥
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		err = fmt.Errorf("failed to parse private key: %v", err)
		return
	}

	// 将私钥转换为rsa.PrivateKey类型
	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		err = fmt.Errorf("not an RSA private key")
		return
	}

	// 计算SHA256哈希值
	hashed := sha256.Sum256([]byte(data))

	// 使用私钥对哈希值进行签名
	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, crypto.SHA256, hashed[:])
	if err != nil {
		err = fmt.Errorf("failed to sign data: %v", err)
		return
	}

	// 将签名转换为Base64编码字符串
	sign = base64.StdEncoding.EncodeToString(signature)
	return
}

// VerifySignature 使用公钥验证数据的签名
func VerifySignature(data string, signatureBase64 string, publicKeyPEM string) (ok bool, err error) {
	publicKeyPEM = fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----", publicKeyPEM)

	// 将公钥字符串转换为PEM块
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		err = fmt.Errorf("failed to decode PEM block containing the public key")
		return
	}

	// 解析PKCS8格式的公钥
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		err = fmt.Errorf("failed to parse public key: %v", err)
		return
	}

	// 将公钥转换为rsa.PublicKey类型
	rsaPublicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		err = fmt.Errorf("not an RSA public key")
		return
	}

	// 计算SHA256哈希值
	hashed := sha256.Sum256([]byte(data))

	// 解码Base64签名
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		err = fmt.Errorf("failed to decode base64 signature: %v", err)
		return
	}

	// 使用公钥验证签名
	err = rsa.VerifyPKCS1v15(rsaPublicKey, crypto.SHA256, hashed[:], signature)
	if err != nil {
		err = fmt.Errorf("signature verification failed: %v", err)
		return
	}

	ok = true
	return
}

// EncryptWithRSA 使用PKCS1格式的公钥对数据进行加密
func EncryptWithRSA(data []byte, publicKey string) (encryptedValue string, err error) {
	publicKeyPEM := fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----", publicKey)

	// 将公钥字符串转换为PEM块
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		err = fmt.Errorf("failed to decode PEM block containing the public key")
		return
	}

	// 解析PKCS8格式的公钥
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		err = fmt.Errorf("failed to parse public key: %v", err)
		return
	}

	// 将公钥转换为rsa.PublicKey类型
	rsaPublicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		err = fmt.Errorf("not an RSA public key")
		return
	}

	// 使用PublicKey对数据进行加密
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, data)
	if err != nil {
		err = fmt.Errorf("failed to encrypt data: %v", err)
		return
	}

	// 将加密数据转换为Base64编码字符串
	encryptedValue = base64.StdEncoding.EncodeToString(encrypted)
	return
}

// DecryptRSAByPublicKey 使用公钥解密数据
func DecryptRSAByPublicKey(data string, publicKeyPEM string) (decrypted []byte, err error) {
	publicKeyPEM = fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----", publicKeyPEM)

	// 将公钥字符串转换为PEM块
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		err = fmt.Errorf("failed to decode PEM block containing the public key")
		return
	}

	// 解析PKCS8格式的公钥
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		err = fmt.Errorf("failed to parse public key: %v", err)
		return
	}

	// 将公钥转换为rsa.PublicKey类型
	rsaPublicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		err = fmt.Errorf("not an RSA public key")
		return
	}

	// 解码Base64数据
	encryptedValue, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		err = fmt.Errorf("failed to decode base64 data: %v", err)
		return
	}

	c := new(big.Int).SetBytes(encryptedValue)
	m := new(big.Int).Exp(c, big.NewInt(int64(rsaPublicKey.E)), rsaPublicKey.N)
	decrypted = m.Bytes()

	// PKCS#1 padding，需要手动处理
	if len(decrypted) < 11 {
		err = fmt.Errorf("decrypted data is too short")
		return
	}

	return
}
