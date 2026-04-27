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

// TaobaoItemDetailArgs 淘宝商品详情参数。
type TaobaoItemDetailArgs struct {
	URL string `json:"url" jsonschema:"商品详情 URL，形如 https://item.taobao.com/item.htm?id=... 或 https://detail.tmall.com/item.htm?id=..."`
}

// XianyuSearchArgs 闲鱼搜索参数。
type XianyuSearchArgs struct {
	Keyword string `json:"keyword" jsonschema:"搜索关键词"`
	Limit   int    `json:"limit,omitempty" jsonschema:"返回条数，默认 20；0 表示首屏全部"`
}

// XianyuItemDetailArgs 闲鱼商品详情参数。idOrURL 接受 id 或完整 URL。
type XianyuItemDetailArgs struct {
	IDOrURL string `json:"id_or_url" jsonschema:"闲鱼商品 ID（如 654321）或完整 URL（h5.m.goofish.com/item.html?id=...）"`
}

// XianyuUserProfileArgs 闲鱼用户主页参数。
type XianyuUserProfileArgs struct {
	UserID string `json:"user_id" jsonschema:"闲鱼用户 ID"`
}

// XianyuItemIDArgs 通用单 itemID 参数（擦亮 / 下架）。
type XianyuItemIDArgs struct {
	ItemID string `json:"item_id" jsonschema:"闲鱼商品 ID（卖家自己在售的商品）"`
}

// XianyuPublishArgs 闲鱼发布参数（stub 用，先定接口）。
type XianyuPublishArgs struct {
	Title       string   `json:"title" jsonschema:"商品标题"`
	Description string   `json:"description" jsonschema:"商品描述"`
	Price       string   `json:"price" jsonschema:"价格，如 99.00"`
	OriginPrice string   `json:"origin_price,omitempty" jsonschema:"原价（划线价），可选"`
	Images      []string `json:"images" jsonschema:"图片本地路径或 URL，按顺序，第一张为主图"`
	Category    string   `json:"category,omitempty" jsonschema:"分类（暂用关键词建议匹配，可留空让闲鱼自动建议）"`
	Location    string   `json:"location,omitempty" jsonschema:"发货地（如 浙江·杭州）"`
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
			Name:        "taobao_get_item_detail",
			Description: "获取淘宝/天猫商品详情：标题、价格、销量、店铺、发货地、主图、SKU 列表。需已登录。",
			Annotations: &mcp.ToolAnnotations{Title: "Taobao: Get Item Detail", ReadOnlyHint: true},
		},
		withPanicRecovery("taobao_get_item_detail", func(ctx context.Context, _ *mcp.CallToolRequest, args TaobaoItemDetailArgs) (*mcp.CallToolResult, any, error) {
			return convertToMCPResult(appServer.handleTaobaoGetItemDetail(ctx, args)), nil, nil
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
			Name:        "xianyu_search",
			Description: "在闲鱼 h5 站搜索商品，返回卡片列表（标题/价格/卖家/发货地/想要数）。游客态可访问。",
			Annotations: &mcp.ToolAnnotations{Title: "Xianyu: Search Items", ReadOnlyHint: true},
		},
		withPanicRecovery("xianyu_search", func(ctx context.Context, _ *mcp.CallToolRequest, args XianyuSearchArgs) (*mcp.CallToolResult, any, error) {
			return convertToMCPResult(appServer.handleXianyuSearch(ctx, args)), nil, nil
		}),
	)

	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "xianyu_get_item_detail",
			Description: "获取闲鱼商品详情：标题、价格、原价、描述、卖家、图片列表、想要数。idOrURL 接受 id 或完整 URL。",
			Annotations: &mcp.ToolAnnotations{Title: "Xianyu: Get Item Detail", ReadOnlyHint: true},
		},
		withPanicRecovery("xianyu_get_item_detail", func(ctx context.Context, _ *mcp.CallToolRequest, args XianyuItemDetailArgs) (*mcp.CallToolResult, any, error) {
			return convertToMCPResult(appServer.handleXianyuGetItemDetail(ctx, args)), nil, nil
		}),
	)

	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "xianyu_get_user_profile",
			Description: "获取闲鱼卖家主页摘要：昵称、信用、在售/已售数、粉丝数 + 最近 20 个商品。",
			Annotations: &mcp.ToolAnnotations{Title: "Xianyu: Get User Profile", ReadOnlyHint: true},
		},
		withPanicRecovery("xianyu_get_user_profile", func(ctx context.Context, _ *mcp.CallToolRequest, args XianyuUserProfileArgs) (*mcp.CallToolResult, any, error) {
			return convertToMCPResult(appServer.handleXianyuGetUserProfile(ctx, args)), nil, nil
		}),
	)

	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "xianyu_refresh",
			Description: "擦亮指定商品（卖家侧操作；把商品刷到 feed 顶部）。需登录，需是自己的商品。",
			Annotations: &mcp.ToolAnnotations{Title: "Xianyu: Refresh Item (擦亮)"},
		},
		withPanicRecovery("xianyu_refresh", func(ctx context.Context, _ *mcp.CallToolRequest, args XianyuItemIDArgs) (*mcp.CallToolResult, any, error) {
			return convertToMCPResult(appServer.handleXianyuRefresh(ctx, args)), nil, nil
		}),
	)

	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "xianyu_delist",
			Description: "下架指定商品（卖家侧操作；商品转为已下架，可重新上架）。需登录。",
			Annotations: &mcp.ToolAnnotations{Title: "Xianyu: Delist Item (下架)", DestructiveHint: boolPtr(true)},
		},
		withPanicRecovery("xianyu_delist", func(ctx context.Context, _ *mcp.CallToolRequest, args XianyuItemIDArgs) (*mcp.CallToolResult, any, error) {
			return convertToMCPResult(appServer.handleXianyuDelist(ctx, args)), nil, nil
		}),
	)

	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "xianyu_publish",
			Description: "发布闲鱼商品（stub，未实现）。当前调用会返回 not implemented；发布是 5 步表单 + 图片上传，需要实机走查后填回选择器。",
			Annotations: &mcp.ToolAnnotations{Title: "Xianyu: Publish Item (NOT IMPLEMENTED)"},
		},
		withPanicRecovery("xianyu_publish", func(ctx context.Context, _ *mcp.CallToolRequest, args XianyuPublishArgs) (*mcp.CallToolResult, any, error) {
			return convertToMCPResult(appServer.handleXianyuPublish(ctx, args)), nil, nil
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

	logrus.Infof("registered %d MCP tools", 12)
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
