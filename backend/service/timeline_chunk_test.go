package service

import "testing"

func TestSplitTimelineRangeSingleChunk(t *testing.T) {
	specs := splitTimelineRange(700, 730, 10)
	if len(specs) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(specs))
	}
	if specs[0].TargetNodes != 10 {
		t.Fatalf("expected 10 nodes, got %d", specs[0].TargetNodes)
	}
}

func TestSplitTimelineRangeMultipleChunks(t *testing.T) {
	specs := splitTimelineRange(700, 800, 30)
	if len(specs) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(specs))
	}
	totalNodes := 0
	for i, s := range specs {
		if s.ChunkIndex != i {
			t.Fatalf("chunk index mismatch")
		}
		if s.TotalChunks != len(specs) {
			t.Fatalf("total chunks mismatch")
		}
		totalNodes += s.TargetNodes
	}
	if totalNodes != 30 {
		t.Fatalf("node count sum %d want 30", totalNodes)
	}
	if specs[0].StartYear != 700 {
		t.Fatalf("first chunk start year")
	}
	if specs[len(specs)-1].EndYear != 800 {
		t.Fatalf("last chunk end year")
	}
}

func TestShouldChunkTimeline(t *testing.T) {
	if shouldChunkTimeline(30, 10) {
		t.Fatal("short span should not chunk")
	}
	if !shouldChunkTimeline(100, 10) {
		t.Fatal("long span should chunk")
	}
	if !shouldChunkTimeline(30, 20) {
		t.Fatal("many nodes should chunk")
	}
}
