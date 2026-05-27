package store

import (
	"testing"
	"time"

	"life-sim/backend/model"
)

func TestVersionCreatesBranch(t *testing.T) {
	cases := []struct {
		label   string
		summary string
		want    bool
	}{
		{"推演后续 · 节点#3", "编辑节点 #2 后推演后续（+2 节点）", true},
		{"更新节点 #3", "重算节点 #2 内心与性格", false},
		{"对话影响 #2", "对话影响：更新节点 #2 内心与性格", false},
		{"", "叙述变更：托孤后插入节点", true},
	}
	for _, c := range cases {
		v := model.TimelineVersion{ParentVersionID: "p", BranchLabel: c.label, ChangeSummary: c.summary}
		if got := VersionCreatesBranch(v); got != c.want {
			t.Fatalf("%q / %q => %v want %v", c.label, c.summary, got, c.want)
		}
	}
}

func TestPrepareBranchTreeVersions_collapsesInnerCurrent(t *testing.T) {
	now := time.Now()
	versions := []model.TimelineVersion{
		{ID: "root", ParentVersionID: "", BranchLabel: "主枝", CreatedAt: now},
		{ID: "inner", ParentVersionID: "root", BranchLabel: "更新节点 #2", ChangeSummary: "重算节点 #2 内心与性格", CreatedAt: now.Add(time.Minute)},
		{ID: "fork", ParentVersionID: "inner", BranchLabel: "推演后续 · 节点#2", ChangeSummary: "编辑节点 #2 后推演后续（+1 节点）", CreatedAt: now.Add(2 * time.Minute)},
	}
	prepared := PrepareBranchTreeVersions(versions)
	if len(prepared) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(prepared))
	}
	if prepared[1].ID != "fork" || prepared[1].ParentVersionID != "root" {
		t.Fatalf("fork should attach to root, got %+v", prepared[1])
	}
}

func TestResolveOverviewVersionID_prefersActiveChain(t *testing.T) {
	now := time.Now()
	versions := []model.TimelineVersion{
		{ID: "root", ParentVersionID: "", CreatedAt: now},
		{ID: "fork", ParentVersionID: "root", BranchLabel: "推演后续 · 节点#1", ChangeSummary: "编辑节点 #1 后推演后续（+1 节点）", CreatedAt: now.Add(time.Minute)},
		{ID: "inner", ParentVersionID: "fork", BranchLabel: "更新节点 #2", ChangeSummary: "重算节点 #2 内心与性格", CreatedAt: now.Add(2 * time.Minute)},
	}
	if got := ResolveOverviewVersionID(versions, "fork", "inner"); got != "inner" {
		t.Fatalf("expected inner, got %s", got)
	}
	if got := ResolveOverviewVersionID(versions, "fork", "root"); got != "fork" {
		t.Fatalf("expected fork when active elsewhere, got %s", got)
	}
}
