package xianyu

import (
	"context"

	"github.com/go-rod/rod"
	"github.com/pkg/errors"
)

// PublishAction 闲鱼发布商品。
//
// 注意：发布是 5 步表单（图片上传 → 标题/描述 → 分类 → 价格 → 定位 → 提交），
// 每步选择器和交互模式都需要实机走一遍才能确定。当前为 stub —— 直接返回未实现错误，
// 由 docs/debug/xianyu-publish-research.md 描述实机研究步骤，跑完之后再回填本文件。
type PublishAction struct {
	page *rod.Page //nolint:unused // reserved for future implementation
}

// NewPublish 构造发布 action（stub，未实现）。
func NewPublish(page *rod.Page) *PublishAction { return &PublishAction{page: page} }

// PublishArgs 发布参数（先定下接口，后填实现）。
type PublishArgs struct {
	Title       string   `json:"title"`        // 商品标题
	Description string   `json:"description"`  // 描述
	Price       string   `json:"price"`        // 价格（如 "99.00"）
	OriginPrice string   `json:"origin_price"` // 原价（划线价，可选）
	Images      []string `json:"images"`       // 本地图片路径或 URL（按顺序，第一张为主图）
	Category    string   `json:"category"`     // 分类（闲鱼分类系统，待研究）
	Location    string   `json:"location"`     // 发货地
}

// Publish 发布商品（未实现）。
//
// 已知未解决的研究项（详见 docs/debug/xianyu-publish-research.md）：
//   - 入口路径：h5 是否有发布入口，还是必须从 app 进
//   - 图片上传组件：input[type=file] 直接 SetFiles 还是要走拖拽 / 调起相机
//   - 分类选择器：闲鱼分类树是否有 ID-based 直达，还是必须层层点
//   - 风控：headless 下提交会不会触发滑块 / 短信验证
//   - 二次确认：提交后的成功页 / 错误提示选择器
func (a *PublishAction) Publish(ctx context.Context, args PublishArgs) (*ManageActionResult, error) {
	return nil, errors.New("xianyu_publish not implemented yet — see docs/debug/xianyu-publish-research.md for the field walkthrough required to fill in selectors")
}
