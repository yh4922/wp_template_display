package datasetVersionDataRouter

import (
	consts "wp_template_display/internal/consts"
	ctxDatasetVersionData "wp_template_display/internal/controller/dataset/version/data"
	mid "wp_template_display/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func init() {
	routerItem := consts.ControllerRouter{
		Path:   "/api/v1/dataset/version/data/batch-delete",
		Method: "DELETE",
		Func:   ctxDatasetVersionData.DatasetVersionDataBatchDelete,
	}

	routerItem.Middle = []func(c *fiber.Ctx) error{
		// 认证中间件，根据需求可以去除或者添加
		mid.JwtLoginMiddleware,
		mid.UserAuthMiddleware([]string{"admin"}),
	}

	consts.ControllerList = append(consts.ControllerList, routerItem)
}
