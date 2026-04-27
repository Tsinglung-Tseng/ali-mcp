package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/Tsinglung-Tseng/ali.mcp/configs"
	"github.com/Tsinglung-Tseng/ali.mcp/cookies"
)

func parseInt(s string) (int, error) { return strconv.Atoi(s) }

// respondError 返回错误响应
func respondError(c *gin.Context, statusCode int, code, message string, details any) {
	logrus.Errorf("%s %s %d %s", c.Request.Method, c.Request.URL.Path, statusCode, code)
	c.JSON(statusCode, ErrorResponse{Error: message, Code: code, Details: details})
}

// respondSuccess 返回成功响应
func respondSuccess(c *gin.Context, data any, message string) {
	c.JSON(http.StatusOK, SuccessResponse{Success: true, Data: data, Message: message})
}

// ---------------- 淘宝 ----------------

func (s *AppServer) taobaoLoginStatusHandler(c *gin.Context) {
	result, err := s.taobao.CheckLoginStatus(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "STATUS_CHECK_FAILED", "检查淘宝登录状态失败", err.Error())
		return
	}
	respondSuccess(c, result, "ok")
}

func (s *AppServer) taobaoLoginQrcodeHandler(c *gin.Context) {
	result, err := s.taobao.GetLoginQrcode(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "QRCODE_FAILED", "获取淘宝登录二维码失败", err.Error())
		return
	}
	respondSuccess(c, result, "ok")
}

func (s *AppServer) taobaoSearchHandler(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		respondError(c, http.StatusBadRequest, "MISSING_KEYWORD", "缺少 q 参数", nil)
		return
	}
	limit := 20
	if v := c.Query("limit"); v != "" {
		if n, err := parseInt(v); err == nil && n > 0 {
			limit = n
		}
	}
	result, err := s.taobao.Search(c.Request.Context(), keyword, limit)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "SEARCH_FAILED", "淘宝搜索失败", err.Error())
		return
	}
	respondSuccess(c, result, "ok")
}

func (s *AppServer) taobaoItemDetailHandler(c *gin.Context) {
	itemURL := c.Query("url")
	if itemURL == "" {
		respondError(c, http.StatusBadRequest, "MISSING_URL", "缺少 url 参数", nil)
		return
	}
	d, err := s.taobao.GetItemDetail(c.Request.Context(), itemURL)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "DETAIL_FAILED", "淘宝商品详情失败", err.Error())
		return
	}
	respondSuccess(c, d, "ok")
}

func (s *AppServer) taobaoDeleteCookiesHandler(c *gin.Context) {
	if err := s.taobao.DeleteCookies(c.Request.Context()); err != nil {
		respondError(c, http.StatusInternalServerError, "DELETE_COOKIES_FAILED", "删除淘宝 cookies 失败", err.Error())
		return
	}
	respondSuccess(c, DeleteCookiesResponse{
		Platform:   string(configs.PlatformTaobao),
		CookiePath: cookies.GetCookiesFilePath(configs.PlatformTaobao),
		Message:    "淘宝 cookies 已删除，下次操作需重新登录",
	}, "ok")
}

// ---------------- 闲鱼 ----------------

func (s *AppServer) xianyuLoginStatusHandler(c *gin.Context) {
	result, err := s.xianyu.CheckLoginStatus(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "STATUS_CHECK_FAILED", "检查闲鱼登录状态失败", err.Error())
		return
	}
	respondSuccess(c, result, "ok")
}

func (s *AppServer) xianyuSearchHandler(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		respondError(c, http.StatusBadRequest, "MISSING_KEYWORD", "缺少 q 参数", nil)
		return
	}
	limit := 20
	if v := c.Query("limit"); v != "" {
		if n, err := parseInt(v); err == nil && n > 0 {
			limit = n
		}
	}
	r, err := s.xianyu.Search(c.Request.Context(), keyword, limit)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "SEARCH_FAILED", "闲鱼搜索失败", err.Error())
		return
	}
	respondSuccess(c, r, "ok")
}

func (s *AppServer) xianyuItemDetailHandler(c *gin.Context) {
	idOrURL := c.Query("id_or_url")
	if idOrURL == "" {
		idOrURL = c.Query("id")
	}
	if idOrURL == "" {
		respondError(c, http.StatusBadRequest, "MISSING_ID", "缺少 id_or_url 参数", nil)
		return
	}
	d, err := s.xianyu.GetItemDetail(c.Request.Context(), idOrURL)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "DETAIL_FAILED", "闲鱼商品详情失败", err.Error())
		return
	}
	respondSuccess(c, d, "ok")
}

func (s *AppServer) xianyuUserProfileHandler(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		respondError(c, http.StatusBadRequest, "MISSING_USER_ID", "缺少 user_id 参数", nil)
		return
	}
	p, err := s.xianyu.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "PROFILE_FAILED", "闲鱼用户主页失败", err.Error())
		return
	}
	respondSuccess(c, p, "ok")
}

func (s *AppServer) xianyuRefreshHandler(c *gin.Context) {
	itemID := c.Query("item_id")
	if itemID == "" {
		respondError(c, http.StatusBadRequest, "MISSING_ITEM_ID", "缺少 item_id 参数", nil)
		return
	}
	r, err := s.xianyu.RefreshItem(c.Request.Context(), itemID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "REFRESH_FAILED", "闲鱼擦亮失败", err.Error())
		return
	}
	respondSuccess(c, r, "ok")
}

func (s *AppServer) xianyuDelistHandler(c *gin.Context) {
	itemID := c.Query("item_id")
	if itemID == "" {
		respondError(c, http.StatusBadRequest, "MISSING_ITEM_ID", "缺少 item_id 参数", nil)
		return
	}
	r, err := s.xianyu.DelistItem(c.Request.Context(), itemID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "DELIST_FAILED", "闲鱼下架失败", err.Error())
		return
	}
	respondSuccess(c, r, "ok")
}

func (s *AppServer) xianyuDeleteCookiesHandler(c *gin.Context) {
	if err := s.xianyu.DeleteCookies(c.Request.Context()); err != nil {
		respondError(c, http.StatusInternalServerError, "DELETE_COOKIES_FAILED", "删除闲鱼 cookies 失败", err.Error())
		return
	}
	respondSuccess(c, DeleteCookiesResponse{
		Platform:   string(configs.PlatformXianyu),
		CookiePath: cookies.GetCookiesFilePath(configs.PlatformXianyu),
		Message:    "闲鱼 cookies 已删除",
	}, "ok")
}

// ---------------- 通用 ----------------

func healthHandler(c *gin.Context) {
	respondSuccess(c, map[string]any{
		"status":  "healthy",
		"service": "ali-mcp",
	}, "服务正常")
}
