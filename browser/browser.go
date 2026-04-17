package browser

import (
	"net/url"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/xpzouying/headless_browser"

	"github.com/Tsinglung-Tseng/ali.mcp/configs"
	"github.com/Tsinglung-Tseng/ali.mcp/cookies"
)

type browserConfig struct {
	binPath  string
	platform configs.Platform // 决定加载哪份 cookie
}

// Option 浏览器构造选项。
type Option func(*browserConfig)

// WithBinPath 指定浏览器二进制路径。
func WithBinPath(binPath string) Option {
	return func(c *browserConfig) { c.binPath = binPath }
}

// WithPlatform 指定平台（决定加载哪份 cookie）。
// 必填：无平台时不加载 cookie。
func WithPlatform(p configs.Platform) Option {
	return func(c *browserConfig) { c.platform = p }
}

// maskProxyCredentials 打日志时屏蔽代理密码。
func maskProxyCredentials(proxyURL string) string {
	u, err := url.Parse(proxyURL)
	if err != nil || u.User == nil {
		return proxyURL
	}
	if _, hasPassword := u.User.Password(); hasPassword {
		u.User = url.UserPassword("***", "***")
	} else {
		u.User = url.User("***")
	}
	return u.String()
}

// NewBrowser 创建一个浏览器实例。
// 若 options 中指定了 platform，则自动加载对应 cookie 文件。
func NewBrowser(headless bool, options ...Option) *headless_browser.Browser {
	cfg := &browserConfig{}
	for _, opt := range options {
		opt(cfg)
	}

	opts := []headless_browser.Option{
		headless_browser.WithHeadless(headless),
	}
	if cfg.binPath != "" {
		opts = append(opts, headless_browser.WithChromeBinPath(cfg.binPath))
	}

	// 代理环境变量：ALI_PROXY（形如 http://user:pass@host:port）
	if proxy := os.Getenv("ALI_PROXY"); proxy != "" {
		opts = append(opts, headless_browser.WithProxy(proxy))
		logrus.Infof("using proxy: %s", maskProxyCredentials(proxy))
	}

	// 加载对应平台的 cookie（若指定了 platform）
	if cfg.platform.IsValid() {
		cookiePath := cookies.GetCookiesFilePath(cfg.platform)
		loader := cookies.NewLoadCookie(cookiePath)
		if data, err := loader.LoadCookies(); err == nil {
			opts = append(opts, headless_browser.WithCookies(string(data)))
			logrus.Debugf("[%s] loaded cookies from %s", cfg.platform, cookiePath)
		} else {
			logrus.Warnf("[%s] no cookie loaded: %v", cfg.platform, err)
		}
	}

	return headless_browser.New(opts...)
}
