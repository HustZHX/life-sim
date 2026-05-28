package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"life-sim/backend/ai"
	"life-sim/backend/model"
)

func timelineConfigFromSnapshot(s model.TimelineJobConfigSnapshot) timelineJobConfig {
	return timelineJobConfig{
		StepYears:         s.StepYears,
		TargetNodeCount:   s.TargetNodeCount,
		StartYear:         s.StartYear,
		EndYear:           s.EndYear,
		Title:             s.Title,
		Instructions:      s.Instructions,
		EraEventsOverride: s.EraEventsOverride,
		CharacterMode:     s.CharacterMode,
		NarrativeDensity:  s.NarrativeDensity,
	}
}

func timelineConfigSnapshot(cfg timelineJobConfig) model.TimelineJobConfigSnapshot {
	return model.TimelineJobConfigSnapshot{
		StepYears:         cfg.StepYears,
		TargetNodeCount:   cfg.TargetNodeCount,
		StartYear:         cfg.StartYear,
		EndYear:           cfg.EndYear,
		Title:             cfg.Title,
		Instructions:      cfg.Instructions,
		EraEventsOverride: cfg.EraEventsOverride,
		CharacterMode:     cfg.CharacterMode,
		NarrativeDensity:  cfg.NarrativeDensity,
	}
}

func (s *CharacterService) RetryJob(ctx context.Context, jobID string) (*model.Job, error) {
	old, err := s.store.GetJob(jobID)
	if err != nil {
		return nil, err
	}
	if old.Status != model.JobFailed {
		return nil, fmt.Errorf("只能重试失败的任务")
	}
	switch old.Type {
	case "timeline_generate":
		return s.retryTimelineGenerate(ctx, old)
	case "timeline_regenerate", "node_inner_current", "node_inner_subsequent":
		return s.retryPatchNode(ctx, old)
	case "timeline_narrative_change":
		return s.retryNarrativeChange(ctx, old)
	default:
		if strings.HasPrefix(old.Type, "narrative_") {
			return nil, fmt.Errorf("请使用叙事任务重试接口")
		}
		return nil, fmt.Errorf("不支持重试该任务类型: %s", old.Type)
	}
}

func (s *CharacterService) retryTimelineGenerate(ctx context.Context, old *model.Job) (*model.Job, error) {
	var req model.TimelineGenerateJobRequest
	if err := json.Unmarshal([]byte(old.RequestJSON), &req); err != nil {
		return nil, fmt.Errorf("无法解析任务参数")
	}
	if req.TimelineID == "" || req.Config.TargetNodeCount == 0 {
		return nil, fmt.Errorf("任务参数不完整，请重新提交生成")
	}
	tl, err := s.store.GetTimeline(req.TimelineID)
	if err != nil {
		return nil, err
	}
	if tl.CurrentVersionID != "" {
		return nil, fmt.Errorf("该时间轴已生成完成，无需重试")
	}
	profile, err := s.store.GetProfile(old.CharacterID)
	if err != nil {
		return nil, err
	}
	modelID := ai.NormalizeModelID(old.Model)
	if _, err := s.resolveAPIModel(modelID); err != nil {
		return nil, err
	}

	job, err := s.store.CreateJobWithRequest(old.CharacterID, old.Type, modelID, old.RequestJSON)
	if err != nil {
		return nil, err
	}
	cfg := timelineConfigFromSnapshot(req.Config)
	go s.runTimelineGenerate(context.Background(), job.ID, old.CharacterID, req.TimelineID, profile, modelID, cfg)
	return job, nil
}

func (s *CharacterService) retryPatchNode(ctx context.Context, old *model.Job) (*model.Job, error) {
	var req model.PatchNodeJobRequest
	if err := json.Unmarshal([]byte(old.RequestJSON), &req); err != nil || req.NodeID == "" {
		return nil, fmt.Errorf("任务参数不完整，请重新提交")
	}
	if req.Patch.Mode == "" {
		switch old.Type {
		case "node_inner_current":
			req.Patch.Mode = model.PatchModeInnerCurrent
		case "node_inner_subsequent":
			req.Patch.Mode = model.PatchModeInnerSubsequent
		default:
			req.Patch.Mode = model.PatchModeFullCascade
		}
	}
	if req.Patch.Model == "" {
		req.Patch.Model = old.Model
	}
	return s.PatchNodeAndRegenerate(ctx, old.CharacterID, req.NodeID, req.Patch)
}

func (s *CharacterService) retryNarrativeChange(ctx context.Context, old *model.Job) (*model.Job, error) {
	var req model.NarrativeChangeJobRequest
	if err := json.Unmarshal([]byte(old.RequestJSON), &req); err != nil || req.Request.Instruction == "" {
		return nil, fmt.Errorf("任务参数不完整，请重新提交")
	}
	if req.Request.Model == "" {
		req.Request.Model = old.Model
	}
	if req.Request.TimelineID == "" {
		req.Request.TimelineID = req.TimelineID
	}
	return s.ApplyNarrativeChange(ctx, old.CharacterID, req.Request)
}
