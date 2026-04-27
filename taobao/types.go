package taobao

// Item 商品卡片（搜索 / 详情共用）。
type Item struct {
	ID        string `json:"id,omitempty"`         // 商品 ID (nid / item_id)
	Title     string `json:"title"`                // 商品标题
	Price     string `json:"price,omitempty"`      // 当前价；淘宝原文（含 "¥"）
	DealCount string `json:"deal_count,omitempty"` // 成交量原文（如 "月销 1.2万"）
	Shop      string `json:"shop,omitempty"`       // 店铺名
	Location  string `json:"location,omitempty"`   // 发货地
	URL       string `json:"url,omitempty"`        // 商品详情 URL
	ImageURL  string `json:"image_url,omitempty"`  // 主图 URL
}
