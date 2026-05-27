package store

import (
	"testing"
)

func TestParsePersonalityImpactNoImpact(t *testing.T) {
	raw := `{"has_impact": false, "summary": "", "thoughts": "", "personality_snapshot": "", "trait_changes": []}`
	impact, err := ParsePersonalityImpact(raw)
	if err != nil {
		t.Fatal(err)
	}
	if impact.HasImpact {
		t.Fatal("expected no impact")
	}
}

func TestParsePersonalityImpactWithTraits(t *testing.T) {
	raw := `{
		"has_impact": true,
		"summary": "对话动摇信念",
		"thoughts": "我开始怀疑",
		"personality_snapshot": "更加谨慎",
		"trait_changes": [{"field": "worldview", "before": "旧", "after": "新", "reason": "对话"}]
	}`
	impact, err := ParsePersonalityImpact(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !impact.HasImpact {
		t.Fatal("expected impact")
	}
	if len(impact.TraitChanges) != 1 {
		t.Fatalf("traits=%d", len(impact.TraitChanges))
	}
	if impact.TraitChanges[0].Field != "世界观" {
		t.Fatalf("field=%q", impact.TraitChanges[0].Field)
	}
}

func TestParsePersonalityImpactFalseWhenEmptyFields(t *testing.T) {
	raw := `{"has_impact": true, "summary": "x", "thoughts": "", "personality_snapshot": "", "trait_changes": []}`
	impact, err := ParsePersonalityImpact(raw)
	if err != nil {
		t.Fatal(err)
	}
	if impact.HasImpact {
		t.Fatal("empty fields should clear has_impact")
	}
}
