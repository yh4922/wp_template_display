package DatasetVersionCategoryRouter

import (
	consts "wp_template_display/internal/consts"
	ctxDatasetVersionCategory "wp_template_display/internal/controller/dataset/version/category"
	mid "wp_template_display/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func init() {
	routerItem := consts.ControllerRouter{
		Path:   "/api/v1/dataset/version/category/delete/:id",
		Method: "DELETE",
		Func:   ctxDatasetVersionCategory.DatasetVersionCategoryDelete,
	}

	routerItem.Middle = []func(c *fiber.Ctx) error{
		// 认证中间件，根据需求可以去除或者添加
		mid.JwtLoginMiddleware,
		mid.UserAuthMiddleware([]string{"admin"}),
	}

	consts.ControllerList = append(consts.ControllerList, routerItem)
}
