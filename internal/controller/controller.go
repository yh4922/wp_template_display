package ctx

import (
	"strconv"
	"strings"
	"time"
	"wp_template_display/pkg"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type RsaKeyItem struct {
	PublicKey  string `json:"public_key"`  // 公钥
	PrivateKey string `json:"private_key"` // 私钥
	Value      string `json:"value"`       // 加密值
	Expire     int64  `json:"expire"`      // 过期时间戳
}

var RsaMap = make(map[string]RsaKeyItem)

var Validate = validator.New(validator.WithRequiredStructEnabled())

// 获取RSA密钥
func GetRsaKey() (string, string, error) {
	publicKey, privateKey, err := pkg.RsaGenerate()
	if err != nil {
		return "", "", err
	}

	rasItem := RsaKeyItem{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Value:      "",
		Expire:     time.Now().Add(time.Minute * 10).UnixMilli(),
	}

	rsaKey := uuid.New().String()
	RsaMap[rsaKey] = rasItem

	// 创建定时器 删除过期的密钥
	time.AfterFunc(time.Minute*1, func() {
		delete(RsaMap, rsaKey)
	})

	list := strings.Split(rasItem.PublicKey, "\n")
	publicKey = strings.Join(list[1:len(list)-2], "\n")

	return rsaKey, publicKey, nil
}

// 成功数据
func CtxSuccess(c *fiber.Ctx, data interface{}, msg ...string) error {
	c.Status(200)
	c.Locals("ResStatus", "0")
	message := "操作成功"
	if len(msg) > 0 {
		message = msg[0]
	}
	c.JSON(fiber.Map{
		"code": 0,
		"msg":  message,
		"data": data,
	})

	c.Next()
	return nil
}

// 失败数据
func CtxError(c *fiber.Ctx, code int, msg string, data interface{}) error {
	c.Status(200)
	c.Locals("ResStatus", strconv.Itoa(code))

	c.JSON(fiber.Map{
		"code": code,
		"msg":  msg,
		"data": data,
	})

	c.Next()
	return nil
}
