package xianyu

// Item 闲鱼商品卡片（搜索 / 用户主页通用）。
type Item struct {
	ID       string `json:"id,omitempty"`        // 商品 ID
	Title    string `json:"title"`               // 标题
	Price    string `json:"price,omitempty"`     // 价格原文（含 "¥"）
	Seller   string `json:"seller,omitempty"`    // 卖家昵称
	Location string `json:"location,omitempty"`  // 发货地
	URL      string `json:"url,omitempty"`       // 商品详情 URL
	ImageURL string `json:"image_url,omitempty"` // 主图 URL
	WantNum  string `json:"want_num,omitempty"`  // 想要数（"123 人想要"）
}

// Detail 闲鱼商品详情。
type Detail struct {
	URL         string   `json:"url"`
	ID          string   `json:"id,omitempty"`
	Title       string   `json:"title"`
	Price       string   `json:"price,omitempty"`
	OriginPrice string   `json:"origin_price,omitempty"` // 原价（划线价）
	Description string   `json:"description,omitempty"`
	Seller      string   `json:"seller,omitempty"`
	SellerURL   string   `json:"seller_url,omitempty"`
	Location    string   `json:"location,omitempty"`
	Images      []string `json:"images,omitempty"`
	WantNum     string   `json:"want_num,omitempty"`
	ViewNum     string   `json:"view_num,omitempty"`
	PostedAt    string   `json:"posted_at,omitempty"`
}

// UserProfile 闲鱼卖家主页摘要。
type UserProfile struct {
	URL          string `json:"url"`
	UserID       string `json:"user_id,omitempty"`
	Nickname     string `json:"nickname"`
	AvatarURL    string `json:"avatar_url,omitempty"`
	Intro        string `json:"intro,omitempty"`         // 个人简介
	CreditScore  string `json:"credit_score,omitempty"`  // 芝麻信用 / 闲鱼信用
	SellingCount string `json:"selling_count,omitempty"` // 在售数
	SoldCount    string `json:"sold_count,omitempty"`    // 已售数
	FollowerNum  string `json:"follower_num,omitempty"`
	RecentItems  []Item `json:"recent_items,omitempty"` // 最近发布的商品
}

// ManageActionResult 卖家侧操作（擦亮/下架）的结果。
type ManageActionResult struct {
	ItemID  string `json:"item_id"`
	Action  string `json:"action"` // refresh | delist | publish
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
