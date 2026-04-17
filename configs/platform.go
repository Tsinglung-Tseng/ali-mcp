package configs

// Platform 平台标识，用于 cookie 分域存储、日志 tag 等。
type Platform string

const (
	PlatformTaobao Platform = "taobao"
	PlatformXianyu Platform = "xianyu"
)

// String 实现 fmt.Stringer 接口。
func (p Platform) String() string { return string(p) }

// IsValid 判断平台标识是否合法。
func (p Platform) IsValid() bool {
	switch p {
	case PlatformTaobao, PlatformXianyu:
		return true
	}
	return false
}

// Username 日志中用作"伪用户名"标识（参考 xiaohongshu-mcp）。
const Username = "ali-mcp"
