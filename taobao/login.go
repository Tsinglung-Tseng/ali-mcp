package taobao

import (
	"context"
	"time"

	"github.com/go-rod/rod"
	"github.com/pkg/errors"
)

// LoginAction 淘宝扫码登录。
// 淘宝/闲鱼共用阿里账号体系，扫码登录统一走淘宝 PC 站。
type LoginAction struct {
	page *rod.Page
}

// NewLogin 构造登录 action。
func NewLogin(page *rod.Page) *LoginAction {
	return &LoginAction{page: page}
}

// 入口 URL 及登录态标识元素选择器。
// TODO(selectors): 以下选择器基于经验假设，首次跑通需实机校验并更新。
const (
	// taobao 首页；已登录时右上角有用户昵称节点。
	homeURL = "https://www.taobao.com"
	// 登录页；未登录访问 my.taobao.com 会被跳到这里。
	loginURL = "https://login.taobao.com/member/login.jhtml"
	// 已登录后用户昵称区域（.site-nav-login-info .site-nav-user 可能变动）。
	loggedInSel = ".site-nav-login-info .site-nav-user"
	// 扫码登录面板中的二维码 img 节点。
	qrcodeSel = "#login .qrcode-img, .iconfont-qrcode + img, canvas.J_QRCodeImg"
)

// CheckLoginStatus 通过访问首页检查登录态。
func (a *LoginAction) CheckLoginStatus(ctx context.Context) (bool, error) {
	pp := a.page.Context(ctx)
	pp.MustNavigate(homeURL).MustWaitLoad()

	time.Sleep(1 * time.Second)

	exists, _, err := pp.Has(loggedInSel)
	if err != nil {
		return false, errors.Wrap(err, "check login status failed")
	}
	return exists, nil
}

// FetchQrcodeImage 拉取扫码登录二维码；若已登录返回 (nil, true, nil)。
// 返回 base64 data URL 或远程 src。
func (a *LoginAction) FetchQrcodeImage(ctx context.Context) (string, bool, error) {
	pp := a.page.Context(ctx)

	pp.MustNavigate(loginURL).MustWaitLoad()
	time.Sleep(2 * time.Second)

	// 可能已登录后会被跳回首页
	if exists, _, _ := pp.Has(loggedInSel); exists {
		return "", true, nil
	}

	el, err := pp.Element(qrcodeSel)
	if err != nil {
		return "", false, errors.Wrap(err, "qrcode element not found")
	}
	src, err := el.Attribute("src")
	if err != nil || src == nil || *src == "" {
		return "", false, errors.Wrap(err, "get qrcode src failed")
	}
	return *src, false, nil
}

// WaitForLogin 轮询登录态；用户扫码成功后返回 true，ctx 超时返回 false。
func (a *LoginAction) WaitForLogin(ctx context.Context) bool {
	pp := a.page.Context(ctx)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			el, err := pp.Element(loggedInSel)
			if err == nil && el != nil {
				return true
			}
		}
	}
}
