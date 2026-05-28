package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	CookieGate    = "gate_token"
	CookieAccess  = "access_token"
	CookieRefresh = "refresh_token"
)

type CookieConfig struct {
	Secure bool
}

func SetTokenCookie(c *gin.Context, cfg CookieConfig, name, value string, ttl time.Duration) {
	maxAge := int(ttl.Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(name, value, maxAge, "/", "", cfg.Secure, true)
}

func ClearAuthCookies(c *gin.Context, cfg CookieConfig) {
	for _, name := range []string{CookieGate, CookieAccess, CookieRefresh} {
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie(name, "", -1, "/", "", cfg.Secure, true)
	}
}
