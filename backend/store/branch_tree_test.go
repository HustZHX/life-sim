package store

import (
	"testing"
	"time"

	"life-sim/backend/model"
)

func TestBuildBranchTree(t *testing.T) {
	now := time.Now()
	versions := []model.TimelineVersion{
		{ID: "a", ParentVersionID: "", BranchLabel: "主枝", CreatedAt: now},
		{ID: "b", ParentVersionID: "a", BranchLabel: "分支1", CreatedAt: now.Add(time.Minute)},
		{ID: "c", ParentVersionID: "a", BranchLabel: "分支2", CreatedAt: now.Add(2 * time.Minute)},
	}
	roots := BuildBranchTree(versions, "c")
	if len(roots) != 1 || roots[0].ID != "a" {
		t.Fatalf("expected single root a, got %+v", roots)
	}
	if len(roots[0].Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(roots[0].Children))
	}
	if !roots[0].Children[1].IsActive {
		t.Fatal("c should be active")
	}
}
