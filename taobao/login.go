package taobao

import (
	"context"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

const (
	// 访问 my.taobao.com 未登录会被重定向到登录页；已登录会留在 my。
	// 用"是否重定向到 login.taobao.com"作为登录态的稳健判断。
	probeURL = "https://i.taobao.com/my_taobao.htm"
	// 登录页；用于主动拉起扫码。
	loginURL = "https://login.taobao.com/member/login.jhtml"
	// 登录页域名片段，用于判断是否仍在登录页。
	loginDomainFragment = "login.taobao.com"
	// 扫码二维码元素；淘宝当前为 canvas，没有 src 属性，这里只用于"能找到"判断。
	// TODO(selectors): 实机确认 canvas 的确切类名。
	qrcodeSel = "canvas, #J_Quick2Static .qrcode-img, .qrcode-img img"
)

// CheckLoginStatus 访问 my.taobao.com 后看最终 URL：
// 若仍在/被跳到 login.taobao.com，则未登录；否则已登录。
func (a *LoginAction) CheckLoginStatus(ctx context.Context) (bool, error) {
	pp := a.page.Context(ctx)
	pp.MustNavigate(probeURL).MustWaitLoad()

	time.Sleep(1 * time.Second)

	info, err := pp.Info()
	if err != nil {
		return false, errors.Wrap(err, "get page info")
	}
	return !strings.Contains(info.URL, loginDomainFragment), nil
}

// FetchQrcodeImage 拉起登录页并等待二维码出现。
// 若访问时就已登录（URL 没停在 login.taobao.com），返回 (nil, true, nil)。
// 当前淘宝用 canvas 画二维码，没有 src；拿不到 src 不算错（返回 ""），
// 用户直接在可见浏览器窗口中扫码即可，登录检测靠 URL 跳转。
func (a *LoginAction) FetchQrcodeImage(ctx context.Context) (string, bool, error) {
	pp := a.page.Context(ctx)

	pp.MustNavigate(loginURL).MustWaitLoad()
	time.Sleep(2 * time.Second)

	info, err := pp.Info()
	if err != nil {
		return "", false, errors.Wrap(err, "get page info after navigate")
	}
	if !strings.Contains(info.URL, loginDomainFragment) {
		// 已经被跳走（可能本就已登录）
		return "", true, nil
	}

	// 尝试抓 img 形式的二维码；拿不到 src 也不 fail，让调用方靠 URL 检测登录态。
	if el, e := pp.Element(qrcodeSel); e == nil && el != nil {
		if src, e2 := el.Attribute("src"); e2 == nil && src != nil && *src != "" {
			return *src, false, nil
		}
	}
	logrus.Debug("[taobao] qrcode src not available (likely canvas); relying on browser window for scan")
	return "", false, nil
}

// WaitForLogin 轮询登录态：扫码成功后页面会跳离 login.taobao.com。
// 每 5 秒打一条 debug log，方便排查用户扫码过程。
func (a *LoginAction) WaitForLogin(ctx context.Context) bool {
	pp := a.page.Context(ctx)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var ticks int
	for {
		select {
		case <-ctx.Done():
			logrus.Info("[taobao] WaitForLogin: context done, timeout")
			return false
		case <-ticker.C:
			ticks++
			info, err := pp.Info()
			if err != nil {
				logrus.Debugf("[taobao] WaitForLogin page.Info err: %v", err)
				continue
			}
			if ticks%5 == 0 {
				logrus.Infof("[taobao] WaitForLogin tick=%ds, url=%s", ticks, info.URL)
			}
			if !strings.Contains(info.URL, loginDomainFragment) {
				logrus.Infof("[taobao] WaitForLogin: login detected, url=%s", info.URL)
				// 页面跳走后再等 2 秒让目标页写完 cookie
				time.Sleep(2 * time.Second)
				return true
			}
		}
	}
}
