package taobao

import (
	"context"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// DetailAction 淘宝商品详情。
// 入参为商品 URL（item.taobao.com / detail.tmall.com 都可），需已登录 cookie。
type DetailAction struct {
	page *rod.Page
}

// NewDetail 构造详情 action。
func NewDetail(page *rod.Page) *DetailAction { return &DetailAction{page: page} }

// 详情页选择器 —— 与搜索页一样，类名带 React hash，前缀匹配。
// TODO(selectors): 实机校验后更新。
const (
	detailTitleSel    = `[class*="ItemTitle"], h3.tb-main-title, [data-spm="1000983"] h3, .tb-main-title`
	detailPriceSel    = `[class*="priceText"], [class*="priceInt"], .tb-rmb-num, strong.tb-rmb-num`
	detailSalesSel    = `[class*="HeaderTip"] [class*="text"], #J_DealCount, .tb-sell-counter strong`
	detailShopSel     = `[class*="shopName"], #J_TShopInfo .tb-shop-name, .shop-name a`
	detailLocationSel = `[class*="location"], #J_DeliveryInfo, .tb-deliver`
	detailImgSel      = `[class*="mainPic"] img, #J_ImgBooth`
	detailSkuSel      = `[class*="skuItem"], [class*="ValueItemText"], .J_TSaleProp li`
)

// Detail 详情结果。
type Detail struct {
	URL       string   `json:"url"`             // 实际访问到的 URL（带 cookie 跳转后的）
	ID        string   `json:"id,omitempty"`    // 从 URL 解析的商品 ID
	Title     string   `json:"title"`           // 商品标题
	Price     string   `json:"price,omitempty"` // 当前价（原文）
	Sales     string   `json:"sales,omitempty"` // 销量原文（如 "已售 1.2 万+"）
	Shop      string   `json:"shop,omitempty"`  // 店铺名
	Location  string   `json:"location,omitempty"`
	MainImage string   `json:"main_image,omitempty"`
	SKUs      []string `json:"skus,omitempty"` // SKU 文案列表（颜色/尺寸/版本等）
}

// GetDetail 访问商品详情页并解析关键字段。
func (a *DetailAction) GetDetail(ctx context.Context, itemURL string) (*Detail, error) {
	if strings.TrimSpace(itemURL) == "" {
		return nil, errors.New("item url is empty")
	}
	if !strings.HasPrefix(itemURL, "http") {
		return nil, errors.New("item url must start with http(s)://")
	}

	pp := a.page.Context(ctx)

	logrus.Infof("[taobao] detail: navigating to %s", itemURL)
	pp.MustNavigate(itemURL).MustWaitLoad()
	time.Sleep(3 * time.Second)

	info, err := pp.Info()
	if err != nil {
		return nil, errors.Wrap(err, "page info")
	}
	if strings.Contains(info.URL, loginDomainFragment) {
		return nil, errors.Errorf("redirected to login: cookie expired? url=%s", info.URL)
	}

	d := &Detail{
		URL:       info.URL,
		ID:        extractItemID(info.URL),
		Title:     textOf(pp.MustElement("body"), detailTitleSel),
		Price:     textOf(pp.MustElement("body"), detailPriceSel),
		Sales:     textOf(pp.MustElement("body"), detailSalesSel),
		Shop:      textOf(pp.MustElement("body"), detailShopSel),
		Location:  textOf(pp.MustElement("body"), detailLocationSel),
		MainImage: imgSrcOf(pp, detailImgSel),
		SKUs:      collectTexts(pp, detailSkuSel),
	}
	if d.Title == "" {
		return nil, errors.New("title not found — selector may be stale or page is not a product page")
	}
	logrus.Infof("[taobao] detail: ok, title=%q price=%q skus=%d", d.Title, d.Price, len(d.SKUs))
	return d, nil
}

// imgSrcOf 取首个匹配元素的 src 属性。
func imgSrcOf(page *rod.Page, sel string) string {
	el, err := page.Element(sel)
	if err != nil || el == nil {
		return ""
	}
	src, err := el.Attribute("src")
	if err != nil || src == nil {
		return ""
	}
	return normalizeURL(*src)
}

// collectTexts 取所有匹配元素的 textContent，去重 + 去空。
func collectTexts(page *rod.Page, sel string) []string {
	els, err := page.Elements(sel)
	if err != nil || len(els) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(els))
	out := make([]string, 0, len(els))
	for _, el := range els {
		t, err := el.Text()
		if err != nil {
			continue
		}
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if _, dup := seen[t]; dup {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}
