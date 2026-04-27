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

// Search 淘宝商品搜索。limit 0 表示返回当页全部。
func (s *TaobaoService) Search(ctx context.Context, keyword string, limit int) (*taobao.SearchResult, error) {
	b := newBrowser(configs.PlatformTaobao)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	return taobao.NewSearch(page).Search(ctx, keyword, limit)
}

// GetItemDetail 淘宝商品详情。需已登录。
func (s *TaobaoService) GetItemDetail(ctx context.Context, itemURL string) (*taobao.Detail, error) {
	b := newBrowser(configs.PlatformTaobao)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	return taobao.NewDetail(page).GetDetail(ctx, itemURL)
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

// Search 闲鱼商品搜索（h5 站，游客态可访问）。
func (s *XianyuService) Search(ctx context.Context, keyword string, limit int) (*xianyu.SearchResult, error) {
	b := newBrowser(configs.PlatformXianyu)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	return xianyu.NewSearch(page).Search(ctx, keyword, limit)
}

// GetItemDetail 闲鱼商品详情。idOrURL 接受商品 ID 或完整 URL。
func (s *XianyuService) GetItemDetail(ctx context.Context, idOrURL string) (*xianyu.Detail, error) {
	b := newBrowser(configs.PlatformXianyu)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	return xianyu.NewDetail(page).GetDetail(ctx, idOrURL)
}

// GetUserProfile 闲鱼卖家主页摘要 + 最近商品。
func (s *XianyuService) GetUserProfile(ctx context.Context, userID string) (*xianyu.UserProfile, error) {
	b := newBrowser(configs.PlatformXianyu)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	return xianyu.NewUser(page).GetProfile(ctx, userID)
}

// RefreshItem 擦亮商品（需登录）。
func (s *XianyuService) RefreshItem(ctx context.Context, itemID string) (*xianyu.ManageActionResult, error) {
	b := newBrowser(configs.PlatformXianyu)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	return xianyu.NewManage(page).Refresh(ctx, itemID)
}

// DelistItem 下架商品（需登录）。
func (s *XianyuService) DelistItem(ctx context.Context, itemID string) (*xianyu.ManageActionResult, error) {
	b := newBrowser(configs.PlatformXianyu)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	return xianyu.NewManage(page).Delist(ctx, itemID)
}

// PublishItem 发布商品（stub，未实现）。
func (s *XianyuService) PublishItem(ctx context.Context, args xianyu.PublishArgs) (*xianyu.ManageActionResult, error) {
	b := newBrowser(configs.PlatformXianyu)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	return xianyu.NewPublish(page).Publish(ctx, args)
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
