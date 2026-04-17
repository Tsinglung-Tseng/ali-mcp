# xianyu（闲鱼 action 层）

## 职责

闲鱼 h5 站的浏览器自动化操作。Phase 1 只实现登录态检查（靠淘宝 SSO 回流获取 cookie）；Phase 2 扩展搜索、商品详情、用户主页。

## 架构

```
xianyu/
  └── login.go — LoginAction（仅登录态检查）
```

## 为何没独立扫码登录

闲鱼和淘宝共用阿里账号体系：
- 淘宝站 cookie scope：`.taobao.com`
- 闲鱼 h5 cookie scope：`.goofish.com`、`.m.goofish.com`

用户在淘宝登录后，访问 `h5.m.goofish.com` 会自动触发 SSO 回流，写入闲鱼域的 cookie。因此本模块只需检查登录态，不需要独立扫码。

## 关键实现

`login.go`：
- 入口 `https://h5.m.goofish.com`
- 登录态识别选择器 `.user-avatar, [class*='avatar'], .mine-entry`（TODO：实机校验）

## 接口

```go
type LoginAction struct{ page *rod.Page }

func NewLogin(page *rod.Page) *LoginAction
func (a *LoginAction) CheckLoginStatus(ctx context.Context) (bool, error)
```

## 依赖

- `go-rod/rod`
- `pkg/errors`

## 已知问题

- 选择器是经验假设，未实机验证
- SSO 回流可能需要先访问淘宝域名触发，当前未做主动"加热"；若首次 CheckLoginStatus 返回 false，可能需要先访问淘宝再回闲鱼
- Phase 2 加入 search/detail 时，需注意闲鱼 h5 是 SPA，数据走 XHR，可能要在 rod 中 hijack 请求抓包
