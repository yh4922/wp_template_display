package datasetVersionRouter

import (
	consts "wp_template_display/internal/consts"
	ctxDatasetVersion "wp_template_display/internal/controller/dataset/version"
	mid "wp_template_display/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func init() {
	routerItem := consts.ControllerRouter{
		Path:   "/api/v1/dataset/version/get/:id",
		Method: "GET",
		Func:   ctxDatasetVersion.DatasetVersionGet,
	}

	routerItem.Middle = []func(c *fiber.Ctx) error{
		// 认证中间件，根据需求可以去除或者添加
		mid.JwtLoginMiddleware,
		mid.UserAuthMiddleware([]string{"admin"}),
	}

	consts.ControllerList = append(consts.ControllerList, routerItem)
}
