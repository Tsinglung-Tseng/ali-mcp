# ali-mcp 开发计划

## 📌 项目定位

**项目描述**：阿里系 MCP 服务器 —— 仿照 `xiaohongshu-mcp` 的架构，先覆盖闲鱼（Xianyu / Goofish）和淘宝（Taobao），通过 go-rod 无头浏览器执行登录、搜索、详情抓取等操作，并以 MCP 协议暴露给 LLM 客户端。

**参考项目**：[`xpzouying/xiaohongshu-mcp`](https://github.com/xpzouying/xiaohongshu-mcp)（架构、分层、登录方案、MCP 工具注册模式全部参照）

**技术栈**：
- Go 1.24
- HTTP: `gin-gonic/gin`
- MCP: `modelcontextprotocol/go-sdk`
- 浏览器: `go-rod/rod` + `xpzouying/headless_browser`
- 日志: `sirupsen/logrus`
- 其它: `avast/retry-go/v4`、`pkg/errors`

## 🗂 模块总览

| 模块 | 核心职责 | 状态 |
|------|---------|------|
| `main` / `app_server` / `routes` | 入口、HTTP + MCP 双协议服务器、优雅关闭 | 已完成（骨架） |
| `configs` | headless / 浏览器路径 / 平台常量等全局配置 | 已完成 |
| `browser` | go-rod Browser 工厂，按平台自动挂 cookie | 已完成 |
| `cookies` | cookie 读写，按平台分文件（`taobao.json` / `xianyu.json`） | 已完成 |
| `taobao` | 登录（骨架）、搜索宝贝、获取宝贝详情 | 进行中（仅登录骨架，选择器待实机校验） |
| `xianyu` | 登录态检查（骨架）、搜索商品、商品详情、用户主页 | 进行中（仅登录态，选择器待实机校验） |
| `mcp_server` / `handlers` | MCP 工具注册、HTTP handler、panic recovery | 已完成（5 个工具，Phase 1 范围） |
| `pkg/` | 通用工具（图片下载、字符串/时间等 util） | 未开始（Phase 2 按需加） |
| `cmd/login/` | 独立的扫码登录 CLI（可选） | 未开始（Phase 2 评估） |

## 🎯 里程碑

### Phase 1 — 骨架 + 登录（当前阶段）
- [x] `go mod init github.com/Tsinglung-Tseng/ali.mcp`
- [x] 基础目录结构（仿 xiaohongshu-mcp）
- [x] configs + browser + cookies 通用层
- [x] 淘宝扫码登录骨架（代码完成，DOM 选择器待实机校验）
- [x] 闲鱼登录态检查骨架（代码完成，DOM 选择器待实机校验）
- [x] MCP 工具：5 个（taobao×3 + xianyu×2）
- [x] HTTP server 启动 + MCP server 注册
- [x] `go build ./...` 通过
- [ ] **实机跑通**：`go run . -headless=false` + 扫码登录 + 选择器校正

### Phase 2 — 只读工具（代码完成 2026-04-27，选择器待实机校正）
- [x] `taobao_search`（入口 `s.taobao.com` 桌面站；React hash class 前缀匹配 + simba 广告过滤 + 去重）
- [x] `taobao_get_item_detail`
- [x] `xianyu_search`（入口 `h5.m.goofish.com`，游客态可访问）
- [x] `xianyu_get_item_detail`
- [x] `xianyu_get_user_profile`（卖家主页摘要 + 最近 20 个商品）

### Phase 3 — 写操作
- [x] **闲鱼擦亮（refresh）**：单步操作（点列表里的擦亮按钮），代码完成
- [x] **闲鱼下架（delist）**：两步（三点菜单 → 下架），代码完成
- [ ] **闲鱼发布宝贝**：5 步表单 + 图片上传，代码 stub，研究指南见 `docs/debug/20260427-xianyu-publish-research.md`
- [ ] 闲鱼私信（未开始）
- [ ] 淘宝加购物车 / 下单（慎重，未开始）

## ⏳ 待完成

- [x] 确认 Go module 路径 → `github.com/Tsinglung-Tseng/ali.mcp`
- [x] Phase 1 骨架编译通过（`go build ./...` 无错误）
- [x] 第一次 git 初始化 + 推到 `github.com/Tsinglung-Tseng/ali-mcp`（2026-04-27 开源，MIT license）
- [x] Phase 2/3 全部代码 + stub 编译通过（`go build ./...` + `go vet ./...` 无错误）
- [ ] **实机首跑**：`go run . -headless=false`，扫淘宝二维码 → 校验所有工具的 DOM 选择器
- [ ] 淘宝登录后 cookie 自动持久化 → 访问闲鱼 h5 自动 SSO → 闲鱼 `check_login_status` 返回 true
- [ ] 在 Claude Code 中配置并调用 ali-mcp 工具（mcp config 指向 `http://localhost:18070/mcp`）
- [ ] 实机走查闲鱼发布流程，回填 `xianyu/publish.go` 的选择器（见 `docs/debug/20260427-xianyu-publish-research.md`）

## 🔑 关键设计决策（待定）

| 决策点 | 选项 | 倾向 | 理由 |
|--------|------|------|------|
| 闲鱼抓取入口 | `www.goofish.com` / `h5.m.goofish.com` | h5 移动站 | DOM 简单，风控相对宽松 |
| 淘宝抓取入口 | `s.taobao.com` / `m.taobao.com` | 移动端优先 | 桌面风控极严，滑块验证多 |
| cookie 存储 | 单文件多域 / 按域分文件 | 按域分文件 | 隔离登录态，互不影响 |
| 反封禁 | 默认 headless=true + stealth / 默认有头 | 有头模式优先 | 电商风控检测 headless 更狠 |

## 📝 开发记录

### 2026-04-17 — 项目启动 + Phase 1 骨架完成
- **已完成**：
  - 调研参考项目 `xiaohongshu-mcp` 架构（13 MCP 工具、MCP + HTTP 双协议、go-rod 分层）
  - 初始化 `docs/` 骨架（`project/` + `modules/` + `decisions/` + `debug/`）
  - 确定 Phase 1~3 里程碑划分
  - `go mod init github.com/Tsinglung-Tseng/ali.mcp`
  - 完成 `configs/` + `browser/` + `cookies/` 通用层（按平台分文件 cookie）
  - 完成 `taobao/login.go` + `xianyu/login.go` 登录骨架
  - 完成 `main.go` + `app_server.go` + `routes.go` + `middleware.go` + `handlers_api.go`
  - 完成 `mcp_server.go` + `mcp_handlers.go`（5 个工具）
  - `go build ./...` + `go vet ./...` 通过，`gofmt` 洁净
- **新增 TODO**：
  - 实机首跑校验 DOM 选择器（taobao `loggedInSel` / `qrcodeSel`，xianyu `loggedInSel`）
  - 在 Claude Code 里接通 ali-mcp，验证 MCP 工具可调用
  - git init + 首次推 GitHub
- **设计决策**：
  - ADR-001：Go module 路径用 `github.com/Tsinglung-Tseng/ali.mcp`（保留点号）
  - ADR-002：cookie 按平台分文件（`cookies/taobao.json` / `cookies/xianyu.json`）
  - 闲鱼不实现独立扫码，走阿里 SSO 从淘宝回流
  - 默认 `-headless=false`（电商风控对 headless 严）
  - HTTP 端口默认 `:18070`（避 xhs 的 `:18060`，方便并存）
- **文档更新**：
  - `docs/modules/app-server.md`
  - `docs/modules/mcp-server.md`
  - `docs/modules/taobao.md`
  - `docs/modules/xianyu.md`
  - `docs/modules/browser-cookies.md`
  - `docs/decisions/001-module-path.md`
  - `docs/decisions/002-per-platform-cookie-scope.md`
  - `CLAUDE.md`（项目规则）
