package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// setupRoutes 注册 HTTP + MCP 路由。
func setupRoutes(appServer *AppServer) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(errorHandlingMiddleware())
	router.Use(corsMiddleware())

	router.GET("/health", healthHandler)

	// MCP Streamable HTTP
	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server { return appServer.mcpServer },
		&mcp.StreamableHTTPOptions{JSONResponse: true},
	)
	router.Any("/mcp", gin.WrapH(mcpHandler))
	router.Any("/mcp/*path", gin.WrapH(mcpHandler))

	api := router.Group("/api/v1")
	{
		tb := api.Group("/taobao")
		{
			tb.GET("/login/status", appServer.taobaoLoginStatusHandler)
			tb.GET("/login/qrcode", appServer.taobaoLoginQrcodeHandler)
			tb.DELETE("/login/cookies", appServer.taobaoDeleteCookiesHandler)
			tb.GET("/search", appServer.taobaoSearchHandler)
			tb.GET("/item", appServer.taobaoItemDetailHandler)
		}
		xy := api.Group("/xianyu")
		{
			xy.GET("/login/status", appServer.xianyuLoginStatusHandler)
			xy.DELETE("/login/cookies", appServer.xianyuDeleteCookiesHandler)
			xy.GET("/search", appServer.xianyuSearchHandler)
			xy.GET("/item", appServer.xianyuItemDetailHandler)
			xy.GET("/user", appServer.xianyuUserProfileHandler)
			xy.POST("/refresh", appServer.xianyuRefreshHandler)
			xy.POST("/delist", appServer.xianyuDelistHandler)
		}
	}

	return router
}
