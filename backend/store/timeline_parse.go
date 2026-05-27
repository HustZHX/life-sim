package store

import (
	"encoding/json"
	"fmt"

	"life-sim/backend/model"
)

func ParseTimelineRecommendations(raw string) ([]model.TimelineRecommendation, error) {
	var resp struct {
		Recommendations []model.TimelineRecommendation `json:"recommendations"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, err
	}
	if len(resp.Recommendations) == 0 {
		return nil, fmt.Errorf("AI 未返回推荐方案")
	}
	out := make([]model.TimelineRecommendation, 0, len(resp.Recommendations))
	for _, r := range resp.Recommendations {
		span := r.EndYear - r.StartYear
		if span < 0 {
			span = 0
		}
		if r.TargetNodeCount <= 0 {
			if r.StepYears > 0 && span > 0 {
				r.TargetNodeCount = span / r.StepYears
			}
			if r.TargetNodeCount <= 0 {
				r.TargetNodeCount = DefaultTargetNodeCount
			}
		}
		r.TargetNodeCount = ClampTargetNodeCount(r.TargetNodeCount)
		r.StepYears = ComputeStepYears(span, r.TargetNodeCount)
		out = append(out, r)
	}
	return out, nil
}

func ParseLifespanRecalc(raw string) (*model.LifespanRecalcResult, error) {
	var r model.LifespanRecalcResult
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		return nil, err
	}
	if r.DeathYear <= 0 {
		return nil, fmt.Errorf("无效的 death_year")
	}
	return &r, nil
}
