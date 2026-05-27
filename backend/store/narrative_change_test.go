package store

import (
	"testing"

	"life-sim/backend/model"
)

func TestApplyNarrativeChangePlan_insertAfter(t *testing.T) {
	orig := []model.LifeNode{
		{Sequence: 1, Year: 220, Title: "A", Events: "a"},
		{Sequence: 2, Year: 221, Title: "B", Events: "b"},
		{Sequence: 3, Year: 222, Title: "托孤", Events: "白帝城托孤"},
		{Sequence: 4, Year: 225, Title: "D", Events: "d"},
	}
	plan := &NarrativeChangePlan{
		ChangeSummary:  "托孤后插入诸葛亮病逝次日",
		AnchorSequence: 4,
		Operations: []NarrativeChangeOperation{
			{
				Op: "insert_after", AfterSequence: 3, Year: 223, Age: 17,
				Title: "丞相驾崩次日", Events: "诸葛亮病逝于五丈原，翌日消息传至...", Thoughts: "天崩地裂",
				PersonalitySnapshot: "惊恸",
			},
		},
	}
	locked, err := ApplyNarrativeChangePlan(orig, plan, "char1", "刘禅")
	if err != nil {
		t.Fatal(err)
	}
	if len(locked) != 4 {
		t.Fatalf("want 4 locked nodes, got %d", len(locked))
	}
	if locked[3].Title != "丞相驾崩次日" {
		t.Fatalf("last node title=%q", locked[3].Title)
	}
	if locked[3].Sequence != 4 {
		t.Fatalf("last node sequence=%d", locked[3].Sequence)
	}
}

func TestApplyNarrativeChangePlan_modify(t *testing.T) {
	orig := []model.LifeNode{
		{Sequence: 1, Title: "A", Events: "a"},
		{Sequence: 2, Title: "B", Events: "b"},
	}
	plan := &NarrativeChangePlan{
		ChangeSummary:  "修改节点2",
		AnchorSequence: 2,
		Operations: []NarrativeChangeOperation{
			{Op: "modify", Sequence: 2, Events: "新经历"},
		},
	}
	locked, err := ApplyNarrativeChangePlan(orig, plan, "c", "X")
	if err != nil {
		t.Fatal(err)
	}
	if locked[1].Events != "新经历" {
		t.Fatalf("events=%q", locked[1].Events)
	}
}
