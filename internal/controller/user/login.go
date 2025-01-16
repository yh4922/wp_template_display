package ctxUser

import (
	"time"
	ctx "wp_template_display/internal/controller"
	m "wp_template_display/internal/models"
	"wp_template_display/pkg"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type UserLoginReq struct {
	MainUsername string `json:"main_username"`
	Username     string `json:"username" validate:"required"`
	Password     string `json:"password" validate:"required"`
	PasswordKey  string `json:"password_key" validate:"required"`
}

type UserLoginResRole struct {
	Id    uint32 `json:"id"`
	Alias string `json:"alias"`
	Label string `json:"label"`
}

type UserLoginRes struct {
	Token string             `json:"token"`
	Info  m.SysUser          `json:"info"`
	Roles []UserLoginResRole `json:"roles"`
}

func UserLogin(c *fiber.Ctx) error {
	// 模型名称
	c.Locals("Action", "/api/v1/user/login")
	c.Locals("Content", "用户登录")

	// 解析请求体
	req := new(UserLoginReq)
	if err := c.BodyParser(req); err != nil {
		return ctx.CtxError(c, 5001, "参数错误", err.Error())
	}

	// 验证字段
	err := ctx.Validate.Struct(req)
	if err != nil {
		return ctx.CtxError(c, 5002, "参数错误", err.Error())
	}

	// {
	// 	// 开发环境下覆盖参数
	// 	rsaKey, _, err := ctx.GetRsaKey()
	// 	if err != nil {
	// 		return ctx.CtxError(c, 5001, "参数错误", err.Error())
	// 	}
	// 	req.PasswordKey = rsaKey
	// 	ras := ctx.RsaMap[req.PasswordKey]

	// 	pwd, err := pkg.RsaEncrypt(ras.PublicKey, "123456")
	// 	if err != nil {
	// 		return ctx.CtxError(c, 5001, "参数错误", err.Error())
	// 	}
	// 	req.Password = pwd
	// }

	// 获取密钥
	rasItem := ctx.RsaMap[req.PasswordKey]
	if rasItem.Expire == 0 {
		return ctx.CtxError(c, 5003, "密钥不存在", nil)
	}

	now := time.Now().UnixMilli()
	if now > rasItem.Expire {
		return ctx.CtxError(c, 5004, "密钥已过期", nil)
	}

	// 解密密码
	password, err := pkg.RsaDecrypt(rasItem.PrivateKey, req.Password)
	if err != nil {
		return ctx.CtxError(c, 5005, "解密失败", err.Error())
	}

	// 查询用户
	user, err := m.FindUserByUsername(req.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ctx.CtxError(c, 5006, "用户不存在", nil)
		}
		return ctx.CtxError(c, 5007, "查询失败", err.Error())
	}

	// 查询主用户
	var mainUser *m.SysUser
	if req.MainUsername != "" {
		mainUser, err = m.FindUserByUsername(req.MainUsername)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return ctx.CtxError(c, 5008, "主用户不存在", nil)
			}
			return ctx.CtxError(c, 5009, "查询失败", err.Error())
		}

		if user.ParentUid != mainUser.Id {
			return ctx.CtxError(c, 5010, "用户不属于主用户", nil)
		}
	}

	// 验证密码
	if !pkg.PwdCompare(password, user.Password) {
		return ctx.CtxError(c, 5011, "密码错误", nil)
	}

	token, err := pkg.JwtGenerateToken(user.Username)
	if err != nil {
		return ctx.CtxError(c, 5012, "生成token失败", err.Error())
	}

	info, roles, err := m.GetUserBaseInfoById(user.Id)
	if err != nil {
		return ctx.CtxError(c, 5013, "查询用户基础信息失败", err.Error())
	}

	// 移除 user 的密码
	res := UserLoginRes{
		Token: token,
		Info:  *info,
	}

	res.Roles = []UserLoginResRole{}
	for _, v := range roles {
		res.Roles = append(res.Roles, UserLoginResRole{
			Id:    v.Id,
			Alias: v.Alias,
			Label: v.Label,
		})
	}

	res.Info.Password = ""
	return ctx.CtxSuccess(c, res)
}
