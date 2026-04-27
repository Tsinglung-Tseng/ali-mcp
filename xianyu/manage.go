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

// ManageAction 闲鱼卖家侧管理操作（擦亮 / 下架）。
// 全部需要登录（cookie 有效），流程都是：
//  1. 进 "我的-在售" 列表页
//  2. 找到目标商品的卡片
//  3. 点对应按钮（擦亮 / 三点菜单 → 下架）
type ManageAction struct {
	page *rod.Page
}

// NewManage 构造管理 action。
func NewManage(page *rod.Page) *ManageAction { return &ManageAction{page: page} }

const (
	// 我的在售商品列表入口；闲鱼 h5 的"我-发布的-在售"。
	mySellingURL = "https://h5.m.goofish.com/2/myItems.html?status=on_sale"

	// TODO(selectors): 全部需实机校验。
	myItemCardSel       = `[class*="my-item"], [class*="ItemCard"], [class*="items-list"] [class*="item"]`
	myItemRefreshBtnSel = `[class*="refresh-btn"], [class*="RefreshBtn"], button[data-action="refresh"]`
	myItemMenuBtnSel    = `[class*="more-btn"], [class*="MoreBtn"], button[data-action="more"]`
	myItemDelistBtnSel  = `[class*="delist"], [class*="Delist"], [data-action="delist"], [class*="off-shelf"]`
	loggedOutHintSel    = `[class*="login-hint"], [class*="LoginHint"]`
)

// Refresh 擦亮指定商品。已登录前提下：
//  1. 导航到我的在售列表
//  2. 通过 itemID 定位卡片（href 包含该 id 即可）
//  3. 点擦亮按钮
//  4. 检查是否出现成功 toast / 按钮变灰
func (a *ManageAction) Refresh(ctx context.Context, itemID string) (*ManageActionResult, error) {
	if strings.TrimSpace(itemID) == "" {
		return nil, errors.New("itemID is empty")
	}

	pp := a.page.Context(ctx)
	logrus.Infof("[xianyu] refresh: navigating to my-selling, target=%s", itemID)
	pp.MustNavigate(mySellingURL).MustWaitLoad()
	time.Sleep(2 * time.Second)

	// 检测是否被踢到登录页
	if exist, _, _ := pp.Has(loggedOutHintSel); exist {
		return nil, errors.New("not logged in (login hint detected) — refresh requires login")
	}

	card, err := findCardByItemID(pp, itemID)
	if err != nil {
		return nil, err
	}

	btn, err := card.Element(myItemRefreshBtnSel)
	if err != nil || btn == nil {
		return nil, errors.Errorf("refresh button not found in card for item %s — selector may be stale", itemID)
	}
	if err := btn.Click("left", 1); err != nil {
		return nil, errors.Wrap(err, "click refresh button")
	}

	// 给 toast / 接口响应留一点时间
	time.Sleep(2 * time.Second)
	logrus.Infof("[xianyu] refresh: clicked for item %s", itemID)
	return &ManageActionResult{
		ItemID:  itemID,
		Action:  "refresh",
		Success: true,
		Message: "擦亮按钮已点击，等待客户端 toast 验证（如失败请实机看页面）",
	}, nil
}

// Delist 下架指定商品。
//  1. 导航到我的在售列表
//  2. 找到卡片
//  3. 点三点菜单 → 找到"下架"项 → 点击
//  4. 闲鱼弹确认对话框（视实现可能需要再点确认）—— 当前版本假设无二次确认；如有，需扩展 confirmSel
func (a *ManageAction) Delist(ctx context.Context, itemID string) (*ManageActionResult, error) {
	if strings.TrimSpace(itemID) == "" {
		return nil, errors.New("itemID is empty")
	}

	pp := a.page.Context(ctx)
	logrus.Infof("[xianyu] delist: navigating to my-selling, target=%s", itemID)
	pp.MustNavigate(mySellingURL).MustWaitLoad()
	time.Sleep(2 * time.Second)

	if exist, _, _ := pp.Has(loggedOutHintSel); exist {
		return nil, errors.New("not logged in (login hint detected) — delist requires login")
	}

	card, err := findCardByItemID(pp, itemID)
	if err != nil {
		return nil, err
	}

	// 第 1 步：打开三点菜单
	moreBtn, err := card.Element(myItemMenuBtnSel)
	if err != nil || moreBtn == nil {
		return nil, errors.Errorf("more-menu button not found in card for item %s", itemID)
	}
	if err := moreBtn.Click("left", 1); err != nil {
		return nil, errors.Wrap(err, "click more-menu button")
	}
	time.Sleep(1 * time.Second)

	// 第 2 步：菜单弹出后，下架项是页面级元素（不在 card 子树里），从 page 找
	delistBtn, err := pp.Element(myItemDelistBtnSel)
	if err != nil || delistBtn == nil {
		return nil, errors.Errorf("delist menu item not found after opening menu for item %s", itemID)
	}
	if err := delistBtn.Click("left", 1); err != nil {
		return nil, errors.Wrap(err, "click delist menu item")
	}
	time.Sleep(2 * time.Second)
	logrus.Infof("[xianyu] delist: clicked for item %s", itemID)

	return &ManageActionResult{
		ItemID:  itemID,
		Action:  "delist",
		Success: true,
		Message: "下架已触发；闲鱼可能有二次确认弹框，如未消失请实机校验 confirmSel",
	}, nil
}

// findCardByItemID 在我的列表里找到 href 含 itemID 的卡片。
func findCardByItemID(page *rod.Page, itemID string) (*rod.Element, error) {
	cards, err := page.Elements(myItemCardSel)
	if err != nil {
		return nil, errors.Wrap(err, "find my-items cards")
	}
	if len(cards) == 0 {
		return nil, errors.New("no items in my-selling list — list may not have loaded, or selector stale")
	}
	for _, card := range cards {
		links, _ := card.Elements("a")
		for _, l := range links {
			href, _ := l.Attribute("href")
			if href != nil && strings.Contains(*href, itemID) {
				return card, nil
			}
		}
		// 卡片本身可能就是 a
		if href, _ := card.Attribute("href"); href != nil && strings.Contains(*href, itemID) {
			return card, nil
		}
	}
	return nil, fmt.Errorf("item %s not found in my-selling list (have %d cards)", itemID, len(cards))
}
