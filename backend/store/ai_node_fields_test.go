package store

import (
	"encoding/json"
	"testing"
)

func TestParseSceneJSON_string(t *testing.T) {
	s, err := ParseSceneJSON(json.RawMessage(`"街亭道旁·黄昏"`))
	if err != nil {
		t.Fatal(err)
	}
	if s == nil || s.Scene != "街亭道旁·黄昏" {
		t.Fatalf("got %+v", s)
	}
}

func TestParseEntitiesJSON_emptyObject(t *testing.T) {
	e, err := ParseEntitiesJSON(json.RawMessage(`{}`), "刘备")
	if err != nil {
		t.Fatal(err)
	}
	if e == nil || len(e.Protagonist) != 1 || e.Protagonist[0] != "刘备" {
		t.Fatalf("got %+v", e)
	}
}

func TestParseTimelineNodes_sceneString(t *testing.T) {
	raw := `{"nodes":[{"sequence":1,"year":200,"age":20,"title":"t","events":"e","thoughts":"","personality_snapshot":"","trait_changes":[],"entities":{},"scene":"军营"}]}`
	nodes, err := ParseTimelineNodes(raw, "c1", "v1", "主角")
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 || nodes[0].Scene == nil || nodes[0].Scene.Scene != "军营" {
		t.Fatalf("nodes=%+v", nodes)
	}
}
