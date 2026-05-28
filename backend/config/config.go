package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server            ServerConfig
	DeepSeek          DeepSeekConfig
	Auth              AuthConfig
	DatabasePath      string
	TimelineStepYears int
	CORSOrigins       []string
}

type ServerConfig struct {
	Port int
}

type DeepSeekConfig struct {
	APIKey     string
	BaseURL    string
	ModelFast  string
	ModelHeavy string
}

type AuthConfig struct {
	Enabled        bool
	JWTSecret      string
	SiteAccessCode string
	BootstrapUsers []BootstrapUser
	AccessTTL      time.Duration
	RefreshTTL     time.Duration
	GateTTL        time.Duration
	RateLimit      int
	CookieSecure   bool
}

type BootstrapUser struct {
	Username string
	Password string
}

func LoadFromEnv() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{Port: 8080},
		DeepSeek: DeepSeekConfig{
			BaseURL:    "https://api.deepseek.com",
			ModelFast:  "deepseek-v4-flash",
			ModelHeavy: "deepseek-v4-pro",
		},
		DatabasePath:      "./data/lifesim.db",
		TimelineStepYears: 3,
		CORSOrigins:       []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		Auth: AuthConfig{
			AccessTTL:  2 * time.Hour,
			RefreshTTL: 720 * time.Hour,
			GateTTL:    24 * time.Hour,
			RateLimit:  5,
		},
	}

	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Server.Port = p
		}
	}
	if v := os.Getenv("DEEPSEEK_API_KEY"); v != "" {
		cfg.DeepSeek.APIKey = v
	}
	if v := os.Getenv("DEEPSEEK_BASE_URL"); v != "" {
		cfg.DeepSeek.BaseURL = strings.TrimRight(v, "/")
	}
	if v := os.Getenv("DEEPSEEK_MODEL_FAST"); v != "" {
		cfg.DeepSeek.ModelFast = v
	}
	if v := os.Getenv("DEEPSEEK_MODEL_HEAVY"); v != "" {
		cfg.DeepSeek.ModelHeavy = v
	}
	if v := os.Getenv("DATABASE_PATH"); v != "" {
		cfg.DatabasePath = v
	}
	if v := os.Getenv("TIMELINE_STEP_YEARS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.TimelineStepYears = n
		}
	}
	if v := os.Getenv("CORS_ORIGINS"); v != "" {
		parts := strings.Split(v, ",")
		cfg.CORSOrigins = make([]string, 0, len(parts))
		for _, p := range parts {
			if s := strings.TrimSpace(p); s != "" {
				cfg.CORSOrigins = append(cfg.CORSOrigins, s)
			}
		}
	}

	cfg.loadAuthFromEnv()

	return cfg, cfg.Validate()
}

func (c *Config) loadAuthFromEnv() {
	if v := os.Getenv("AUTH_ENABLED"); v != "" {
		c.Auth.Enabled = strings.EqualFold(v, "true") || v == "1"
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		c.Auth.JWTSecret = v
	}
	if v := os.Getenv("SITE_ACCESS_CODE"); v != "" {
		c.Auth.SiteAccessCode = v
	}
	if v := os.Getenv("AUTH_BOOTSTRAP_USERS"); v != "" {
		for _, part := range strings.Split(v, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			idx := strings.Index(part, ":")
			if idx <= 0 || idx >= len(part)-1 {
				continue
			}
			c.Auth.BootstrapUsers = append(c.Auth.BootstrapUsers, BootstrapUser{
				Username: strings.TrimSpace(part[:idx]),
				Password: strings.TrimSpace(part[idx+1:]),
			})
		}
	}
	if v := os.Getenv("AUTH_ACCESS_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			c.Auth.AccessTTL = d
		}
	}
	if v := os.Getenv("AUTH_REFRESH_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			c.Auth.RefreshTTL = d
		}
	}
	if v := os.Getenv("AUTH_GATE_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			c.Auth.GateTTL = d
		}
	}
	if v := os.Getenv("AUTH_RATE_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.Auth.RateLimit = n
		}
	}
	if v := os.Getenv("AUTH_COOKIE_SECURE"); v != "" {
		c.Auth.CookieSecure = strings.EqualFold(v, "true") || v == "1"
	}
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.DeepSeek.APIKey) == "" {
		return fmt.Errorf("缺少必填配置: DEEPSEEK_API_KEY")
	}
	if c.TimelineStepYears <= 0 {
		c.TimelineStepYears = 3
	}
	if c.Auth.Enabled {
		if strings.TrimSpace(c.Auth.JWTSecret) == "" {
			return fmt.Errorf("AUTH_ENABLED=true 时缺少 JWT_SECRET")
		}
		if strings.TrimSpace(c.Auth.SiteAccessCode) == "" {
			return fmt.Errorf("AUTH_ENABLED=true 时缺少 SITE_ACCESS_CODE")
		}
	}
	return nil
}
