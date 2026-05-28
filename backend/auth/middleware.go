package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"life-sim/backend/config"
)

const ContextUserKey = "authUser"

type UserContext struct {
	ID          string
	Username    string
	DisplayName string
}

type Middleware struct {
	cfg    *config.AuthConfig
	tokens *TokenIssuer
}

func NewMiddleware(cfg *config.AuthConfig) *Middleware {
	var tokens *TokenIssuer
	if cfg.Enabled && cfg.JWTSecret != "" {
		tokens = NewTokenIssuer(cfg.JWTSecret)
	}
	return &Middleware{cfg: cfg, tokens: tokens}
}

func abortUnauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": msg})
	c.Abort()
}

func (m *Middleware) GateRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.cfg.Enabled {
			c.Next()
			return
		}
		token, err := c.Cookie(CookieGate)
		if err != nil || token == "" {
			abortUnauthorized(c, "gate_required")
			return
		}
		if _, err := m.tokens.Verify(token, TokenKindGate); err != nil {
			abortUnauthorized(c, "gate_required")
			return
		}
		c.Next()
	}
}

func (m *Middleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.cfg.Enabled {
			c.Next()
			return
		}
		token, err := c.Cookie(CookieAccess)
		if err != nil || token == "" {
			abortUnauthorized(c, "login_required")
			return
		}
		claims, err := m.tokens.Verify(token, TokenKindAccess)
		if err != nil {
			msg := "login_required"
			if err == ErrTokenExpired {
				msg = "token_expired"
			}
			abortUnauthorized(c, msg)
			return
		}
		c.Set(ContextUserKey, UserContext{
			ID:       claims.UserID,
			Username: claims.Username,
		})
		c.Next()
	}
}

func UserFromContext(c *gin.Context) (UserContext, bool) {
	v, ok := c.Get(ContextUserKey)
	if !ok {
		return UserContext{}, false
	}
	u, ok := v.(UserContext)
	return u, ok
}
