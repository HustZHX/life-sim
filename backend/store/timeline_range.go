package store

import (
	"fmt"
	"time"

	"life-sim/backend/model"
)

const (
	DefaultTargetNodeCount = 20
	MinTargetNodeCount     = 1
	MaxTargetNodeCount     = 25
	// LivingProfileStepYears 未完结人生按节点推演时的参考间隔（年），与推演余生一致。
	LivingProfileStepYears = 3
)

// ClampTargetNodeCount 将目标节点数限制在余生推演粒度 1～25。
func ClampTargetNodeCount(n int) int {
	if n <= 0 {
		return DefaultTargetNodeCount
	}
	if n < MinTargetNodeCount {
		return MinTargetNodeCount
	}
	if n > MaxTargetNodeCount {
		return MaxTargetNodeCount
	}
	return n
}

// TimelineRangeConfig 时间轴生成区间与节点规模。
type TimelineRangeConfig struct {
	TargetNodeCount int
	StepYears       int
	StartYear       int
	EndYear         int
}

// ComputeStepYears 按跨度与目标节点数计算参考间隔（年）。
func ComputeStepYears(spanYears, targetNodes int) int {
	if targetNodes <= 0 {
		targetNodes = DefaultTargetNodeCount
	}
	if spanYears <= 0 {
		return 1
	}
	step := spanYears / targetNodes
	if step < 1 {
		return 1
	}
	return step
}

// EffectiveDeathYear 档案未标注卒年或仍在世时，用当前年份作为时间轴上界。
func EffectiveDeathYear(birthYear, deathYear int) int {
	if deathYear > birthYear {
		return deathYear
	}
	y := time.Now().Year()
	if y > birthYear {
		return y
	}
	if birthYear > 0 {
		return birthYear + 1
	}
	return y
}

// IsLivingProfile 卒年未填或不大于生年，视为人生未完结。
func IsLivingProfile(profile *model.Profile) bool {
	if profile == nil {
		return false
	}
	return profile.DeathYear <= 0 || profile.DeathYear <= profile.BirthYear
}

// ResolveLivingTimelineRange 未完结人生：只推演从出生起的前 N 个节点，不拉伸至当前年份。
func ResolveLivingTimelineRange(targetNodeCount, birthYear int) (startYear, endYear, stepYears int) {
	targetNodeCount = ClampTargetNodeCount(targetNodeCount)
	stepYears = LivingProfileStepYears
	startYear = birthYear
	if startYear <= 0 {
		startYear = 1
	}
	endYear = startYear + stepYears*targetNodeCount
	if endYear <= startYear {
		endYear = startYear + stepYears
	}
	return startYear, endYear, stepYears
}

// ResolveTimelineRange 补全并校验时间轴生成区间，按目标节点数推算参考间隔。
func ResolveTimelineRange(cfg TimelineRangeConfig, profile *model.Profile) (TimelineRangeConfig, error) {
	out := cfg
	out.TargetNodeCount = ClampTargetNodeCount(out.TargetNodeCount)
	if profile == nil {
		return out, fmt.Errorf("档案不存在")
	}
	if IsLivingProfile(profile) {
		out.StartYear, out.EndYear, out.StepYears = ResolveLivingTimelineRange(out.TargetNodeCount, profile.BirthYear)
		return out, nil
	}
	effectiveDeath := EffectiveDeathYear(profile.BirthYear, profile.DeathYear)
	if out.StartYear <= 0 {
		out.StartYear = profile.BirthYear
	}
	if out.EndYear <= 0 {
		out.EndYear = effectiveDeath
	}
	if out.StartYear > out.EndYear {
		return out, fmt.Errorf("起始年份不能晚于结束年份")
	}
	if out.StartYear < profile.BirthYear {
		out.StartYear = profile.BirthYear
	}
	if out.EndYear > effectiveDeath {
		out.EndYear = effectiveDeath
	}
	span := out.EndYear - out.StartYear
	out.StepYears = ComputeStepYears(span, out.TargetNodeCount)
	return out, nil
}

// ResolveRegenerateTailRange 推演余生：节点数由用户配置，不与寿命跨度机械对应。
func ResolveRegenerateTailRange(_anchorYear, _deathYear, targetNodeCount int) (stepYears, targetNodes int) {
	targetNodeCount = ClampTargetNodeCount(targetNodeCount)
	return LivingProfileStepYears, targetNodeCount
}
