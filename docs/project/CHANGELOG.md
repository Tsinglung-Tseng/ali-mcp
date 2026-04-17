# CHANGELOG

项目变更时间线，按日期追加。

### 2026-04-17
- [discovery] docs: 初始化 `docs/` 骨架（`project/` + `modules/` + `decisions/` + `debug/`），创建 `dev-plan.md`
- [discovery] project: 调研参考项目 `xiaohongshu-mcp`，确定 ali-mcp 三阶段里程碑
- [decision] project: 确认 Go module 路径 `github.com/Tsinglung-Tseng/ali.mcp`（目录名保留点号，匹配 GitHub 仓库）
- [feature] configs: 新增 `Platform` 类型（`PlatformTaobao` / `PlatformXianyu`）和全局 headless / binPath 配置
- [feature] cookies: 按平台分文件存储（`cookies/taobao.json` / `cookies/xianyu.json`），支持 `ALI_COOKIES_DIR` 环境变量覆盖
- [feature] browser: 按 `WithPlatform()` 自动加载对应平台 cookie，代理走 `ALI_PROXY`（带凭据 mask）
- [feature] taobao: `LoginAction` 骨架（`CheckLoginStatus` / `FetchQrcodeImage` / `WaitForLogin`），DOM 选择器 TODO 待实机校验
- [feature] xianyu: `LoginAction` 骨架（仅 `CheckLoginStatus`，依赖阿里 SSO 从淘宝回流 cookie）
- [feature] app-server: Gin HTTP + MCP Streamable HTTP 双协议，默认监听 `:18070`
- [feature] mcp-server: 注册 5 个工具（taobao 3 + xianyu 2），泛型 `withPanicRecovery` 保护
- [decision] app-server: CLI `-headless` 默认 `false`（电商风控对 headless 严，与 xhs 默认 true 不同）
