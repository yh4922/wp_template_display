package setup

import (
	"flag"
	"os"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/json5"
)

func SetupConfig() {
	config.AddDriver(json5.Driver)

	// 获取命令行参数 -c
	cfg := flag.String("c", "./configs/config.json5", "config file")
	flag.Parse()

	// 加载配置文件
	err := config.LoadFiles(*cfg)
	if err != nil {
		panic(err)
	}

	// 运行时目录
	runtimePath := config.String("RuntimePath", "runtime")
	// 创建运行时目录(如不存在)
	if err := os.MkdirAll(runtimePath, 0755); err != nil {
		panic(err)
	}

	// 清空临时目录 ???
	// tempPath := filepath.Join(runtimePath, "temp")
	// _ = os.RemoveAll(tempPath)
}
