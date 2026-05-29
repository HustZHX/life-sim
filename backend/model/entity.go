package model

import "time"

const (
	ModeFamous = "famous"
	ModeRandom = "random"

	PlayStyleSimulation = "simulation"
	PlayStyleGame       = "game"

	StatusDraft         = "draft"
	StatusResolved      = "resolved"
	StatusConfirmed     = "confirmed"
	StatusProfileReady  = "profile_ready"
	StatusTimelineReady = "timeline_ready"

	JobPending   = "pending"
	JobRunning   = "running"
	JobCompleted = "completed"
	JobFailed    = "failed"
)

type Character struct {
	ID                 string    `json:"id"`
	Mode               string    `json:"mode"`
	PlayStyle          string    `json:"play_style,omitempty"`
	GameConfig         *GameConfig `json:"game_config,omitempty"`
	DisplayName        string    `json:"display_name"`
	Status             string    `json:"status"`
	ResolveQuery       string    `json:"resolve_query,omitempty"`
	ConfirmedIdentity  string    `json:"confirmed_identity,omitempty"`
	CurrentVersionID   string    `json:"current_version_id,omitempty"`
	CurrentTimelineID  string    `json:"current_timeline_id,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// GameConfig 人生游戏模式配置（存于 characters.game_config_json）。
type GameConfig struct {
	EraDescription  string          `json:"era_description,omitempty"`
	EraOptions      []GameEraOption `json:"era_options,omitempty"`
	SelectedEra     string          `json:"selected_era,omitempty"`
	StartYearHint   int             `json:"start_year_hint,omitempty"`
	BirthBackground string          `json:"birth_background,omitempty"`
}

// GameEraOption AI 生成的历史时期选项。
type GameEraOption struct {
	Label       string `json:"label"`
	YearRange   string `json:"year_range"`
	StartYear   int    `json:"start_year"`
	Description string `json:"description"`
}

// GameEraOptionsRequest 生成时期选项。
type GameEraOptionsRequest struct {
	Model          string `json:"model"`
	EraDescription string `json:"era_description,omitempty"`
	Regenerate     bool   `json:"regenerate,omitempty"`
}

// GameEraOptionsResponse 时期选项列表。
type GameEraOptionsResponse struct {
	Options []GameEraOption `json:"options"`
}

// GameProfileGenerateRequest 游戏模式档案生成。
type GameProfileGenerateRequest struct {
	Model           string `json:"model"`
	DisplayName     string `json:"display_name,omitempty"`
	BirthBackground string `json:"birth_background,omitempty"`
	// 名人：需先写入 GameConfig（含 selected_era / start_year_hint）
}

// GameTimelineStartRequest 开始游戏时间轴（首节点）。
type GameTimelineStartRequest struct {
	Model string `json:"model"`
	Title string `json:"title,omitempty"`
}

// GameTimelineStartJobRequest Job 请求体。
type GameTimelineStartJobRequest struct {
	TimelineID string `json:"timeline_id"`
}

// GameChoiceOption 节点抉择选项。
type GameChoiceOption struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	Description  string `json:"description"`
	IsHistorical bool   `json:"is_historical,omitempty"`
}

// GameChoicesResponse 抉择选项响应。
type GameChoicesResponse struct {
	Options      []GameChoiceOption `json:"options"`
	ChosenID     string             `json:"chosen_id,omitempty"`
	CustomText   string             `json:"custom_text,omitempty"`
	AlreadyChosen bool              `json:"already_chosen,omitempty"`
}

// GameChooseRequest 应用抉择。
type GameChooseRequest struct {
	Model      string `json:"model"`
	ChoiceID   string `json:"choice_id,omitempty"`
	CustomText string `json:"custom_text,omitempty"`
}

// GameChooseJobRequest Job 请求体。
type GameChooseJobRequest struct {
	TimelineID string `json:"timeline_id"`
	NodeID     string `json:"node_id"`
	ChoiceID   string `json:"choice_id,omitempty"`
	CustomText string `json:"custom_text,omitempty"`
}

// GameNodeChoiceRecord 节点已缓存的抉择（含选项与已选）。
type GameNodeChoiceRecord struct {
	Options    []GameChoiceOption `json:"options"`
	ChosenID   string             `json:"chosen_id,omitempty"`
	CustomText string             `json:"custom_text,omitempty"`
}

// GameNodeChoiceDisplay 时间轴上展示的已选抉择（位于两节点之间）。
type GameNodeChoiceDisplay struct {
	NodeID       string `json:"node_id"`
	NodeSequence int    `json:"node_sequence"`
	DisplayText  string `json:"display_text"`
}

// GameUpdateConfigRequest 更新游戏配置（选时期等）。
type GameUpdateConfigRequest struct {
	GameConfig GameConfig `json:"game_config" binding:"required"`
}

type ResolveCandidate struct {
	Name        string  `json:"name"`
	BirthYear   int     `json:"birth_year"`
	DeathYear   int     `json:"death_year"`
	Nationality string  `json:"nationality"`
	Summary     string  `json:"summary"`
	Confidence  float64 `json:"confidence"`
}

type Profile struct {
	CharacterID        string `json:"character_id"`
	DisplayName        string `json:"display_name"`
	BirthYear          int    `json:"birth_year"`
	DeathYear          int    `json:"death_year"`
	Era                string `json:"era"`
	EraBackground      string `json:"era_background"`
	PersonalityInitial string `json:"personality_initial"`
	BeliefsMotto       string `json:"beliefs_motto"`
	Experiences        string `json:"experiences"`
	Literature         string `json:"literature"`
	BirthPlace         string `json:"birth_place"`
	Nationality        string `json:"nationality"`
	SocialClass        string `json:"social_class"`
	Occupation         string `json:"occupation"`
	Education          string `json:"education"`
	FamilyRelations    string `json:"family_relations"`
	BeliefsPolitics    string `json:"beliefs_politics"`
	Appearance         string `json:"appearance"`
	MajorWorks         string `json:"major_works"`
	Controversies      string `json:"controversies"`
	DeathCause         string `json:"death_cause"`
	SourcesNote        string `json:"sources_note"`
	TemplateSource     string `json:"template_source,omitempty"`
	RawJSON            string `json:"-"`
}

type TraitChange struct {
	Field  string `json:"field"`
	Before string `json:"before"`
	After  string `json:"after"`
	Reason string `json:"reason"`
}

// NodeEntities 由 AI 标注的节点实体，供前端精确高亮人名与地名。
type NodeEntities struct {
	Protagonist []string `json:"protagonist,omitempty"`
	Persons     []string `json:"persons,omitempty"`
	Places      []string `json:"places,omitempty"`
}

// NodeScene 节点沉浸感上下文；仅在有依据或合理推断时填写，不确定则省略。
type NodeScene struct {
	DateTime  string `json:"datetime,omitempty"`    // 具体日期或年月日
	TimeOfDay string `json:"time_of_day,omitempty"` // 时辰，如「午后」「黄昏」
	Season    string `json:"season,omitempty"`    // 季节
	Weather   string `json:"weather,omitempty"`     // 天气
	Scene     string `json:"scene,omitempty"`     // 场景/环境，如「隆中草庐」「街亭道旁」
}

type LifeNode struct {
	ID                  string        `json:"id"`
	CharacterID         string        `json:"character_id"`
	VersionID           string        `json:"version_id"`
	Sequence            int           `json:"sequence"`
	Year                int           `json:"year"`
	Age                 int           `json:"age"`
	Title               string        `json:"title"`
	Events              string        `json:"events"`
	Thoughts            string        `json:"thoughts"`
	PersonalitySnapshot string        `json:"personality_snapshot"`
	TraitChanges        []TraitChange `json:"trait_changes,omitempty"`
	Entities            *NodeEntities `json:"entities,omitempty"`
	Scene               *NodeScene    `json:"scene,omitempty"`
}

type TimelineVersion struct {
	ID                 string    `json:"id"`
	CharacterID        string    `json:"character_id"`
	TimelineID         string    `json:"timeline_id,omitempty"`
	ParentVersionID    string    `json:"parent_version_id,omitempty"`
	TriggerNodeID      string    `json:"trigger_node_id,omitempty"`
	ChangeSummary      string    `json:"change_summary,omitempty"`
	BranchLabel        string    `json:"branch_label,omitempty"`
	ForkSequence       int       `json:"fork_sequence,omitempty"`
	ForkNodeID         string    `json:"fork_node_id,omitempty"`
	DeathYearSnapshot  int       `json:"death_year_snapshot,omitempty"`
	DeathCauseSnapshot string    `json:"death_cause_snapshot,omitempty"`
	NodeCount          int       `json:"node_count,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

// BranchNode 分支树节点（API 展示用）。
type BranchNode struct {
	ID                 string       `json:"id"`
	ParentID           string       `json:"parent_id,omitempty"`
	Label              string       `json:"label"`
	ChangeSummary      string       `json:"change_summary,omitempty"`
	ForkSequence       int          `json:"fork_sequence,omitempty"`
	ForkNodeID         string       `json:"fork_node_id,omitempty"`
	NodeCount          int          `json:"node_count"`
	DeathYearSnapshot  int          `json:"death_year_snapshot,omitempty"`
	DeathCauseSnapshot string       `json:"death_cause_snapshot,omitempty"`
	IsActive           bool         `json:"is_active"`
	CreatesBranch      bool         `json:"creates_branch,omitempty"`
	CreatedAt          time.Time    `json:"created_at"`
	Children           []BranchNode `json:"children,omitempty"`
}

// BranchTreeResponse 时间轴分支树。
type BranchTreeResponse struct {
	TimelineID            string       `json:"timeline_id"`
	ActiveVersionID       string       `json:"active_version_id"`
	DisplayActiveBranchID string       `json:"display_active_branch_id,omitempty"`
	Roots                 []BranchNode `json:"roots"`
}

// BranchOverviewNodeLite 总览视图节点（年份、标题与性格变更摘要）。
type BranchOverviewNodeLite struct {
	Sequence     int           `json:"sequence"`
	Year         int           `json:"year"`
	Title        string        `json:"title"`
	TraitChanges []TraitChange `json:"trait_changes,omitempty"`
}

// BranchOverviewEntry 总览中的单条分支。
type BranchOverviewEntry struct {
	VersionID         string                   `json:"version_id"`
	Label             string                   `json:"label"`
	IsActive          bool                     `json:"is_active"`
	CreatesBranch     bool                     `json:"creates_branch,omitempty"`
	ForkSequence      int                      `json:"fork_sequence,omitempty"`
	DeathYearSnapshot int                      `json:"death_year_snapshot,omitempty"`
	Nodes             []BranchOverviewNodeLite `json:"nodes"`
}

// BranchOverviewResponse 全部分支总览。
type BranchOverviewResponse struct {
	TimelineID      string                `json:"timeline_id"`
	ActiveVersionID string                `json:"active_version_id"`
	Branches        []BranchOverviewEntry `json:"branches"`
}

// Timeline 人物下的独立时间轴实例（可有多条）；版本链归属同一条时间轴。
type Timeline struct {
	ID               string             `json:"id"`
	CharacterID      string             `json:"character_id"`
	Title            string             `json:"title"`
	CurrentVersionID string             `json:"current_version_id,omitempty"`
	NodeCount        int                `json:"node_count,omitempty"`
	VersionCount     int                `json:"version_count,omitempty"`
	WorldLine        *WorldLine         `json:"world_line,omitempty"`
	GenerationStatus string             `json:"generation_status,omitempty"` // ready | generating | failed
	GenerationError  string             `json:"generation_error,omitempty"`
	ActiveJob        *TimelineActiveJob `json:"active_job,omitempty"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

// WorldLineEvent 世界线上的历史大事节点。
type WorldLineEvent struct {
	Year              int    `json:"year"`
	Name              string `json:"name"`
	Description       string `json:"description,omitempty"`
	Impact            string `json:"impact,omitempty"`
	DivergenceNote    string `json:"divergence_note,omitempty"`
	CausedByNodeSeq   *int   `json:"caused_by_node_sequence,omitempty"`
}

// WorldLine 时间轴绑定的世界线（历史走向与天下大事）。
type WorldLine struct {
	TimelineID       string           `json:"timeline_id"`
	EraSummary       string           `json:"era_summary,omitempty"`
	HistoricalTrend  string           `json:"historical_trend"`
	DailyLifeContext string           `json:"daily_life_context,omitempty"`
	Events           []WorldLineEvent `json:"events"`
	StartYear        int              `json:"start_year,omitempty"`
	EndYear          int              `json:"end_year,omitempty"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// WorldLineUpdateRequest 更新世界线并可选择同步到人生节点。
type WorldLineUpdateRequest struct {
	WorldLine    WorldLine `json:"world_line"`
	Model        string    `json:"model"`
	ApplyToNodes bool      `json:"apply_to_nodes"`
}

// WorldLineRefreshRequest 重算当前时间轴世界线。
type WorldLineRefreshRequest struct {
	Model string `json:"model"`
}

const (
	TimelineGenReady      = "ready"
	TimelineGenGenerating = "generating"
	TimelineGenFailed     = "failed"
)

// TimelineActiveJob 时间轴关联的进行中任务摘要。
type TimelineActiveJob struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	Progress  int    `json:"progress"`
	StageText string `json:"stage_text,omitempty"`
}

type Job struct {
	ID          string    `json:"id"`
	CharacterID string    `json:"character_id"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Progress    int       `json:"progress"`
	StageText   string    `json:"stage_text,omitempty"`
	Model       string    `json:"model,omitempty"`
	RequestJSON string    `json:"request_json,omitempty"`
	Result      string    `json:"result,omitempty"`
	Error       string    `json:"error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

const (
	PatchModeFullCascade       = "full_cascade"
	PatchModeAppendNext        = "append_next"
	PatchModeInnerCurrent      = "inner_current"
	PatchModeInnerSubsequent   = "inner_subsequent"
	PatchModeDialogueImpact    = "dialogue_impact"

	DialogueIdentityReader = "命运读者"

	NarrativeStandard = "standard"
	NarrativeRich     = "rich"

	NarrativeKindLightNovel = "light_novel"
	NarrativeKindChronicle  = "chronicle"
	NarrativeKindDiary      = "diary"
	NarrativeKindLetter     = "letter"
	NarrativeKindArchive    = "archive"

	LightNovelPersonFirst  = "first"
	LightNovelPersonSecond = "second"
	LightNovelPersonThird  = "third"
)

// NormalizeLightNovelPerson 规范轻小说叙述人称，默认第一人称。
func NormalizeLightNovelPerson(p string) string {
	switch p {
	case LightNovelPersonSecond, LightNovelPersonThird:
		return p
	default:
		return LightNovelPersonFirst
	}
}

// NormalizeNarrativeDensity 规范叙事密度，默认 rich（两阶段厚叙事）。
func NormalizeNarrativeDensity(d string) string {
	if d == NarrativeStandard {
		return NarrativeStandard
	}
	return NarrativeRich
}

// PatchNodeRequest 编辑节点请求
type PatchNodeRequest struct {
	Title               string `json:"title"`
	Events              string `json:"events"`
	Thoughts            string `json:"thoughts"`
	PersonalitySnapshot string `json:"personality_snapshot"`
	Mode                string `json:"mode"`
	Model               string `json:"model"`
	TargetNodeCount     int    `json:"target_node_count"`
	NextNodeTitle       string `json:"next_node_title,omitempty"`
	StepYears           int    `json:"step_years,omitempty"`
	ConfirmedDeathYear  int    `json:"confirmed_death_year,omitempty"`
	ConfirmedDeathCause string `json:"confirmed_death_cause,omitempty"`
	LifespanReasoning   string `json:"lifespan_reasoning,omitempty"`
	ChangeSummary       string `json:"change_summary,omitempty"`
}

// RollbackToNodeRequest 回退至指定节点（截断后续）
type RollbackToNodeRequest struct {
	ChangeSummary string `json:"change_summary,omitempty"`
}

// TimelineJobConfigSnapshot 时间轴生成任务配置快照（用于重试）
type TimelineJobConfigSnapshot struct {
	StepYears         int    `json:"step_years"`
	TargetNodeCount   int    `json:"target_node_count"`
	StartYear         int    `json:"start_year"`
	EndYear           int    `json:"end_year"`
	Title             string `json:"title"`
	Instructions      string `json:"instructions"`
	EraEventsOverride string `json:"era_events_override,omitempty"`
	CharacterMode     string `json:"character_mode"`
	NarrativeDensity  string `json:"narrative_density"`
}

// TimelineGenerateJobRequest timeline_generate 任务完整参数
type TimelineGenerateJobRequest struct {
	TimelineID string                    `json:"timeline_id"`
	Generate   TimelineGenerateRequest   `json:"generate"`
	Config     TimelineJobConfigSnapshot `json:"config"`
}

// PatchNodeJobRequest 节点编辑/推演任务完整参数
type PatchNodeJobRequest struct {
	TimelineID string           `json:"timeline_id"`
	NodeID     string           `json:"node_id"`
	Patch      PatchNodeRequest `json:"patch"`
}

// NarrativeChangeJobRequest 叙述变更任务完整参数
type NarrativeChangeJobRequest struct {
	TimelineID string                 `json:"timeline_id"`
	Request    NarrativeChangeRequest `json:"request"`
}

// NodeNarrativeJobRequest 节点叙事任务完整参数
type NodeNarrativeJobRequest struct {
	NodeID  string                `json:"node_id"`
	Kind    string                `json:"kind"`
	Request NodeNarrativeRequest  `json:"request"`
}

// RegenerateNodeEventsRequest 根据标题重新生成节点经历
type RegenerateNodeEventsRequest struct {
	Title string `json:"title"`
	Model string `json:"model"`
}

type RegenerateNodeEventsResponse struct {
	Events              string        `json:"events"`
	Thoughts            string        `json:"thoughts,omitempty"`
	PersonalitySnapshot string        `json:"personality_snapshot,omitempty"`
	TraitChanges        []TraitChange `json:"trait_changes,omitempty"`
}

// NarrativeChangeRequest 用自然语言叙述变更时间轴
type NarrativeChangeRequest struct {
	TimelineID      string `json:"timeline_id"`
	Instruction     string `json:"instruction"`
	Model           string `json:"model"`
	TargetNodeCount int    `json:"target_node_count,omitempty"`
}

// LifespanPreviewRequest 预览寿命（编辑节点后）
type LifespanPreviewRequest struct {
	Title               string `json:"title"`
	Events              string `json:"events"`
	Thoughts            string `json:"thoughts"`
	PersonalitySnapshot string `json:"personality_snapshot"`
	Model               string `json:"model"`
}

// LifespanPreviewResponse 寿命预览结果
type LifespanPreviewResponse struct {
	CurrentDeathYear  int                  `json:"current_death_year"`
	AnchorYear        int                  `json:"anchor_year"`
	Preview           LifespanRecalcResult `json:"preview"`
	ContextTokenHint  int                  `json:"context_token_hint,omitempty"`
}

// ProfileGenerateRequest 生成档案请求
type ProfileGenerateRequest struct {
	Model        string `json:"model"`
	DisplayName  string `json:"display_name,omitempty"`
	Background   string `json:"background,omitempty"`
	Introduction string `json:"introduction,omitempty"`
}

// ProfileRandomizeFieldRequest 随机重生成档案单字段
type ProfileRandomizeFieldRequest struct {
	Field string `json:"field"`
	Model string `json:"model"`
}

// SuggestNamesRequest AI 推荐随机人物姓名
type SuggestNamesRequest struct {
	Model        string `json:"model"`
	Background   string `json:"background,omitempty"`
	Introduction string `json:"introduction,omitempty"`
}

// TimelineGenerateRequest 生成时间轴请求
type TimelineGenerateRequest struct {
	Model           string `json:"model"`
	Title           string `json:"title,omitempty"`
	Instructions    string `json:"instructions,omitempty"`
	// EraEvents 时代背景与大事记；留空则 AI 按区间自动生成（史实或架空世界观）
	EraEvents string `json:"era_events,omitempty"`
	TargetNodeCount int    `json:"target_node_count"`
	StepYears       int    `json:"step_years,omitempty"`
	StartYear         int    `json:"start_year"`
	EndYear           int    `json:"end_year"`
	NarrativeDensity  string `json:"narrative_density,omitempty"`
}

// TimelineRecommendation AI 推荐的时间轴生成配置
type TimelineRecommendation struct {
	Label           string `json:"label"`
	Description     string `json:"description"`
	TargetNodeCount int    `json:"target_node_count"`
	StepYears       int    `json:"step_years,omitempty"`
	StartYear       int    `json:"start_year"`
	EndYear         int    `json:"end_year"`
	FocusPhase      string `json:"focus_phase,omitempty"`
}

// LifespanRecalcResult 经历变更后重算寿命
type LifespanRecalcResult struct {
	DeathYear   int    `json:"death_year"`
	DeathCause  string `json:"death_cause"`
	Reasoning   string `json:"reasoning"`
	HealthNotes string `json:"health_notes,omitempty"`
}

type NodeFieldChange struct {
	NodeID   string `json:"node_id"`
	Sequence int    `json:"sequence"`
	Year     int    `json:"year"`
	Field    string `json:"field"`
	Before   string `json:"before"`
	After    string `json:"after"`
}

type VersionDiff struct {
	VersionID       string            `json:"version_id"`
	ParentVersionID string            `json:"parent_version_id,omitempty"`
	Changes         []NodeFieldChange `json:"changes"`
}

// CharacterHistoryItem 生成历史列表项（含摘要信息）
type CharacterHistoryItem struct {
	ID               string    `json:"id"`
	Mode             string    `json:"mode"`
	PlayStyle        string    `json:"play_style,omitempty"`
	DisplayName      string    `json:"display_name"`
	Status           string    `json:"status"`
	ResolveQuery     string    `json:"resolve_query,omitempty"`
	CurrentVersionID string    `json:"current_version_id,omitempty"`
	Era              string    `json:"era,omitempty"`
	BirthYear        int       `json:"birth_year,omitempty"`
	DeathYear        int       `json:"death_year,omitempty"`
	NodeCount        int       `json:"node_count"`
	TimelineCount    int       `json:"timeline_count"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// NarrativeArtifact 按需生成的叙事文本缓存
type NarrativeArtifact struct {
	ID           string    `json:"id"`
	CharacterID  string    `json:"character_id"`
	VersionID    string    `json:"version_id"`
	NodeID       string    `json:"node_id,omitempty"`
	Kind         string    `json:"kind"`
	FromSequence int       `json:"from_sequence,omitempty"`
	ToSequence   int       `json:"to_sequence,omitempty"`
	Person       string    `json:"person,omitempty"`
	ContentKey   string    `json:"content_key,omitempty"`
	Content      string    `json:"content"`
	Model        string    `json:"model,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// NarrativeArtifactQuery 缓存查询键
type NarrativeArtifactQuery struct {
	VersionID    string
	NodeID       string
	Kind         string
	FromSequence int
	ToSequence   int
	Person       string
}

// SavedLightNovel 磁盘上保存的轻小说（data/light-novels/<branch>/）。
type SavedLightNovel struct {
	ID           string    `json:"id"`
	CharacterID  string    `json:"character_id"`
	DisplayName  string    `json:"display_name,omitempty"`
	VersionID    string    `json:"version_id"`
	FromSequence int       `json:"from_sequence"`
	ToSequence   int       `json:"to_sequence"`
	Person       string    `json:"person,omitempty"`
	Model        string    `json:"model,omitempty"`
	Branch       string    `json:"branch,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	Content      string    `json:"content"`
}

// SavedLightNovelMeta 列表项（不含正文）。
type SavedLightNovelMeta struct {
	ID           string    `json:"id"`
	CharacterID  string    `json:"character_id"`
	DisplayName  string    `json:"display_name,omitempty"`
	VersionID    string    `json:"version_id"`
	FromSequence int       `json:"from_sequence"`
	ToSequence   int       `json:"to_sequence"`
	Person       string    `json:"person,omitempty"`
	Model        string    `json:"model,omitempty"`
	Branch       string    `json:"branch,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// LightNovelRequest 生成轻小说
type LightNovelRequest struct {
	Model        string `json:"model"`
	VersionID    string `json:"version_id"`
	FromSequence int    `json:"from_sequence"`
	ToSequence   int    `json:"to_sequence"`
	Person       string `json:"person,omitempty"`
	Force        bool   `json:"force,omitempty"`
}

// SavedChronicle 磁盘上保存的史书（data/chronicles/<branch>/）。
type SavedChronicle struct {
	ID           string    `json:"id"`
	CharacterID  string    `json:"character_id"`
	DisplayName  string    `json:"display_name,omitempty"`
	VersionID    string    `json:"version_id"`
	FromSequence int       `json:"from_sequence"`
	ToSequence   int       `json:"to_sequence"`
	Model        string    `json:"model,omitempty"`
	Branch       string    `json:"branch,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	Content      string    `json:"content"`
}

// SavedChronicleMeta 史书列表项（不含正文）。
type SavedChronicleMeta struct {
	ID           string    `json:"id"`
	CharacterID  string    `json:"character_id"`
	DisplayName  string    `json:"display_name,omitempty"`
	VersionID    string    `json:"version_id"`
	FromSequence int       `json:"from_sequence"`
	ToSequence   int       `json:"to_sequence"`
	Model        string    `json:"model,omitempty"`
	Branch       string    `json:"branch,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// ChronicleRequest 编纂史书
type ChronicleRequest struct {
	Model        string `json:"model"`
	VersionID    string `json:"version_id"`
	FromSequence int    `json:"from_sequence"`
	ToSequence   int    `json:"to_sequence"`
	Force        bool   `json:"force,omitempty"`
}

// NodeNarrativeRequest 单节点多视角生成
type NodeNarrativeRequest struct {
	Model string `json:"model"`
	Force bool   `json:"force,omitempty"`
}

// CharacterMemory 人物对用户的持久记忆（绑定版本分支）。
type CharacterMemory struct {
	ID              string    `json:"id"`
	CharacterID     string    `json:"character_id"`
	VersionID       string    `json:"version_id"`
	SourceNodeID    string    `json:"source_node_id"`
	SourceSequence  int       `json:"source_sequence"`
	SpeakerIdentity string    `json:"speaker_identity"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// DialogueSession 节点对话会话。
type DialogueSession struct {
	ID              string    `json:"id"`
	CharacterID     string    `json:"character_id"`
	VersionID       string    `json:"version_id"`
	NodeID          string    `json:"node_id"`
	NodeSequence    int       `json:"node_sequence"`
	SpeakerIdentity string    `json:"speaker_identity"`
	Model           string    `json:"model"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Messages        []DialogueMessage `json:"messages,omitempty"`
}

// DialogueSessionSummary 历史对话列表项。
type DialogueSessionSummary struct {
	ID              string    `json:"id"`
	CharacterID     string    `json:"character_id"`
	VersionID       string    `json:"version_id"`
	NodeID          string    `json:"node_id"`
	NodeSequence    int       `json:"node_sequence"`
	NodeYear        int       `json:"node_year,omitempty"`
	NodeAge         int       `json:"node_age,omitempty"`
	NodeTitle       string    `json:"node_title,omitempty"`
	SpeakerIdentity string    `json:"speaker_identity"`
	Model           string    `json:"model"`
	MessageCount    int       `json:"message_count"`
	LastMessage     string    `json:"last_message,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// DialogueMessage 对话消息。
type DialogueMessage struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// DialogueIdentityOption AI 生成的可选身份。
type DialogueIdentityOption struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}

// PersonalityImpact 对话对性格/思想的影响检测结果。
type PersonalityImpact struct {
	HasImpact           bool          `json:"has_impact"`
	Summary             string        `json:"summary,omitempty"`
	Thoughts            string        `json:"thoughts,omitempty"`
	PersonalitySnapshot string        `json:"personality_snapshot,omitempty"`
	TraitChanges        []TraitChange `json:"trait_changes,omitempty"`
}

// CreateDialogueSessionRequest 创建对话会话。
type CreateDialogueSessionRequest struct {
	Identity string `json:"identity" binding:"required"`
	Model    string `json:"model"`
	Resume   bool   `json:"resume"`
}

// SendDialogueMessageRequest 发送对话消息。
type SendDialogueMessageRequest struct {
	Content string `json:"content" binding:"required"`
	Model   string `json:"model"`
}

// SendDialogueMessageResponse 发送消息响应。
type SendDialogueMessageResponse struct {
	UserMessage      DialogueMessage   `json:"user_message"`
	AssistantMessage DialogueMessage   `json:"assistant_message"`
	Impact           *PersonalityImpact `json:"impact,omitempty"`
}

// CreateMemoryRequest 写入记忆。
type CreateMemoryRequest struct {
	VersionID       string `json:"version_id" binding:"required"`
	SourceNodeID    string `json:"source_node_id" binding:"required"`
	SourceSequence  int    `json:"source_sequence"`
	SpeakerIdentity string `json:"speaker_identity"`
	Content         string `json:"content" binding:"required"`
}

// UpdateMemoryRequest 更新记忆内容。
type UpdateMemoryRequest struct {
	Content string `json:"content" binding:"required"`
}

// ApplyDialogueImpactRequest 确认性格影响写入节点。
type ApplyDialogueImpactRequest struct {
	Thoughts            string        `json:"thoughts" binding:"required"`
	PersonalitySnapshot string        `json:"personality_snapshot" binding:"required"`
	TraitChanges        []TraitChange `json:"trait_changes"`
	Summary             string        `json:"summary"`
}

// ApplyDialogueImpactResponse 性格影响写入结果。
type ApplyDialogueImpactResponse struct {
	VersionID string    `json:"version_id"`
	Node      *LifeNode `json:"node"`
}
