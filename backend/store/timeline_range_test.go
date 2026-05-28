package store

import (
	"life-sim/backend/model"
	"testing"
)

func TestClampTargetNodeCount(t *testing.T) {
	if got := ClampTargetNodeCount(0); got != DefaultTargetNodeCount {
		t.Fatalf("0 -> default %d, got %d", DefaultTargetNodeCount, got)
	}
	if got := ClampTargetNodeCount(1); got != 1 {
		t.Fatalf("min 1, got %d", got)
	}
	if got := ClampTargetNodeCount(25); got != 25 {
		t.Fatalf("max 25, got %d", got)
	}
	if got := ClampTargetNodeCount(100); got != MaxTargetNodeCount {
		t.Fatalf("clamp max, got %d", got)
	}
}

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

func TestResolveTimelineRangeLivingProfile(t *testing.T) {
	profile := &model.Profile{BirthYear: 180, DeathYear: 0}
	cfg, err := ResolveTimelineRange(TimelineRangeConfig{TargetNodeCount: 5}, profile)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.StartYear != 180 {
		t.Fatalf("start=%d want 180", cfg.StartYear)
	}
	if cfg.EndYear != 195 {
		t.Fatalf("end=%d want 195 (180+3*5), not stretched to current year", cfg.EndYear)
	}
	if cfg.StepYears != LivingProfileStepYears {
		t.Fatalf("step=%d want %d", cfg.StepYears, LivingProfileStepYears)
	}
	if cfg.TargetNodeCount != 5 {
		t.Fatalf("target=%d want 5", cfg.TargetNodeCount)
	}
}

func TestResolveLivingTimelineRange(t *testing.T) {
	start, end, step := ResolveLivingTimelineRange(5, 2000)
	if start != 2000 || end != 2015 || step != 3 {
		t.Fatalf("got %d-%d step=%d", start, end, step)
	}
}
