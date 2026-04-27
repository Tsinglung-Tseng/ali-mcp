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

### 淘宝（Taobao / Tmall）

| 工具 | 类型 | 说明 |
|------|------|------|
| `taobao_check_login_status` | read | 检查淘宝登录态 |
| `taobao_get_login_qrcode` | read | 获取扫码登录二维码（base64 PNG） |
| `taobao_search` | read | 搜索商品，返回前 N 条（标题/价格/店铺/发货地/URL） |
| `taobao_get_item_detail` | read | 商品详情（标题/价格/销量/店铺/SKU/主图） |
| `taobao_delete_cookies` | destructive | 删除 cookie 重置登录 |

### 闲鱼（Xianyu / Goofish）

| 工具 | 类型 | 说明 |
|------|------|------|
| `xianyu_check_login_status` | read | 检查 h5.m.goofish.com 登录态 |
| `xianyu_search` | read | 搜索商品（游客态可访问） |
| `xianyu_get_item_detail` | read | 商品详情（标题/价格/原价/描述/卖家/图片/想要数） |
| `xianyu_get_user_profile` | read | 卖家主页（昵称/信用/在售已售数 + 最近 20 个商品） |
| `xianyu_refresh` | write | 擦亮商品（卖家把商品刷到 feed 顶部，需登录） |
| `xianyu_delist` | destructive | 下架商品（需登录，自己的商品） |
| `xianyu_publish` | write | 发布商品（**stub 未实现**，等实机走查后填回） |
| `xianyu_delete_cookies` | destructive | 删除 cookie |

> 闲鱼无独立扫码：阿里 SSO 会从淘宝回流，先登录淘宝再访问闲鱼 h5 即可。
>
> ⚠️ 所有 DOM 选择器为初版假设，**需实机校正**才能可靠工作（参见每个 `*.go` 文件顶部的 `TODO(selectors)` 注释）。
>
> 📋 `xianyu_publish` 当前是 stub —— 实机研究指南见 [`docs/debug/20260427-xianyu-publish-research.md`](docs/debug/20260427-xianyu-publish-research.md)。

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
