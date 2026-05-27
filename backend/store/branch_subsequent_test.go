package store

import (
	"testing"

	"life-sim/backend/model"
)

func TestTrimSubsequentTail(t *testing.T) {
	raw := []model.LifeNode{
		{Sequence: 99, Year: 10, Title: "a"},
		{Sequence: 100, Year: 15, Title: "b"},
		{Sequence: 101, Year: 20, Title: "c"},
	}
	out := TrimSubsequentTail(raw, 2, 2000, 1)
	if len(out) != 1 {
		t.Fatalf("want 1 node got %d", len(out))
	}
	if out[0].Sequence != 3 || out[0].Year != 10 {
		t.Fatalf("got seq=%d year=%d", out[0].Sequence, out[0].Year)
	}
}
