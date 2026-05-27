package model

import "time"

const (
	ModeFamous = "famous"
	ModeRandom = "random"

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
	DisplayName        string    `json:"display_name"`
	Status             string    `json:"status"`
	ResolveQuery       string    `json:"resolve_query,omitempty"`
	ConfirmedIdentity  string    `json:"confirmed_identity,omitempty"`
	CurrentVersionID   string    `json:"current_version_id,omitempty"`
	CurrentTimelineID  string    `json:"current_timeline_id,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
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
	ID              string    `json:"id"`
	CharacterID     string    `json:"character_id"`
	TimelineID      string    `json:"timeline_id,omitempty"`
	ParentVersionID string    `json:"parent_version_id,omitempty"`
	TriggerNodeID   string    `json:"trigger_node_id,omitempty"`
	ChangeSummary   string    `json:"change_summary,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// Timeline 人物下的独立时间轴实例（可有多条）；版本链归属同一条时间轴。
type Timeline struct {
	ID               string    `json:"id"`
	CharacterID      string    `json:"character_id"`
	Title            string    `json:"title"`
	CurrentVersionID string    `json:"current_version_id,omitempty"`
	NodeCount        int       `json:"node_count,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Job struct {
	ID          string    `json:"id"`
	CharacterID string    `json:"character_id"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Progress    int       `json:"progress"`
	StageText   string    `json:"stage_text,omitempty"`
	Model       string    `json:"model,omitempty"`
	Result      string    `json:"result,omitempty"`
	Error       string    `json:"error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

const (
	PatchModeFullCascade       = "full_cascade"
	PatchModeInnerCurrent      = "inner_current"
	PatchModeInnerSubsequent   = "inner_subsequent"

	NarrativeStandard = "standard"
	NarrativeRich     = "rich"

	NarrativeKindLightNovel = "light_novel"
	NarrativeKindDiary      = "diary"
	NarrativeKindLetter     = "letter"
	NarrativeKindArchive    = "archive"
)

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
	StepYears           int    `json:"step_years,omitempty"`
	ConfirmedDeathYear  int    `json:"confirmed_death_year,omitempty"`
	ConfirmedDeathCause string `json:"confirmed_death_cause,omitempty"`
	LifespanReasoning   string `json:"lifespan_reasoning,omitempty"`
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
}

// LightNovelRequest 生成轻小说
type LightNovelRequest struct {
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
