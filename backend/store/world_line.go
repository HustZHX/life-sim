package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"life-sim/backend/model"
)

type worldLineJSON struct {
	TimelineID       string               `json:"timeline_id"`
	EraSummary       string               `json:"era_summary,omitempty"`
	HistoricalTrend  string               `json:"historical_trend"`
	DailyLifeContext string               `json:"daily_life_context,omitempty"`
	Events           []worldLineEventJSON `json:"events"`
	StartYear        int                  `json:"start_year,omitempty"`
	EndYear          int                  `json:"end_year,omitempty"`
	UpdatedAt        time.Time            `json:"updated_at,omitempty"`
}

type worldLineEventJSON struct {
	Year            int             `json:"year"`
	Name            string          `json:"name"`
	Description     string          `json:"description,omitempty"`
	Impact          string          `json:"impact,omitempty"`
	DivergenceNote  string          `json:"divergence_note,omitempty"`
	CausedByNodeSeq json.RawMessage `json:"caused_by_node_sequence,omitempty"`
}

// parseFlexibleIntPtr 兼容 AI 返回的 null、整数、浮点或整数数组（取首个有效值）。
func parseFlexibleIntPtr(raw json.RawMessage) *int {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil
	}
	var i int
	if err := json.Unmarshal(raw, &i); err == nil {
		return &i
	}
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		v := int(f)
		return &v
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err == nil {
		for _, item := range arr {
			if v := parseFlexibleIntPtr(item); v != nil {
				return v
			}
		}
		return nil
	}
	return nil
}

func worldLineFromJSON(parsed worldLineJSON) model.WorldLine {
	wl := model.WorldLine{
		TimelineID:       parsed.TimelineID,
		EraSummary:       parsed.EraSummary,
		HistoricalTrend:  parsed.HistoricalTrend,
		DailyLifeContext: parsed.DailyLifeContext,
		StartYear:        parsed.StartYear,
		EndYear:          parsed.EndYear,
		UpdatedAt:        parsed.UpdatedAt,
	}
	for _, ev := range parsed.Events {
		wl.Events = append(wl.Events, model.WorldLineEvent{
			Year:            ev.Year,
			Name:            ev.Name,
			Description:     ev.Description,
			Impact:          ev.Impact,
			DivergenceNote:  ev.DivergenceNote,
			CausedByNodeSeq: parseFlexibleIntPtr(ev.CausedByNodeSeq),
		})
	}
	return wl
}

func ParseWorldLine(raw, timelineID string) (*model.WorldLine, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("世界线为空")
	}
	var parsed worldLineJSON
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, fmt.Errorf("世界线解析失败: %w", err)
	}
	wl := worldLineFromJSON(parsed)
	wl.TimelineID = timelineID
	normalizeWorldLine(&wl)
	return &wl, nil
}

func MustWorldLineJSON(wl *model.WorldLine) string {
	if wl == nil {
		return ""
	}
	normalizeWorldLine(wl)
	b, _ := json.Marshal(wl)
	return string(b)
}

func normalizeWorldLine(wl *model.WorldLine) {
	if wl == nil {
		return
	}
	if strings.TrimSpace(wl.HistoricalTrend) == "" {
		wl.HistoricalTrend = wl.EraSummary
	}
	sort.Slice(wl.Events, func(i, j int) bool {
		if wl.Events[i].Year != wl.Events[j].Year {
			return wl.Events[i].Year < wl.Events[j].Year
		}
		return wl.Events[i].Name < wl.Events[j].Name
	})
	if wl.EndYear <= 0 && len(wl.Events) > 0 {
		wl.EndYear = wl.Events[len(wl.Events)-1].Year
	}
	if wl.StartYear <= 0 && len(wl.Events) > 0 {
		wl.StartYear = wl.Events[0].Year
	}
}

func WorldLineFromEraContext(timelineID string, startYear, endYear int, era *eraContextResult) *model.WorldLine {
	if era == nil {
		return nil
	}
	wl := &model.WorldLine{
		TimelineID:       timelineID,
		EraSummary:       era.EraSummary,
		HistoricalTrend:  era.EraSummary,
		DailyLifeContext: era.DailyLifeContext,
		StartYear:        startYear,
		EndYear:          endYear,
		UpdatedAt:        time.Now(),
	}
	for _, e := range era.MajorEvents {
		wl.Events = append(wl.Events, model.WorldLineEvent{
			Year:        e.Year,
			Name:        e.Name,
			Impact:      e.Impact,
			Description: e.Impact,
		})
	}
	normalizeWorldLine(wl)
	return wl
}

func ParseEraContextStruct(raw string) (*eraContextResult, error) {
	var r eraContextResult
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func MarshalNodesForWorldLine(nodes []model.LifeNode) string {
	type lite struct {
		Sequence int    `json:"sequence"`
		Year     int    `json:"year"`
		Title    string `json:"title"`
		Events   string `json:"events"`
	}
	out := make([]lite, len(nodes))
	for i, n := range nodes {
		events := n.Events
		if len([]rune(events)) > 200 {
			events = string([]rune(events)[:200]) + "…"
		}
		out[i] = lite{Sequence: n.Sequence, Year: n.Year, Title: n.Title, Events: events}
	}
	b, _ := json.Marshal(out)
	return string(b)
}

func MarshalWorldLineForPrompt(wl *model.WorldLine) string {
	if wl == nil {
		return "[]"
	}
	b, _ := json.Marshal(wl)
	return string(b)
}

type worldLineNodeUpdate struct {
	Sequence            int    `json:"sequence"`
	Title               string `json:"title,omitempty"`
	Events              string `json:"events,omitempty"`
	Thoughts            string `json:"thoughts,omitempty"`
	PersonalitySnapshot string `json:"personality_snapshot,omitempty"`
}

func ParseWorldLineNodeUpdates(raw string) ([]worldLineNodeUpdate, error) {
	var resp struct {
		NodeUpdates []worldLineNodeUpdate `json:"node_updates"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, err
	}
	return resp.NodeUpdates, nil
}

func ApplyWorldLineNodeUpdates(nodes []model.LifeNode, updates []worldLineNodeUpdate) []model.LifeNode {
	if len(updates) == 0 {
		return nodes
	}
	bySeq := make(map[int]worldLineNodeUpdate, len(updates))
	for _, u := range updates {
		bySeq[u.Sequence] = u
	}
	out := make([]model.LifeNode, len(nodes))
	for i, n := range nodes {
		out[i] = n
		if u, ok := bySeq[n.Sequence]; ok {
			if s := strings.TrimSpace(u.Title); s != "" {
				out[i].Title = s
			}
			if s := strings.TrimSpace(u.Events); s != "" {
				out[i].Events = s
			}
			if s := strings.TrimSpace(u.Thoughts); s != "" {
				out[i].Thoughts = s
			}
			if s := strings.TrimSpace(u.PersonalitySnapshot); s != "" {
				out[i].PersonalitySnapshot = s
			}
		}
	}
	return out
}
