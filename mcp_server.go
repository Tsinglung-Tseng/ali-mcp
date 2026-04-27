package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"runtime/debug"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// boolPtr 方便 ToolAnnotations 传指针。
func boolPtr(b bool) *bool { return &b }

// TaobaoSearchArgs 淘宝搜索工具参数。
type TaobaoSearchArgs struct {
	Keyword string `json:"keyword" jsonschema:"搜索关键词，如 健康的炒锅"`
	Limit   int    `json:"limit,omitempty" jsonschema:"返回条数，默认 20；0 表示本页全部"`
}

// InitMCPServer 初始化 MCP Server 并注册所有工具。
func InitMCPServer(appServer *AppServer) *mcp.Server {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "ali-mcp",
			Version: "0.1.0",
		},
		nil,
	)
	registerTools(server, appServer)
	logrus.Info("MCP Server initialized (ali-mcp)")
	return server
}

// withPanicRecovery 工具 panic 时返回结构化错误，避免整个 server 崩溃。
func withPanicRecovery[T any](
	toolName string,
	handler func(context.Context, *mcp.CallToolRequest, T) (*mcp.CallToolResult, any, error),
) func(context.Context, *mcp.CallToolRequest, T) (*mcp.CallToolResult, any, error) {

	return func(ctx context.Context, req *mcp.CallToolRequest, args T) (result *mcp.CallToolResult, resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithFields(logrus.Fields{
					"tool":  toolName,
					"panic": r,
				}).Error("tool panic")
				logrus.Errorf("stack:\n%s", debug.Stack())

				result = &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("工具 %s 执行时发生内部错误: %v\n\n详见服务端日志。", toolName, r),
						},
					},
					IsError: true,
				}
				resp = nil
				err = nil
			}
		}()
		return handler(ctx, req, args)
	}
}

// registerTools 注册所有 MCP 工具。
func registerTools(server *mcp.Server, appServer *AppServer) {
	// ---------------- 淘宝 ----------------

	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "taobao_check_login_status",
			Description: "检查淘宝登录状态",
			Annotations: &mcp.ToolAnnotations{Title: "Taobao: Check Login Status", ReadOnlyHint: true},
		},
		withPanicRecovery("taobao_check_login_status", func(ctx context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
			return convertToMCPResult(appServer.handleTaobaoCheckLoginStatus(ctx)), nil, nil
		}),
	)

	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "taobao_get_login_qrcode",
			Description: "获取淘宝扫码登录二维码（返回 Base64 图片和超时时间）",
			Annotations: &mcp.ToolAnnotations{Title: "Taobao: Get Login QR Code", ReadOnlyHint: true},
		},
		withPanicRecovery("taobao_get_login_qrcode", func(ctx context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
			return convertToMCPResult(appServer.handleTaobaoGetLoginQrcode(ctx)), nil, nil
		}),
	)

	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "taobao_search",
			Description: "在淘宝搜索商品，返回前 N 个商品的标题/价格/店铺/发货地/商品 URL。需已登录。",
			Annotations: &mcp.ToolAnnotations{Title: "Taobao: Search Items", ReadOnlyHint: true},
		},
		withPanicRecovery("taobao_search", func(ctx context.Context, _ *mcp.CallToolRequest, args TaobaoSearchArgs) (*mcp.CallToolResult, any, error) {
			return convertToMCPResult(appServer.handleTaobaoSearch(ctx, args)), nil, nil
		}),
	)

	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "taobao_delete_cookies",
			Description: "删除淘宝 cookies 文件，重置登录状态",
			Annotations: &mcp.ToolAnnotations{Title: "Taobao: Delete Cookies", DestructiveHint: boolPtr(true)},
		},
		withPanicRecovery("taobao_delete_cookies", func(ctx context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
			return convertToMCPResult(appServer.handleTaobaoDeleteCookies(ctx)), nil, nil
		}),
	)

	// ---------------- 闲鱼 ----------------

	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "xianyu_check_login_status",
			Description: "检查闲鱼登录状态（h5.m.goofish.com）",
			Annotations: &mcp.ToolAnnotations{Title: "Xianyu: Check Login Status", ReadOnlyHint: true},
		},
		withPanicRecovery("xianyu_check_login_status", func(ctx context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
			return convertToMCPResult(appServer.handleXianyuCheckLoginStatus(ctx)), nil, nil
		}),
	)

	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "xianyu_delete_cookies",
			Description: "删除闲鱼 cookies 文件",
			Annotations: &mcp.ToolAnnotations{Title: "Xianyu: Delete Cookies", DestructiveHint: boolPtr(true)},
		},
		withPanicRecovery("xianyu_delete_cookies", func(ctx context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
			return convertToMCPResult(appServer.handleXianyuDeleteCookies(ctx)), nil, nil
		}),
	)

	logrus.Infof("registered %d MCP tools", 6)
}

// convertToMCPResult 将内部 MCPToolResult 转成 SDK 格式。
func convertToMCPResult(result *MCPToolResult) *mcp.CallToolResult {
	var contents []mcp.Content
	for _, c := range result.Content {
		switch c.Type {
		case "text":
			contents = append(contents, &mcp.TextContent{Text: c.Text})
		case "image":
			data, err := base64.StdEncoding.DecodeString(c.Data)
			if err != nil {
				logrus.WithError(err).Error("decode base64 image failed")
				contents = append(contents, &mcp.TextContent{Text: "图片数据解码失败: " + err.Error()})
			} else {
				contents = append(contents, &mcp.ImageContent{Data: data, MIMEType: c.MimeType})
			}
		}
	}
	return &mcp.CallToolResult{Content: contents, IsError: result.IsError}
}
