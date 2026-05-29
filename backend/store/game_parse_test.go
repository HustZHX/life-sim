package store

import (
	"testing"

	"life-sim/backend/model"
)

func TestShouldSkipWorldLineDelta_explicitNoChange(t *testing.T) {
	f := false
	p := &gameApplyChoicePayload{
		WorldLineChanged: &f,
		MajorEvents:      nil,
	}
	if !ShouldSkipWorldLineDelta(p, 200) {
		t.Fatal("expected skip when world_line_changed=false and no delta")
	}
}

func TestParseGameApplyChoice_profileUpdatesNumber(t *testing.T) {
	raw := `{
		"next_node": {
			"year": 1910, "age": 8, "title": "童年", "events": "读书。",
			"thoughts": "", "personality_snapshot": "",
			"trait_changes": [], "entities": {}, "scene": {}
		},
		"profile_updates": { "social_class": 0, "occupation": "童生" },
		"world_line_changed": false,
		"world_line_delta": {}
	}`
	p, err := ParseGameApplyChoice(raw)
	if err != nil {
		t.Fatal(err)
	}
	prof := &model.Profile{}
	if err := ApplyProfileUpdatesFromJSON(prof, p.ProfileUpdates); err != nil {
		t.Fatal(err)
	}
	if prof.SocialClass != "0" || prof.Occupation != "童生" {
		t.Fatalf("prof=%+v", prof)
	}
}

func TestParseGameApplyChoice_sceneString(t *testing.T) {
	raw := `{
		"next_node": {
			"year": 1910, "age": 8, "title": "童年一日", "events": "在私塾读书。",
			"thoughts": "想出去玩", "personality_snapshot": "顽皮",
			"trait_changes": [], "entities": {}, "scene": "隆中草庐·午后晴"
		},
		"profile_updates": {},
		"world_line_changed": false,
		"world_line_delta": {}
	}`
	p, err := ParseGameApplyChoice(raw)
	if err != nil {
		t.Fatal(err)
	}
	if p.NextNode.Scene == nil || p.NextNode.Scene.Scene != "隆中草庐·午后晴" {
		t.Fatalf("scene=%+v", p.NextNode.Scene)
	}
}

func TestParseGameApplyChoice_traitChangesString(t *testing.T) {
	raw := `{
		"next_node": {
			"year": 1910, "age": 8, "title": "童年一日", "events": "在私塾读书。",
			"thoughts": "想出去玩", "personality_snapshot": "顽皮",
			"trait_changes": "性格更加沉稳"
		},
		"profile_updates": {},
		"world_line_changed": false,
		"world_line_delta": {}
	}`
	p, err := ParseGameApplyChoice(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.NextNode.TraitChanges) != 1 {
		t.Fatalf("traits=%d", len(p.NextNode.TraitChanges))
	}
}

func TestShouldSkipWorldLineDelta_newEvent(t *testing.T) {
	f := false
	p := &gameApplyChoicePayload{
		WorldLineChanged: &f,
		MajorEvents:      []string{"赤壁之战"},
	}
	if ShouldSkipWorldLineDelta(p, 200) {
		t.Fatal("expected apply when major_events non-empty")
	}
	p2 := &gameApplyChoicePayload{WorldLineChanged: &f}
	p2.WorldLineDelta.Events = []model.WorldLineEvent{{Year: 208, Name: "新事件"}}
	if ShouldSkipWorldLineDelta(p2, 200) {
		t.Fatal("expected apply when appendable events present")
	}
}
