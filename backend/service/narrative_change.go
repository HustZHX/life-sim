package service

import (
	"context"
	"fmt"
	"strings"

	"life-sim/backend/ai"
	"life-sim/backend/model"
	"life-sim/backend/store"
)

func (s *CharacterService) ApplyNarrativeChange(ctx context.Context, characterID string, req model.NarrativeChangeRequest) (*model.Job, error) {
	instruction := strings.TrimSpace(req.Instruction)
	if instruction == "" {
		return nil, fmt.Errorf("请填写叙述变更内容")
	}
	if req.TimelineID == "" {
		return nil, fmt.Errorf("缺少 timeline_id")
	}

	ch, err := s.store.GetCharacter(characterID)
	if err != nil {
		return nil, err
	}
	timeline, err := s.store.GetTimeline(req.TimelineID)
	if err != nil {
		return nil, err
	}
	if timeline.CharacterID != characterID {
		return nil, fmt.Errorf("时间轴不属于该角色")
	}
	if timeline.CurrentVersionID == "" {
		return nil, fmt.Errorf("时间轴尚无版本")
	}

	oldNodes, err := s.store.GetNodesByVersion(timeline.CurrentVersionID)
	if err != nil {
		return nil, err
	}
	if len(oldNodes) == 0 {
		return nil, fmt.Errorf("当前版本无节点")
	}

	modelID := ai.NormalizeModelID(req.Model)
	if _, err := s.resolveAPIModel(modelID); err != nil {
		return nil, err
	}

	job, err := s.store.CreateJob(characterID, "timeline_narrative_change", modelID)
	if err != nil {
		return nil, err
	}

	go s.runNarrativeChange(context.Background(), job.ID, characterID, ch, timeline, oldNodes, instruction, req)

	return job, nil
}

func (s *CharacterService) runNarrativeChange(
	ctx context.Context,
	jobID, characterID string,
	ch *model.Character,
	timeline *model.Timeline,
	oldNodes []model.LifeNode,
	instruction string,
	req model.NarrativeChangeRequest,
) {
	modelID := req.Model
	job, _ := s.store.GetJob(jobID)
	job.Status = model.JobRunning
	job.Progress = 10
	job.Model = modelID
	job.StageText = "正在理解叙述变更…"
	_ = s.store.UpdateJob(job)

	profile, err := s.store.GetProfile(characterID)
	if err != nil {
		s.failJob(job, err.Error())
		return
	}

	system, err := s.ai.LoadPrompt(planNarrativeChangePromptName(ch.Mode))
	if err != nil {
		s.failJob(job, err.Error())
		return
	}

	user := fmt.Sprintf(
		"人物档案：\n%s\n\n当前时间轴全部节点（sequence 递增）：\n%s\n\n用户叙述变更（须落实）：\n%s",
		store.ProfileJSONTimeline(profile),
		store.MarshalNodesForNarrative(oldNodes),
		instruction,
	)

	setJobStage(s.store, jobID, 15, "AI 正在规划节点变更…")
	apiModel, err := s.resolveAPIModel(modelID)
	if err != nil {
		s.failJob(job, err.Error())
		return
	}

	done := make(chan struct{})
	go tickJobProgress(s.store, jobID, 18, 42, done)

	raw, err := s.ai.ChatJSONModel(ctx, apiModel, system, user)
	close(done)
	if err != nil {
		s.failJob(job, fmt.Sprintf("规划失败: %v", err))
		return
	}

	plan, err := store.ParseNarrativeChangePlan(raw)
	if err != nil {
		s.failJob(job, "规划解析失败: "+err.Error())
		return
	}

	setJobStage(s.store, jobID, 45, "正在应用节点变更…")
	locked, err := store.ApplyNarrativeChangePlan(oldNodes, plan, characterID, profile.DisplayName)
	if err != nil {
		s.failJob(job, "应用变更失败: "+err.Error())
		return
	}
	anchor := locked[len(locked)-1]

	patchReq := model.PatchNodeRequest{
		Model:               modelID,
		Mode:                model.PatchModeFullCascade,
		TargetNodeCount:     req.TargetNodeCount,
		Title:               anchor.Title,
		Events:              anchor.Events,
		Thoughts:            anchor.Thoughts,
		PersonalitySnapshot: anchor.PersonalitySnapshot,
		ChangeSummary:       plan.ChangeSummary + "；重算后续节点",
	}
	setJobStage(s.store, jobID, 50, "正在推演后续节点…")
	s.runRegenerate(ctx, jobID, characterID, ch, timeline, locked, &anchor, oldNodes, patchReq)
}
