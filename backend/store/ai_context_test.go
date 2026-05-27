package store

import (
	"encoding/json"
	"strings"
	"testing"

	"life-sim/backend/model"
)

func sampleProfile() *model.Profile {
	return &model.Profile{
		CharacterID: "c1", DisplayName: "李白", BirthYear: 701, DeathYear: 762,
		Era: "盛唐", PersonalityInitial: "豪放", BeliefsMotto: "天生我材必有用",
		SourcesNote: "史料有争议", TemplateSource: "杜甫",
	}
}

func sampleNode(seq int) model.LifeNode {
	return model.LifeNode{
		ID: "id-" + string(rune('a'+seq)), CharacterID: "c1", VersionID: "v1",
		Sequence: seq, Year: 700 + seq*10, Age: seq * 10,
		Title: "节点", Events: "经历", Thoughts: "想法",
		PersonalitySnapshot: "性格",
		Entities:            &model.NodeEntities{Persons: []string{"杜甫"}},
		Scene:               &model.NodeScene{Weather: "晴"},
	}
}

func TestProfileJSONInnerSmallerThanFull(t *testing.T) {
	p := sampleProfile()
	full := ProfileJSONFull(p)
	inner := ProfileJSONInner(p)
	if len(inner) >= len(full) {
		t.Fatalf("inner should be smaller: full=%d inner=%d", len(full), len(full))
	}
	if strings.Contains(inner, "sources_note") {
		t.Fatal("inner should not contain sources_note key context")
	}
}

func TestMarshalNodesLiteOmitsIDs(t *testing.T) {
	nodes := []model.LifeNode{sampleNode(0), sampleNode(1), sampleNode(2)}
	raw := MarshalNodesLiteBeforeSequence(nodes, 2)
	if strings.Contains(raw, `"id"`) {
		t.Fatal("lite nodes should not contain id")
	}
	if strings.Contains(raw, "entities") {
		t.Fatal("lite nodes should not contain entities")
	}
	var parsed []nodeContextLite
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		t.Fatal(err)
	}
	if len(parsed) != 2 {
		t.Fatalf("expected 2 prior nodes, got %d", len(parsed))
	}
}

func TestMarshalNodesLockedHasThoughts(t *testing.T) {
	raw := MarshalNodesLocked([]model.LifeNode{sampleNode(0)})
	if !strings.Contains(raw, "thoughts") {
		t.Fatal("locked nodes should include thoughts")
	}
	if strings.Contains(raw, `"id"`) {
		t.Fatal("locked nodes should not contain id")
	}
}

func TestMarshalNodesTailInputMinimal(t *testing.T) {
	raw := MarshalNodesTailInput([]model.LifeNode{sampleNode(1)})
	if strings.Contains(raw, "personality_snapshot") {
		t.Fatal("tail input should not include personality_snapshot")
	}
	if strings.Contains(raw, "thoughts") {
		t.Fatal("tail input should not include thoughts")
	}
}

func TestMergeTimelineNodeChunks(t *testing.T) {
	c1 := []model.LifeNode{{Sequence: 99, Year: 1}, {Sequence: 99, Year: 2}}
	c2 := []model.LifeNode{{Sequence: 99, Year: 3}}
	merged := MergeTimelineNodeChunks(c1, c2)
	if len(merged) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(merged))
	}
	for i, n := range merged {
		if n.Sequence != i {
			t.Fatalf("sequence %d want %d", n.Sequence, i)
		}
	}
}

func TestProfileHashStable(t *testing.T) {
	p := sampleProfile()
	h1 := ProfileHash(p)
	h2 := ProfileHash(p)
	if h1 != h2 {
		t.Fatal("hash should be stable")
	}
	p2 := *p
	p2.DeathYear = 800
	if ProfileHash(&p2) == h1 {
		t.Fatal("hash should change when profile changes")
	}
}
