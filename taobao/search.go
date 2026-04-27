package taobao

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// SearchAction 淘宝商品搜索。
// 需已登录（cookie 有效）。未登录会被跳到登录页。
type SearchAction struct {
	page *rod.Page
}

// NewSearch 构造搜索 action。
func NewSearch(page *rod.Page) *SearchAction { return &SearchAction{page: page} }

// 选择器 —— 淘宝 s.taobao.com 搜索页当前版本（2026-04）。
// DOM 类名带 hash 后缀（React 模块样式），用前缀匹配。
// TODO(selectors): 实机验证，若改版立即更新。
const (
	searchBaseURL = "https://s.taobao.com/search"

	// 商品卡片容器（带 hash，前缀匹配）
	itemCardSel = `[class*="Card--doubleCardWrapper"], [class*="doubleCard"], .item.J_MouserOnverReq`
	// 卡片内部：标题 / 价格 / 店铺 / 地址 / 图片
	titleSel    = `[class*="title"], .title .J_ClickStat, .row-title a`
	priceSel    = `[class*="priceInt"], .price .J_Price, strong`
	dealSel     = `[class*="realSales"], .deal-cnt`
	shopSel     = `[class*="shopNameText"], .shopname, .shop .J_ShopInfo`
	locationSel = `[class*="procity"], .location`
	linkSel     = `a`
	imgSel      = `img`
)

// SearchResult 搜索结果。
type SearchResult struct {
	Keyword string `json:"keyword"`
	Items   []Item `json:"items"`
	Count   int    `json:"count"`
	PageURL string `json:"page_url"` // 实际访问到的 URL（含 cookie 带的跳转结果）
}

// Search 执行关键词搜索，返回前 limit 条（0 表示返回本页全部去重后）。
//
// 过滤策略：
//   - 跳过 simba 广告位（URL 包含 click.simba.taobao.com）
//   - 按 item ID（若能解析）或 URL 去重
//
// 为了跳过首屏全广告的情况，加载后会向下滚动 2 次让"自然结果"出现。
func (a *SearchAction) Search(ctx context.Context, keyword string, limit int) (*SearchResult, error) {
	if strings.TrimSpace(keyword) == "" {
		return nil, errors.New("keyword is empty")
	}

	pp := a.page.Context(ctx)

	u := fmt.Sprintf("%s?q=%s", searchBaseURL, url.QueryEscape(keyword))
	logrus.Infof("[taobao] search: navigating to %s", u)
	pp.MustNavigate(u).MustWaitLoad()
	time.Sleep(3 * time.Second)

	// 滚动加载更多（触发 lazy-load 的自然结果）。
	scrollToLoadMore(pp, 2)

	info, err := pp.Info()
	if err != nil {
		return nil, errors.Wrap(err, "page info")
	}
	if strings.Contains(info.URL, loginDomainFragment) {
		return nil, errors.Errorf("redirected to login: cookie expired? url=%s", info.URL)
	}

	cards, err := pp.Elements(itemCardSel)
	if err != nil {
		return nil, errors.Wrap(err, "find item cards")
	}
	logrus.Infof("[taobao] search: found %d card(s)", len(cards))

	var (
		items   = make([]Item, 0, len(cards))
		seen    = make(map[string]struct{}, len(cards))
		skipAd  int
		skipDup int
	)
	for i, card := range cards {
		if limit > 0 && len(items) >= limit {
			break
		}
		// 卡片内任何链接命中广告 → 整卡跳过（广告卡片会同时含 simba 主链接 + amos 客服链接）
		if cardIsAd(card) {
			skipAd++
			continue
		}
		item := extractItem(card)
		if item.Title == "" {
			continue
		}
		key := item.ID
		if key == "" {
			key = item.URL
		}
		if key == "" {
			key = item.Title + "|" + item.Price
		}
		if _, dup := seen[key]; dup {
			skipDup++
			continue
		}
		seen[key] = struct{}{}
		items = append(items, item)
		logrus.Debugf("[taobao] kept #%d: %s", i, item.Title)
	}
	logrus.Infof("[taobao] search: kept=%d, skip_ad=%d, skip_dup=%d", len(items), skipAd, skipDup)

	return &SearchResult{
		Keyword: keyword,
		Items:   items,
		Count:   len(items),
		PageURL: info.URL,
	}, nil
}

// scrollToLoadMore 触发 N 次"滚到底"，每次间隔 1s 让新内容加载。
func scrollToLoadMore(page *rod.Page, rounds int) {
	for i := 0; i < rounds; i++ {
		_, err := page.Eval(`() => window.scrollTo(0, document.body.scrollHeight)`)
		if err != nil {
			logrus.Debugf("[taobao] scroll err: %v", err)
			return
		}
		time.Sleep(1 * time.Second)
	}
}

// isAdURL 判断商品 URL 是否为 simba/联盟广告跳转（非真实详情页）。
func isAdURL(u string) bool {
	return strings.Contains(u, "click.simba.taobao.com") ||
		strings.Contains(u, "x.alimama.com") ||
		strings.Contains(u, "alimama.com/common/code")
}

// extractItem 从单张卡片 DOM 节点提取商品字段。
// 允许任意字段抓不到（返回 ""），不 fail —— 淘宝卡片多模板，字段缺失很常见。
func extractItem(card *rod.Element) Item {
	var item Item

	item.Title = textOf(card, titleSel)
	item.Price = textOf(card, priceSel)
	item.DealCount = textOf(card, dealSel)
	item.Shop = textOf(card, shopSel)
	item.Location = textOf(card, locationSel)

	item.URL = pickItemLink(card)
	if img, err := card.Element(imgSel); err == nil && img != nil {
		if src, _ := img.Attribute("src"); src != nil && *src != "" {
			item.ImageURL = normalizeURL(*src)
		} else if src, _ := img.Attribute("data-src"); src != nil {
			item.ImageURL = normalizeURL(*src)
		}
	}
	// 从 URL 里挖 item id
	if item.URL != "" {
		item.ID = extractItemID(item.URL)
	}
	return item
}

// cardIsAd 卡片内任一链接命中广告域即判定为广告卡。
func cardIsAd(card *rod.Element) bool {
	links, err := card.Elements("a")
	if err != nil {
		return false
	}
	for _, l := range links {
		href, _ := l.Attribute("href")
		if href != nil && isAdURL(*href) {
			return true
		}
	}
	return false
}

// pickItemLink 从卡片所有 <a> 中挑商品详情链接。
// 优先级：item.taobao.com / detail.tmall.com > chaoshi.detail.tmall.com
// > 店铺主页 view_shop.htm（作为 fallback）。
// 过滤：simba 广告、amos 客服、登录页、javascript:void(0) 等。
func pickItemLink(card *rod.Element) string {
	links, err := card.Elements("a")
	if err != nil {
		return ""
	}
	var shopFallback string
	for _, l := range links {
		hrefAttr, _ := l.Attribute("href")
		if hrefAttr == nil || *hrefAttr == "" {
			continue
		}
		h := normalizeURL(*hrefAttr)
		if isLinkNoise(h) {
			continue
		}
		if strings.Contains(h, "item.taobao.com/item.htm") ||
			strings.Contains(h, "detail.tmall.com/item.htm") ||
			strings.Contains(h, "chaoshi.detail.tmall.com") {
			return h
		}
		if shopFallback == "" && strings.Contains(h, "view_shop.htm") {
			shopFallback = h
		}
	}
	return shopFallback
}

// isLinkNoise 判断链接是否为我们不关心的噪声（广告 / 客服 / 登录 / 空协议）。
func isLinkNoise(u string) bool {
	if u == "" {
		return true
	}
	if strings.HasPrefix(u, "javascript:") || strings.HasPrefix(u, "#") {
		return true
	}
	return isAdURL(u) ||
		strings.Contains(u, "amos.alicdn.com") ||
		strings.Contains(u, "login.taobao.com") ||
		strings.Contains(u, "login.tmall.com")
}

// textOf 取第一个命中选择器的元素 textContent；无结果返回 ""。
func textOf(root *rod.Element, sel string) string {
	el, err := root.Element(sel)
	if err != nil || el == nil {
		return ""
	}
	t, err := el.Text()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(t)
}

// normalizeURL 把 "//..." 补成 "https://..."；相对路径补上域名。
func normalizeURL(u string) string {
	u = strings.TrimSpace(u)
	switch {
	case strings.HasPrefix(u, "//"):
		return "https:" + u
	case strings.HasPrefix(u, "/"):
		return "https://s.taobao.com" + u
	}
	return u
}

// extractItemID 从商品 URL 里挖 nid / id 参数，失败返回 ""。
func extractItemID(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	q := parsed.Query()
	for _, k := range []string{"id", "nid", "item_id"} {
		if v := q.Get(k); v != "" {
			return v
		}
	}
	return ""
}
