package service

import (
	"context"
	"fmt"
	"strings"

	"life-sim/backend/model"
	"life-sim/backend/store"
)

const expandJSONCompactSuffix = "\n【重要】请缩短每节点 events/thoughts 篇幅，务必输出完整闭合 JSON，nodes 数组须包含本批全部 sequence，不可截断。"

func buildTimelineExpandUser(
	profileJSON string,
	batch []store.TimelineSkeletonNode,
	contextPrior []model.LifeNode,
	extra string,
) string {
	user := fmt.Sprintf(
		"人物档案：\n%s\n\n骨架节点（须逐条扩写，sequence 不可变）：\n%s",
		profileJSON, store.MarshalSkeletonBatch(batch),
	)
	if len(contextPrior) > 0 {
		user += "\n\n前置已扩写节点（性格衔接参考）：\n" + store.MarshalSkeletonContextPrior(contextPrior, 2)
	}
	if extra != "" {
		user += extra
	}
	return user
}

func expandSkeletonBatchComplete(batch []store.TimelineSkeletonNode, expanded map[int]store.TimelineExpandNode) bool {
	for _, sk := range batch {
		if _, ok := expanded[sk.Sequence]; !ok {
			return false
		}
	}
	return len(expanded) > 0
}

func (s *CharacterService) tryExpandSkeletonBatch(
	ctx context.Context,
	apiModel, expandPrompt, displayName, user string,
	batch []store.TimelineSkeletonNode,
) (map[int]store.TimelineExpandNode, error) {
	var lastErr error
	for _, compact := range []bool{false, true} {
		u := user
		if compact {
			u += expandJSONCompactSuffix
		}
		raw, err := s.ai.ChatJSONModelLong(ctx, apiModel, expandPrompt, u)
		if err != nil {
			lastErr = err
			continue
		}
		if strings.TrimSpace(raw) == "" {
			lastErr = fmt.Errorf("AI 返回空 JSON")
			continue
		}
		expanded, err := store.ParseTimelineExpandBatch(raw, displayName)
		if err != nil {
			lastErr = err
			continue
		}
		if !expandSkeletonBatchComplete(batch, expanded) {
			lastErr = fmt.Errorf("扩写结果缺少本批部分 sequence")
			continue
		}
		return expanded, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("扩写失败")
	}
	return nil, lastErr
}

// expandSkeletonBatchWithContext 叙事扩写一批骨架；解析失败时紧凑重试，仍失败则拆半批递归。
func (s *CharacterService) expandSkeletonBatchWithContext(
	ctx context.Context,
	apiModel, expandPrompt, displayName, profileJSON, extra string,
	batch []store.TimelineSkeletonNode,
	contextPrior []model.LifeNode,
) (map[int]store.TimelineExpandNode, error) {
	user := buildTimelineExpandUser(profileJSON, batch, contextPrior, extra)
	expanded, err := s.tryExpandSkeletonBatch(ctx, apiModel, expandPrompt, displayName, user, batch)
	if err == nil {
		return expanded, nil
	}
	if len(batch) <= 1 {
		return nil, err
	}
	mid := len(batch) / 2
	left, errL := s.expandSkeletonBatchWithContext(ctx, apiModel, expandPrompt, displayName, profileJSON, extra, batch[:mid], contextPrior)
	if errL != nil {
		return nil, errL
	}
	right, errR := s.expandSkeletonBatchWithContext(ctx, apiModel, expandPrompt, displayName, profileJSON, extra, batch[mid:], contextPrior)
	if errR != nil {
		return nil, errR
	}
	merged := make(map[int]store.TimelineExpandNode, len(left)+len(right))
	for k, v := range left {
		merged[k] = v
	}
	for k, v := range right {
		merged[k] = v
	}
	return merged, nil
}
