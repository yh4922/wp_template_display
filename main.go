package main

import (
	"fmt"
	"log"
	"wp_template_display/cmd"

	"github.com/gookit/config/v2"
	"github.com/joho/godotenv"
)

func init() {
	// 加载环境变量
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// 启动主程序
	cmd.Start()
}

func main() {
	// 启动服务
	log.Fatal(cmd.App.Listen(
		fmt.Sprintf(
			"%s:%d",
			config.String("Server.Host", "0.0.0.0"),
			config.Int("Server.Port", 3000),
		),
	))
}
