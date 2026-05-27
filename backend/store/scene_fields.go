package store

import (
	"encoding/json"
	"strings"

	"life-sim/backend/model"
)

func NormalizeNodeScene(s *model.NodeScene) *model.NodeScene {
	if s == nil {
		return nil
	}
	out := &model.NodeScene{
		DateTime:  strings.TrimSpace(s.DateTime),
		TimeOfDay: strings.TrimSpace(s.TimeOfDay),
		Season:    strings.TrimSpace(s.Season),
		Weather:   strings.TrimSpace(s.Weather),
		Scene:     strings.TrimSpace(s.Scene),
	}
	if out.DateTime == "" && out.TimeOfDay == "" && out.Season == "" && out.Weather == "" && out.Scene == "" {
		return nil
	}
	return out
}

func marshalScene(s *model.NodeScene) string {
	if s == nil {
		return "{}"
	}
	b, err := json.Marshal(s)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func parseSceneJSON(raw string) *model.NodeScene {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		return nil
	}
	var s model.NodeScene
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return nil
	}
	return NormalizeNodeScene(&s)
}
