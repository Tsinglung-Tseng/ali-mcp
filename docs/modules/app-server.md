# app-server（入口层）

## 职责

聚合两个 service（TaobaoService / XianyuService），同时暴露 HTTP 和 MCP 两种协议；负责启动、优雅关闭。

## 架构

```
main.go
  └── NewAppServer(taobao, xianyu)
         ├── mcpServer (registerTools via mcp_server.go)
         └── httpServer (routes.go → gin.Engine)
```

## 关键实现

- `main.go` — CLI flags：`-headless`（默认 false，电商风控对 headless 严）、`-bin`、`-port`（默认 `:18070`，避 xhs 的 `:18060`）
- `app_server.go` — `AppServer` 结构体 + `Start(port)` 阻塞启动 + SIGINT/SIGTERM 优雅关闭
- `routes.go` — Gin 路由分组：`/api/v1/taobao/*`、`/api/v1/xianyu/*`；MCP 走 `Streamable HTTP`
- `middleware.go` — CORS + panic-recovery 中间件
- `handlers_api.go` — HTTP 层薄包装，调用 service 后 `respondSuccess / respondError`

## 接口

HTTP：
- `GET /health`
- `GET /api/v1/taobao/login/status`
- `GET /api/v1/taobao/login/qrcode`
- `DELETE /api/v1/taobao/login/cookies`
- `GET /api/v1/xianyu/login/status`
- `DELETE /api/v1/xianyu/login/cookies`
- `ANY /mcp`、`ANY /mcp/*path`

MCP 工具（见 `mcp-server.md`）。

## 依赖

- `gin-gonic/gin` HTTP 框架
- `modelcontextprotocol/go-sdk/mcp` MCP 协议
- `sirupsen/logrus` 日志

## 已知问题

- 无 graceful reload，修改配置需重启
- 无 per-request auth，本地调试用
