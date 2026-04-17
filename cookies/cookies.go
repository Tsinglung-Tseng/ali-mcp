package cookies

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/Tsinglung-Tseng/ali.mcp/configs"
)

// Cookier cookie 读写接口。
type Cookier interface {
	LoadCookies() ([]byte, error)
	SaveCookies(data []byte) error
	DeleteCookies() error
	Path() string
}

type localCookie struct {
	path string
}

// NewLoadCookie 按绝对路径构造 cookie loader。
// path 必须非空，缺失直接 panic（fail loud）。
func NewLoadCookie(path string) Cookier {
	if path == "" {
		panic("cookie path is required")
	}
	return &localCookie{path: path}
}

func (c *localCookie) Path() string { return c.path }

// LoadCookies 从文件加载 cookies。
func (c *localCookie) LoadCookies() ([]byte, error) {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return nil, errors.Wrap(err, "read cookie file")
	}
	return data, nil
}

// SaveCookies 保存 cookies 到文件。
func (c *localCookie) SaveCookies(data []byte) error {
	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return errors.Wrap(err, "mkdir cookie dir")
	}
	return os.WriteFile(c.path, data, 0o600)
}

// DeleteCookies 删除 cookies 文件；不存在视为已删除。
func (c *localCookie) DeleteCookies() error {
	if _, err := os.Stat(c.path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(c.path)
}

// CookieDir 返回 cookie 存放目录。
// 优先读 ALI_COOKIES_DIR；未设置则 fallback 到 <cwd>/.cookies。
// 注 1：此 fallback 是"明确文档化的默认值"，便于本地调试，属全局规则例外。
// 注 2：用 `.cookies`（带点前缀）而非 `cookies`，避免和源码包 `cookies/` 目录冲突。
func CookieDir() string {
	if dir := os.Getenv("ALI_COOKIES_DIR"); dir != "" {
		return dir
	}
	return ".cookies"
}

// GetCookiesFilePath 返回指定平台的 cookie 文件绝对/相对路径。
// 非法平台直接 panic，避免写错参数造成 cookie 串流。
func GetCookiesFilePath(p configs.Platform) string {
	if !p.IsValid() {
		panic(fmt.Sprintf("invalid platform: %q", p))
	}
	return filepath.Join(CookieDir(), fmt.Sprintf("%s.json", p))
}
