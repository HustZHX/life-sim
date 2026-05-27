package store

import (
	"encoding/json"
	"strings"

	"life-sim/backend/model"
)

func NormalizeNodeEntities(e *model.NodeEntities, defaultProtagonist string) *model.NodeEntities {
	if e == nil {
		e = &model.NodeEntities{}
	}
	out := &model.NodeEntities{
		Protagonist: dedupeStrings(e.Protagonist),
		Persons:     dedupeStrings(e.Persons),
		Places:      dedupeStrings(e.Places),
	}
	if len(out.Protagonist) == 0 && strings.TrimSpace(defaultProtagonist) != "" {
		out.Protagonist = []string{strings.TrimSpace(defaultProtagonist)}
	}
	protSet := make(map[string]struct{}, len(out.Protagonist))
	for _, p := range out.Protagonist {
		protSet[p] = struct{}{}
	}
	filtered := out.Persons[:0]
	for _, p := range out.Persons {
		if _, ok := protSet[p]; ok {
			continue
		}
		filtered = append(filtered, p)
	}
	out.Persons = filtered
	if len(out.Protagonist) == 0 && len(out.Persons) == 0 && len(out.Places) == 0 {
		return nil
	}
	return out
}

func dedupeStrings(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, raw := range items {
		s := strings.TrimSpace(raw)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func marshalEntities(e *model.NodeEntities) string {
	if e == nil {
		return "{}"
	}
	b, err := json.Marshal(e)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func parseEntitiesJSON(raw string) *model.NodeEntities {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		return nil
	}
	var e model.NodeEntities
	if err := json.Unmarshal([]byte(raw), &e); err != nil {
		return nil
	}
	return NormalizeNodeEntities(&e, "")
}

func sortEntitiesByLength(items []string) []string {
	if len(items) == 0 {
		return items
	}
	out := append([]string(nil), items...)
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if len(out[j]) > len(out[i]) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}
