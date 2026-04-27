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

// UserAction 闲鱼用户主页操作（卖家侧只读视图）。
type UserAction struct {
	page *rod.Page
}

// NewUser 构造用户 action。
func NewUser(page *rod.Page) *UserAction { return &UserAction{page: page} }

const (
	userProfileURLPattern = "https://h5.m.goofish.com/personal.html?userId=%s"

	// TODO(selectors): 实机校验。
	xyUserNickSel     = `[class*="user-nick"], [class*="user-name"], .nickname`
	xyUserAvatarSel   = `[class*="user-avatar"] img, [class*="avatar"] img`
	xyUserIntroSel    = `[class*="user-intro"], [class*="user-desc"]`
	xyUserCreditSel   = `[class*="credit"], [class*="zhima"]`
	xyUserSellingSel  = `[class*="selling-count"]`
	xyUserSoldSel     = `[class*="sold-count"]`
	xyUserFollowerSel = `[class*="follower"], [class*="fans"]`
)

// GetProfile 访问卖家主页并解析摘要 + 最近商品。
func (a *UserAction) GetProfile(ctx context.Context, userID string) (*UserProfile, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("userID is empty")
	}

	pp := a.page.Context(ctx)
	target := fmt.Sprintf(userProfileURLPattern, userID)
	logrus.Infof("[xianyu] user-profile: navigating to %s", target)
	pp.MustNavigate(target).MustWaitLoad()
	time.Sleep(3 * time.Second)

	// 滚动一次让"在售商品"列表加载
	_, _ = pp.Eval(`() => window.scrollTo(0, document.body.scrollHeight)`)
	time.Sleep(1 * time.Second)

	info, _ := pp.Info()
	body := pp.MustElement("body")

	p := &UserProfile{
		URL:          info.URL,
		UserID:       userID,
		Nickname:     textOfEl(body, xyUserNickSel),
		AvatarURL:    imgSrcOfEl(body, xyUserAvatarSel),
		Intro:        textOfEl(body, xyUserIntroSel),
		CreditScore:  textOfEl(body, xyUserCreditSel),
		SellingCount: textOfEl(body, xyUserSellingSel),
		SoldCount:    textOfEl(body, xyUserSoldSel),
		FollowerNum:  textOfEl(body, xyUserFollowerSel),
	}
	if p.Nickname == "" {
		return nil, errors.Errorf("nickname not found at %s — profile selector may be stale or user does not exist", info.URL)
	}

	// 用户主页的商品卡片选择器与搜索页基本一致
	cards, err := pp.Elements(xySearchCardSel)
	if err == nil {
		for i, card := range cards {
			if i >= 20 { // 主页只取前 20
				break
			}
			it := extractSearchItem(card)
			if it.Title == "" {
				continue
			}
			it.Seller = p.Nickname // 主页所有商品都是这个卖家
			p.RecentItems = append(p.RecentItems, it)
		}
	}
	logrus.Infof("[xianyu] user-profile: ok, nick=%q items=%d", p.Nickname, len(p.RecentItems))
	return p, nil
}
