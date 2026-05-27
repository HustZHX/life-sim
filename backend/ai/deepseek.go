package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"life-sim/backend/config"
)

type Client struct {
	cfg        *config.DeepSeekConfig
	httpClient *http.Client
	promptDir  string
}

func NewClient(cfg *config.DeepSeekConfig, promptDir string) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 300 * time.Second,
		},
		promptDir: promptDir,
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
	Temperature    float64         `json:"temperature,omitempty"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
	ReasoningEffort string         `json:"reasoning_effort,omitempty"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Client) LoadPrompt(name string) (string, error) {
	path := filepath.Join(c.promptDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("读取 prompt %s: %w", name, err)
	}
	return string(data), nil
}

// ChatJSON 调用 DeepSeek 并解析 JSON。model 为空时 heavy 决定默认模型。
func (c *Client) ChatJSON(ctx context.Context, model, systemPrompt, userContent string, heavy bool) (string, error) {
	if model == "" {
		if heavy {
			model = c.cfg.ModelHeavy
		} else {
			model = c.cfg.ModelFast
		}
	}
	return c.chatJSONWithModel(ctx, model, systemPrompt, userContent)
}

// ChatJSONModel 使用指定 API 模型 ID（deepseek-v4-flash / deepseek-v4-pro）。
func (c *Client) ChatJSONModel(ctx context.Context, apiModel, systemPrompt, userContent string) (string, error) {
	if apiModel == "" {
		apiModel = c.cfg.ModelFast
	}
	return c.chatJSONWithModel(ctx, apiModel, systemPrompt, userContent)
}

func (c *Client) chatJSONWithModel(ctx context.Context, model, systemPrompt, userContent string) (string, error) {
	reqBody := chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userContent},
		},
		MaxTokens:   8192,
		Temperature: 0.7,
		ResponseFormat: &responseFormat{Type: "json_object"},
	}

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		content, err := c.doChat(ctx, reqBody)
		if err == nil {
			return extractJSON(content), nil
		}
		lastErr = err
		time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
	}
	return "", lastErr
}

func (c *Client) doChat(ctx context.Context, reqBody chatRequest) (string, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := c.cfg.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("DeepSeek API %d: %s", resp.StatusCode, string(respBody))
	}

	var cr chatResponse
	if err := json.Unmarshal(respBody, &cr); err != nil {
		return "", err
	}
	if cr.Error != nil {
		return "", fmt.Errorf("DeepSeek: %s", cr.Error.Message)
	}
	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("DeepSeek: 空响应")
	}
	return cr.Choices[0].Message.Content, nil
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		lines := strings.Split(s, "\n")
		var b strings.Builder
		for _, line := range lines {
			if strings.HasPrefix(line, "```") {
				continue
			}
			b.WriteString(line)
			b.WriteByte('\n')
		}
		return strings.TrimSpace(b.String())
	}
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	if idx := strings.Index(s, "["); idx >= 0 {
		endIdx := strings.LastIndex(s, "]")
		if endIdx > idx {
			return s[idx : endIdx+1]
		}
	}
	return s
}
