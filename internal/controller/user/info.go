package ctxUser

import (
	ctx "wp_template_display/internal/controller"
	m "wp_template_display/internal/models"

	"github.com/gofiber/fiber/v2"
)

func UserInfo(c *fiber.Ctx) error {
	// 查询用户基本信息
	user := c.Locals("User").(*m.SysUser)

	// 查询用户角色
	info, roles, err := m.GetUserBaseInfoById(user.Id)
	if err != nil {
		return ctx.CtxError(c, 5001, "用户不存在", nil)
	}

	info.Password = ""

	rolesRes := []UserLoginResRole{}
	for _, v := range roles {
		rolesRes = append(rolesRes, UserLoginResRole{
			Id:    v.Id,
			Alias: v.Alias,
			Label: v.Label,
		})
	}

	return ctx.CtxSuccess(c, fiber.Map{
		"info":  info,
		"roles": rolesRes,
	})
}
