package userRouter

import (
	consts "wp_template_display/internal/consts"
	ctxUser "wp_template_display/internal/controller/user"
	mid "wp_template_display/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func init() {
	routerItem := consts.ControllerRouter{
		Path:   "/api/v1/user/info",
		Method: "GET",
		Func:   ctxUser.UserInfo,
	}

	routerItem.Middle = []func(c *fiber.Ctx) error{
		// 认证中间件，根据需求可以去除或者添加
		mid.JwtLoginMiddleware,
	}

	consts.ControllerList = append(consts.ControllerList, routerItem)
}
