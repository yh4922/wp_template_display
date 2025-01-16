package cmd

import (
	m "wp_template_display/internal/models"
	"wp_template_display/internal/setup"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gookit/config/v2"
)

var (
	App *fiber.App
)

// 初始化
func Start() {
	// 加载配置
	setup.SetupConfig()

	// 加载数据库
	setup.SetupDatabase()

	// 初始化 Fiber 应用
	App = fiber.New(fiber.Config{
		AppName:      config.String("AppName", "WordpressTemplateDisplay"),
		ErrorHandler: errorHandler,
		BodyLimit:    55 * 1024 * 1024, // 55MB 超过50M
	})

	// 配置跨域
	App.Use(cors.New())

	// 配置路由
	setup.SetupRouter(App)

	App.Get("/test", func(c *fiber.Ctx) error {
		ver, err := m.GetDatasetVerById(3)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		taskPath, _, err := m.ExportDatasetVerForTrain(ver)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		return c.JSON(taskPath)
	})
}

// 错误处理
func errorHandler(ctx *fiber.Ctx, err error) error {
	// 状态码 默认500
	code := fiber.StatusInternalServerError

	// 获取自定义状态码
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// 统一错误处理
	return ctx.Status(code).JSON(fiber.Map{
		"code": code,
		"msg":  err.Error(),
	})
}
