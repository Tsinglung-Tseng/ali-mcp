package main

// HTTP API 响应通用类型

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details any    `json:"details,omitempty"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data"`
	Message string `json:"message,omitempty"`
}

// MCP 工具内部类型（见 mcp_server.go 的 convertToMCPResult）

// MCPToolResult MCP 工具结果（内部使用）
type MCPToolResult struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

// MCPContent MCP 内容（内部使用）
type MCPContent struct {
	Type     string `json:"type"`     // text | image
	Text     string `json:"text"`     // text 类型的正文
	MimeType string `json:"mimeType"` // image 类型的 MIME
	Data     string `json:"data"`     // image 类型的 base64
}

// Service 层通用响应

// LoginStatusResponse 登录状态响应
type LoginStatusResponse struct {
	Platform   string `json:"platform"`
	IsLoggedIn bool   `json:"is_logged_in"`
	Username   string `json:"username,omitempty"`
}

// LoginQrcodeResponse 登录扫码二维码响应
type LoginQrcodeResponse struct {
	Platform   string `json:"platform"`
	Timeout    string `json:"timeout"`
	IsLoggedIn bool   `json:"is_logged_in"`
	Img        string `json:"img,omitempty"`
}

// DeleteCookiesResponse 删除 cookie 响应
type DeleteCookiesResponse struct {
	Platform   string `json:"platform"`
	CookiePath string `json:"cookie_path"`
	Message    string `json:"message"`
}
