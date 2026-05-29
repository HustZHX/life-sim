package store

import (
	"encoding/json"
	"fmt"
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
	s, _ := ParseSceneJSON(json.RawMessage(raw))
	return s
}

// ParseSceneJSON 容错解析 AI 返回的 scene（对象、JSON 字符串或纯文本场景描述）。
func ParseSceneJSON(data json.RawMessage) (*model.NodeScene, error) {
	if len(data) == 0 || string(data) == "null" {
		return nil, nil
	}
	var s model.NodeScene
	if err := json.Unmarshal(data, &s); err == nil {
		return NormalizeNodeScene(&s), nil
	}
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		str = strings.TrimSpace(str)
		if str == "" {
			return nil, nil
		}
		if strings.HasPrefix(str, "{") {
			return parseSceneJSON(str), nil
		}
		return NormalizeNodeScene(&model.NodeScene{Scene: str}), nil
	}
	return nil, fmt.Errorf("invalid scene format")
}
