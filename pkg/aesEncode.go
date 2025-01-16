package pkg

import (
	"github.com/dromara/dongle"
)

const (
	aesKey = "m87fAensCq4VBUfi" // key 长度必须是 16、24 或 32 字节
	aesIv  = "yG6Z7puxc8qk4Fvp" // iv 长度必须是 16 字节，ECB 模式不需要设置 iv
)

// 加密
func AesEncrypt(data string) string {
	cipher := dongle.NewCipher()
	cipher.SetMode(dongle.CBC)
	cipher.SetPadding(dongle.PKCS7)
	cipher.SetKey(aesKey)
	cipher.SetIV(aesIv)
	// 加密
	return dongle.Encrypt.FromString(data).ByAes(cipher).ToBase64String()
}

// 解密
func AesDecrypt(rawStr string) string {
	cipher := dongle.NewCipher()
	cipher.SetMode(dongle.CBC)
	cipher.SetPadding(dongle.PKCS7)
	cipher.SetKey(aesKey)
	cipher.SetIV(aesIv)
	// 解密
	return dongle.Decrypt.FromBase64String(rawStr).ByAes(cipher).ToString()
}
