package pkg

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/dromara/dongle"
	"github.com/gookit/config/v2"
)

func genNonce() string {
	return dongle.Encrypt.FromString(time.Now().String()).ByMd5().ToHexString()
}

// 生成JSON Web Token
func JwtGenerateToken(username string) (string, error) {
	cipher := dongle.NewCipher()
	cipher.SetMode(dongle.CBC)
	cipher.SetPadding(dongle.PKCS7)

	// 获取密钥并设置
	cipher.SetKey(config.String("JwtSecret.Key", "LdMn3VYDbv7yZQQL3WVZzy46"))
	cipher.SetIV(config.String("JwtSecret.IV", "VLu8JTxrsBwmKwnk"))

	var maps = make(map[string]interface{})
	maps[genNonce()] = genNonce()
	maps["username"] = username
	maps["expires"] = time.Now().Add(time.Hour * 1).Unix()
	maps[genNonce()] = genNonce()

	var jsonString []byte
	jsonString, err := json.Marshal(maps)
	if err != nil {
		return "", err
	}

	rawString := dongle.Encrypt.FromBytes(jsonString).ByAes(cipher).ToHexString()

	return rawString, nil
}

// 验证Token
func JwtCheckToken(token string) (string, error) {
	cipher := dongle.NewCipher()
	cipher.SetMode(dongle.CBC)
	cipher.SetPadding(dongle.PKCS7)

	// 获取密钥并设置
	cipher.SetKey(config.String("JwtSecret.Key", "rkD7Nt6y6p4bnUTru61QeHm7"))
	cipher.SetIV(config.String("JwtSecret.IV", "DE150HXMHnQ5hLUg"))

	jsonString := dongle.Decrypt.FromHexString(token).ByAes(cipher).ToBytes()

	var maps map[string]interface{}
	err := json.Unmarshal(jsonString, &maps)
	if err != nil {
		return "", err
	}

	// 判断超时
	now := time.Now().Unix()
	expires := int64(maps["expires"].(float64))
	if now > expires {
		return "", errors.New("token expired")
	}

	return maps["username"].(string), nil
}

// 刷新token  有效期内刷新
func JwtRefreshToken(token string) (string, error) {
	username, err := JwtCheckToken(token)
	if err != nil {
		return "", err
	}
	return JwtGenerateToken(username)
}
