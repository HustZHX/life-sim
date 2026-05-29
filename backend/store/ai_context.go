package store

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"life-sim/backend/model"
)

// profileTimeline 供时间轴推荐/生成使用的精简档案（去掉 AI 不需要的元数据）。
type profileTimeline struct {
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
}

// profileInner 供内心/寿命重算使用的性格相关字段。
type profileInner struct {
	DisplayName        string `json:"display_name"`
	BirthYear          int    `json:"birth_year"`
	DeathYear          int    `json:"death_year"`
	Era                string `json:"era"`
	PersonalityInitial string `json:"personality_initial"`
	BeliefsMotto       string `json:"beliefs_motto"`
	BeliefsPolitics    string `json:"beliefs_politics"`
	Experiences        string `json:"experiences"`
}

type nodeContextLite struct {
	Sequence            int    `json:"sequence"`
	Year                int    `json:"year"`
	Age                 int    `json:"age"`
	Title               string `json:"title"`
	Events              string `json:"events"`
	PersonalitySnapshot string `json:"personality_snapshot"`
}

type nodeContextLocked struct {
	nodeContextLite
	Thoughts     string              `json:"thoughts"`
	TraitChanges []model.TraitChange `json:"trait_changes,omitempty"`
}

type nodeContextNarrative struct {
	nodeContextLocked
	Entities *model.NodeEntities `json:"entities,omitempty"`
	Scene    *model.NodeScene    `json:"scene,omitempty"`
}

type nodeTailInput struct {
	Sequence int    `json:"sequence"`
	Year     int    `json:"year"`
	Title    string `json:"title"`
	Events   string `json:"events"`
}

// ProfileJSONFull 完整档案 JSON（与 MustProfileJSON 等价）。
func ProfileJSONFull(p *model.Profile) string {
	return MustProfileJSON(p)
}

// ProfileJSONTimeline 时间轴场景精简档案。
func ProfileJSONTimeline(p *model.Profile) string {
	b, _ := json.Marshal(profileFromTimeline(p))
	return string(b)
}

// MarshalGameConfigForPrompt 游戏模式配置（时期、开局年份等）。
func MarshalGameConfigForPrompt(cfg *model.GameConfig) string {
	if cfg == nil {
		return "{}"
	}
	b, _ := json.Marshal(cfg)
	return string(b)
}

// MarshalNodesJourneyForChoices 已历人生节点摘要（含当前节点），供抉择生成参考过往经历。
func MarshalNodesJourneyForChoices(nodes []model.LifeNode, currentSeq int) string {
	type lite struct {
		Sequence            int    `json:"sequence"`
		Year                int    `json:"year"`
		Age                 int    `json:"age"`
		Title               string `json:"title"`
		Events              string `json:"events"`
		PersonalitySnapshot string `json:"personality_snapshot,omitempty"`
	}
	out := make([]lite, 0, len(nodes))
	for _, n := range nodes {
		if currentSeq >= 0 && n.Sequence > currentSeq {
			continue
		}
		snap := ""
		if n.Sequence >= currentSeq-1 {
			snap = TruncateRunes(n.PersonalitySnapshot, 160)
		}
		out = append(out, lite{
			Sequence:            n.Sequence,
			Year:                n.Year,
			Age:                 n.Age,
			Title:               n.Title,
			Events:              TruncateRunes(n.Events, 300),
			PersonalitySnapshot: snap,
		})
	}
	b, _ := json.Marshal(out)
	return string(b)
}

// MarshalGameNodeForChoices 当前节点上下文（供抉择生成）。
func MarshalGameNodeForChoices(n *model.LifeNode) string {
	if n == nil {
		return "{}"
	}
	events := TruncateRunes(n.Events, 480)
	b, _ := json.Marshal(map[string]any{
		"sequence":             n.Sequence,
		"year":                 n.Year,
		"age":                  n.Age,
		"title":                n.Title,
		"events":               events,
		"personality_snapshot": TruncateRunes(n.PersonalitySnapshot, 200),
	})
	return string(b)
}

// TruncateRunes 按 rune 截断字符串。
func TruncateRunes(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "…"
}

// ProfileJSONInner 内心/寿命重算精简档案。
func ProfileJSONInner(p *model.Profile) string {
	b, _ := json.Marshal(profileFromInner(p))
	return string(b)
}

// ProfileHash 用于缓存/session 失效检测。
func ProfileHash(p *model.Profile) string {
	sum := sha256.Sum256([]byte(ProfileJSONTimeline(p)))
	return hex.EncodeToString(sum[:8])
}

func profileFromTimeline(p *model.Profile) profileTimeline {
	return profileTimeline{
		DisplayName: p.DisplayName, BirthYear: p.BirthYear, DeathYear: p.DeathYear,
		Era: p.Era, EraBackground: p.EraBackground, PersonalityInitial: p.PersonalityInitial,
		BeliefsMotto: p.BeliefsMotto, Experiences: p.Experiences, Literature: p.Literature,
		BirthPlace: p.BirthPlace, Nationality: p.Nationality, SocialClass: p.SocialClass,
		Occupation: p.Occupation, Education: p.Education, FamilyRelations: p.FamilyRelations,
		BeliefsPolitics: p.BeliefsPolitics, Appearance: p.Appearance, MajorWorks: p.MajorWorks,
		Controversies: p.Controversies, DeathCause: p.DeathCause,
	}
}

func profileFromInner(p *model.Profile) profileInner {
	return profileInner{
		DisplayName: p.DisplayName, BirthYear: p.BirthYear, DeathYear: p.DeathYear,
		Era: p.Era, PersonalityInitial: p.PersonalityInitial, BeliefsMotto: p.BeliefsMotto,
		BeliefsPolitics: p.BeliefsPolitics, Experiences: p.Experiences,
	}
}

func toNodeLite(n model.LifeNode) nodeContextLite {
	return nodeContextLite{
		Sequence: n.Sequence, Year: n.Year, Age: n.Age,
		Title: n.Title, Events: n.Events, PersonalitySnapshot: n.PersonalitySnapshot,
	}
}

func toNodeLocked(n model.LifeNode) nodeContextLocked {
	return nodeContextLocked{
		nodeContextLite: toNodeLite(n),
		Thoughts:        n.Thoughts,
		TraitChanges:    n.TraitChanges,
	}
}

func toNodeTail(n model.LifeNode) nodeTailInput {
	return nodeTailInput{
		Sequence: n.Sequence, Year: n.Year, Title: n.Title, Events: n.Events,
	}
}

// MarshalNodesLiteBeforeSequence 序列化 sequence < beforeSeq 的精简前置节点。
func MarshalNodesLiteBeforeSequence(nodes []model.LifeNode, beforeSeq int) string {
	out := make([]nodeContextLite, 0)
	for _, n := range nodes {
		if n.Sequence < beforeSeq {
			out = append(out, toNodeLite(n))
		}
	}
	b, _ := json.Marshal(out)
	return string(b)
}

// MarshalNodesLocked 序列化锁定节点（不含 id/entities/scene）。
func MarshalNodesLocked(nodes []model.LifeNode) string {
	out := make([]nodeContextLocked, len(nodes))
	for i, n := range nodes {
		out[i] = toNodeLocked(n)
	}
	b, _ := json.Marshal(out)
	return string(b)
}

// MarshalNodesForNarrative 序列化节点供轻小说/多视角生成（含 scene、entities）。
func MarshalNodesForNarrative(nodes []model.LifeNode) string {
	out := make([]nodeContextNarrative, len(nodes))
	for i, n := range nodes {
		out[i] = nodeContextNarrative{
			nodeContextLocked: toNodeLocked(n),
			Entities:          n.Entities,
			Scene:               n.Scene,
		}
	}
	b, _ := json.Marshal(out)
	return string(b)
}

// MarshalSingleNodeForNarrative 单节点完整叙事上下文。
func MarshalSingleNodeForNarrative(n model.LifeNode) string {
	b, _ := json.Marshal(nodeContextNarrative{
		nodeContextLocked: toNodeLocked(n),
		Entities:          n.Entities,
		Scene:             n.Scene,
	})
	return string(b)
}

// BatchLifeNodeSlices 按固定大小分批节点。
func BatchLifeNodeSlices(nodes []model.LifeNode, size int) [][]model.LifeNode {
	if size <= 0 {
		size = ExpandBatchSize()
	}
	var batches [][]model.LifeNode
	for i := 0; i < len(nodes); i += size {
		end := i + size
		if end > len(nodes) {
			end = len(nodes)
		}
		batches = append(batches, nodes[i:end])
	}
	return batches
}

// MarshalNodesTailInput 序列化待重算内心的后续节点（仅保留经历）。
func MarshalNodesTailInput(nodes []model.LifeNode) string {
	out := make([]nodeTailInput, len(nodes))
	for i, n := range nodes {
		out[i] = toNodeTail(n)
	}
	b, _ := json.Marshal(out)
	return string(b)
}

// LastNodesLite 取最后 n 个节点的 Lite 表示，用于分块衔接。
func LastNodesLite(nodes []model.LifeNode, count int) string {
	if count <= 0 || len(nodes) == 0 {
		return "[]"
	}
	start := len(nodes) - count
	if start < 0 {
		start = 0
	}
	out := make([]nodeContextLite, 0, len(nodes)-start)
	for _, n := range nodes[start:] {
		out = append(out, toNodeLite(n))
	}
	b, _ := json.Marshal(out)
	return string(b)
}

// EstimateContextBytes 估算上下文字节数（供前端提示）。
func EstimateContextBytes(parts ...string) int {
	n := 0
	for _, p := range parts {
		n += len(p)
	}
	return n
}

// RenumberTimelineNodes 将节点 sequence 重排为从 startSeq 递增。
func RenumberTimelineNodes(nodes []model.LifeNode, startSeq int) {
	for i := range nodes {
		nodes[i].Sequence = startSeq + i
	}
}

// MergeTimelineNodeChunks 合并分块生成的节点并连续编号。
func MergeTimelineNodeChunks(chunks ...[]model.LifeNode) []model.LifeNode {
	var all []model.LifeNode
	seq := 0
	for _, chunk := range chunks {
		for _, n := range chunk {
			nn := n
			nn.Sequence = seq
			all = append(all, nn)
			seq++
		}
	}
	return all
}

// FormatChunkStageText 分块进度文案。
func FormatChunkStageText(chunkIndex, totalChunks, startYear, endYear int) string {
	return fmt.Sprintf("正在生成 %d—%d 年（第 %d/%d 段）…", startYear, endYear, chunkIndex+1, totalChunks)
}
