package service

import (
	"context"
	"fmt"
	"strings"

	"life-sim/backend/model"
	"life-sim/backend/store"
)

const regenerateTailBatchSize = 6

func buildRegenerateTailUser(
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
			"M>1 时单节点宜简练（events 80～150 字），务必输出完整合法 JSON。\n"+
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
