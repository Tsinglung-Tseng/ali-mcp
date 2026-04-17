package configs

// 浏览器相关的全局配置（由 main 初始化，后续只读）。
var (
	useHeadless = true
	binPath     = "" // 浏览器二进制路径，空则走 headless_browser 自动下载
)

// InitHeadless 初始化 headless 模式开关。
func InitHeadless(h bool) {
	useHeadless = h
}

// IsHeadless 是否无头模式。
func IsHeadless() bool {
	return useHeadless
}

// SetBinPath 设置浏览器二进制路径。
func SetBinPath(b string) {
	binPath = b
}

// GetBinPath 获取浏览器二进制路径。
func GetBinPath() string {
	return binPath
}
