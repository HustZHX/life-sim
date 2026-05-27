package store

import "testing"

func TestComputeStepYears(t *testing.T) {
	if got := ComputeStepYears(1000, 20); got != 50 {
		t.Fatalf("1000y/20 nodes = 50, got %d", got)
	}
	if got := ComputeStepYears(60, 20); got != 3 {
		t.Fatalf("60y/20 nodes = 3, got %d", got)
	}
	if got := ComputeStepYears(10, 20); got != 1 {
		t.Fatalf("10y/20 nodes min 1, got %d", got)
	}
}
