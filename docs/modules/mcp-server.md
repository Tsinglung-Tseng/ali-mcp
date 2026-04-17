# mcp-server（MCP 协议层）

## 职责

将 service 能力注册为 MCP tool，供 LLM 客户端（Claude Code、Cursor 等）调用。工具命名用 `<platform>_<action>` 双前缀，避免淘宝/闲鱼混淆。

## 架构

```
mcp_server.go
  ├── InitMCPServer(appServer) → *mcp.Server
  ├── registerTools() — 向 server 注册所有工具
  ├── withPanicRecovery[T]() — 工具 panic 时返回 IsError 结果，不中断 server
  └── convertToMCPResult() — 内部 MCPToolResult → SDK CallToolResult

mcp_handlers.go
  └── 每个工具一个 handle* 方法，调用对应 service 并拼 MCPToolResult
```

## 已注册工具（Phase 1：5 个）

| 工具名 | 类型 | 说明 |
|--------|------|------|
| `taobao_check_login_status` | ReadOnly | 检查淘宝登录态 |
| `taobao_get_login_qrcode` | ReadOnly | 获取淘宝扫码二维码（base64 PNG） |
| `taobao_delete_cookies` | Destructive | 删除淘宝 cookie |
| `xianyu_check_login_status` | ReadOnly | 检查闲鱼 h5 登录态 |
| `xianyu_delete_cookies` | Destructive | 删除闲鱼 cookie |

> 闲鱼没有独立扫码登录：阿里 SSO 从淘宝回流，用户登录淘宝后访问闲鱼 h5 会自动带上闲鱼 cookie。

## 关键实现

- **panic recovery**：`withPanicRecovery[T]` 泛型包装，仿 xiaohongshu-mcp，所有 handler 强制套一层
- **image content**：`qrcodeResult()` 返回 `text + image` 双 content，Claude Code 可直接在对话里渲染二维码
- **base64 处理**：`convertToMCPResult` 对 image 类型 base64-decode 到 `[]byte`，MCP SDK 要原始字节

## 接口

对外：`*mcp.Server`（挂载到 `/mcp` endpoint）。

## 依赖

- `modelcontextprotocol/go-sdk v1.5.0`

## 已知问题

- 工具 description 太简，LLM 调用时理解可能不到位，后续扩充
- 尚未实现 search / detail 等只读业务工具（Phase 2）
