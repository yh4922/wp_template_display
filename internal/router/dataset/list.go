package datasetRouter

import (
	consts "wp_template_display/internal/consts"
	ctxDataset "wp_template_display/internal/controller/dataset"
	mid "wp_template_display/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func init() {
	routerItem := consts.ControllerRouter{
		Path:   "/api/v1/dataset/list",
		Method: "GET",
		Func:   ctxDataset.DatasetList,
	}

	routerItem.Middle = []func(c *fiber.Ctx) error{
		// 认证中间件，根据需求可以去除或者添加
		mid.JwtLoginMiddleware,
		mid.UserAuthMiddleware([]string{"admin"}),
	}

	consts.ControllerList = append(consts.ControllerList, routerItem)
}
