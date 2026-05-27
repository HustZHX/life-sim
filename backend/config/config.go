package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server           ServerConfig
	DeepSeek         DeepSeekConfig
	DatabasePath     string
	TimelineStepYears int
	CORSOrigins      []string
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

	return cfg, cfg.Validate()
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.DeepSeek.APIKey) == "" {
		return fmt.Errorf("缺少必填配置: DEEPSEEK_API_KEY")
	}
	if c.TimelineStepYears <= 0 {
		c.TimelineStepYears = 3
	}
	return nil
}
