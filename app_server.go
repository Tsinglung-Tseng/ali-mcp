package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// AppServer 聚合所有 service、MCP server、HTTP server。
type AppServer struct {
	taobao *TaobaoService
	xianyu *XianyuService

	mcpServer  *mcp.Server
	router     *gin.Engine
	httpServer *http.Server
}

// NewAppServer 构造 AppServer。
func NewAppServer(taobao *TaobaoService, xianyu *XianyuService) *AppServer {
	app := &AppServer{taobao: taobao, xianyu: xianyu}
	app.mcpServer = InitMCPServer(app)
	return app
}

// Start 启动服务器，阻塞直到收到 SIGINT / SIGTERM。
func (s *AppServer) Start(port string) error {
	s.router = setupRoutes(s)
	s.httpServer = &http.Server{Addr: port, Handler: s.router}

	go func() {
		logrus.Infof("HTTP server listening on %s", port)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Errorf("server error: %v", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		logrus.Warnf("graceful shutdown timed out: %v", err)
	} else {
		logrus.Info("server stopped")
	}
	return nil
}
