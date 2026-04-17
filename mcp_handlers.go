package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Tsinglung-Tseng/ali.mcp/configs"
	"github.com/Tsinglung-Tseng/ali.mcp/cookies"
)

// ---------------- 淘宝 MCP handlers ----------------

func (s *AppServer) handleTaobaoCheckLoginStatus(ctx context.Context) *MCPToolResult {
	logrus.Info("MCP: taobao check_login_status")
	status, err := s.taobao.CheckLoginStatus(ctx)
	if err != nil {
		return errText("检查淘宝登录状态失败: " + err.Error())
	}
	if status.IsLoggedIn {
		return okText(fmt.Sprintf("✅ 淘宝已登录\n用户标识: %s", status.Username))
	}
	return okText("❌ 淘宝未登录\n\n使用 taobao_get_login_qrcode 获取二维码扫码登录。")
}

func (s *AppServer) handleTaobaoGetLoginQrcode(ctx context.Context) *MCPToolResult {
	logrus.Info("MCP: taobao get_login_qrcode")
	result, err := s.taobao.GetLoginQrcode(ctx)
	if err != nil {
		return errText("获取淘宝登录二维码失败: " + err.Error())
	}
	if result.IsLoggedIn {
		return okText("淘宝当前已处于登录状态")
	}
	return qrcodeResult("请用手机淘宝扫码登录 👇", result.Img, result.Timeout)
}

func (s *AppServer) handleTaobaoDeleteCookies(ctx context.Context) *MCPToolResult {
	logrus.Info("MCP: taobao delete_cookies")
	if err := s.taobao.DeleteCookies(ctx); err != nil {
		return errText("删除淘宝 cookies 失败: " + err.Error())
	}
	path := cookies.GetCookiesFilePath(configs.PlatformTaobao)
	return okText(fmt.Sprintf("淘宝 cookies 已删除，下次操作需重新登录。\n\n路径: %s", path))
}

// ---------------- 闲鱼 MCP handlers ----------------

func (s *AppServer) handleXianyuCheckLoginStatus(ctx context.Context) *MCPToolResult {
	logrus.Info("MCP: xianyu check_login_status")
	status, err := s.xianyu.CheckLoginStatus(ctx)
	if err != nil {
		return errText("检查闲鱼登录状态失败: " + err.Error())
	}
	if status.IsLoggedIn {
		return okText(fmt.Sprintf("✅ 闲鱼已登录\n用户标识: %s", status.Username))
	}
	return okText("❌ 闲鱼未登录\n\n闲鱼走阿里 SSO，请先通过 taobao_get_login_qrcode 登录淘宝；若仍未同步请访问 h5.m.goofish.com 触发 SSO 回流。")
}

func (s *AppServer) handleXianyuDeleteCookies(ctx context.Context) *MCPToolResult {
	logrus.Info("MCP: xianyu delete_cookies")
	if err := s.xianyu.DeleteCookies(ctx); err != nil {
		return errText("删除闲鱼 cookies 失败: " + err.Error())
	}
	path := cookies.GetCookiesFilePath(configs.PlatformXianyu)
	return okText(fmt.Sprintf("闲鱼 cookies 已删除。\n\n路径: %s", path))
}

// ---------------- helpers ----------------

func okText(s string) *MCPToolResult {
	return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: s}}}
}

func errText(s string) *MCPToolResult {
	return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: s}}, IsError: true}
}

// qrcodeResult 返回提示文本 + 二维码图片（base64 去掉 data URL 前缀）。
// timeout 形如 "4m0s"，渲染绝对过期时间；解析失败则按当前时间渲染。
func qrcodeResult(prompt, imgSrc, timeout string) *MCPToolResult {
	deadline := time.Now()
	if d, err := time.ParseDuration(timeout); err == nil {
		deadline = deadline.Add(d)
	}
	text := fmt.Sprintf("%s（请在 %s 前完成）", prompt, deadline.Format("2006-01-02 15:04:05"))

	imgData := strings.TrimPrefix(imgSrc, "data:image/png;base64,")
	return &MCPToolResult{
		Content: []MCPContent{
			{Type: "text", Text: text},
			{Type: "image", MimeType: "image/png", Data: imgData},
		},
	}
}
