package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"life-sim/backend/ai"
	"life-sim/backend/model"
)

func (s *NarrativeService) RetryJob(ctx context.Context, jobID string) (*model.Job, error) {
	old, err := s.store.GetJob(jobID)
	if err != nil {
		return nil, err
	}
	if old.Status != model.JobFailed {
		return nil, fmt.Errorf("只能重试失败的任务")
	}
	switch old.Type {
	case "narrative_light_novel":
		return s.retryLightNovel(ctx, old)
	case "narrative_diary", "narrative_letter", "narrative_archive":
		return s.retryNodeNarrative(ctx, old)
	default:
		return nil, fmt.Errorf("不支持重试该任务类型: %s", old.Type)
	}
}

func (s *NarrativeService) retryLightNovel(ctx context.Context, old *model.Job) (*model.Job, error) {
	var req model.LightNovelRequest
	if err := json.Unmarshal([]byte(old.RequestJSON), &req); err != nil {
		return nil, fmt.Errorf("任务参数不完整，请重新提交")
	}
	if req.Model == "" {
		req.Model = ai.NormalizeModelID(old.Model)
	}
	req.Force = true
	job, _, err := s.StartLightNovelJob(ctx, old.CharacterID, req)
	return job, err
}

func (s *NarrativeService) retryNodeNarrative(ctx context.Context, old *model.Job) (*model.Job, error) {
	var req model.NodeNarrativeJobRequest
	if err := json.Unmarshal([]byte(old.RequestJSON), &req); err != nil || req.NodeID == "" {
		return nil, fmt.Errorf("任务参数不完整，请重新提交")
	}
	kind := req.Kind
	if kind == "" {
		kind = strings.TrimPrefix(old.Type, "narrative_")
	}
	if req.Request.Model == "" {
		req.Request.Model = ai.NormalizeModelID(old.Model)
	}
	req.Request.Force = true
	job, _, err := s.StartNodeNarrativeJob(ctx, old.CharacterID, req.NodeID, kind, req.Request)
	return job, err
}
