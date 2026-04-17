# ali-mcp Project Guidelines

## 本地开发规范

- **格式化**：每次修改完 Go 源码后自动 `gofmt -w` / `goimports -w`。
- **测试/构建中间产物**：用完即删，不留残渣。
- **分支开发**：所有 feature 变更走分支；未经明确同意不得推送远程。
- **Review 流程**：本地 review → PR review → 合并。
- **代码风格**：不过度设计，保持简洁易读；中文注释简洁明了，专业名词可用英文。

## 项目专属原则（来自参考项目 xiaohongshu-mcp）

- **优先 go-rod，慎用 JS 注入**：PR 中若出现大量 JS 注入，检查是否可用 go-rod 原生 API 替代；不能替代再注入，并注释理由。
- **登录态按域隔离**：淘宝/闲鱼账号共用阿里系，但 cookie 按域分文件存放（`cookies/taobao.json`、`cookies/xianyu.json`），避免 scope 污染。
- **电商风控**：默认提供 `--headless=false` 选项；headless 模式下务必启用 `go-rod/stealth`。
- **反模式警告**：所有"检查不到元素就重试 N 次"的逻辑，必须带超时和 stop 信号，不允许死循环。

## 配置读取铁律（来自 user global rules）

- **禁止硬编码**：API Key / Token、服务地址端口（含内网 IP）、用户路径（`/Users/<name>/...`）、数据库连接串一律从 env 或 config 读取，缺失即 fail loud，不静默降级。
- **No fallback on config & field reads**：读取 env / config / dict / struct 字段时，禁止 `or default` / `.get(k, default)` / `try/except` 静默兜底。字段缺失直接 panic，让 bug 立即暴露，不允许下游才爆。
- **例外**：用户输入校验、明确文档化的可选字段、外部系统兼容层 —— 就地注释说明为何可容忍缺失。

## Go 工具链

- **模块代理**：`GOPROXY=https://goproxy.cn,direct`（国内网络，走七牛镜像）
- **依赖**：只添加真实使用的依赖；定期 `go mod tidy`。
- **测试**：关键 action 写 `*_test.go`，跑 `go test ./...` 需 pass。

## 文档结构（dev-framework 管理）

- `docs/project/dev-plan.md` — 开发计划（活文档）
- `docs/project/CHANGELOG.md` — 变更时间线
- `docs/modules/<module>.md` — 模块活文档（真相来源）
- `docs/decisions/NNN-<desc>.md` — 架构决策记录（只增不改）
- `docs/debug/YYYYMMDD-<desc>.md` — 调试记录（一次性）
