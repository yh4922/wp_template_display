package resource

import (
	"embed"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

//go:embed client
var clientFs embed.FS

func InitFileServer(app *fiber.App) {
	// 静态资源文件
	app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(clientFs),
		PathPrefix: "client",
		Browse:     true,
	}))
}
