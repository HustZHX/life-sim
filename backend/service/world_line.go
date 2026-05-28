package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"life-sim/backend/ai"
	"life-sim/backend/model"
	"life-sim/backend/store"
)

func worldLineGeneratePromptName(mode string) string {
	if mode == model.ModeFamous {
		return "generate_world_line_famous.txt"
	}
	return "generate_world_line_random.txt"
}

func worldLineRecalcPromptName(mode string) string {
	if mode == model.ModeFamous {
		return "recalc_world_line_from_nodes_famous.txt"
	}
	return "recalc_world_line_from_nodes_random.txt"
}

func syncNodesFromWorldLinePromptName(mode string) string {
	if mode == model.ModeFamous {
		return "sync_nodes_from_world_line_famous.txt"
	}
	return "sync_nodes_from_world_line_random.txt"
}

func (s *CharacterService) generateWorldLineForTimeline(
	ctx context.Context,
	characterMode string,
	profile *model.Profile,
	apiModel string,
	startYear, endYear int,
	nodes []model.LifeNode,
	eraContext string,
) (*model.WorldLine, error) {
	system, err := s.ai.LoadPrompt(worldLineGeneratePromptName(characterMode))
	if err != nil {
		return nil, err
	}
	user := fmt.Sprintf(
		"start_year=%d\nend_year=%d\n人物档案：\n%s\n时代背景参考：\n%s\n人生节点：\n%s",
		startYear, endYear,
		store.ProfileJSONTimeline(profile),
		eraContext,
		store.MarshalNodesForWorldLine(nodes),
	)
	raw, err := s.ai.ChatJSONModelLong(ctx, apiModel, system, user)
	if err != nil {
		return nil, err
	}
	return store.ParseWorldLine(raw, "")
}

func (s *CharacterService) recalcWorldLineFromNodes(
	ctx context.Context,
	characterMode string,
	profile *model.Profile,
	apiModel string,
	endYear int,
	nodes []model.LifeNode,
	previous *model.WorldLine,
) (*model.WorldLine, error) {
	system, err := s.ai.LoadPrompt(worldLineRecalcPromptName(characterMode))
	if err != nil {
		return nil, err
	}
	prevJSON := "{}"
	if previous != nil {
		prevJSON = store.MarshalWorldLineForPrompt(previous)
	}
	user := fmt.Sprintf(
		"end_year=%d（重算至最后一个人生节点年份）\n人物档案：\n%s\n当前人生节点：\n%s\n旧世界线：\n%s",
		endYear,
		store.ProfileJSONTimeline(profile),
		store.MarshalNodesForWorldLine(nodes),
		prevJSON,
	)
	raw, err := s.ai.ChatJSONModelLong(ctx, apiModel, system, user)
	if err != nil {
		return nil, err
	}
	wl, err := store.ParseWorldLine(raw, "")
	if err != nil {
		return nil, err
	}
	if previous != nil && previous.StartYear > 0 {
		wl.StartYear = previous.StartYear
	} else if len(nodes) > 0 {
		wl.StartYear = nodes[0].Year
	}
	wl.EndYear = endYear
	return wl, nil
}

func (s *CharacterService) syncNodesFromWorldLineAI(
	ctx context.Context,
	characterMode string,
	profile *model.Profile,
	apiModel string,
	nodes []model.LifeNode,
	worldLine *model.WorldLine,
) ([]model.LifeNode, error) {
	system, err := s.ai.LoadPrompt(syncNodesFromWorldLinePromptName(characterMode))
	if err != nil {
		return nil, err
	}
	user := fmt.Sprintf(
		"人物档案：\n%s\n世界线：\n%s\n当前人生节点：\n%s",
		store.ProfileJSONTimeline(profile),
		store.MarshalWorldLineForPrompt(worldLine),
		store.MarshalNodesLocked(nodes),
	)
	raw, err := s.ai.ChatJSONModelLong(ctx, apiModel, system, user)
	if err != nil {
		return nil, err
	}
	updates, err := store.ParseWorldLineNodeUpdates(raw)
	if err != nil {
		return nil, fmt.Errorf("解析节点同步结果失败: %w", err)
	}
	if len(updates) == 0 {
		return nil, fmt.Errorf("AI 未返回需调整的节点")
	}
	return store.ApplyWorldLineNodeUpdates(nodes, updates), nil
}

func (s *CharacterService) persistWorldLine(timelineID string, wl *model.WorldLine) error {
	if wl == nil {
		return nil
	}
	wl.TimelineID = timelineID
	return s.store.SaveWorldLine(timelineID, wl)
}

func (s *CharacterService) maybeRecalcWorldLineAfterRegenerate(
	ctx context.Context,
	jobID, characterID, timelineID string,
	ch *model.Character,
	profile *model.Profile,
	allNodes []model.LifeNode,
	apiModel string,
) {
	endYear := store.MaxNodeYear(allNodes)
	if endYear <= 0 {
		return
	}
	setJobStage(s.store, jobID, 92, "正在同步世界线…")
	prev, _ := s.store.GetWorldLine(timelineID)
	wl, err := s.recalcWorldLineFromNodes(ctx, ch.Mode, profile, apiModel, endYear, allNodes, prev)
	if err != nil {
		log.Printf("[worldline] 重算失败 timeline=%s: %v", timelineID, err)
		return
	}
	if err := s.persistWorldLine(timelineID, wl); err != nil {
		log.Printf("[worldline] 保存失败 timeline=%s: %v", timelineID, err)
	}
}

func (s *CharacterService) UpdateWorldLine(ctx context.Context, characterID, timelineID string, req model.WorldLineUpdateRequest) (*model.Job, *model.WorldLine, error) {
	ch, err := s.store.GetCharacter(characterID)
	if err != nil {
		return nil, nil, err
	}
	tid, err := s.resolveTimelineScope(ch, timelineID)
	if err != nil {
		return nil, nil, err
	}
	tl, err := s.store.GetTimeline(tid)
	if err != nil {
		return nil, nil, err
	}
	req.WorldLine.TimelineID = tid
	if err := s.persistWorldLine(tid, &req.WorldLine); err != nil {
		return nil, nil, err
	}
	if !req.ApplyToNodes {
		return nil, &req.WorldLine, nil
	}
	if tl.CurrentVersionID == "" {
		return nil, &req.WorldLine, fmt.Errorf("时间轴尚无版本")
	}
	modelID := ai.NormalizeModelID(req.Model)
	if _, err := s.resolveAPIModel(modelID); err != nil {
		return nil, nil, err
	}
	reqJSON, _ := json.Marshal(map[string]string{"timeline_id": tid})
	job, err := s.store.CreateJobWithRequest(characterID, "world_line_sync", modelID, string(reqJSON))
	if err != nil {
		return nil, nil, err
	}
	go s.runWorldLineSync(context.Background(), job.ID, characterID, ch, tl, req.WorldLine, modelID)
	return job, &req.WorldLine, nil
}

func (s *CharacterService) StartRefreshWorldLineJob(ctx context.Context, characterID, timelineID, modelID string) (*model.Job, error) {
	ch, err := s.store.GetCharacter(characterID)
	if err != nil {
		return nil, err
	}
	tid, err := s.resolveTimelineScope(ch, timelineID)
	if err != nil {
		return nil, err
	}
	tl, err := s.store.GetTimeline(tid)
	if err != nil {
		return nil, err
	}
	if tl.CurrentVersionID == "" {
		return nil, fmt.Errorf("时间轴尚无版本")
	}
	modelID = ai.NormalizeModelID(modelID)
	if _, err := s.resolveAPIModel(modelID); err != nil {
		return nil, err
	}
	reqJSON, _ := json.Marshal(map[string]string{"timeline_id": tid})
	job, err := s.store.CreateJobWithRequest(characterID, "world_line_refresh", modelID, string(reqJSON))
	if err != nil {
		return nil, err
	}
	go s.runRefreshWorldLine(context.Background(), job.ID, characterID, ch, tl, modelID)
	return job, nil
}

func (s *CharacterService) runRefreshWorldLine(
	ctx context.Context,
	jobID, characterID string,
	ch *model.Character,
	timeline *model.Timeline,
	modelID string,
) {
	job, _ := s.store.GetJob(jobID)
	job.Status = model.JobRunning
	job.Progress = 10
	job.StageText = "正在重算世界线…"
	_ = s.store.UpdateJob(job)

	profile, err := s.store.GetProfile(characterID)
	if err != nil {
		s.failJob(job, err.Error())
		return
	}
	nodes, err := s.store.GetNodesByVersion(timeline.CurrentVersionID)
	if err != nil {
		s.failJob(job, err.Error())
		return
	}
	if len(nodes) == 0 {
		s.failJob(job, "当前时间轴无节点")
		return
	}
	endYear := store.MaxNodeYear(nodes)
	if endYear <= 0 {
		s.failJob(job, "无法确定节点年份上界")
		return
	}

	apiModel, err := s.resolveAPIModel(modelID)
	if err != nil {
		s.failJob(job, err.Error())
		return
	}

	setJobStage(s.store, jobID, 30, "AI 正在整理天下大事与历史走向…")
	prev, _ := s.store.GetWorldLine(timeline.ID)
	wl, err := s.recalcWorldLineFromNodes(ctx, ch.Mode, profile, apiModel, endYear, nodes, prev)
	if err != nil {
		s.failJob(job, err.Error())
		return
	}
	if err := s.persistWorldLine(timeline.ID, wl); err != nil {
		s.failJob(job, err.Error())
		return
	}

	wlJSON, _ := json.Marshal(wl)
	job.Status = model.JobCompleted
	job.Progress = 100
	job.StageText = "世界线已更新"
	job.Result = string(wlJSON)
	_ = s.store.UpdateJob(job)
}

func (s *CharacterService) runWorldLineSync(
	ctx context.Context,
	jobID, characterID string,
	ch *model.Character,
	timeline *model.Timeline,
	worldLine model.WorldLine,
	modelID string,
) {
	job, _ := s.store.GetJob(jobID)
	job.Status = model.JobRunning
	job.Progress = 15
	job.StageText = "正在根据世界线调整人生节点…"
	_ = s.store.UpdateJob(job)

	profile, err := s.store.GetProfile(characterID)
	if err != nil {
		s.failJob(job, err.Error())
		return
	}
	oldNodes, err := s.store.GetNodesByVersion(timeline.CurrentVersionID)
	if err != nil {
		s.failJob(job, err.Error())
		return
	}
	apiModel, err := s.resolveAPIModel(modelID)
	if err != nil {
		s.failJob(job, err.Error())
		return
	}

	setJobStage(s.store, jobID, 35, "AI 正在对齐人生节点…")
	updated, err := s.syncNodesFromWorldLineAI(ctx, ch.Mode, profile, apiModel, oldNodes, &worldLine)
	if err != nil {
		s.failJob(job, err.Error())
		return
	}

	newVersionID := uuid.New().String()
	allNodes := make([]model.LifeNode, len(updated))
	for i, n := range updated {
		nn := n
		nn.VersionID = newVersionID
		nn.ID = uuid.New().String()
		allNodes[i] = nn
	}

	version := &model.TimelineVersion{
		ID:                newVersionID,
		CharacterID:       characterID,
		TimelineID:        timeline.ID,
		ParentVersionID:   timeline.CurrentVersionID,
		BranchLabel:       "世界线同步",
		ChangeSummary:     "根据世界线变更调整人生节点",
		DeathYearSnapshot: store.MaxNodeYear(allNodes),
		CreatedAt:         time.Now(),
	}
	if err := s.store.CreateVersion(version); err != nil {
		s.failJob(job, err.Error())
		return
	}
	if err := s.store.SaveNodes(allNodes); err != nil {
		s.failJob(job, err.Error())
		return
	}
	if err := s.store.UpdateTimelineCurrentVersion(timeline.ID, newVersionID); err != nil {
		s.failJob(job, err.Error())
		return
	}
	ch.CurrentVersionID = newVersionID
	ch.CurrentTimelineID = timeline.ID
	_ = s.store.UpdateCharacter(ch)
	_ = s.store.CopyMemoriesWithNewVersion(timeline.CurrentVersionID, newVersionID)
	_ = s.store.CopyDialogueIdentityPresetsWithNewVersion(timeline.CurrentVersionID, newVersionID)

	if err := s.persistWorldLine(timeline.ID, &worldLine); err != nil {
		log.Printf("[worldline] 保存失败: %v", err)
	}

	diff := store.ComputeVersionDiff(oldNodes, allNodes, "")
	diffJSON, _ := json.Marshal(model.VersionDiff{
		VersionID: newVersionID, ParentVersionID: timeline.CurrentVersionID, Changes: diff,
	})
	job.Status = model.JobCompleted
	job.Progress = 100
	job.StageText = "世界线同步完成"
	job.Result = string(diffJSON)
	_ = s.store.UpdateJob(job)
}
