package service

import (
	"fmt"
)

const (
	timelineChunkMaxNodes   = 12
	timelineChunkMinNodes   = 15
	timelineChunkMinSpanYrs = 60
)

type timelineChunkSpec struct {
	StartYear     int
	EndYear       int
	TargetNodes   int
	ChunkIndex    int
	TotalChunks   int
	SequenceStart int
}

func shouldChunkTimeline(span, targetNodes int) bool {
	return targetNodes > timelineChunkMinNodes || span > timelineChunkMinSpanYrs
}

// splitTimelineRange 将时间轴生成任务拆分为 1~4 段。
func splitTimelineRange(startYear, endYear, targetNodes int) []timelineChunkSpec {
	span := endYear - startYear
	if span < 0 {
		span = 0
	}
	if !shouldChunkTimeline(span, targetNodes) {
		return []timelineChunkSpec{{
			StartYear: startYear, EndYear: endYear, TargetNodes: targetNodes,
			ChunkIndex: 0, TotalChunks: 1, SequenceStart: 0,
		}}
	}

	numChunks := 2
	if targetNodes > 24 || span > 120 {
		numChunks = 3
	}
	if targetNodes > 36 || span > 200 {
		numChunks = 4
	}

	chunks := make([]timelineChunkSpec, numChunks)
	nodesPerChunk := targetNodes / numChunks
	extraNodes := targetNodes % numChunks
	yearStep := span / numChunks
	if yearStep < 1 {
		yearStep = 1
	}

	seq := 0
	curYear := startYear
	for i := 0; i < numChunks; i++ {
		chunkNodes := nodesPerChunk
		if i < extraNodes {
			chunkNodes++
		}
		chunkEnd := endYear
		if i < numChunks-1 {
			chunkEnd = curYear + yearStep
			if chunkEnd > endYear {
				chunkEnd = endYear
			}
		}
		chunks[i] = timelineChunkSpec{
			StartYear: curYear, EndYear: chunkEnd, TargetNodes: chunkNodes,
			ChunkIndex: i, TotalChunks: numChunks, SequenceStart: seq,
		}
		seq += chunkNodes
		curYear = chunkEnd
		if curYear >= endYear && i < numChunks-1 {
			curYear = endYear - 1
		}
	}
	chunks[len(chunks)-1].EndYear = endYear
	return chunks
}

func estimateTimelineChunks(targetNodes, span int) int {
	if !shouldChunkTimeline(span, targetNodes) {
		return 1
	}
	specs := splitTimelineRange(0, span, targetNodes)
	return len(specs)
}

func appendTimelineExtras(user string, cfg timelineJobConfig) string {
	if cfg.EraContext != "" {
		user += "\n\n" + cfg.EraContext + "\n（生成节点时须让主角人生与上述时代大事相呼应；节点 year 宜对齐相关大事年份或其后影响期）"
	}
	if cfg.Instructions != "" {
		user += fmt.Sprintf("\n\n用户特殊要求/备注（须优先满足）：\n%s", cfg.Instructions)
	}
	return user
}

func buildTimelineUserContent(profileJSON string, cfg timelineJobConfig, span int) string {
	user := fmt.Sprintf(
		"目标约 %d 个节点，生成区间 start_year=%d, end_year=%d（跨度 %d 年），参考间隔约 %d 年（非强制，以关键事件为准）\n人物档案：\n%s",
		cfg.TargetNodeCount, cfg.StartYear, cfg.EndYear, span, cfg.StepYears, profileJSON,
	)
	return appendTimelineExtras(user, cfg)
}

func buildTimelineUserWithoutProfile(cfg timelineJobConfig, span int) string {
	user := fmt.Sprintf(
		"请根据上文人物档案生成时间轴。目标约 %d 个节点，生成区间 start_year=%d, end_year=%d（跨度 %d 年），参考间隔约 %d 年（非强制，以关键事件为准）",
		cfg.TargetNodeCount, cfg.StartYear, cfg.EndYear, span, cfg.StepYears,
	)
	return appendTimelineExtras(user, cfg)
}

func buildTimelineChunkUserWithoutProfile(cfg timelineJobConfig, chunk timelineChunkSpec, span int) string {
	user := fmt.Sprintf(
		"请根据上文人物档案生成时间轴。【第 %d/%d 段】目标约 %d 个节点，本段区间 start_year=%d, end_year=%d（全轴跨度 %d 年），sequence 从 %d 起递增，参考间隔约 %d 年",
		chunk.ChunkIndex+1, chunk.TotalChunks, chunk.TargetNodes,
		chunk.StartYear, chunk.EndYear, span, chunk.SequenceStart, cfg.StepYears,
	)
	return appendTimelineExtras(user, cfg)
}

func buildTimelineChunkUser(profileJSON string, cfg timelineJobConfig, chunk timelineChunkSpec, span int) string {
	user := fmt.Sprintf(
		"【第 %d/%d 段】目标约 %d 个节点，本段区间 start_year=%d, end_year=%d（全轴跨度 %d 年），sequence 从 %d 起递增，参考间隔约 %d 年\n人物档案：\n%s",
		chunk.ChunkIndex+1, chunk.TotalChunks, chunk.TargetNodes,
		chunk.StartYear, chunk.EndYear, span, chunk.SequenceStart, cfg.StepYears, profileJSON,
	)
	return appendTimelineExtras(user, cfg)
}

func buildTimelineContinuationUser(profileJSON string, cfg timelineJobConfig, chunk timelineChunkSpec, prevTailJSON string) string {
	user := fmt.Sprintf(
		"【第 %d/%d 段续段】目标约 %d 个节点，本段区间 start_year=%d, end_year=%d，sequence 从 %d 起递增\n人物档案：\n%s\n上一段末节点（须衔接）：\n%s",
		chunk.ChunkIndex+1, chunk.TotalChunks, chunk.TargetNodes,
		chunk.StartYear, chunk.EndYear, chunk.SequenceStart, profileJSON, prevTailJSON,
	)
	return appendTimelineExtras(user, cfg)
}
