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
	var wire struct {
		Events              string          `json:"events"`
		Thoughts            string          `json:"thoughts"`
		PersonalitySnapshot string          `json:"personality_snapshot"`
		TraitChanges        json.RawMessage `json:"trait_changes"`
	}
	if err := json.Unmarshal([]byte(raw), &wire); err != nil {
		return nil, err
	}
	if wire.Events == "" {
		return nil, fmt.Errorf("AI 未返回 events")
	}
	traits, err := ParseTraitChangesJSON(wire.TraitChanges)
	if err != nil {
		return nil, err
	}
	return &RegenerateNodeFromTitle{
		Events:              wire.Events,
		Thoughts:            wire.Thoughts,
		PersonalitySnapshot: wire.PersonalitySnapshot,
		TraitChanges:        traits,
	}, nil
}

// ParseRegenerateEventsResult 兼容仅返回 events 的旧响应。
func ParseRegenerateEventsResult(raw string) (string, error) {
	r, err := ParseRegenerateNodeFromTitle(raw)
	if err != nil {
		return "", err
	}
	return r.Events, nil
}
