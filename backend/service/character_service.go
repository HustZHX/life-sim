package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"life-sim/backend/ai"
	"life-sim/backend/config"
	"life-sim/backend/model"
	"life-sim/backend/store"
)

type CharacterService struct {
	store      *store.Store
	ai         *ai.Client
	stepYears  int
	pending    map[string][]model.ResolveCandidate
	aiCache    *aiContextCache
}

func NewCharacterService(st *store.Store, aiClient *ai.Client, cfg *config.Config) *CharacterService {
	return &CharacterService{
		store:     st,
		ai:        aiClient,
		stepYears: cfg.TimelineStepYears,
		pending:   make(map[string][]model.ResolveCandidate),
		aiCache:   newAIContextCache(),
	}
}

func (s *CharacterService) resolveAPIModel(modelID string) (string, error) {
	return ai.ResolveAPIModel(modelID)
}

func (s *CharacterService) ListModels() []ai.ModelOption {
	return ai.ListModels()
}

func (s *CharacterService) CreateCharacter(mode string) (*model.Character, error) {
	if mode != model.ModeFamous && mode != model.ModeRandom {
		return nil, fmt.Errorf("无效模式: %s", mode)
	}
	return s.store.CreateCharacter(mode)
}

func (s *CharacterService) GetCharacter(id string) (*model.Character, error) {
	return s.store.GetCharacter(id)
}

func (s *CharacterService) ListHistory(limit int) ([]model.CharacterHistoryItem, error) {
	return s.store.ListCharacterHistory(limit)
}

func (s *CharacterService) ResolvePerson(ctx context.Context, characterID, query string) ([]model.ResolveCandidate, error) {
	ch, err := s.store.GetCharacter(characterID)
	if err != nil {
		return nil, err
	}
	if ch.Mode != model.ModeFamous {
		return nil, fmt.Errorf("仅名人模式支持消歧")
	}

	system, err := s.ai.LoadPrompt("resolve_person.txt")
	if err != nil {
		return nil, err
	}
	user := fmt.Sprintf("用户输入的名字：%s", query)
	raw, err := s.ai.ChatJSON(ctx, "", system, user, false)
	if err != nil {
		return nil, err
	}

	candidates, err := store.ParseResolveResult(raw)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("未找到匹配的历史人物")
	}

	s.pending[characterID] = candidates
	ch.ResolveQuery = query
	ch.Status = model.StatusResolved
	_ = s.store.UpdateCharacter(ch)
	return candidates, nil
}

func (s *CharacterService) ConfirmPerson(characterID string, candidateIndex int) (*model.Character, error) {
	candidates, ok := s.pending[characterID]
	if !ok || candidateIndex < 0 || candidateIndex >= len(candidates) {
		return nil, fmt.Errorf("无效的候选索引")
	}
	c := candidates[candidateIndex]
	identity, _ := json.Marshal(c)

	ch, err := s.store.GetCharacter(characterID)
	if err != nil {
		return nil, err
	}
	ch.DisplayName = c.Name
	ch.ConfirmedIdentity = string(identity)
	ch.Status = model.StatusConfirmed
	if err := s.store.UpdateCharacter(ch); err != nil {
		return nil, err
	}
	delete(s.pending, characterID)
	return ch, nil
}

func (s *CharacterService) GenerateProfile(ctx context.Context, characterID string, req model.ProfileGenerateRequest) (*model.Profile, error) {
	ch, err := s.store.GetCharacter(characterID)
	if err != nil {
		return nil, err
	}
	modelID := ai.NormalizeModelID(req.Model)
	apiModel, err := s.resolveAPIModel(modelID)
	if err != nil {
		return nil, err
	}

	var system, user string
	switch ch.Mode {
	case model.ModeFamous:
		if ch.Status != model.StatusConfirmed {
			return nil, fmt.Errorf("请先确认人物身份")
		}
		system, err = s.ai.LoadPrompt("fetch_profile_famous.txt")
		if err != nil {
			return nil, err
		}
		user = fmt.Sprintf("已确认身份：%s\n用户原始查询：%s", ch.ConfirmedIdentity, ch.ResolveQuery)
	case model.ModeRandom:
		system, err = s.ai.LoadPrompt("fetch_profile_random.txt")
		if err != nil {
			return nil, err
		}
		name := req.DisplayName
		if name == "" {
			name = "小明"
		}
		user = fmt.Sprintf("指定姓名：%s", name)
		if req.Background != "" {
			user += fmt.Sprintf("\n背景设定：%s", req.Background)
		}
		if req.Introduction != "" {
			user += fmt.Sprintf("\n一句话介绍：%s", req.Introduction)
		}
		if req.Background == "" && req.Introduction == "" {
			user += "\n请自由发挥，生成一个有趣的普通人。"
		}
	default:
		return nil, fmt.Errorf("未知模式")
	}

	raw, err := s.ai.ChatJSONModel(ctx, apiModel, system, user)
	if err != nil {
		return nil, err
	}
	profile, err := store.ParseProfile(raw, characterID)
	if err != nil {
		return nil, err
	}
	if profile.DisplayName == "" && req.DisplayName != "" {
		profile.DisplayName = req.DisplayName
	}
	if profile.DisplayName == "" && ch.DisplayName != "" {
		profile.DisplayName = ch.DisplayName
	}
	if profile.DisplayName != "" {
		ch.DisplayName = profile.DisplayName
	}
	if err := s.store.SaveProfile(profile); err != nil {
		return nil, err
	}
	s.aiCache.invalidate(characterID)
	ch.Status = model.StatusProfileReady
	_ = s.store.UpdateCharacter(ch)
	return profile, nil
}

func (s *CharacterService) UpdateProfile(characterID string, incoming *model.Profile) (*model.Profile, error) {
	if _, err := s.store.GetProfile(characterID); err != nil {
		return nil, err
	}
	incoming.CharacterID = characterID
	if err := s.store.SaveProfile(incoming); err != nil {
		return nil, err
	}
	s.aiCache.invalidate(characterID)
	if incoming.DisplayName != "" {
		ch, err := s.store.GetCharacter(characterID)
		if err == nil {
			ch.DisplayName = incoming.DisplayName
			_ = s.store.UpdateCharacter(ch)
		}
	}
	return incoming, nil
}

func (s *CharacterService) RandomizeProfileField(ctx context.Context, characterID string, req model.ProfileRandomizeFieldRequest) (*model.Profile, error) {
	if !store.ValidProfileField(req.Field) {
		return nil, fmt.Errorf("不支持随机化的字段: %s", req.Field)
	}
	profile, err := s.store.GetProfile(characterID)
	if err != nil {
		return nil, err
	}
	modelID := ai.NormalizeModelID(req.Model)
	apiModel, err := s.resolveAPIModel(modelID)
	if err != nil {
		return nil, err
	}
	system, err := s.ai.LoadPrompt("randomize_profile_field.txt")
	if err != nil {
		return nil, err
	}
	user := fmt.Sprintf("需重新生成的字段：%s\n当前完整档案：\n%s", req.Field, store.ProfileJSONFull(profile))
	raw, err := s.ai.ChatJSONModel(ctx, apiModel, system, user)
	if err != nil {
		return nil, err
	}
	valueRaw, err := store.ParseRandomizeFieldResult(raw, req.Field)
	if err != nil {
		return nil, err
	}
	if err := store.ApplyProfileFieldValue(profile, req.Field, valueRaw); err != nil {
		return nil, err
	}
	if err := s.store.SaveProfile(profile); err != nil {
		return nil, err
	}
	s.aiCache.invalidate(characterID)
	if req.Field == "display_name" && profile.DisplayName != "" {
		ch, _ := s.store.GetCharacter(characterID)
		if ch != nil {
			ch.DisplayName = profile.DisplayName
			_ = s.store.UpdateCharacter(ch)
		}
	}
	return profile, nil
}

func (s *CharacterService) ImportTemplate(ctx context.Context, characterID, famousQuery string) (*model.Profile, error) {
	ch, err := s.store.GetCharacter(characterID)
	if err != nil {
		return nil, err
	}
	if ch.Mode != model.ModeRandom {
		return nil, fmt.Errorf("仅随机模式支持导入名人模板")
	}

	resolveSystem, err := s.ai.LoadPrompt("resolve_person.txt")
	if err != nil {
		return nil, err
	}
	resolveRaw, err := s.ai.ChatJSON(ctx, "", resolveSystem, "用户输入："+famousQuery, false)
	if err != nil {
		return nil, err
	}
	candidates, err := store.ParseResolveResult(resolveRaw)
	if err != nil || len(candidates) == 0 {
		return nil, fmt.Errorf("未能识别名人: %s", famousQuery)
	}

	importSystem, err := s.ai.LoadPrompt("import_template.txt")
	if err != nil {
		return nil, err
	}
	importUser := fmt.Sprintf("历史名人：%s\n简介：%s", candidates[0].Name, candidates[0].Summary)
	importRaw, err := s.ai.ChatJSON(ctx, "", importSystem, importUser, false)
	if err != nil {
		return nil, err
	}

	var patch map[string]string
	if err := json.Unmarshal([]byte(importRaw), &patch); err != nil {
		return nil, err
	}

	profile, err := s.store.GetProfile(characterID)
	if err != nil {
		profile, err = s.GenerateProfile(ctx, characterID, model.ProfileGenerateRequest{})
		if err != nil {
			return nil, err
		}
	}
	store.MergeProfile(profile, patch)
	if err := s.store.SaveProfile(profile); err != nil {
		return nil, err
	}
	return profile, nil
}

func (s *CharacterService) GetProfile(characterID string) (*model.Profile, error) {
	return s.store.GetProfile(characterID)
}

func (s *CharacterService) SuggestRandomNames(ctx context.Context, req model.SuggestNamesRequest) ([]string, error) {
	modelID := ai.NormalizeModelID(req.Model)
	apiModel, err := s.resolveAPIModel(modelID)
	if err != nil {
		return nil, err
	}
	system, err := s.ai.LoadPrompt("suggest_random_names.txt")
	if err != nil {
		return nil, err
	}
	bg := strings.TrimSpace(req.Background)
	intro := strings.TrimSpace(req.Introduction)
	var user string
	if bg == "" && intro == "" {
		user = "用户未提供背景与介绍，请随机生成 5 个风格各异的中文姓名。"
	} else {
		user = "请根据以下信息生成 5 个姓名备选："
		if bg != "" {
			user += fmt.Sprintf("\n背景设定：%s", bg)
		}
		if intro != "" {
			user += fmt.Sprintf("\n一句话介绍：%s", intro)
		}
	}
	raw, err := s.ai.ChatJSONModel(ctx, apiModel, system, user)
	if err != nil {
		return nil, err
	}
	return store.ParseSuggestNamesResult(raw)
}

func (s *CharacterService) StartTimelineJob(ctx context.Context, characterID string, req model.TimelineGenerateRequest) (*model.Job, error) {
	ch, err := s.store.GetCharacter(characterID)
	if err != nil {
		return nil, err
	}
	if ch.Status != model.StatusProfileReady && ch.Status != model.StatusTimelineReady {
		return nil, fmt.Errorf("请先生成人物档案")
	}
	profile, err := s.store.GetProfile(characterID)
	if err != nil {
		return nil, err
	}
	modelID := ai.NormalizeModelID(req.Model)
	if _, err := s.resolveAPIModel(modelID); err != nil {
		return nil, err
	}
	rangeCfg, err := store.ResolveTimelineRange(store.TimelineRangeConfig{
		TargetNodeCount: req.TargetNodeCount,
		StartYear:       req.StartYear,
		EndYear:         req.EndYear,
	}, profile)
	if err != nil {
		return nil, err
	}

	timelineID := uuid.New().String()
	title := defaultTimelineTitle(req.Title, rangeCfg.StartYear, rangeCfg.EndYear, rangeCfg.TargetNodeCount)
	if err := s.store.CreateTimeline(&model.Timeline{
		ID: timelineID, CharacterID: characterID, Title: title,
	}); err != nil {
		return nil, err
	}

	job, err := s.store.CreateJob(characterID, "timeline_generate", modelID)
	if err != nil {
		return nil, err
	}

	cfg := timelineJobConfig{
		StepYears:        rangeCfg.StepYears,
		TargetNodeCount:  rangeCfg.TargetNodeCount,
		StartYear:        rangeCfg.StartYear,
		EndYear:          rangeCfg.EndYear,
		Title:            title,
		Instructions:     req.Instructions,
		CharacterMode:    ch.Mode,
		NarrativeDensity: model.NormalizeNarrativeDensity(req.NarrativeDensity),
	}
	go s.runTimelineGenerate(context.Background(), job.ID, characterID, timelineID, profile, modelID, cfg)

	return job, nil
}

func defaultTimelineTitle(custom string, startYear, endYear, targetNodes int) string {
	if custom != "" {
		return custom
	}
	if startYear > 0 && endYear > 0 {
		return fmt.Sprintf("%d—%d · 约%d节点", startYear, endYear, targetNodes)
	}
	return fmt.Sprintf("时间轴 · 约%d节点", targetNodes)
}

type timelineJobConfig struct {
	StepYears        int
	TargetNodeCount  int
	StartYear        int
	EndYear          int
	Title            string
	Instructions     string
	CharacterMode    string
	NarrativeDensity string
}

func (s *CharacterService) runTimelineGenerate(ctx context.Context, jobID, characterID, timelineID string, profile *model.Profile, modelID string, cfg timelineJobConfig) {
	span := cfg.EndYear - cfg.StartYear
	log.Printf("[timeline] 开始生成 character=%s %s (%d-%d) target=%d nodes span=%d step=%d range=%d-%d",
		characterID, profile.DisplayName, profile.BirthYear, profile.DeathYear,
		cfg.TargetNodeCount, span, cfg.StepYears, cfg.StartYear, cfg.EndYear)

	job, _ := s.store.GetJob(jobID)
	job.Status = model.JobRunning
	job.Progress = 10
	job.StageText = "正在准备生成任务…"
	_ = s.store.UpdateJob(job)

	profileJSON := s.aiCache.getProfileTimelineJSON(characterID, profile)
	profileHash := store.ProfileHash(profile)

	apiModel, err := s.resolveAPIModel(modelID)
	if err != nil {
		s.failJob(job, err.Error())
		return
	}
	job.Model = modelID
	_ = s.store.UpdateJob(job)

	versionID := uuid.New().String()
	var allNodes []model.LifeNode
	var genErr error

	if cfg.NarrativeDensity == model.NarrativeRich {
		setJobStage(s.store, jobID, 12, "细腻模式：先规划骨架，再叙事扩写…")
		allNodes, genErr = s.generateRichTimelineNodes(ctx, jobID, characterID, cfg.CharacterMode, profile, profileJSON, apiModel, cfg, versionID)
		if genErr != nil {
			log.Printf("[timeline] 细腻模式生成失败: %v", genErr)
			s.failJob(job, genErr.Error())
			return
		}
	} else {
		allNodes, genErr = s.generateStandardTimelineNodes(ctx, jobID, characterID, profile, profileJSON, profileHash, apiModel, cfg, versionID)
		if genErr != nil {
			log.Printf("[timeline] 标准模式生成失败: %v", genErr)
			s.failJob(job, genErr.Error())
			return
		}
	}

	setJobStage(s.store, jobID, 90, "正在保存版本…")

	version := &model.TimelineVersion{
		ID:              versionID,
		CharacterID:     characterID,
		TimelineID:      timelineID,
		ParentVersionID: "",
		ChangeSummary:   fmt.Sprintf("生成时间轴（%d-%d，约%d个节点）", cfg.StartYear, cfg.EndYear, cfg.TargetNodeCount),
		CreatedAt:       time.Now(),
	}
	if err := s.store.CreateVersion(version); err != nil {
		s.failJob(job, err.Error())
		return
	}
	if err := s.store.SaveNodes(allNodes); err != nil {
		s.failJob(job, err.Error())
		return
	}

	if err := s.store.UpdateTimelineCurrentVersion(timelineID, versionID); err != nil {
		s.failJob(job, err.Error())
		return
	}

	ch, _ := s.store.GetCharacter(characterID)
	ch.CurrentVersionID = versionID
	ch.CurrentTimelineID = timelineID
	ch.Status = model.StatusTimelineReady
	_ = s.store.UpdateCharacter(ch)

	result, _ := json.Marshal(map[string]string{"timeline_id": timelineID, "version_id": versionID})
	job.Status = model.JobCompleted
	job.Progress = 100
	job.StageText = "生成完成"
	job.Result = string(result)
	_ = s.store.UpdateJob(job)
	log.Printf("[timeline] 完成 character=%s nodes=%d mode=%s density=%s",
		characterID, len(allNodes), cfg.CharacterMode, cfg.NarrativeDensity)
}

func (s *CharacterService) failJob(job *model.Job, msg string) {
	log.Printf("[job] 失败 id=%s: %s", job.ID, msg)
	if err := s.store.MarkJobFailed(job.ID, msg); err != nil {
		log.Printf("标记任务失败写入失败 job=%s: %v", job.ID, err)
	}
}

func (s *CharacterService) ListTimelines(characterID string) ([]model.Timeline, error) {
	return s.store.ListTimelines(characterID)
}

func (s *CharacterService) resolveTimelineScope(ch *model.Character, timelineID string) (string, error) {
	if timelineID != "" {
		tl, err := s.store.GetTimeline(timelineID)
		if err != nil {
			return "", err
		}
		if tl.CharacterID != ch.ID {
			return "", fmt.Errorf("时间轴不属于该角色")
		}
		return timelineID, nil
	}
	if ch.CurrentTimelineID != "" {
		return ch.CurrentTimelineID, nil
	}
	list, err := s.store.ListTimelines(ch.ID)
	if err != nil {
		return "", err
	}
	if len(list) == 0 {
		return "", fmt.Errorf("尚无时间轴")
	}
	return list[0].ID, nil
}

func (s *CharacterService) GetTimeline(characterID, timelineID, versionID string) (*model.Timeline, []model.LifeNode, *model.TimelineVersion, error) {
	ch, err := s.store.GetCharacter(characterID)
	if err != nil {
		return nil, nil, nil, err
	}
	tid, err := s.resolveTimelineScope(ch, timelineID)
	if err != nil {
		return nil, nil, nil, err
	}
	tl, err := s.store.GetTimeline(tid)
	if err != nil {
		return nil, nil, nil, err
	}
	if versionID == "" || versionID == "latest" {
		versionID = tl.CurrentVersionID
	}
	if versionID == "" {
		return tl, nil, nil, fmt.Errorf("该时间轴尚无版本")
	}
	version, err := s.store.GetVersion(versionID)
	if err != nil {
		return tl, nil, nil, err
	}
	if version.TimelineID != tid {
		return tl, nil, nil, fmt.Errorf("版本不属于该时间轴")
	}
	nodes, err := s.store.GetNodesByVersion(versionID)
	return tl, nodes, version, err
}

func (s *CharacterService) GetNode(nodeID string) (*model.LifeNode, error) {
	return s.store.GetNode(nodeID)
}

func (s *CharacterService) PatchNodeAndRegenerate(ctx context.Context, characterID, nodeID string, req model.PatchNodeRequest) (*model.Job, error) {
	node, err := s.store.GetNode(nodeID)
	if err != nil {
		return nil, err
	}
	if node.CharacterID != characterID {
		return nil, fmt.Errorf("节点不属于该角色")
	}
	req.Model = ai.NormalizeModelID(req.Model)
	if _, err := s.resolveAPIModel(req.Model); err != nil {
		return nil, err
	}

	node.Title = req.Title
	node.Events = req.Events
	node.Thoughts = req.Thoughts
	node.PersonalitySnapshot = req.PersonalitySnapshot

	mode := req.Mode
	if mode == "" {
		mode = model.PatchModeFullCascade
	}

	ch, err := s.store.GetCharacter(characterID)
	if err != nil {
		return nil, err
	}
	timeline, err := s.store.GetTimelineByVersionID(node.VersionID)
	if err != nil {
		return nil, fmt.Errorf("无法定位节点所属时间轴")
	}
	currentVersionID := timeline.CurrentVersionID
	oldNodes, err := s.store.GetNodesByVersion(currentVersionID)
	if err != nil {
		return nil, err
	}

	jobType := "timeline_regenerate"
	switch mode {
	case model.PatchModeInnerCurrent:
		jobType = "node_inner_current"
	case model.PatchModeInnerSubsequent:
		jobType = "node_inner_subsequent"
	}
	job, err := s.store.CreateJob(characterID, jobType, req.Model)
	if err != nil {
		return nil, err
	}

	switch mode {
	case model.PatchModeInnerCurrent:
		go s.runInnerCurrent(context.Background(), job.ID, characterID, ch, timeline, oldNodes, node, req.Model)
	case model.PatchModeInnerSubsequent:
		locked := store.FilterNodesFromSequence(oldNodes, node.Sequence)
		for i := range locked {
			if locked[i].ID == node.ID {
				locked[i] = *node
				break
			}
		}
		go s.runInnerSubsequent(context.Background(), job.ID, characterID, ch, timeline, locked, node, oldNodes, req.Model)
	default:
		locked := store.FilterNodesFromSequence(oldNodes, node.Sequence)
		for i := range locked {
			if locked[i].ID == node.ID {
				locked[i] = *node
				break
			}
		}
		go s.runRegenerate(context.Background(), job.ID, characterID, ch, timeline, locked, node, oldNodes, req)
	}

	return job, nil
}

func (s *CharacterService) runInnerCurrent(ctx context.Context, jobID, characterID string, ch *model.Character, timeline *model.Timeline, oldNodes []model.LifeNode, edited *model.LifeNode, modelID string) {
	job, _ := s.store.GetJob(jobID)
	job.Status = model.JobRunning
	job.Progress = 15
	job.Model = modelID
	_ = s.store.UpdateJob(job)

	profile, err := s.store.GetProfile(characterID)
	if err != nil {
		s.failJob(job, err.Error())
		return
	}
	system, err := s.ai.LoadPrompt(innerCurrentPromptName(ch.Mode))
	if err != nil {
		s.failJob(job, err.Error())
		return
	}
	oldOne := findNodeBySeq(oldNodes, edited.Sequence)
	oldThoughts, oldPersonality := "", ""
	if oldOne != nil {
		oldThoughts = oldOne.Thoughts
		oldPersonality = oldOne.PersonalitySnapshot
	}
	user := fmt.Sprintf(
		"人物档案（含初始性格、信念、经历摘要等原形）：\n%s\n"+
			"前置人生节点（sequence<%d，含各阶段经历/想法/性格快照）：\n%s\n"+
			"本节点原内心/性格（trait_changes 的 before 参考）：\nthoughts=%s\npersonality_snapshot=%s\n"+
			"本节点（新经历，勿改 events）：sequence=%d year=%d title=%s\nevents=%s",
		store.ProfileJSONInner(profile),
		edited.Sequence, store.MarshalNodesLiteBeforeSequence(oldNodes, edited.Sequence),
		oldThoughts, oldPersonality,
		edited.Sequence, edited.Year, edited.Title, edited.Events,
	)

	setJobStage(s.store, jobID, 30, "正在调用 AI 重算本节点内心…")
	done := make(chan struct{})
	go tickJobProgress(s.store, jobID, 32, 85, done)

	apiModel, _ := s.resolveAPIModel(modelID)
	raw, err := s.ai.ChatJSONModel(ctx, apiModel, system, user)
	close(done)
	if err != nil {
		s.failJob(job, err.Error())
		return
	}
	thoughts, personality, traits, entities, scene, err := store.ParseInnerCurrent(raw, profile.DisplayName)
	if err != nil {
		s.failJob(job, err.Error())
		return
	}

	newVersionID := uuid.New().String()
	allNodes := store.CopyNodesWithNewVersion(oldNodes, newVersionID)
	for i := range allNodes {
		if allNodes[i].Sequence == edited.Sequence {
			allNodes[i].Title = edited.Title
			allNodes[i].Events = edited.Events
			allNodes[i].Thoughts = thoughts
			allNodes[i].PersonalitySnapshot = personality
			allNodes[i].TraitChanges = traits
			allNodes[i].Entities = entities
			if scene != nil {
				allNodes[i].Scene = scene
			}
			break
		}
	}

	version := &model.TimelineVersion{
		ID: newVersionID, CharacterID: characterID, TimelineID: timeline.ID, ParentVersionID: timeline.CurrentVersionID,
		TriggerNodeID: edited.ID, ChangeSummary: fmt.Sprintf("重算节点 #%d 内心与性格", edited.Sequence),
		CreatedAt: time.Now(),
	}
	if err := s.store.CreateVersion(version); err != nil {
		s.failJob(job, err.Error())
		return
	}
	if err := s.store.SaveNodes(allNodes); err != nil {
		s.failJob(job, err.Error())
		return
	}

	diff := store.ComputeVersionDiff(oldNodes, allNodes, edited.ID)
	s.finishRegenerateJob(job, ch, timeline.ID, timeline.CurrentVersionID, newVersionID, diff)
}

func (s *CharacterService) runInnerSubsequent(ctx context.Context, jobID, characterID string, ch *model.Character, timeline *model.Timeline, locked []model.LifeNode, edited *model.LifeNode, oldNodes []model.LifeNode, modelID string) {
	job, _ := s.store.GetJob(jobID)
	job.Status = model.JobRunning
	job.Progress = 15
	job.Model = modelID
	_ = s.store.UpdateJob(job)

	profile, err := s.store.GetProfile(characterID)
	if err != nil {
		s.failJob(job, err.Error())
		return
	}
	tail := store.NodesAfterSequence(oldNodes, edited.Sequence)
	if len(tail) == 0 {
		s.runInnerCurrent(ctx, jobID, characterID, ch, timeline, oldNodes, edited, modelID)
		return
	}

	system, err := s.ai.LoadPrompt(innerSubsequentPromptName(ch.Mode))
	if err != nil {
		s.failJob(job, err.Error())
		return
	}
	user := fmt.Sprintf(
		"人物档案（含初始性格、信念、经历摘要等原形）：\n%s\n"+
			"已发生人生节点（含修改后的锚点，sequence<=%d）：\n%s\n"+
			"后续节点（保留 title/events/year，仅重算内心与性格）：\n%s",
		store.ProfileJSONInner(profile), edited.Sequence,
		store.MarshalNodesLocked(locked), store.MarshalNodesTailInput(tail),
	)

	setJobStage(s.store, jobID, 30, "正在调用 AI 重算后续节点内心…")
	done := make(chan struct{})
	go tickJobProgress(s.store, jobID, 32, 88, done)

	apiModel, _ := s.resolveAPIModel(modelID)
	raw, err := s.ai.ChatJSONModel(ctx, apiModel, system, user)
	close(done)
	if err != nil {
		s.failJob(job, err.Error())
		return
	}
	innerMap, err := store.ParseInnerSubsequent(raw, profile.DisplayName)
	if err != nil {
		s.failJob(job, err.Error())
		return
	}

	newVersionID := uuid.New().String()
	allNodes := make([]model.LifeNode, 0, len(oldNodes))
	for _, n := range locked {
		nn := n
		nn.VersionID = newVersionID
		nn.ID = uuid.New().String()
		allNodes = append(allNodes, nn)
	}
	for _, n := range tail {
		nn := n
		nn.VersionID = newVersionID
		nn.ID = uuid.New().String()
		if inner, ok := innerMap[n.Sequence]; ok {
			nn.Thoughts = inner.Thoughts
			nn.PersonalitySnapshot = inner.PersonalitySnapshot
			nn.TraitChanges = inner.TraitChanges
			nn.Entities = inner.Entities
			if inner.Scene != nil {
				nn.Scene = inner.Scene
			}
		}
		allNodes = append(allNodes, nn)
	}

	version := &model.TimelineVersion{
		ID: newVersionID, CharacterID: characterID, TimelineID: timeline.ID, ParentVersionID: timeline.CurrentVersionID,
		TriggerNodeID: edited.ID,
		ChangeSummary: fmt.Sprintf("重算节点 #%d 之后内心与性格（保留经历）", edited.Sequence),
		CreatedAt: time.Now(),
	}
	if err := s.store.CreateVersion(version); err != nil {
		s.failJob(job, err.Error())
		return
	}
	if err := s.store.SaveNodes(allNodes); err != nil {
		s.failJob(job, err.Error())
		return
	}

	diff := store.ComputeVersionDiff(oldNodes, allNodes, edited.ID)
	s.finishRegenerateJob(job, ch, timeline.ID, timeline.CurrentVersionID, newVersionID, diff)
}

func findNodeBySeq(nodes []model.LifeNode, seq int) *model.LifeNode {
	for i := range nodes {
		if nodes[i].Sequence == seq {
			return &nodes[i]
		}
	}
	return nil
}

func (s *CharacterService) finishRegenerateJob(job *model.Job, ch *model.Character, timelineID, parentVersionID, newVersionID string, diff []model.NodeFieldChange) {
	if diff == nil {
		diff = []model.NodeFieldChange{}
	}
	diffJSON, _ := json.Marshal(model.VersionDiff{
		VersionID: newVersionID, ParentVersionID: parentVersionID, Changes: diff,
	})
	ch.CurrentVersionID = newVersionID
	ch.CurrentTimelineID = timelineID
	_ = s.store.UpdateCharacter(ch)
	_ = s.store.UpdateTimelineCurrentVersion(timelineID, newVersionID)
	job.Status = model.JobCompleted
	job.Progress = 100
	job.StageText = "完成"
	job.Result = string(diffJSON)
	_ = s.store.UpdateJob(job)
}

func (s *CharacterService) runRegenerate(ctx context.Context, jobID, characterID string, ch *model.Character, timeline *model.Timeline, locked []model.LifeNode, edited *model.LifeNode, oldNodes []model.LifeNode, req model.PatchNodeRequest) {
	modelID := req.Model
	job, _ := s.store.GetJob(jobID)
	job.Status = model.JobRunning
	if job.Progress < 25 {
		job.Progress = 15
	}
	job.Model = modelID
	_ = s.store.UpdateJob(job)

	profile, err := s.store.GetProfile(characterID)
	if err != nil {
		s.failJob(job, err.Error())
		return
	}

	// 1. 确定寿命（预览已确认则沿用，否则现场重算）
	job.Progress = 25
	_ = s.store.UpdateJob(job)
	var life *model.LifespanRecalcResult
	if req.ConfirmedDeathYear > 0 {
		if req.ConfirmedDeathYear <= edited.Year {
			s.failJob(job, "确认卒年须大于锚点年份")
			return
		}
		reasoning := req.LifespanReasoning
		if reasoning == "" {
			reasoning = "用户已在预览步骤确认"
		}
		life = &model.LifespanRecalcResult{
			DeathYear:  req.ConfirmedDeathYear,
			DeathCause: req.ConfirmedDeathCause,
			Reasoning:  reasoning,
		}
	} else {
		var errLife error
		life, errLife = s.recalculateLifespan(ctx, profile, locked, edited, modelID)
		if errLife != nil {
			s.failJob(job, "寿命重算失败: "+errLife.Error())
			return
		}
	}
	if life.DeathYear > edited.Year {
		profile.DeathYear = life.DeathYear
		if life.DeathCause != "" {
			profile.DeathCause = life.DeathCause
		}
		if err := s.store.SaveProfile(profile); err != nil {
			s.failJob(job, err.Error())
			return
		}
	}

	stepYears, targetNodes := store.ResolveRegenerateTailRange(edited.Year, profile.DeathYear, req.TargetNodeCount)

	// 2. 完全重新生成后续时间轴
	system, err := s.ai.LoadPrompt(fullTailPromptName(ch.Mode))
	if err != nil {
		s.failJob(job, err.Error())
		return
	}

	lockedJSON := store.MarshalNodesLocked(locked)
	remainingSpan := profile.DeathYear - edited.Year
	if remainingSpan < 0 {
		remainingSpan = 0
	}
	user := fmt.Sprintf(
		"目标约 %d 个后续节点，锚点 year=%d，death_year=%d（剩余跨度 %d 年），参考间隔约 %d 年（非强制，以关键事件为准）\n"+
			"人物档案（含初始性格、信念、经历摘要等原形）：\n%s\n"+
			"已锁定人生节点（含锚点及之前全部经历/想法/性格，须作为后续衔接基准）：\n%s\n"+
			"锚点节点（编辑后）：sequence=%d year=%d age=%d title=%s\nevents=%s\n"+
			"寿命重算说明：%s",
		targetNodes, edited.Year, profile.DeathYear, remainingSpan, stepYears,
		store.ProfileJSONTimeline(profile), lockedJSON,
		edited.Sequence, edited.Year, edited.Age, edited.Title, edited.Events,
		life.Reasoning,
	)

	setJobStage(s.store, jobID, 40, "正在调用 AI 生成后续时间轴…")

	apiModel, err := s.resolveAPIModel(modelID)
	if err != nil {
		s.failJob(job, err.Error())
		return
	}

	done := make(chan struct{})
	go tickJobProgress(s.store, jobID, 42, 88, done)

	raw, err := s.ai.ChatJSONModel(ctx, apiModel, system, user)
	close(done)
	if err != nil {
		s.failJob(job, err.Error())
		return
	}

	newVersionID := uuid.New().String()
	newTail, err := store.ParseTimelineNodes(raw, characterID, newVersionID, profile.DisplayName)
	if err != nil {
		s.failJob(job, err.Error())
		return
	}

	allNodes := make([]model.LifeNode, 0, len(locked)+len(newTail))
	for _, n := range locked {
		nn := n
		nn.VersionID = newVersionID
		nn.ID = uuid.New().String()
		allNodes = append(allNodes, nn)
	}
	for _, n := range newTail {
		n.VersionID = newVersionID
		n.ID = uuid.New().String()
		allNodes = append(allNodes, n)
	}
	version := &model.TimelineVersion{
		ID:              newVersionID,
		CharacterID:     characterID,
		TimelineID:      timeline.ID,
		ParentVersionID: timeline.CurrentVersionID,
		TriggerNodeID:   edited.ID,
		ChangeSummary:   regenerateVersionSummary(req, edited.Sequence, profile.DeathYear),
		CreatedAt:       time.Now(),
	}

	if err := s.store.CreateVersion(version); err != nil {
		s.failJob(job, err.Error())
		return
	}
	if err := s.store.SaveNodes(allNodes); err != nil {
		s.failJob(job, err.Error())
		return
	}

	diff := store.ComputeVersionDiff(oldNodes, allNodes, edited.ID)
	s.finishRegenerateJob(job, ch, timeline.ID, timeline.CurrentVersionID, newVersionID, diff)
}

func (s *CharacterService) ListVersions(characterID, timelineID string) ([]model.TimelineVersion, error) {
	ch, err := s.store.GetCharacter(characterID)
	if err != nil {
		return nil, err
	}
	tid, err := s.resolveTimelineScope(ch, timelineID)
	if err != nil {
		return nil, err
	}
	return s.store.ListVersions(characterID, tid)
}

func (s *CharacterService) GetVersionDiff(characterID, versionID string) (*model.VersionDiff, error) {
	version, err := s.store.GetVersion(versionID)
	if err != nil {
		return nil, err
	}
	if version.CharacterID != characterID {
		return nil, fmt.Errorf("版本不属于该角色")
	}
	if version.ParentVersionID == "" {
		return &model.VersionDiff{VersionID: versionID, ParentVersionID: "", Changes: []model.NodeFieldChange{}}, nil
	}

	newNodes, err := s.store.GetNodesByVersion(versionID)
	if err != nil {
		return nil, err
	}
	oldNodes, err := s.store.GetNodesByVersion(version.ParentVersionID)
	if err != nil {
		return nil, err
	}

	changes := store.ComputeVersionDiff(oldNodes, newNodes, version.TriggerNodeID)
	return &model.VersionDiff{
		VersionID:       versionID,
		ParentVersionID: version.ParentVersionID,
		Changes:         changes,
	}, nil
}

func (s *CharacterService) Rollback(characterID, versionID string) (*model.Character, error) {
	version, err := s.store.GetVersion(versionID)
	if err != nil {
		return nil, err
	}
	if version.CharacterID != characterID {
		return nil, fmt.Errorf("版本不属于该角色")
	}
	if version.TimelineID == "" {
		return nil, fmt.Errorf("版本未关联时间轴")
	}
	ch, err := s.store.GetCharacter(characterID)
	if err != nil {
		return nil, err
	}
	ch.CurrentVersionID = versionID
	ch.CurrentTimelineID = version.TimelineID
	if err := s.store.UpdateCharacter(ch); err != nil {
		return nil, err
	}
	if err := s.store.UpdateTimelineCurrentVersion(version.TimelineID, versionID); err != nil {
		return nil, err
	}
	return ch, nil
}

func (s *CharacterService) GetJob(jobID string) (*model.Job, error) {
	return s.store.GetJob(jobID)
}

func regenerateVersionSummary(req model.PatchNodeRequest, anchorSeq, deathYear int) string {
	if req.ChangeSummary != "" {
		return req.ChangeSummary
	}
	return fmt.Sprintf("编辑节点 #%d 后重算寿命并全新生成后续（卒于%d年）", anchorSeq, deathYear)
}
