package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"life-sim/backend/model"
	"life-sim/backend/store"
)

const regenerateTailBatchSize = 6

func buildRegenerateTailContext(
	head string,
	batchTarget int,
	profile *model.Profile,
	lockedJSON string,
	anchor *model.LifeNode,
	req model.PatchNodeRequest,
) string {
	user := fmt.Sprintf(
		"%s恰好生成 M=%d 个全新后续节点；anchor_sequence=%d；anchor_year=%d；birth_year=%d\n"+
			"（勿写到某一卒年；year 由事件决定；M=1 时只生成下一个成长/事件阶段）\n"+
			"人物档案：\n%s\n"+
			"已锁定节点（含锚点，须衔接）：\n%s\n"+
			"锚点节点：sequence=%d year=%d age=%d title=%s\nevents=%s",
		head, batchTarget, anchor.Sequence, anchor.Year, profile.BirthYear,
		store.ProfileJSONTimeline(profile), lockedJSON,
		anchor.Sequence, anchor.Year, anchor.Age, anchor.Title, anchor.Events,
	)
	if req.NextNodeTitle != "" {
		user += fmt.Sprintf("\n【下一节点标题】用户指定，title 须使用该标题或与其等价表述：%s\n", req.NextNodeTitle)
	} else if req.Mode == model.PatchModeAppendNext {
		user += "\n【下一节点标题】用户未指定，请自拟贴切标题。\n"
	}
	return user
}

func buildRegenerateTailUser(
	head string,
	batchTarget int,
	profile *model.Profile,
	lockedJSON string,
	anchor *model.LifeNode,
	req model.PatchNodeRequest,
) string {
	user := buildRegenerateTailContext(head, batchTarget, profile, lockedJSON, anchor, req)
	return user + "\nM>1 时单节点宜简练（events 80～150 字），务必输出完整合法 JSON。\n"
}

func buildRegenerateTailSkeletonUser(
	head string,
	batchTarget int,
	profile *model.Profile,
	lockedJSON string,
	anchor *model.LifeNode,
	req model.PatchNodeRequest,
) string {
	return buildRegenerateTailContext(head, batchTarget, profile, lockedJSON, anchor, req)
}

func (s *CharacterService) generateRichRegenerateTail(
	ctx context.Context,
	jobID string,
	apiModel string,
	characterMode string,
	characterID, versionID, displayName string,
	profile *model.Profile,
	head string,
	batchTarget int,
	lockedJSON string,
	anchor *model.LifeNode,
	req model.PatchNodeRequest,
	priorTail []model.LifeNode,
) ([]model.LifeNode, error) {
	skeletonPrompt, err := s.ai.LoadPrompt(tailSkeletonPromptName(characterMode))
	if err != nil {
		return nil, err
	}
	expandPrompt, err := s.ai.LoadPrompt(expandPromptName(characterMode))
	if err != nil {
		return nil, err
	}

	setJobStage(s.store, jobID, 30, fmt.Sprintf("正在规划后续骨架（%d 个节点）…", batchTarget))
	skUser := "【骨架阶段】只输出后续人生锚点，不写长叙事。\n" +
		buildRegenerateTailSkeletonUser(head, batchTarget, profile, lockedJSON, anchor, req)
	raw, err := s.ai.ChatJSONModel(ctx, apiModel, skeletonPrompt, skUser)
	if err != nil {
		return nil, fmt.Errorf("后续骨架生成失败: %w", err)
	}
	skeleton, err := store.ParseTimelineSkeleton(raw)
	if err != nil {
		return nil, fmt.Errorf("后续骨架解析失败: %w", err)
	}
	skeleton = trimTailSkeleton(skeleton, anchor.Sequence, profile.BirthYear, anchor.Year, batchTarget)
	if len(skeleton) == 0 {
		return nil, fmt.Errorf("AI 未返回有效后续骨架")
	}

	profileJSON := store.ProfileJSONTimeline(profile)
	batches := store.BatchSkeletonSlices(skeleton, store.ExpandBatchSize())
	var merged []model.LifeNode
	expandBase := 42

	for bi, batch := range batches {
		progress := expandBase + (bi * 40 / max(len(batches), 1))
		setJobStage(s.store, jobID, progress, fmt.Sprintf("正在叙事扩写后续节点（第 %d/%d 批）…", bi+1, len(batches)))

		contextNodes := append([]model.LifeNode(nil), priorTail...)
		contextNodes = append(contextNodes, merged...)

		done := make(chan struct{})
		go tickJobProgress(s.store, jobID, progress+1, progress+35/max(len(batches), 1), done)

		expanded, err := s.expandSkeletonBatchWithContext(ctx, apiModel, expandPrompt, displayName, profileJSON, "", batch, contextNodes)
		close(done)
		if err != nil {
			return nil, fmt.Errorf("后续扩写失败（第 %d 批）: %w", bi+1, err)
		}
		part, err := store.MergeSkeletonAndExpand(batch, expanded, characterID, versionID, displayName)
		if err != nil {
			return nil, err
		}
		merged = append(merged, part...)
	}
	return merged, nil
}

func trimTailSkeleton(
	nodes []store.TimelineSkeletonNode,
	anchorSeq, birthYear, anchorYear, target int,
) []store.TimelineSkeletonNode {
	if target <= 0 || len(nodes) == 0 {
		return nil
	}
	sorted := append([]store.TimelineSkeletonNode(nil), nodes...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Year != sorted[j].Year {
			return sorted[i].Year < sorted[j].Year
		}
		return sorted[i].Sequence < sorted[j].Sequence
	})
	var filtered []store.TimelineSkeletonNode
	for _, n := range sorted {
		if anchorYear > 0 && n.Year <= anchorYear {
			continue
		}
		filtered = append(filtered, n)
	}
	if len(filtered) > target {
		filtered = filtered[:target]
	}
	out := make([]store.TimelineSkeletonNode, len(filtered))
	for i := range filtered {
		out[i] = filtered[i]
		out[i].Sequence = anchorSeq + 1 + i
		if birthYear > 0 && out[i].Year >= birthYear {
			out[i].Age = out[i].Year - birthYear
		}
	}
	return out
}

func (s *CharacterService) generateRegenerateTail(
	ctx context.Context,
	apiModel, system, user string,
	characterID, versionID, displayName string,
	anchor *model.LifeNode,
	birthYear, batchTarget int,
) ([]model.LifeNode, error) {
	raw, err := s.callRegenerateTailAI(ctx, apiModel, system, user, false)
	if err != nil {
		return nil, err
	}
	batch, err := parseRegenerateTailNodes(raw, characterID, versionID, displayName, anchor.Sequence, birthYear, batchTarget)
	if err != nil {
		raw, err = s.callRegenerateTailAI(ctx, apiModel, system, user, true)
		if err != nil {
			return nil, err
		}
		batch, err = parseRegenerateTailNodes(raw, characterID, versionID, displayName, anchor.Sequence, birthYear, batchTarget)
		if err != nil {
			return nil, err
		}
	}
	return batch, nil
}

func (s *CharacterService) callRegenerateTailAI(ctx context.Context, apiModel, system, user string, compactRetry bool) (string, error) {
	if compactRetry {
		user += "\n【重要】请缩短每节点篇幅，务必输出完整闭合 JSON，nodes 数组不可截断。"
	}
	raw, err := s.ai.ChatJSONModelLong(ctx, apiModel, system, user)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(raw) == "" {
		return "", fmt.Errorf("AI 返回空 JSON")
	}
	return raw, nil
}

func parseRegenerateTailNodes(
	raw, characterID, versionID, displayName string,
	anchorSeq, birthYear, batchTarget int,
) ([]model.LifeNode, error) {
	nodes, err := store.ParseTimelineNodes(raw, characterID, versionID, displayName)
	if err != nil {
		return nil, fmt.Errorf("解析 AI 响应失败（输出可能不完整，请减少节点数后重试）: %w", err)
	}
	return store.TrimSubsequentTail(nodes, anchorSeq, birthYear, batchTarget), nil
}
