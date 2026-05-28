package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"life-sim/backend/auth"
	"life-sim/backend/model"
	"life-sim/backend/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Status(c *gin.Context) {
	cfg := h.svc.Config()
	resp := model.AuthStatusResponse{Enabled: cfg.Enabled}
	if !cfg.Enabled {
		OK(c, resp)
		return
	}
	if token, err := c.Cookie(auth.CookieGate); err == nil && token != "" {
		if _, err := h.svc.Tokens().Verify(token, auth.TokenKindGate); err == nil {
			resp.GatePassed = true
		}
	}
	if token, err := c.Cookie(auth.CookieAccess); err == nil && token != "" {
		if _, err := h.svc.Tokens().Verify(token, auth.TokenKindAccess); err == nil {
			resp.LoggedIn = true
		}
	}
	OK(c, resp)
}

func (h *AuthHandler) Gate(c *gin.Context) {
	if !h.svc.Config().Enabled {
		OK(c, gin.H{"enabled": false})
		return
	}
	if !h.svc.AllowRate("gate:" + clientIP(c)) {
		Fail(c, http.StatusTooManyRequests, 429, "请求过于频繁，请稍后再试")
		return
	}
	var req model.GateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	token, err := h.svc.VerifyGate(strings.TrimSpace(req.Code))
	if err != nil {
		Fail(c, http.StatusUnauthorized, 401, err.Error())
		return
	}
	auth.SetTokenCookie(c, h.svc.CookieConfig(), auth.CookieGate, token, h.svc.Config().GateTTL)
	OK(c, gin.H{"ok": true})
}

func (h *AuthHandler) Login(c *gin.Context) {
	if !h.svc.Config().Enabled {
		OK(c, gin.H{"enabled": false})
		return
	}
	if !h.svc.AllowRate("login:" + clientIP(c)) {
		Fail(c, http.StatusTooManyRequests, 429, "请求过于频繁，请稍后再试")
		return
	}
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	access, refresh, user, err := h.svc.Login(strings.TrimSpace(req.Username), req.Password)
	if err != nil {
		Fail(c, http.StatusUnauthorized, 401, err.Error())
		return
	}
	cfg := h.svc.Config()
	auth.SetTokenCookie(c, h.svc.CookieConfig(), auth.CookieAccess, access, cfg.AccessTTL)
	auth.SetTokenCookie(c, h.svc.CookieConfig(), auth.CookieRefresh, refresh, cfg.RefreshTTL)
	OK(c, model.AuthUserResponse{
		Username:    user.Username,
		DisplayName: user.DisplayName,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	if !h.svc.Config().Enabled {
		OK(c, gin.H{"enabled": false})
		return
	}
	refreshToken, err := c.Cookie(auth.CookieRefresh)
	if err != nil || refreshToken == "" {
		Fail(c, http.StatusUnauthorized, 401, "login_required")
		return
	}
	access, refresh, err := h.svc.Refresh(refreshToken)
	if err != nil {
		msg := err.Error()
		Fail(c, http.StatusUnauthorized, 401, msg)
		return
	}
	cfg := h.svc.Config()
	auth.SetTokenCookie(c, h.svc.CookieConfig(), auth.CookieAccess, access, cfg.AccessTTL)
	auth.SetTokenCookie(c, h.svc.CookieConfig(), auth.CookieRefresh, refresh, cfg.RefreshTTL)
	OK(c, gin.H{"ok": true})
}

func (h *AuthHandler) Me(c *gin.Context) {
	u, ok := auth.UserFromContext(c)
	if !ok || u.ID == "" {
		Fail(c, http.StatusUnauthorized, 401, "login_required")
		return
	}
	me, err := h.svc.GetMe(u.ID)
	if err != nil {
		Fail(c, http.StatusUnauthorized, 401, "login_required")
		return
	}
	OK(c, me)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	auth.ClearAuthCookies(c, h.svc.CookieConfig())
	OK(c, gin.H{"ok": true})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	u, ok := auth.UserFromContext(c)
	if !ok || u.ID == "" {
		Fail(c, http.StatusUnauthorized, 401, "login_required")
		return
	}
	var req model.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	if err := h.svc.ChangePassword(u.ID, req.OldPassword, req.NewPassword); err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	OK(c, gin.H{"ok": true})
}

func clientIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}
	return c.ClientIP()
}
