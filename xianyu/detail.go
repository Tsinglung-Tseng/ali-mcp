package xianyu

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// DetailAction 闲鱼商品详情页操作。
type DetailAction struct {
	page *rod.Page
}

// NewDetail 构造详情 action。
func NewDetail(page *rod.Page) *DetailAction { return &DetailAction{page: page} }

const (
	itemDetailURLPattern = "https://h5.m.goofish.com/item.html?id=%s"

	// TODO(selectors): 实机校验。
	xyDetailTitleSel    = `[class*="item-title"], [class*="ItemTitle"], h1`
	xyDetailPriceSel    = `[class*="item-price"], [class*="ItemPrice"]:not([class*="origin"]), .price-current`
	xyDetailOriginSel   = `[class*="origin-price"], [class*="OriginPrice"], .price-origin, del`
	xyDetailDescSel     = `[class*="item-desc"], [class*="ItemDesc"], .desc`
	xyDetailSellerSel   = `[class*="seller-nick"], [class*="user-nick"], .seller-name`
	xyDetailLocationSel = `[class*="seller-area"], [class*="location"]`
	xyDetailImgsSel     = `[class*="item-image"] img, [class*="banner"] img, .item-pic img`
	xyDetailWantSel     = `[class*="want-num"]`
	xyDetailViewSel     = `[class*="view-num"]`
	xyDetailPostedSel   = `[class*="posted-at"], [class*="time"]`
)

// GetDetail 访问商品详情页并解析。itemID 或完整 URL 都接受。
func (a *DetailAction) GetDetail(ctx context.Context, idOrURL string) (*Detail, error) {
	if strings.TrimSpace(idOrURL) == "" {
		return nil, errors.New("item id or url is empty")
	}
	target := idOrURL
	if !strings.HasPrefix(target, "http") {
		target = fmt.Sprintf(itemDetailURLPattern, target)
	}

	pp := a.page.Context(ctx)
	logrus.Infof("[xianyu] detail: navigating to %s", target)
	pp.MustNavigate(target).MustWaitLoad()
	time.Sleep(3 * time.Second)

	info, err := pp.Info()
	if err != nil {
		return nil, errors.Wrap(err, "page info")
	}

	body := pp.MustElement("body")
	d := &Detail{
		URL:         info.URL,
		ID:          extractItemID(info.URL),
		Title:       textOfEl(body, xyDetailTitleSel),
		Price:       textOfEl(body, xyDetailPriceSel),
		OriginPrice: textOfEl(body, xyDetailOriginSel),
		Description: textOfEl(body, xyDetailDescSel),
		Seller:      textOfEl(body, xyDetailSellerSel),
		Location:    textOfEl(body, xyDetailLocationSel),
		WantNum:     textOfEl(body, xyDetailWantSel),
		ViewNum:     textOfEl(body, xyDetailViewSel),
		PostedAt:    textOfEl(body, xyDetailPostedSel),
		Images:      collectImgSrcs(pp, xyDetailImgsSel),
	}
	if d.Title == "" {
		return nil, errors.Errorf("title not found at %s — page may not be a valid item page or selector is stale", info.URL)
	}
	logrus.Infof("[xianyu] detail: ok, title=%q price=%q imgs=%d", d.Title, d.Price, len(d.Images))
	return d, nil
}
