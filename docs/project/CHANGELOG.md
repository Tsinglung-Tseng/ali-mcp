# CHANGELOG

项目变更时间线，按日期追加。

### 2026-04-27
- [decision] project: 项目开源，public repo `github.com/Tsinglung-Tseng/ali-mcp`，MIT license
- [feature] taobao: `taobao_search` MCP 工具落地（s.taobao.com，simba 广告过滤 + 去重 + 滚动加载）
- [feature] taobao: `taobao_get_item_detail` MCP 工具（item.taobao.com / detail.tmall.com，标题/价格/销量/店铺/SKU/主图）
- [feature] xianyu: `xianyu_search` MCP 工具（h5.m.goofish.com，游客态可访问）
- [feature] xianyu: `xianyu_get_item_detail` MCP 工具（标题/价格/原价/描述/卖家/图片列表/想要数）
- [feature] xianyu: `xianyu_get_user_profile` MCP 工具（卖家主页摘要 + 最近 20 个商品）
- [feature] xianyu: `xianyu_refresh` 擦亮工具（卖家把商品刷到 feed 顶部，单步操作）
- [feature] xianyu: `xianyu_delist` 下架工具（两步：三点菜单 → 下架）
- [feature] xianyu: `xianyu_publish` stub（5 步表单太复杂闭门盲写不可靠，留 `docs/debug/20260427-xianyu-publish-research.md` 实机走查指南）
- [discovery] xianyu: 卖家侧管理操作必须经过"我的-在售列表"入口（`/2/myItems.html?status=on_sale`），通过 itemID 在卡片 href 中匹配定位
- [feature] handlers: 12 个 MCP 工具全部注册并通过 `go vet`；HTTP API 端点对称扩展（`/api/v1/taobao/item`、`/api/v1/xianyu/{search,item,user,refresh,delist}`）
- [decision] project: 选择器全部用 `[class*="X"]` 前缀匹配应对 React hash class，每个动作文件顶部带 `TODO(selectors)` 注释提醒实机校正

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
