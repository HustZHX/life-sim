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
	"life-sim/backend/model"
	"life-sim/backend/store"
)

type GameService struct {
	store *store.Store
	ai    *ai.Client
	char  *CharacterService
}

func NewGameService(st *store.Store, aiClient *ai.Client, charSvc *CharacterService) *GameService {
	return &GameService{store: st, ai: aiClient, char: charSvc}
}

func gamePromptSuffix(mode string) string {
	if mode == model.ModeFamous {
		return "famous"
	}
	return "random"
}

func (g *GameService) ensureGameCharacter(characterID string) (*model.Character, error) {
	ch, err := g.store.GetCharacter(characterID)
	if err != nil {
		return nil, err
	}
	if ch.PlayStyle != model.PlayStyleGame {
		return nil, fmt.Errorf("该角色不是人生游戏模式")
	}
	if ch.GameConfig == nil {
		ch.GameConfig = &model.GameConfig{}
	}
	return ch, nil
}

func (g *GameService) CreateGameCharacter(mode string) (*model.Character, error) {
	if mode != model.ModeFamous && mode != model.ModeRandom {
		return nil, fmt.Errorf("无效模式: %s", mode)
	}
	return g.store.CreateCharacterWithStyle(mode, model.PlayStyleGame)
}

func (g *GameService) UpdateGameConfig(characterID string, cfg model.GameConfig) (*model.Character, error) {
	ch, err := g.ensureGameCharacter(characterID)
	if err != nil {
		return nil, err
	}
	merged := mergeGameConfig(ch.GameConfig, cfg)
	ch.GameConfig = &merged
	if err := g.store.UpdateCharacter(ch); err != nil {
		return nil, err
	}
	return ch, nil
}

func mergeGameConfig(prev *model.GameConfig, patch model.GameConfig) model.GameConfig {
	var out model.GameConfig
	if prev != nil {
		out = *prev
	}
	if patch.EraDescription != "" {
		out.EraDescription = patch.EraDescription
	}
	if len(patch.EraOptions) > 0 {
		out.EraOptions = patch.EraOptions
	}
	if patch.SelectedEra != "" {
		out.SelectedEra = patch.SelectedEra
	}
	if patch.StartYearHint > 0 {
		out.StartYearHint = patch.StartYearHint
	}
	if patch.BirthBackground != "" {
		out.BirthBackground = patch.BirthBackground
	}
	return out
}

func (g *GameService) GenerateEraOptions(ctx context.Context, characterID string, req model.GameEraOptionsRequest) ([]model.GameEraOption, error) {
	ch, err := g.ensureGameCharacter(characterID)
	if err != nil {
		return nil, err
	}
	if ch.Mode != model.ModeFamous {
		return nil, fmt.Errorf("仅历史名人支持时期选项")
	}
	if ch.Status != model.StatusConfirmed && ch.Status != model.StatusProfileReady && ch.Status != model.StatusTimelineReady {
		return nil, fmt.Errorf("请先确认人物身份")
	}
	desc := strings.TrimSpace(req.EraDescription)
	if ch.GameConfig == nil {
		ch.GameConfig = &model.GameConfig{}
	}
	if !req.Regenerate && len(ch.GameConfig.EraOptions) > 0 && strings.TrimSpace(ch.GameConfig.EraDescription) == desc {
		return ch.GameConfig.EraOptions, nil
	}

	modelID := ai.NormalizeModelID(req.Model)
	apiModel, err := g.char.resolveAPIModel(modelID)
	if err != nil {
		return nil, err
	}
	system, err := g.ai.LoadPrompt("game_era_options.txt")
	if err != nil {
		return nil, err
	}
	var user string
	if desc == "" {
		user = fmt.Sprintf(`历史人物：%s
已确认身份：%s

【用户想从什么时期开始】
（未填写 — 请基于该人物一生，给出 3～4 个最具代表性的游戏起点时期）

【输出要求】
- 只返回 3 或 4 个 options
- 不要输出「从出生开始」类选项（系统会单独添加）
- 选项须符合史实、年份准确、彼此区分明显，勿用童年/青年/壮年/晚年等泛泛套话凑数`,
			ch.DisplayName, ch.ConfirmedIdentity)
	} else {
		user = fmt.Sprintf(`历史人物：%s
已确认身份：%s

【用户想从什么时期开始 — 必须严格据此筛选】
%s

【输出要求】
- 只返回 3 或 4 个 options，且每个都必须与用户上述描述直接相关
- 不要输出「从出生开始」类选项（系统会单独添加）
- 禁止返回与该描述无关的其他人生阶段或泛泛的生平分期
- 若用户描述具体（如某战役前后、某作品问世时），选项应围绕该主题，不要凑满一生各时期`,
			ch.DisplayName, ch.ConfirmedIdentity, desc)
	}
	raw, err := g.ai.ChatJSONModel(ctx, apiModel, system, user)
	if err != nil {
		return nil, err
	}
	opts, err := store.ParseGameEraOptions(raw)
	if err != nil {
		return nil, err
	}
	birthYear := g.resolveBirthYear(characterID, ch)
	opts = ensureBirthEraOption(opts, birthYear)
	ch.GameConfig.EraDescription = desc
	ch.GameConfig.EraOptions = opts
	if err := g.store.UpdateCharacter(ch); err != nil {
		return nil, err
	}
	return opts, nil
}

const birthEraLabel = "从出生开始"

func birthYearFromConfirmedIdentity(identityJSON string) int {
	if strings.TrimSpace(identityJSON) == "" {
		return 0
	}
	var c model.ResolveCandidate
	if err := json.Unmarshal([]byte(identityJSON), &c); err != nil {
		return 0
	}
	return c.BirthYear
}

func (g *GameService) resolveBirthYear(characterID string, ch *model.Character) int {
	if y := birthYearFromConfirmedIdentity(ch.ConfirmedIdentity); y > 0 {
		return y
	}
	profile, err := g.store.GetProfile(characterID)
	if err == nil && profile != nil && profile.BirthYear > 0 {
		return profile.BirthYear
	}
	return 0
}

func ensureBirthEraOption(opts []model.GameEraOption, birthYear int) []model.GameEraOption {
	filtered := make([]model.GameEraOption, 0, len(opts))
	for _, o := range opts {
		if isBirthEraOption(o, birthYear) {
			continue
		}
		filtered = append(filtered, o)
	}
	if len(filtered) > 4 {
		filtered = filtered[:4]
	}
	yearRange := "出生年起"
	if birthYear > 0 {
		yearRange = fmt.Sprintf("%d—", birthYear)
	}
	birth := model.GameEraOption{
		Label:       birthEraLabel,
		YearRange:   yearRange,
		StartYear:   birthYear,
		Description: "从出生起完整体验其一生，从幼年经历、成长与抉择一路展开（史实生卒年起点）。",
	}
	out := make([]model.GameEraOption, 0, len(filtered)+1)
	out = append(out, birth)
	return append(out, filtered...)
}

func isBirthEraOption(o model.GameEraOption, birthYear int) bool {
	label := strings.TrimSpace(o.Label)
	if label == birthEraLabel {
		return true
	}
	if strings.Contains(label, "出生") && strings.Contains(label, "开始") {
		return true
	}
	if birthYear > 0 && o.StartYear == birthYear {
		if strings.Contains(label, "出生") || strings.Contains(o.Description, "从出生") {
			return true
		}
	}
	return false
}

func parseConfirmedIdentity(identityJSON string) *model.ResolveCandidate {
	if strings.TrimSpace(identityJSON) == "" {
		return nil
	}
	var c model.ResolveCandidate
	if err := json.Unmarshal([]byte(identityJSON), &c); err != nil {
		return nil
	}
	return &c
}

func findSelectedEraOption(cfg *model.GameConfig) *model.GameEraOption {
	if cfg == nil || cfg.SelectedEra == "" {
		return nil
	}
	for i := range cfg.EraOptions {
		if cfg.EraOptions[i].Label == cfg.SelectedEra {
			return &cfg.EraOptions[i]
		}
	}
	return nil
}

func isFamousBirthStart(ch *model.Character) bool {
	if ch.GameConfig == nil {
		return false
	}
	birthYear := birthYearFromConfirmedIdentity(ch.ConfirmedIdentity)
	if birthYear <= 0 {
		return false
	}
	if ch.GameConfig.SelectedEra == birthEraLabel {
		return true
	}
	if ch.GameConfig.StartYearHint > 0 && ch.GameConfig.StartYearHint == birthYear {
		return true
	}
	if opt := findSelectedEraOption(ch.GameConfig); opt != nil && isBirthEraOption(*opt, birthYear) {
		return true
	}
	return false
}

func buildFamousBirthProfileUserPrompt(ch *model.Character) string {
	var b strings.Builder
	b.WriteString("【从出生开始 — 仅出生设定】\n")
	if cand := parseConfirmedIdentity(ch.ConfirmedIdentity); cand != nil {
		b.WriteString(fmt.Sprintf("姓名：%s\n史实生年：%d\n", cand.Name, cand.BirthYear))
		if cand.DeathYear > 0 {
			b.WriteString(fmt.Sprintf("史实卒年：%d\n", cand.DeathYear))
		}
		if cand.Nationality != "" {
			b.WriteString(fmt.Sprintf("国籍/朝代：%s\n", cand.Nationality))
		}
		if cand.Summary != "" {
			b.WriteString(fmt.Sprintf("身份摘要：%s\n", cand.Summary))
		}
	} else {
		b.WriteString(fmt.Sprintf("已确认身份：%s\n", ch.ConfirmedIdentity))
	}
	b.WriteString("请按出生时刻填写档案；性格、经历、格言、职业、主要成就等保持空字符串。\n")
	return b.String()
}

func buildFamousStageProfileUserPrompt(ch *model.Character) string {
	cfg := ch.GameConfig
	var b strings.Builder
	b.WriteString("【已确认身份】\n")
	if cand := parseConfirmedIdentity(ch.ConfirmedIdentity); cand != nil {
		b.WriteString(fmt.Sprintf("姓名：%s\n史实生年：%d\n", cand.Name, cand.BirthYear))
		if cand.DeathYear > 0 {
			b.WriteString(fmt.Sprintf("史实卒年：%d\n", cand.DeathYear))
		}
		if cand.Nationality != "" {
			b.WriteString(fmt.Sprintf("国籍/朝代：%s\n", cand.Nationality))
		}
		if cand.Summary != "" {
			b.WriteString(fmt.Sprintf("身份摘要：%s\n", cand.Summary))
		}
	} else {
		b.WriteString(ch.ConfirmedIdentity + "\n")
	}
	b.WriteString("\n【游戏起点时期 — 档案内容截止于此，不得超出】\n")
	b.WriteString(fmt.Sprintf("时期名称：%s\n", cfg.SelectedEra))
	if cfg.StartYearHint > 0 {
		b.WriteString(fmt.Sprintf("起点年份（约）：%d\n", cfg.StartYearHint))
	}
	if opt := findSelectedEraOption(cfg); opt != nil {
		if opt.YearRange != "" {
			b.WriteString(fmt.Sprintf("年份范围：%s\n", opt.YearRange))
		}
		if opt.Description != "" {
			b.WriteString(fmt.Sprintf("时期说明：%s\n", opt.Description))
		}
	}
	if desc := strings.TrimSpace(cfg.EraDescription); desc != "" {
		b.WriteString(fmt.Sprintf("用户补充描述：%s\n", desc))
	}
	b.WriteString("\n【生成要求】\n")
	b.WriteString("personality_initial、experiences、beliefs_motto、family_relations、occupation、social_class、major_works、education 等必须仅反映截至起点年份当时的状态；禁止写入该年之后的任何事迹、作品或关系。\n")
	return b.String()
}

func (g *GameService) GenerateGameProfile(ctx context.Context, characterID string, req model.GameProfileGenerateRequest) (*model.Profile, error) {
	ch, err := g.ensureGameCharacter(characterID)
	if err != nil {
		return nil, err
	}
	modelID := ai.NormalizeModelID(req.Model)
	apiModel, err := g.char.resolveAPIModel(modelID)
	if err != nil {
		return nil, err
	}

	var system, user string
	switch ch.Mode {
	case model.ModeFamous:
		if ch.Status != model.StatusConfirmed {
			return nil, fmt.Errorf("请先确认人物身份")
		}
		if ch.GameConfig == nil || ch.GameConfig.SelectedEra == "" {
			return nil, fmt.Errorf("请先选择历史时期")
		}
		if ch.GameConfig.StartYearHint <= 0 {
			return nil, fmt.Errorf("请先保存所选时期的起点年份")
		}
		if isFamousBirthStart(ch) {
			system, err = g.ai.LoadPrompt("fetch_profile_famous_birth.txt")
			if err != nil {
				return nil, err
			}
			user = buildFamousBirthProfileUserPrompt(ch)
		} else {
			system, err = g.ai.LoadPrompt("fetch_profile_famous_stage.txt")
			if err != nil {
				return nil, err
			}
			user = buildFamousStageProfileUserPrompt(ch)
		}
	case model.ModeRandom:
		system, err = g.ai.LoadPrompt("fetch_profile_random_birth.txt")
		if err != nil {
			return nil, err
		}
		bg := req.BirthBackground
		if bg == "" && ch.GameConfig != nil {
			bg = ch.GameConfig.BirthBackground
		}
		name := req.DisplayName
		if name == "" {
			name = "小明"
		}
		user = fmt.Sprintf("指定姓名：%s\n出生背景：%s", name, bg)
		if ch.GameConfig != nil && ch.GameConfig.BirthBackground != "" {
			ch.GameConfig.BirthBackground = bg
		}
	default:
		return nil, fmt.Errorf("未知模式")
	}

	raw, err := g.ai.ChatJSONModel(ctx, apiModel, system, user)
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
	if err := g.store.SaveProfile(profile); err != nil {
		return nil, err
	}
	g.char.aiCache.invalidate(characterID)
	ch.Status = model.StatusProfileReady
	_ = g.store.UpdateCharacter(ch)
	return profile, nil
}

func (g *GameService) StartTimelineJob(ctx context.Context, characterID string, req model.GameTimelineStartRequest) (*model.Job, error) {
	ch, err := g.ensureGameCharacter(characterID)
	if err != nil {
		return nil, err
	}
	if ch.Status != model.StatusProfileReady && ch.Status != model.StatusTimelineReady {
		return nil, fmt.Errorf("请先生成人物档案")
	}
	profile, err := g.store.GetProfile(characterID)
	if err != nil {
		return nil, err
	}
	modelID := ai.NormalizeModelID(req.Model)
	if _, err := g.char.resolveAPIModel(modelID); err != nil {
		return nil, err
	}

	timelineID := uuid.New().String()
	title := req.Title
	if title == "" {
		title = "人生游戏"
	}
	if err := g.store.CreateTimeline(&model.Timeline{ID: timelineID, CharacterID: characterID, Title: title}); err != nil {
		return nil, err
	}

	jobJSON, _ := json.Marshal(model.GameTimelineStartJobRequest{TimelineID: timelineID})
	job, err := g.store.CreateJobWithRequest(characterID, "game_timeline_start", modelID, string(jobJSON))
	if err != nil {
		return nil, err
	}
	go g.runGameTimelineStart(context.Background(), job.ID, characterID, timelineID, profile, modelID, title)
	return job, nil
}

func (g *GameService) runGameTimelineStart(ctx context.Context, jobID, characterID, timelineID string, profile *model.Profile, modelID, title string) {
	job, _ := g.store.GetJob(jobID)
	job.Status = model.JobRunning
	job.Progress = 10
	job.StageText = "正在生成首节点…"
	_ = g.store.UpdateJob(job)

	ch, err := g.store.GetCharacter(characterID)
	if err != nil {
		g.char.failJob(job, err.Error())
		return
	}
	apiModel, err := g.char.resolveAPIModel(modelID)
	if err != nil {
		g.char.failJob(job, err.Error())
		return
	}

	profileJSON := g.char.aiCache.getProfileTimelineJSON(characterID, profile)
	profileHash := store.ProfileHash(profile)
	birthYear := profile.BirthYear
	targetYear := birthYear
	if ch.Mode == model.ModeFamous && ch.GameConfig != nil && ch.GameConfig.StartYearHint > 0 {
		targetYear = ch.GameConfig.StartYearHint
	}
	wlStart := birthYear
	if wlStart <= 0 {
		wlStart = targetYear
	}
	endYear := profile.DeathYear
	if endYear <= 0 {
		endYear = targetYear + 80
	}
	cfg := timelineJobConfig{
		StartYear:       wlStart,
		EndYear:         endYear,
		TargetNodeCount: gameTimelineStageTargetCount(birthYear, targetYear),
		CharacterMode:   ch.Mode,
	}
	gameCfgJSON, _ := json.Marshal(ch.GameConfig)
	eraCtx, eraErr := resolveEraContext(ctx, g.char, characterID, ch.Mode, profileJSON, profileHash, apiModel, cfg, "")
	if eraErr != nil {
		log.Printf("[game] 时代背景失败: %v", eraErr)
	}
	eraCtx += "\n【游戏配置】" + string(gameCfgJSON)

	versionID := uuid.New().String()
	multiStage := ch.Mode == model.ModeFamous && targetYear > birthYear && birthYear > 0
	var nodes []model.LifeNode
	var genErr error
	if multiStage {
		setJobStage(g.store, jobID, 25, fmt.Sprintf("正在推演至 %d 年阶段（%d 个节点）…", targetYear, cfg.TargetNodeCount))
		nodes, genErr = g.generateGameTimelineToStage(ctx, apiModel, ch, profile, characterID, versionID, birthYear, targetYear, string(gameCfgJSON))
	} else {
		setJobStage(g.store, jobID, 25, "正在推演开场节点…")
		startYear := birthYear
		if startYear <= 0 {
			startYear = targetYear
		}
		nodes, genErr = g.generateGameFirstNode(ctx, apiModel, ch, profile, characterID, versionID, startYear, string(gameCfgJSON))
	}
	if genErr != nil {
		g.char.failJob(job, genErr.Error())
		return
	}

	changeSummary := "人生游戏 · 首节点"
	if len(nodes) > 1 {
		changeSummary = fmt.Sprintf("人生游戏 · 开局至 %d 年（%d 节点）", store.MaxNodeYear(nodes), len(nodes))
	}
	version := &model.TimelineVersion{
		ID:              versionID,
		CharacterID:     characterID,
		TimelineID:      timelineID,
		BranchLabel:     "游戏开局",
		ChangeSummary:   changeSummary,
		DeathYearSnapshot: store.MaxNodeYear(nodes),
		CreatedAt:       time.Now(),
	}
	if err := g.store.CreateVersion(version); err != nil {
		g.char.failJob(job, err.Error())
		return
	}
	if err := g.store.SaveNodes(nodes); err != nil {
		g.char.failJob(job, err.Error())
		return
	}

	setJobStage(g.store, jobID, 70, "正在生成世界线…")
	wlEnd := store.MaxNodeYear(nodes)
	if wl, wlErr := g.char.generateWorldLineForTimeline(ctx, ch.Mode, profile, apiModel, wlStart, wlEnd, nodes, eraCtx); wlErr != nil {
		log.Printf("[game] 世界线生成失败: %v", wlErr)
	} else {
		wl.TimelineID = timelineID
		_ = g.char.persistWorldLine(timelineID, wl)
	}

	lastSeq := nodes[len(nodes)-1].Sequence
	_ = g.store.SaveGameProfileSnapshot(characterID, timelineID, lastSeq, profile)

	if err := g.store.UpdateTimelineCurrentVersion(timelineID, versionID); err != nil {
		g.char.failJob(job, err.Error())
		return
	}
	ch.CurrentVersionID = versionID
	ch.CurrentTimelineID = timelineID
	ch.Status = model.StatusTimelineReady
	_ = g.store.UpdateCharacter(ch)

	result, _ := json.Marshal(map[string]string{"timeline_id": timelineID, "version_id": versionID})
	job.Status = model.JobCompleted
	job.Progress = 100
	job.StageText = "完成"
	job.Result = string(result)
	_ = g.store.UpdateJob(job)

	if len(nodes) > 0 {
		last := nodes[len(nodes)-1]
		go g.prefetchNodeChoices(context.Background(), characterID, last.ID, versionID, modelID)
	}
}

func (g *GameService) prefetchNodeChoices(ctx context.Context, characterID, nodeID, versionID, modelID string) {
	if err := g.generateAndSaveNodeChoices(ctx, characterID, nodeID, versionID, modelID); err != nil {
		log.Printf("[game] 预生成抉择选项失败 node=%s: %v", nodeID, err)
	}
}

func (g *GameService) generateAndSaveNodeChoices(ctx context.Context, characterID, nodeID, versionID, modelID string) error {
	ch, err := g.ensureGameCharacter(characterID)
	if err != nil {
		return err
	}
	node, err := g.store.GetNode(nodeID)
	if err != nil {
		return err
	}
	if node.VersionID != versionID {
		return fmt.Errorf("节点版本不匹配")
	}
	profile, err := g.store.GetProfile(characterID)
	if err != nil {
		return err
	}
	apiModel, err := g.char.resolveAPIModel(ai.NormalizeModelID(modelID))
	if err != nil {
		return err
	}
	promptName := "game_node_choices_" + gamePromptSuffix(ch.Mode)
	system, err := g.ai.LoadPrompt(promptName)
	if err != nil {
		return err
	}
	user := buildGameNodeChoicesUser(ch, profile, node)
	raw, err := g.ai.ChatJSONModel(ctx, apiModel, system, user)
	if err != nil {
		return err
	}
	opts, err := store.ParseGameChoices(raw)
	if err != nil {
		return err
	}
	age := resolveNodeAge(profile, node)
	if err := store.ValidateGameChoicesForAge(age, opts); err != nil {
		retryUser := user + "\n\n【纠正】上次选项违反年龄/身份约束：" + err.Error() +
			"。请重新生成 3～5 个选项，每个均须对当前年龄与社会角色可行。"
		raw, err = g.ai.ChatJSONModel(ctx, apiModel, system, retryUser)
		if err != nil {
			return err
		}
		opts, err = store.ParseGameChoices(raw)
		if err != nil {
			return err
		}
		if err2 := store.ValidateGameChoicesForAge(age, opts); err2 != nil {
			log.Printf("[game] 抉择选项仍不符合年龄约束 node=%s age=%d: %v", nodeID, age, err2)
		}
	}
	return g.store.SaveGameNodeChoices(characterID, nodeID, versionID, node.Sequence, &model.GameNodeChoiceRecord{Options: opts})
}

func (g *GameService) ListChoiceDisplays(versionID string, nodes []model.LifeNode) ([]model.GameNodeChoiceDisplay, error) {
	rows, err := g.store.ListGameNodeChoicesByVersion(versionID)
	if err != nil {
		return nil, err
	}
	out := make([]model.GameNodeChoiceDisplay, 0)
	for _, row := range rows {
		text := resolveGameChoiceDisplayText(&row.Record)
		if text == "" {
			continue
		}
		nodeID := row.NodeID
		if id := findNodeIDBySequence(nodes, row.NodeSequence); id != "" {
			nodeID = id
		}
		out = append(out, model.GameNodeChoiceDisplay{
			NodeID: nodeID, NodeSequence: row.NodeSequence, DisplayText: text,
		})
	}
	return out, nil
}

func resolveGameChoiceDisplayText(rec *model.GameNodeChoiceRecord) string {
	if rec == nil {
		return ""
	}
	if t := strings.TrimSpace(rec.CustomText); t != "" {
		return t
	}
	if rec.ChosenID == "" {
		return ""
	}
	for _, o := range rec.Options {
		if o.ID == rec.ChosenID {
			if o.Description != "" {
				return o.Label + "：" + o.Description
			}
			return o.Label
		}
	}
	return rec.ChosenID
}

func findNodeIDBySequence(nodes []model.LifeNode, seq int) string {
	for i := range nodes {
		if nodes[i].Sequence == seq {
			return nodes[i].ID
		}
	}
	return ""
}

func (g *GameService) GetNodeChoices(ctx context.Context, characterID, nodeID, modelID string, regenerate bool) (*model.GameChoicesResponse, error) {
	if _, err := g.ensureGameCharacter(characterID); err != nil {
		return nil, err
	}
	node, err := g.store.GetNode(nodeID)
	if err != nil {
		return nil, fmt.Errorf("节点不存在")
	}
	if node.CharacterID != characterID {
		return nil, fmt.Errorf("节点不属于该角色")
	}
	timeline, err := g.store.GetTimelineByVersionID(node.VersionID)
	if err != nil {
		return nil, err
	}
	if timeline.CurrentVersionID != node.VersionID {
		return nil, fmt.Errorf("只能对当前激活分支的末节点生成抉择")
	}
	nodes, err := g.store.GetNodesByVersion(node.VersionID)
	if err != nil {
		return nil, err
	}
	last := nodes[len(nodes)-1]
	if last.ID != nodeID {
		return nil, fmt.Errorf("只能对时间轴最后一个节点生成抉择")
	}

	if !regenerate {
		if rec, _, err := g.store.GetGameNodeChoices(nodeID, node.VersionID); err == nil {
			resp := &model.GameChoicesResponse{Options: rec.Options}
			if rec.ChosenID != "" || rec.CustomText != "" {
				resp.ChosenID = rec.ChosenID
				resp.CustomText = rec.CustomText
				resp.AlreadyChosen = true
			}
			return resp, nil
		}
	}

	if regenerate {
		if err := g.generateAndSaveNodeChoices(ctx, characterID, nodeID, node.VersionID, modelID); err != nil {
			return nil, err
		}
	} else if _, _, err := g.store.GetGameNodeChoices(nodeID, node.VersionID); err != nil {
		return nil, fmt.Errorf("抉择选项尚未生成，请稍后点击「刷新选项」")
	}
	rec, _, err := g.store.GetGameNodeChoices(nodeID, node.VersionID)
	if err != nil {
		return nil, err
	}
	resp := &model.GameChoicesResponse{Options: rec.Options}
	if rec.ChosenID != "" || rec.CustomText != "" {
		resp.ChosenID = rec.ChosenID
		resp.CustomText = rec.CustomText
		resp.AlreadyChosen = true
	}
	return resp, nil
}

func (g *GameService) ChooseJob(ctx context.Context, characterID, nodeID string, req model.GameChooseRequest) (*model.Job, error) {
	ch, err := g.ensureGameCharacter(characterID)
	if err != nil {
		return nil, err
	}
	if req.ChoiceID == "" && strings.TrimSpace(req.CustomText) == "" {
		return nil, fmt.Errorf("请选择选项或填写自定义抉择")
	}
	node, err := g.store.GetNode(nodeID)
	if err != nil {
		return nil, fmt.Errorf("节点不存在")
	}
	timeline, err := g.store.GetTimelineByVersionID(node.VersionID)
	if err != nil {
		return nil, err
	}
	if timeline.CurrentVersionID != node.VersionID {
		return nil, fmt.Errorf("当前不在激活分支")
	}
	nodes, err := g.store.GetNodesByVersion(node.VersionID)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 || nodes[len(nodes)-1].ID != nodeID {
		return nil, fmt.Errorf("只能对末节点做出抉择")
	}
	if rec, _, err := g.store.GetGameNodeChoices(nodeID, node.VersionID); err == nil && (rec.ChosenID != "" || rec.CustomText != "") {
		return nil, fmt.Errorf("该节点已做出抉择，请回退后重试")
	}

	modelID := ai.NormalizeModelID(req.Model)
	jobJSON, _ := json.Marshal(model.GameChooseJobRequest{
		TimelineID: timeline.ID, NodeID: nodeID, ChoiceID: req.ChoiceID, CustomText: req.CustomText,
	})
	job, err := g.store.CreateJobWithRequest(characterID, "game_choose", modelID, string(jobJSON))
	if err != nil {
		return nil, err
	}
	go g.runGameChoose(context.Background(), job.ID, characterID, timeline, node, nodes, req, ch.Mode)
	return job, nil
}

func (g *GameService) runGameChoose(ctx context.Context, jobID, characterID string, timeline *model.Timeline, anchor *model.LifeNode, locked []model.LifeNode, req model.GameChooseRequest, mode string) {
	job, _ := g.store.GetJob(jobID)
	job.Status = model.JobRunning
	job.Progress = 15
	job.StageText = "正在推演抉择后果…"
	_ = g.store.UpdateJob(job)

	profile, err := g.store.GetProfile(characterID)
	if err != nil {
		g.char.failJob(job, err.Error())
		return
	}
	ch, err := g.ensureGameCharacter(characterID)
	if err != nil {
		ch = &model.Character{Mode: mode}
	}
	apiModel, err := g.char.resolveAPIModel(job.Model)
	if err != nil {
		g.char.failJob(job, err.Error())
		return
	}

	choiceLabel := strings.TrimSpace(req.CustomText)
	if req.ChoiceID != "" {
		if rec, _, err := g.store.GetGameNodeChoices(anchor.ID, anchor.VersionID); err == nil {
			for _, o := range rec.Options {
				if o.ID == req.ChoiceID {
					choiceLabel = o.Label + "：" + o.Description
					break
				}
			}
		}
	}
	if choiceLabel == "" {
		choiceLabel = req.ChoiceID
	}

	promptName := "game_apply_choice_" + gamePromptSuffix(mode)
	system, err := g.ai.LoadPrompt(promptName)
	if err != nil {
		g.char.failJob(job, err.Error())
		return
	}
	prevWL, _ := g.store.GetWorldLine(timeline.ID)
	lastWLYear := 0
	if prevWL != nil {
		for _, ev := range prevWL.Events {
			if ev.Year > lastWLYear {
				lastWLYear = ev.Year
			}
		}
	}
	wlJSON := "{}"
	if prevWL != nil {
		wlJSON = store.MarshalWorldLineForPrompt(prevWL)
	}
	user := buildGameApplyChoiceUser(choiceLabel, lastWLYear, wlJSON, ch, profile, anchor)

	setJobStage(g.store, jobID, 35, "正在调用 AI…")
	raw, err := g.ai.ChatJSONModelLong(ctx, apiModel, system, user)
	if err != nil {
		g.char.failJob(job, err.Error())
		return
	}
	payload, err := store.ParseGameApplyChoice(raw)
	if err != nil {
		g.char.failJob(job, err.Error())
		return
	}

	if err := store.ApplyProfileUpdatesFromJSON(profile, payload.ProfileUpdates); err != nil {
		g.char.failJob(job, err.Error())
		return
	}
	if err := g.store.SaveProfile(profile); err != nil {
		g.char.failJob(job, err.Error())
		return
	}
	g.char.aiCache.invalidate(characterID)

	nextSeq := anchor.Sequence + 1
	nextNode := model.LifeNode{
		ID:                  uuid.New().String(),
		CharacterID:         characterID,
		Sequence:            nextSeq,
		Year:                payload.NextNode.Year,
		Age:                 payload.NextNode.Age,
		Title:               payload.NextNode.Title,
		Events:              payload.NextNode.Events,
		Thoughts:            payload.NextNode.Thoughts,
		PersonalitySnapshot: payload.NextNode.PersonalitySnapshot,
		TraitChanges:        store.NormalizeTraitChanges(payload.NextNode.TraitChanges),
		Entities:            store.NormalizeNodeEntities(payload.NextNode.Entities, profile.DisplayName),
		Scene:               payload.NextNode.Scene,
	}
	if nextNode.Age <= 0 && profile.BirthYear > 0 && nextNode.Year > 0 {
		nextNode.Age = nextNode.Year - profile.BirthYear
	}

	newVersionID := uuid.New().String()
	nextNode.VersionID = newVersionID
	allNodes := make([]model.LifeNode, 0, len(locked)+1)
	for _, n := range locked {
		nn := n
		nn.VersionID = newVersionID
		nn.ID = uuid.New().String()
		allNodes = append(allNodes, nn)
	}
	allNodes = append(allNodes, nextNode)

	version := &model.TimelineVersion{
		ID:                newVersionID,
		CharacterID:       characterID,
		TimelineID:        timeline.ID,
		ParentVersionID:   timeline.CurrentVersionID,
		TriggerNodeID:     anchor.ID,
		BranchLabel:       "抉择 · " + truncateLabel(choiceLabel, 24),
		ForkSequence:      anchor.Sequence,
		ForkNodeID:        anchor.ID,
		ChangeSummary:     "人生抉择 · +" + fmt.Sprintf("%d", 1),
		DeathYearSnapshot: store.MaxNodeYear(allNodes),
		CreatedAt:         time.Now(),
	}
	if err := g.store.CreateVersion(version); err != nil {
		g.char.failJob(job, err.Error())
		return
	}
	if err := g.store.SaveNodes(allNodes); err != nil {
		g.char.failJob(job, err.Error())
		return
	}

	if !store.ShouldSkipWorldLineDelta(payload, lastWLYear) {
		if prevWL == nil {
			prevWL = &model.WorldLine{TimelineID: timeline.ID}
		}
		delta := &model.WorldLine{
			EraSummary:       payload.WorldLineDelta.EraSummary,
			HistoricalTrend:  payload.WorldLineDelta.HistoricalTrend,
			DailyLifeContext: payload.WorldLineDelta.DailyLifeContext,
			Events:           payload.WorldLineDelta.Events,
		}
		store.AppendWorldLineDelta(prevWL, delta, nextSeq)
		prevWL.UpdatedAt = time.Now()
		_ = g.char.persistWorldLine(timeline.ID, prevWL)
	}

	_ = g.store.SaveGameProfileSnapshot(characterID, timeline.ID, nextSeq, profile)
	anchorNewID := findNodeIDBySequence(allNodes, anchor.Sequence)
	if anchorNewID == "" {
		anchorNewID = anchor.ID
	}
	var savedOpts []model.GameChoiceOption
	if rec, _, err := g.store.GetGameNodeChoices(anchor.ID, anchor.VersionID); err == nil {
		savedOpts = rec.Options
	} else if rec, _, err := g.store.GetGameNodeChoices(anchorNewID, timeline.CurrentVersionID); err == nil {
		savedOpts = rec.Options
	}
	_ = g.store.SaveGameNodeChoices(characterID, anchorNewID, newVersionID, anchor.Sequence, &model.GameNodeChoiceRecord{
		Options:    savedOpts,
		ChosenID:   req.ChoiceID,
		CustomText: req.CustomText,
	})

	parentVersionID := timeline.CurrentVersionID
	_ = g.store.CopyMemoriesWithNewVersion(parentVersionID, newVersionID)
	_ = g.store.CopyDialogueIdentityPresetsWithNewVersion(parentVersionID, newVersionID)

	if refreshed, err := g.store.GetCharacter(characterID); err == nil {
		ch = refreshed
	}
	ch.CurrentVersionID = newVersionID
	ch.CurrentTimelineID = timeline.ID
	_ = g.store.UpdateCharacter(ch)
	_ = g.store.UpdateTimelineCurrentVersion(timeline.ID, newVersionID)

	result, _ := json.Marshal(map[string]string{"version_id": newVersionID, "node_id": nextNode.ID})
	job.Status = model.JobCompleted
	job.Progress = 100
	job.StageText = "完成"
	job.Result = string(result)
	_ = g.store.UpdateJob(job)

	go g.prefetchNodeChoices(context.Background(), characterID, nextNode.ID, newVersionID, job.Model)
}

func truncateLabel(s string, max int) string {
	s = strings.TrimSpace(s)
	if len([]rune(s)) <= max {
		return s
	}
	r := []rune(s)
	return string(r[:max]) + "…"
}

// ApplyGameRollback 回退后恢复游戏档案与世界线、清理抉择缓存。
func (g *GameService) ApplyGameRollback(characterID string, timelineID string, targetSeq, targetYear int) error {
	if snap, err := g.store.GetGameProfileSnapshot(timelineID, targetSeq); err == nil {
		snap.CharacterID = characterID
		if err := g.store.SaveProfile(snap); err != nil {
			return err
		}
		g.char.aiCache.invalidate(characterID)
	}
	_ = g.store.DeleteGameProfileSnapshotsAfter(timelineID, targetSeq)
	_ = g.store.DeleteGameNodeChoicesAfter(characterID, targetSeq)
	if wl, err := g.store.GetWorldLine(timelineID); err == nil && wl != nil {
		store.TruncateWorldLineEvents(wl, targetYear)
		_ = g.char.persistWorldLine(timelineID, wl)
	}
	return nil
}
