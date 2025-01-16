package datasetVersionDataRouter

import (
	consts "wp_template_display/internal/consts"
	ctxDatasetVersionData "wp_template_display/internal/controller/dataset/version/data"
	mid "wp_template_display/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func init() {
	routerItem := consts.ControllerRouter{
		Path:   "/api/v1/dataset/version/data/preview/:id",
		Method: "GET",
		Func:   ctxDatasetVersionData.DatasetVersionDataPreview,
	}

	routerItem.Middle = []func(c *fiber.Ctx) error{
		// 认证中间件，根据需求可以去除或者添加
		mid.JwtLoginMiddlewareByUrl,
		mid.UserAuthMiddleware([]string{"admin", "mark"}),
	}

	consts.ControllerList = append(consts.ControllerList, routerItem)
}
