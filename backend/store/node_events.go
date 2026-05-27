package store

import (
	"encoding/json"
	"fmt"

	"life-sim/backend/model"
)

type RegenerateNodeFromTitle struct {
	Events              string              `json:"events"`
	Thoughts            string              `json:"thoughts"`
	PersonalitySnapshot string              `json:"personality_snapshot"`
	TraitChanges        []model.TraitChange `json:"trait_changes"`
}

func ParseRegenerateNodeFromTitle(raw string) (*RegenerateNodeFromTitle, error) {
	var resp RegenerateNodeFromTitle
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, err
	}
	if resp.Events == "" {
		return nil, fmt.Errorf("AI 未返回 events")
	}
	return &resp, nil
}

// ParseRegenerateEventsResult 兼容仅返回 events 的旧响应。
func ParseRegenerateEventsResult(raw string) (string, error) {
	r, err := ParseRegenerateNodeFromTitle(raw)
	if err != nil {
		return "", err
	}
	return r.Events, nil
}
