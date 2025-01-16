package pkg

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

const (
	bits = 1024
)

// 生成RSA密钥对
func RsaGenerate() (string, string, error) {
	// 生成RSA密钥对
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", err
	}

	// 将私钥编码为PKCS1格式
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// 将公钥编码为PKIX格式
	publicKey := &privateKey.PublicKey
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", "", err
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(publicKeyPEM), string(privateKeyPEM), nil
}

// 公钥加密
func RsaEncrypt(publicKeyPEM string, plaintext string) (string, error) {
	// 解码公钥PEM
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return "", fmt.Errorf("无法解码公钥PEM")
	}

	// 解析公钥
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("解析公钥时发生错误: %v", err)
	}

	// 类型断言为RSA公钥
	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("无法获取RSA公钥")
	}

	// 加密明文
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, []byte(plaintext))
	if err != nil {
		return "", fmt.Errorf("加密时发生错误: %v", err)
	}

	// 返回密文的Base64编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// 私钥解密
func RsaDecrypt(privateKeyPEM string, ciphertext string) (string, error) {
	// 解码私钥PEM
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", fmt.Errorf("无法解码私钥PEM")
	}

	// 解析私钥
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("解析私钥时发生错误: %v", err)
	}

	// 解码密文的Base64编码
	encryptedBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("解码密文时发生错误: %v", err)
	}

	// 解密密文
	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedBytes)
	if err != nil {
		return "", fmt.Errorf("解密时发生错误: %v", err)
	}

	return string(plaintext), nil
}
