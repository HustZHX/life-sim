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
