package store

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type eraMajorEvent struct {
	Year   int    `json:"year"`
	Name   string `json:"name"`
	Impact string `json:"impact"`
}

type eraContextResult struct {
	EraSummary        string          `json:"era_summary"`
	MajorEvents       []eraMajorEvent `json:"major_events"`
	DailyLifeContext  string          `json:"daily_life_context"`
}

func ParseEraContextResult(raw string) (string, error) {
	var r eraContextResult
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		return "", fmt.Errorf("时代背景解析失败: %w", err)
	}
	if strings.TrimSpace(r.EraSummary) == "" && len(r.MajorEvents) == 0 {
		return "", fmt.Errorf("时代背景为空")
	}
	return FormatEraContextForPrompt(&r), nil
}

func FormatEraContextForPrompt(r *eraContextResult) string {
	var b strings.Builder
	b.WriteString("【时代背景与大事记】\n")
	if s := strings.TrimSpace(r.EraSummary); s != "" {
		b.WriteString("时代总述：")
		b.WriteString(s)
		b.WriteByte('\n')
	}
	if len(r.MajorEvents) > 0 {
		events := append([]eraMajorEvent(nil), r.MajorEvents...)
		sort.Slice(events, func(i, j int) bool {
			if events[i].Year != events[j].Year {
				return events[i].Year < events[j].Year
			}
			return events[i].Name < events[j].Name
		})
		b.WriteString("影响生活的大事（按年）：\n")
		for _, e := range events {
			if strings.TrimSpace(e.Name) == "" {
				continue
			}
			year := e.Year
			if year > 0 {
				fmt.Fprintf(&b, "- %d年 %s", year, e.Name)
			} else {
				fmt.Fprintf(&b, "- %s", e.Name)
			}
			if imp := strings.TrimSpace(e.Impact); imp != "" {
				b.WriteString("：")
				b.WriteString(imp)
			}
			b.WriteByte('\n')
		}
	}
	if s := strings.TrimSpace(r.DailyLifeContext); s != "" {
		b.WriteString("日常生活背景：")
		b.WriteString(s)
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}
