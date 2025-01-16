package setup

import (
	"time"
	consts "wp_template_display/internal/consts"
	mid "wp_template_display/internal/middleware"
	"wp_template_display/resource"

	"github.com/gofiber/fiber/v2"

	// 导出路由
	_ "wp_template_display/internal/router/dataset"
	_ "wp_template_display/internal/router/dataset/version"
	_ "wp_template_display/internal/router/dataset/version/category"
	_ "wp_template_display/internal/router/dataset/version/data"
	_ "wp_template_display/internal/router/dataset/version/markup"
	_ "wp_template_display/internal/router/user"
)

func SetupRouter(app *fiber.App) {
	// 初始化文件服务
	resource.InitFileServer(app)

	// // 写入一些区域变量方便后续请求使用
	app.Use(func(c *fiber.Ctx) error {
		// 当前请求的触发时间
		c.Locals("TriggerTime", time.Now())
		return c.Next()
	})

	// 遍历路由列表
	for _, item := range consts.ControllerList {
		switch item.Method {
		case "GET":
			app.Get(item.Path, append(item.Middle, item.Func, mid.RequestLogMiddleware)...)
		case "POST":
			app.Post(item.Path, append(item.Middle, item.Func, mid.RequestLogMiddleware)...)
		case "PUT":
			app.Put(item.Path, append(item.Middle, item.Func, mid.RequestLogMiddleware)...)
		case "DELETE":
			app.Delete(item.Path, append(item.Middle, item.Func, mid.RequestLogMiddleware)...)
		}
	}
}
