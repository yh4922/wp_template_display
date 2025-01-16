package userRouter

import (
	consts "wp_template_display/internal/consts"
	ctxUser "wp_template_display/internal/controller/user"

	"github.com/gofiber/fiber/v2"
)

func init() {
	routerItem := consts.ControllerRouter{
		Path:   "/api/v1/user/login-pwd-key",
		Method: "GET",
		Func:   ctxUser.UserLoginPwdKey,
	}

	routerItem.Middle = []func(c *fiber.Ctx) error{}

	consts.ControllerList = append(consts.ControllerList, routerItem)
}
