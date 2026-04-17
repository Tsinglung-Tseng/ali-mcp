package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
	"github.com/xpzouying/headless_browser"

	"github.com/Tsinglung-Tseng/ali.mcp/browser"
	"github.com/Tsinglung-Tseng/ali.mcp/configs"
	"github.com/Tsinglung-Tseng/ali.mcp/cookies"
	"github.com/Tsinglung-Tseng/ali.mcp/taobao"
	"github.com/Tsinglung-Tseng/ali.mcp/xianyu"
)

// ---------------- TaobaoService ----------------

// TaobaoService 淘宝业务服务。
type TaobaoService struct{}

func NewTaobaoService() *TaobaoService { return &TaobaoService{} }

// CheckLoginStatus 检查淘宝登录状态。
func (s *TaobaoService) CheckLoginStatus(ctx context.Context) (*LoginStatusResponse, error) {
	b := newBrowser(configs.PlatformTaobao)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	action := taobao.NewLogin(page)
	ok, err := action.CheckLoginStatus(ctx)
	if err != nil {
		return nil, err
	}
	return &LoginStatusResponse{
		Platform:   string(configs.PlatformTaobao),
		IsLoggedIn: ok,
		Username:   configs.Username,
	}, nil
}

// GetLoginQrcode 获取淘宝扫码登录二维码；扫码成功后后台异步持久化 cookie。
func (s *TaobaoService) GetLoginQrcode(ctx context.Context) (*LoginQrcodeResponse, error) {
	b := newBrowser(configs.PlatformTaobao)
	page := b.NewPage()

	closeAll := func() {
		_ = page.Close()
		b.Close()
	}

	action := taobao.NewLogin(page)
	img, loggedIn, err := action.FetchQrcodeImage(ctx)
	if err != nil || loggedIn {
		defer closeAll()
	}
	if err != nil {
		return nil, err
	}

	timeout := 4 * time.Minute
	if !loggedIn {
		go func() {
			ctxTimeout, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			defer closeAll()

			if action.WaitForLogin(ctxTimeout) {
				if err := saveCookies(page, configs.PlatformTaobao); err != nil {
					logrus.Errorf("[taobao] save cookies failed: %v", err)
				}
			}
		}()
	}

	return &LoginQrcodeResponse{
		Platform: string(configs.PlatformTaobao),
		Timeout: func() string {
			if loggedIn {
				return "0s"
			}
			return timeout.String()
		}(),
		Img:        img,
		IsLoggedIn: loggedIn,
	}, nil
}

// DeleteCookies 删除淘宝 cookie。
func (s *TaobaoService) DeleteCookies(ctx context.Context) error {
	path := cookies.GetCookiesFilePath(configs.PlatformTaobao)
	return cookies.NewLoadCookie(path).DeleteCookies()
}

// ---------------- XianyuService ----------------

// XianyuService 闲鱼业务服务。
type XianyuService struct{}

func NewXianyuService() *XianyuService { return &XianyuService{} }

// CheckLoginStatus 检查闲鱼 h5 登录状态。
func (s *XianyuService) CheckLoginStatus(ctx context.Context) (*LoginStatusResponse, error) {
	b := newBrowser(configs.PlatformXianyu)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	action := xianyu.NewLogin(page)
	ok, err := action.CheckLoginStatus(ctx)
	if err != nil {
		return nil, err
	}
	return &LoginStatusResponse{
		Platform:   string(configs.PlatformXianyu),
		IsLoggedIn: ok,
		Username:   configs.Username,
	}, nil
}

// DeleteCookies 删除闲鱼 cookie。
func (s *XianyuService) DeleteCookies(ctx context.Context) error {
	path := cookies.GetCookiesFilePath(configs.PlatformXianyu)
	return cookies.NewLoadCookie(path).DeleteCookies()
}

// ---------------- helpers ----------------

// newBrowser 为指定平台创建一个浏览器实例（自动挂对应 cookie）。
func newBrowser(p configs.Platform) *headless_browser.Browser {
	return browser.NewBrowser(
		configs.IsHeadless(),
		browser.WithBinPath(configs.GetBinPath()),
		browser.WithPlatform(p),
	)
}

// saveCookies 将当前浏览器的 cookie 导出到对应平台的文件。
func saveCookies(page *rod.Page, p configs.Platform) error {
	cks, err := page.Browser().GetCookies()
	if err != nil {
		return fmt.Errorf("get cookies: %w", err)
	}
	data, err := json.Marshal(cks)
	if err != nil {
		return fmt.Errorf("marshal cookies: %w", err)
	}
	loader := cookies.NewLoadCookie(cookies.GetCookiesFilePath(p))
	return loader.SaveCookies(data)
}
