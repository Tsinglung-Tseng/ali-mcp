# ali-mcp

阿里系 MCP 服务器，目前支持**淘宝**和**闲鱼**。架构仿照 [`xiaohongshu-mcp`](https://github.com/xpzouying/xiaohongshu-mcp)，通过 `go-rod` 无头浏览器执行登录 / 搜索 / 详情等操作，并以 MCP 协议暴露给 LLM 客户端。

## 快速开始

```bash
# 构建
go build -o ali-mcp .

# 启动（首次建议用有头模式，电商风控对 headless 严）
./ali-mcp -headless=false -port=:18070
```

服务起来后：

- HTTP API：`http://localhost:18070/api/v1/...`
- MCP 端点：`http://localhost:18070/mcp`
- 健康检查：`curl http://localhost:18070/health`

## MCP 工具

| 工具 | 说明 |
|------|------|
| `taobao_check_login_status` | 检查淘宝登录态 |
| `taobao_get_login_qrcode` | 获取淘宝扫码登录二维码 |
| `taobao_delete_cookies` | 删除淘宝 cookie（重置登录） |
| `xianyu_check_login_status` | 检查闲鱼 h5 登录态 |
| `xianyu_delete_cookies` | 删除闲鱼 cookie |

> 闲鱼无独立扫码：阿里 SSO 会从淘宝回流，先登录淘宝再访问闲鱼 h5 即可。

## 接入 Claude Code

在 Claude Code 的 MCP 配置中增加：

```json
{
  "mcpServers": {
    "ali-mcp": {
      "type": "http",
      "url": "http://localhost:18070/mcp"
    }
  }
}
```

或者用 `claude mcp add` 命令行：

```bash
claude mcp add --transport http ali-mcp http://localhost:18070/mcp
```

## 环境变量

| 变量 | 用途 | 默认 |
|------|------|------|
| `ALI_COOKIES_DIR` | cookie 存放目录 | `./.cookies` |
| `ALI_PROXY` | HTTP 代理（形如 `http://user:pass@host:port`） | 无 |
| `ROD_BROWSER_BIN` | 浏览器二进制路径 | 无（rod 自动下载） |

## 项目文档

- 开发计划：[`docs/project/dev-plan.md`](docs/project/dev-plan.md)
- 变更日志：[`docs/project/CHANGELOG.md`](docs/project/CHANGELOG.md)
- 模块文档：[`docs/modules/`](docs/modules/)
- 架构决策：[`docs/decisions/`](docs/decisions/)

## 参考

- [`xpzouying/xiaohongshu-mcp`](https://github.com/xpzouying/xiaohongshu-mcp) — 架构原型
- [`xpzouying/headless_browser`](https://github.com/xpzouying/headless_browser) — 浏览器封装
- [`modelcontextprotocol/go-sdk`](https://github.com/modelcontextprotocol/go-sdk) — MCP 协议实现
