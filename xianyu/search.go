package xianyu

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

// SearchAction 闲鱼商品搜索。
// 走 h5.m.goofish.com（移动站，DOM 简单 + 风控相对宽松）。
// 公开搜索游客态可访问，但翻多页 / 个性化推荐需要登录。
type SearchAction struct {
	page *rod.Page
}

// NewSearch 构造搜索 action。
func NewSearch(page *rod.Page) *SearchAction { return &SearchAction{page: page} }

const (
	searchURLPattern = "https://h5.m.goofish.com/2/search.html?q=%s"

	// 卡片容器及内部字段；闲鱼 h5 也是 React 组件，类名带 hash。
	// TODO(selectors): 实机校验。
	xySearchCardSel   = `[class*="feeds-item"], [class*="card-feed"], .feeds-item-wrap`
	xySearchTitleSel  = `[class*="feeds-item-title"], [class*="card-title"], .item-title`
	xySearchPriceSel  = `[class*="feeds-item-price"], [class*="card-price"], .price`
	xySearchSellerSel = `[class*="seller-nick"], [class*="user-nick"], .nick`
	xySearchLocSel    = `[class*="seller-area"], [class*="location"], .area`
	xySearchWantSel   = `[class*="want-num"], [class*="wanted"], .want`
	xySearchImgSel    = `img`
)

// Search 在闲鱼 h5 站搜索商品。limit=0 返回首屏全部。
func (a *SearchAction) Search(ctx context.Context, keyword string, limit int) (*SearchResult, error) {
	if strings.TrimSpace(keyword) == "" {
		return nil, errors.New("keyword is empty")
	}

	pp := a.page.Context(ctx)
	u := fmt.Sprintf(searchURLPattern, url.QueryEscape(keyword))
	logrus.Infof("[xianyu] search: navigating to %s", u)
	pp.MustNavigate(u).MustWaitLoad()
	time.Sleep(3 * time.Second)

	// 滚动两次让懒加载的卡片都出现
	for i := 0; i < 2; i++ {
		_, _ = pp.Eval(`() => window.scrollTo(0, document.body.scrollHeight)`)
		time.Sleep(1 * time.Second)
	}

	info, _ := pp.Info()
	cards, err := pp.Elements(xySearchCardSel)
	if err != nil {
		return nil, errors.Wrap(err, "find cards")
	}
	logrus.Infof("[xianyu] search: found %d cards", len(cards))

	items := make([]Item, 0, len(cards))
	seen := make(map[string]struct{}, len(cards))
	for _, card := range cards {
		if limit > 0 && len(items) >= limit {
			break
		}
		it := extractSearchItem(card)
		if it.Title == "" {
			continue
		}
		key := it.ID
		if key == "" {
			key = it.URL
		}
		if key == "" {
			key = it.Title + "|" + it.Price
		}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		items = append(items, it)
	}

	return &SearchResult{
		Keyword: keyword,
		Items:   items,
		Count:   len(items),
		PageURL: info.URL,
	}, nil
}

// SearchResult 搜索结果。
type SearchResult struct {
	Keyword string `json:"keyword"`
	Items   []Item `json:"items"`
	Count   int    `json:"count"`
	PageURL string `json:"page_url"`
}

// extractSearchItem 从一张卡片提取商品字段。
func extractSearchItem(card *rod.Element) Item {
	var it Item
	it.Title = textOfEl(card, xySearchTitleSel)
	it.Price = textOfEl(card, xySearchPriceSel)
	it.Seller = textOfEl(card, xySearchSellerSel)
	it.Location = textOfEl(card, xySearchLocSel)
	it.WantNum = textOfEl(card, xySearchWantSel)

	// 链接：取卡片自身或第一个 a
	if href, _ := card.Attribute("href"); href != nil && *href != "" {
		it.URL = normalizeURL(*href)
	} else if a, err := card.Element("a"); err == nil && a != nil {
		if h, _ := a.Attribute("href"); h != nil {
			it.URL = normalizeURL(*h)
		}
	}
	if it.URL != "" {
		it.ID = extractItemID(it.URL)
	}

	if img, err := card.Element(xySearchImgSel); err == nil && img != nil {
		if src, _ := img.Attribute("src"); src != nil && *src != "" {
			it.ImageURL = normalizeURL(*src)
		} else if src, _ := img.Attribute("data-src"); src != nil {
			it.ImageURL = normalizeURL(*src)
		}
	}
	return it
}
