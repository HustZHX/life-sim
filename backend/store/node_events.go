package store

import (
	"encoding/json"
	"fmt"
)

func ParseRegenerateEventsResult(raw string) (string, error) {
	var resp struct {
		Events string `json:"events"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return "", err
	}
	if resp.Events == "" {
		return "", fmt.Errorf("AI 未返回 events")
	}
	return resp.Events, nil
}
