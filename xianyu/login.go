package xianyu

import (
	"context"
	"time"

	"github.com/go-rod/rod"
	"github.com/pkg/errors"
)

// LoginAction 闲鱼登录态操作。
// 注意：闲鱼账号体系和淘宝共享（阿里 SSO），但 cookie scope 独立：
//   - 淘宝站 cookie scope: .taobao.com
//   - 闲鱼 h5 站 cookie scope: .goofish.com / .m.goofish.com
//
// 因此扫码登录走淘宝，然后在闲鱼域下再跑一次 SSO 回流获取闲鱼 cookie。
type LoginAction struct {
	page *rod.Page
}

// NewLogin 构造登录 action。
func NewLogin(page *rod.Page) *LoginAction {
	return &LoginAction{page: page}
}

// 入口 URL 及登录态标识元素选择器。
// TODO(selectors): 首次跑通需实机校验并更新。
const (
	// 闲鱼 h5 首页；结构简单、风控相对宽松。
	h5HomeURL = "https://h5.m.goofish.com"
	// 已登录后"我的"入口节点（头像/昵称）。
	loggedInSel = ".user-avatar, [class*='avatar'], .mine-entry"
)

// CheckLoginStatus 访问闲鱼 h5 首页检查登录态。
func (a *LoginAction) CheckLoginStatus(ctx context.Context) (bool, error) {
	pp := a.page.Context(ctx)
	pp.MustNavigate(h5HomeURL).MustWaitLoad()

	time.Sleep(1 * time.Second)

	exists, _, err := pp.Has(loggedInSel)
	if err != nil {
		return false, errors.Wrap(err, "check xianyu login status failed")
	}
	return exists, nil
}
