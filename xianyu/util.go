package xianyu

import (
	"net/url"
	"strings"

	"github.com/go-rod/rod"
)

// textOfEl 在元素子树里取首个匹配选择器的 textContent。
// 抓不到返回 ""。
func textOfEl(root *rod.Element, sel string) string {
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

// imgSrcOfEl 取元素子树里首个匹配元素的 src。
func imgSrcOfEl(root *rod.Element, sel string) string {
	el, err := root.Element(sel)
	if err != nil || el == nil {
		return ""
	}
	src, err := el.Attribute("src")
	if err != nil || src == nil {
		return ""
	}
	return normalizeURL(*src)
}

// collectImgSrcs 取页面所有匹配的 img 元素 src。
func collectImgSrcs(page *rod.Page, sel string) []string {
	els, err := page.Elements(sel)
	if err != nil || len(els) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(els))
	out := make([]string, 0, len(els))
	for _, el := range els {
		src, err := el.Attribute("src")
		if err != nil || src == nil || *src == "" {
			// 闲鱼有时用 data-src 做 lazy load
			if d, _ := el.Attribute("data-src"); d != nil && *d != "" {
				src = d
			} else {
				continue
			}
		}
		u := normalizeURL(*src)
		if _, dup := seen[u]; dup {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}

// normalizeURL 处理协议相对 URL 与相对路径。
func normalizeURL(u string) string {
	u = strings.TrimSpace(u)
	switch {
	case strings.HasPrefix(u, "//"):
		return "https:" + u
	case strings.HasPrefix(u, "/"):
		return "https://h5.m.goofish.com" + u
	}
	return u
}

// extractItemID 从闲鱼商品 URL 解析 id 参数。
func extractItemID(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	for _, k := range []string{"id", "itemId", "item_id"} {
		if v := parsed.Query().Get(k); v != "" {
			return v
		}
	}
	return ""
}
