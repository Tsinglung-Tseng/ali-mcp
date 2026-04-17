package main

import (
	"flag"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/Tsinglung-Tseng/ali.mcp/configs"
)

func main() {
	var (
		headless bool
		binPath  string
		port     string
	)
	flag.BoolVar(&headless, "headless", false, "是否无头模式；电商风控对 headless 更严，默认关闭")
	flag.StringVar(&binPath, "bin", "", "浏览器二进制路径（可通过 ROD_BROWSER_BIN 环境变量设置）")
	flag.StringVar(&port, "port", ":18070", "HTTP 监听端口")
	flag.Parse()

	if binPath == "" {
		binPath = os.Getenv("ROD_BROWSER_BIN")
	}

	configs.InitHeadless(headless)
	configs.SetBinPath(binPath)

	app := NewAppServer(NewTaobaoService(), NewXianyuService())
	if err := app.Start(port); err != nil {
		logrus.Fatalf("failed to run server: %v", err)
	}
}
