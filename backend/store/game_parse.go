package store

import (
	"encoding/json"
	"fmt"
	"strings"

	"life-sim/backend/model"
)

type gameEraOptionsPayload struct {
	Options []model.GameEraOption `json:"options"`
}

func ParseGameEraOptions(raw string) ([]model.GameEraOption, error) {
	var p gameEraOptionsPayload
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return nil, err
	}
	if len(p.Options) == 0 {
		return nil, fmt.Errorf("未返回时期选项")
	}
	if len(p.Options) > 4 {
		p.Options = p.Options[:4]
	}
	return p.Options, nil
}

type gameChoicesPayload struct {
	Options []struct {
		ID           string `json:"id"`
		Label        string `json:"label"`
		Description  string `json:"description"`
		IsHistorical bool   `json:"is_historical"`
	} `json:"options"`
}

func ParseGameProfileRefresh(raw string) (json.RawMessage, error) {
	var j struct {
		ProfileUpdates json.RawMessage `json:"profile_updates"`
	}
	if err := json.Unmarshal([]byte(raw), &j); err != nil {
		return nil, err
	}
	if len(j.ProfileUpdates) == 0 || string(j.ProfileUpdates) == "null" {
		return nil, fmt.Errorf("未返回 profile_updates")
	}
	return j.ProfileUpdates, nil
}

func ParseGameChoices(raw string) ([]model.GameChoiceOption, error) {
	var p gameChoicesPayload
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return nil, err
	}
	if len(p.Options) < 3 {
		return nil, fmt.Errorf("抉择选项不足")
	}
	out := make([]model.GameChoiceOption, 0, len(p.Options))
	for i, o := range p.Options {
		id := o.ID
		if id == "" {
			id = fmt.Sprintf("choice_%d", i+1)
		}
		out = append(out, model.GameChoiceOption{
			ID:           id,
			Label:        o.Label,
			Description:  o.Description,
			IsHistorical: o.IsHistorical,
		})
	}
	return out, nil
}

type gameApplyNextNode struct {
	Year                int                 `json:"year"`
	Age                 int                 `json:"age"`
	Title               string              `json:"title"`
	Events              string              `json:"events"`
	Thoughts            string              `json:"thoughts"`
	PersonalitySnapshot string              `json:"personality_snapshot"`
	TraitChanges json.RawMessage `json:"trait_changes"`
	Entities     json.RawMessage `json:"entities"`
	Scene        json.RawMessage `json:"scene"`
}

// GameApplyChoicePayload 为 game_apply_choice AI 响应解析结果。
type GameApplyChoicePayload struct {
	NextNode         gameApplyNextNodeParsed `json:"next_node"`
	ProfileUpdates   json.RawMessage         `json:"profile_updates"`
	AffectsHistory   bool                    `json:"affects_history"`
	WorldLineChanged *bool                   `json:"world_line_changed"`
	MajorEvents      []string                `json:"major_events"`
	WorldLineDelta   struct {
		EraSummary           string                 `json:"era_summary"`
		HistoricalTrend      string                 `json:"historical_trend_patch"`
		HistoricalTrendPatch string                 `json:"historical_trend"` // alias
		DailyLifeContext     string                 `json:"daily_life_context"`
		Events               []model.WorldLineEvent `json:"events"`
	} `json:"world_line_delta"`
}

// gameApplyNextNodeParsed 由 ParseGameApplyChoice 填充（trait_changes 已容错解析）。
type gameApplyNextNodeParsed struct {
	Year                int
	Age                 int
	Title               string
	Events              string
	Thoughts            string
	PersonalitySnapshot string
	TraitChanges        []model.TraitChange
	Entities            *model.NodeEntities
	Scene               *model.NodeScene
}

func ParseGameApplyChoice(raw string) (*GameApplyChoicePayload, error) {
	var j struct {
		NextNode         gameApplyNextNode `json:"next_node"`
		ProfileUpdates   json.RawMessage `json:"profile_updates"`
		AffectsHistory   bool              `json:"affects_history"`
		WorldLineChanged *bool             `json:"world_line_changed"`
		MajorEvents      []string          `json:"major_events"`
		WorldLineDelta   struct {
			EraSummary           string                 `json:"era_summary"`
			HistoricalTrend      string                 `json:"historical_trend_patch"`
			HistoricalTrendPatch string                 `json:"historical_trend"`
			DailyLifeContext     string                 `json:"daily_life_context"`
			Events               []model.WorldLineEvent `json:"events"`
		} `json:"world_line_delta"`
	}
	if err := json.Unmarshal([]byte(raw), &j); err != nil {
		return nil, err
	}
	if j.NextNode.Title == "" || j.NextNode.Events == "" {
		return nil, fmt.Errorf("缺少下一节点内容")
	}
	traits, err := ParseTraitChangesJSON(j.NextNode.TraitChanges)
	if err != nil {
		return nil, fmt.Errorf("trait_changes: %w", err)
	}
	entities, err := ParseEntitiesJSON(j.NextNode.Entities, "")
	if err != nil {
		return nil, fmt.Errorf("entities: %w", err)
	}
	scene, err := ParseSceneJSON(j.NextNode.Scene)
	if err != nil {
		return nil, fmt.Errorf("scene: %w", err)
	}
	p := &GameApplyChoicePayload{
		NextNode: gameApplyNextNodeParsed{
			Year:                j.NextNode.Year,
			Age:                 j.NextNode.Age,
			Title:               j.NextNode.Title,
			Events:              j.NextNode.Events,
			Thoughts:            j.NextNode.Thoughts,
			PersonalitySnapshot: j.NextNode.PersonalitySnapshot,
			TraitChanges:        traits,
			Entities:            entities,
			Scene:               scene,
		},
		ProfileUpdates:   j.ProfileUpdates,
		AffectsHistory:   j.AffectsHistory,
		WorldLineChanged: j.WorldLineChanged,
		MajorEvents:      j.MajorEvents,
		WorldLineDelta:   j.WorldLineDelta,
	}
	if p.WorldLineDelta.HistoricalTrend == "" {
		p.WorldLineDelta.HistoricalTrend = p.WorldLineDelta.HistoricalTrendPatch
	}
	return p, nil
}

func hasAppendableWorldLineEvents(delta *GameApplyChoicePayload, lastWLYear int) bool {
	if delta == nil {
		return false
	}
	for _, ev := range delta.WorldLineDelta.Events {
		if ev.Year > lastWLYear {
			return true
		}
	}
	return false
}

func hasWorldLineTextPatches(delta *GameApplyChoicePayload) bool {
	if delta == nil {
		return false
	}
	d := delta.WorldLineDelta
	return strings.TrimSpace(d.EraSummary) != "" ||
		strings.TrimSpace(d.HistoricalTrend) != "" ||
		strings.TrimSpace(d.DailyLifeContext) != ""
}

// SanitizeGameWorldLineDelta 抉择未改变世界走向时，清除事件上的分歧标注，避免误归因于本节点。
func SanitizeGameWorldLineDelta(p *GameApplyChoicePayload) {
	if p == nil {
		return
	}
	changed := p.AffectsHistory
	if p.WorldLineChanged != nil {
		changed = *p.WorldLineChanged
	}
	if changed {
		return
	}
	for i := range p.WorldLineDelta.Events {
		p.WorldLineDelta.Events[i].DivergenceNote = ""
		p.WorldLineDelta.Events[i].CausedByNodeSeq = nil
	}
}

// ShouldSkipWorldLineDelta 无宏观变化、无背景大事增量、无 era 补丁时跳过世界线写入。
func ShouldSkipWorldLineDelta(p *GameApplyChoicePayload, lastWLYear int) bool {
	if p == nil {
		return true
	}
	if hasAppendableWorldLineEvents(p, lastWLYear) || hasWorldLineTextPatches(p) {
		return false
	}
	if p.WorldLineChanged != nil {
		return !*p.WorldLineChanged && len(p.MajorEvents) == 0
	}
	return !p.AffectsHistory
}
