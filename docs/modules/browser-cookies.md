# browser + cookies（通用层）

## 职责

- `configs/` — 进程级全局配置（headless / 浏览器路径 / 平台常量）
- `browser/` — 基于 `xpzouying/headless_browser` 的浏览器工厂，按平台自动挂 cookie
- `cookies/` — 按平台分文件的 cookie loader

## 为何把 browser + cookies 绑一起

平台隔离的核心机制在这里：`browser.NewBrowser(headless, WithPlatform(p))` 会调 `cookies.GetCookiesFilePath(p)` 加载对应平台的 cookie，淘宝/闲鱼 session 互不污染。

## 关键实现

### configs

- `configs.go` — 单例 `useHeadless` + `binPath` 的 getter/setter
- `platform.go` — `Platform` 类型 + `PlatformTaobao` / `PlatformXianyu` 常量 + `IsValid()`

### browser

- `Option` pattern：`WithBinPath()`、`WithPlatform()`
- 代理走环境变量 `ALI_PROXY`（格式 `http://user:pass@host:port`），日志打印时自动 mask 凭据
- 未传 `WithPlatform` 时不加载 cookie（匿名访问场景）

### cookies

- `Cookier` 接口：`LoadCookies / SaveCookies / DeleteCookies / Path`
- `CookieDir()` 读 `ALI_COOKIES_DIR`，未设置则 fallback 到 `./.cookies`（点前缀避免与源码包 `cookies/` 冲突）
- `GetCookiesFilePath(platform)` 返回 `<cookie_dir>/<platform>.json`
- 非法 platform → `panic`（fail loud，避免 cookie 串流）
- `NewLoadCookie("")` → `panic`（path 缺失是编程错误）
- `SaveCookies` 写入前自动 `MkdirAll`，权限 `0600`

## 接口

```go
// configs
configs.InitHeadless(bool); configs.IsHeadless() bool
configs.SetBinPath(string); configs.GetBinPath() string
configs.Platform; PlatformTaobao / PlatformXianyu
configs.Username  // ali-mcp（日志标识）

// browser
browser.NewBrowser(headless bool, opts ...Option) *headless_browser.Browser
browser.WithBinPath(string) Option
browser.WithPlatform(configs.Platform) Option

// cookies
cookies.NewLoadCookie(path string) Cookier
cookies.GetCookiesFilePath(p configs.Platform) string
cookies.CookieDir() string
```

## 环境变量

| 变量 | 用途 | 默认 |
|------|------|------|
| `ALI_COOKIES_DIR` | cookie 文件存放目录 | `./.cookies` |
| `ALI_PROXY` | HTTP 代理（可含 user:pass） | 无 |
| `ROD_BROWSER_BIN` | 浏览器二进制路径 | 无（走 rod 自动下载） |

## 依赖

- `xpzouying/headless_browser v0.3.0`
- `go-rod/rod v0.116.2`
- `pkg/errors v0.9.1`

## 已知问题

- cookie 文件明文存储；本地调试可接受，生产须加密
- 单进程全局 `configs.useHeadless` 单例，多租户场景需改成 per-request context
