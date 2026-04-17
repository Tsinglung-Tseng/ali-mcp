# taobao（淘宝 action 层）

## 职责

淘宝站内的浏览器自动化操作。Phase 1 只实现扫码登录；Phase 2 扩展搜索、详情。

## 架构

```
taobao/
  └── login.go — LoginAction（扫码登录 + 登录态检查）
```

## 关键实现

`login.go`：
- `NewLogin(page) *LoginAction`
- `CheckLoginStatus(ctx) (bool, error)` — 访问 `https://www.taobao.com`，查 `.site-nav-login-info .site-nav-user`
- `FetchQrcodeImage(ctx) (src string, alreadyLoggedIn bool, err error)` — 访问 `https://login.taobao.com/member/login.jhtml`，抓二维码 `src`
- `WaitForLogin(ctx) bool` — 500ms 轮询登录态元素出现，ctx 超时返回 false

**TODO(selectors)**：所有 DOM 选择器（`loggedInSel`、`qrcodeSel`）基于经验假设，首次实机跑通后需校验更新。

## 接口

```go
type LoginAction struct{ page *rod.Page }

func NewLogin(page *rod.Page) *LoginAction
func (a *LoginAction) CheckLoginStatus(ctx context.Context) (bool, error)
func (a *LoginAction) FetchQrcodeImage(ctx context.Context) (string, bool, error)
func (a *LoginAction) WaitForLogin(ctx context.Context) bool
```

## 依赖

- `go-rod/rod` 浏览器自动化
- `pkg/errors` 错误链

## 已知问题

- 淘宝对 headless 风控极严，建议 `-headless=false` 运行
- 淘宝 PC 登录页可能跳出滑块验证码，当前 action 未处理
- 选择器未实机验证，首次跑通需要调试
