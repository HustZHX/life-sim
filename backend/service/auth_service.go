package service

import (
	"errors"
	"fmt"
	"time"

	"life-sim/backend/auth"
	"life-sim/backend/config"
	"life-sim/backend/model"
	"life-sim/backend/store"
)

type AuthService struct {
	store  *store.Store
	cfg    *config.AuthConfig
	tokens *auth.TokenIssuer
	limit  *auth.RateLimiter
}

func NewAuthService(st *store.Store, cfg *config.AuthConfig) *AuthService {
	var tokens *auth.TokenIssuer
	if cfg.Enabled && cfg.JWTSecret != "" {
		tokens = auth.NewTokenIssuer(cfg.JWTSecret)
	}
	limit := auth.NewRateLimiter(cfg.RateLimit, time.Minute)
	return &AuthService{store: st, cfg: cfg, tokens: tokens, limit: limit}
}

func (s *AuthService) CookieConfig() auth.CookieConfig {
	return auth.CookieConfig{Secure: s.cfg.CookieSecure}
}

func (s *AuthService) VerifyGate(code string) (string, error) {
	if !s.cfg.Enabled {
		return "", nil
	}
	if code != s.cfg.SiteAccessCode {
		return "", fmt.Errorf("进门密码错误")
	}
	return s.tokens.Sign(auth.TokenKindGate, "", "", s.cfg.GateTTL)
}

func (s *AuthService) Login(username, password string) (access, refresh string, user *model.User, err error) {
	if !s.cfg.Enabled {
		return "", "", nil, fmt.Errorf("鉴权未启用")
	}
	u, err := s.store.GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			return "", "", nil, fmt.Errorf("用户名或密码错误")
		}
		return "", "", nil, err
	}
	if !auth.CheckPassword(u.PasswordHash, password) {
		return "", "", nil, fmt.Errorf("用户名或密码错误")
	}
	access, err = s.tokens.Sign(auth.TokenKindAccess, u.ID, u.Username, s.cfg.AccessTTL)
	if err != nil {
		return "", "", nil, err
	}
	refresh, err = s.tokens.Sign(auth.TokenKindRefresh, u.ID, u.Username, s.cfg.RefreshTTL)
	if err != nil {
		return "", "", nil, err
	}
	return access, refresh, u, nil
}

func (s *AuthService) Refresh(refreshToken string) (access, refresh string, err error) {
	if !s.cfg.Enabled {
		return "", "", fmt.Errorf("鉴权未启用")
	}
	claims, err := s.tokens.Verify(refreshToken, auth.TokenKindRefresh)
	if err != nil {
		if errors.Is(err, auth.ErrTokenExpired) {
			return "", "", fmt.Errorf("token_expired")
		}
		return "", "", fmt.Errorf("login_required")
	}
	u, err := s.store.GetUserByID(claims.UserID)
	if err != nil {
		return "", "", fmt.Errorf("login_required")
	}
	access, err = s.tokens.Sign(auth.TokenKindAccess, u.ID, u.Username, s.cfg.AccessTTL)
	if err != nil {
		return "", "", err
	}
	refresh, err = s.tokens.Sign(auth.TokenKindRefresh, u.ID, u.Username, s.cfg.RefreshTTL)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func (s *AuthService) GetMe(userID string) (*model.AuthUserResponse, error) {
	u, err := s.store.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	return &model.AuthUserResponse{
		Username:    u.Username,
		DisplayName: u.DisplayName,
	}, nil
}

func (s *AuthService) ChangePassword(userID, oldPassword, newPassword string) error {
	u, err := s.store.GetUserByID(userID)
	if err != nil {
		return err
	}
	if !auth.CheckPassword(u.PasswordHash, oldPassword) {
		return fmt.Errorf("原密码错误")
	}
	if len(newPassword) < 6 {
		return fmt.Errorf("新密码至少 6 位")
	}
	return s.store.UpdateUserPassword(userID, newPassword)
}

func (s *AuthService) AllowRate(key string) bool {
	if s.cfg.RateLimit <= 0 {
		return true
	}
	return s.limit.Allow(key)
}

func (s *AuthService) Tokens() *auth.TokenIssuer {
	return s.tokens
}

func (s *AuthService) Config() *config.AuthConfig {
	return s.cfg
}
