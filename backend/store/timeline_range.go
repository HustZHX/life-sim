package store

import (
	"fmt"

	"life-sim/backend/model"
)

const DefaultTargetNodeCount = 20

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

// ResolveTimelineRange 补全并校验时间轴生成区间，按目标节点数推算参考间隔。
func ResolveTimelineRange(cfg TimelineRangeConfig, profile *model.Profile) (TimelineRangeConfig, error) {
	out := cfg
	if out.TargetNodeCount <= 0 {
		out.TargetNodeCount = DefaultTargetNodeCount
	}
	if out.TargetNodeCount < 5 {
		out.TargetNodeCount = 5
	}
	if out.TargetNodeCount > 50 {
		out.TargetNodeCount = 50
	}
	if profile == nil {
		return out, fmt.Errorf("档案不存在")
	}
	if out.StartYear <= 0 {
		out.StartYear = profile.BirthYear
	}
	if out.EndYear <= 0 {
		out.EndYear = profile.DeathYear
	}
	if out.StartYear > out.EndYear {
		return out, fmt.Errorf("起始年份不能晚于结束年份")
	}
	if out.StartYear < profile.BirthYear {
		out.StartYear = profile.BirthYear
	}
	if out.EndYear > profile.DeathYear {
		out.EndYear = profile.DeathYear
	}
	span := out.EndYear - out.StartYear
	out.StepYears = ComputeStepYears(span, out.TargetNodeCount)
	return out, nil
}

// ResolveRegenerateTailRange 重算后续：从锚点年份到去世，按目标节点数推算间隔。
func ResolveRegenerateTailRange(anchorYear, deathYear, targetNodeCount int) (stepYears, targetNodes int) {
	if targetNodeCount <= 0 {
		targetNodeCount = DefaultTargetNodeCount
	}
	if targetNodeCount < 5 {
		targetNodeCount = 5
	}
	if targetNodeCount > 50 {
		targetNodeCount = 50
	}
	span := deathYear - anchorYear
	if span < 0 {
		span = 0
	}
	return ComputeStepYears(span, targetNodeCount), targetNodeCount
}
