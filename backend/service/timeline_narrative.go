package service

import (
	"context"
	"fmt"

	"life-sim/backend/model"
	"life-sim/backend/store"
)

func skeletonPromptName(mode string) string {
	if mode == model.ModeFamous {
		return "generate_timeline_skeleton_famous.txt"
	}
	return "generate_timeline_skeleton_random.txt"
}

func expandPromptName(mode string) string {
	if mode == model.ModeFamous {
		return "expand_timeline_nodes_famous.txt"
	}
	return "expand_timeline_nodes_random.txt"
}

func (s *CharacterService) generateRichTimelineNodes(
	ctx context.Context,
	jobID string,
	characterID string,
	characterMode string,
	profile *model.Profile,
	profileJSON string,
	apiModel string,
	cfg timelineJobConfig,
	versionID string,
) ([]model.LifeNode, error) {
	span := cfg.EndYear - cfg.StartYear
	yearChunks := splitTimelineRange(cfg.StartYear, cfg.EndYear, cfg.TargetNodeCount)

	skeletonPrompt, err := s.ai.LoadPrompt(skeletonPromptName(characterMode))
	if err != nil {
		return nil, err
	}
	expandPrompt, err := s.ai.LoadPrompt(expandPromptName(characterMode))
	if err != nil {
		return nil, err
	}

	var allSkeleton []store.TimelineSkeletonNode
	seqOffset := 0

	for i, yc := range yearChunks {
		progress := 10 + (i * 25 / max(len(yearChunks), 1))
		setJobStage(s.store, jobID, progress, fmt.Sprintf("正在规划史实骨架（第 %d/%d 段）…", i+1, len(yearChunks)))

		chunkCfg := cfg
		chunkCfg.StartYear = yc.StartYear
		chunkCfg.EndYear = yc.EndYear
		chunkCfg.TargetNodeCount = yc.TargetNodes

		var user string
		if yc.TotalChunks == 1 {
			user = buildTimelineUserContent(profileJSON, chunkCfg, span)
		} else {
			user = buildTimelineChunkUser(profileJSON, chunkCfg, yc, span)
		}
		user = "【骨架阶段】只输出史实/人生锚点，不写长叙事。\n" + user

		raw, err := s.ai.ChatJSONModel(ctx, apiModel, skeletonPrompt, user)
		if err != nil {
			return nil, fmt.Errorf("骨架生成失败: %w", err)
		}
		chunkSk, err := store.ParseTimelineSkeleton(raw)
		if err != nil {
			return nil, fmt.Errorf("骨架解析失败: %w", err)
		}
		for j := range chunkSk {
			chunkSk[j].Sequence = seqOffset + j
		}
		seqOffset += len(chunkSk)
		allSkeleton = append(allSkeleton, chunkSk...)
	}

	batches := store.BatchSkeletonSlices(allSkeleton, store.ExpandBatchSize())
	var merged []model.LifeNode
	expandBase := 35

	for bi, batch := range batches {
		progress := expandBase + (bi * 50 / max(len(batches), 1))
		setJobStage(s.store, jobID, progress, fmt.Sprintf("正在叙事扩写（第 %d/%d 批，共 %d 节点）…", bi+1, len(batches), len(batch)))

		user := fmt.Sprintf(
			"人物档案：\n%s\n\n骨架节点（须逐条扩写，sequence 不可变）：\n%s",
			profileJSON, store.MarshalSkeletonBatch(batch),
		)
		if len(merged) > 0 {
			user += "\n\n前置已扩写节点（性格衔接参考）：\n" + store.MarshalSkeletonContextPrior(merged, 2)
		}
		if cfg.Instructions != "" {
			user += "\n\n用户特殊要求/备注（须优先满足）：\n" + cfg.Instructions
		}

		done := make(chan struct{})
		go tickJobProgress(s.store, jobID, progress+1, progress+45/max(len(batches), 1), done)

		raw, err := s.ai.ChatJSONModel(ctx, apiModel, expandPrompt, user)
		close(done)
		if err != nil {
			return nil, fmt.Errorf("叙事扩写失败（第 %d 批）: %w", bi+1, err)
		}
		expanded, err := store.ParseTimelineExpandBatch(raw, profile.DisplayName)
		if err != nil {
			return nil, fmt.Errorf("扩写解析失败（第 %d 批）: %w", bi+1, err)
		}
		part, err := store.MergeSkeletonAndExpand(batch, expanded, characterID, versionID, profile.DisplayName)
		if err != nil {
			return nil, err
		}
		merged = append(merged, part...)
	}

	return merged, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
