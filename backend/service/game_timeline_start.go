package service

import (
	"context"
	"encoding/json"
	"fmt"

	"life-sim/backend/model"
	"life-sim/backend/store"
)

const gameTimelineStageMaxNodes = 25

func gameTimelineStageTargetCount(birthYear, targetYear int) int {
	if birthYear <= 0 || targetYear <= birthYear {
		return 1
	}
	span := targetYear - birthYear
	n := span/6 + 2
	if n < 3 {
		n = 3
	}
	return store.ClampTargetNodeCount(n)
}

func buildGameTimelineToStageUser(
	batchTarget, birthYear, targetYear int,
	profile *model.Profile,
	locked []model.LifeNode,
	anchor *model.LifeNode,
	gameCfgJSON string,
) string {
	lockedJSON := "[]"
	if len(locked) > 0 {
		b, _ := json.Marshal(locked)
		lockedJSON = string(b)
	}
	head := fmt.Sprintf(
		"【人生游戏开局】从出生后推演至所选阶段。目标阶段年份约 %d（最后一节点 year 须落在 %d±3 年内）。\n",
		targetYear, targetYear,
	)
	user := fmt.Sprintf(
		"%s恰好生成 M=%d 个全新节点；anchor_sequence=%d；anchor_year=%d；birth_year=%d\n"+
			"M>1 时单节点宜简练，务必输出完整合法 JSON。\n"+
			"人物档案：\n%s\n"+
			"已锁定节点（含锚点，须衔接）：\n%s\n"+
			"锚点节点：sequence=%d year=%d age=%d title=%s\nevents=%s\n"+
			"游戏配置：%s",
		head, batchTarget, anchor.Sequence, anchor.Year, birthYear,
		store.ProfileJSONTimeline(profile), lockedJSON,
		anchor.Sequence, anchor.Year, anchor.Age, anchor.Title, anchor.Events,
		gameCfgJSON,
	)
	return user
}

func (g *GameService) generateGameTimelineToStage(
	ctx context.Context,
	apiModel string,
	ch *model.Character,
	profile *model.Profile,
	characterID, versionID string,
	birthYear, targetYear int,
	gameCfgJSON string,
) ([]model.LifeNode, error) {
	system, err := g.ai.LoadPrompt(gameTimelineToStagePromptName(ch.Mode))
	if err != nil {
		return nil, err
	}
	maxNodes := gameTimelineStageTargetCount(birthYear, targetYear)
	anchor := &model.LifeNode{
		Sequence: 0,
		Year:     birthYear,
		Title:    "出生",
		Events:   "出生",
	}
	var allNodes []model.LifeNode
	for len(allNodes) < maxNodes {
		remaining := maxNodes - len(allNodes)
		batchTarget := remaining
		if batchTarget > regenerateTailBatchSize {
			batchTarget = regenerateTailBatchSize
		}
		user := buildGameTimelineToStageUser(
			batchTarget, birthYear, targetYear, profile, allNodes, anchor, gameCfgJSON,
		)
		batch, err := g.char.generateRegenerateTail(
			ctx, apiModel, system, user,
			characterID, versionID, profile.DisplayName,
			anchor, birthYear, batchTarget,
		)
		if err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			break
		}
		allNodes = append(allNodes, batch...)
		anchor = &allNodes[len(allNodes)-1]
		if anchor.Year >= targetYear-3 {
			break
		}
		if len(batch) < batchTarget {
			break
		}
	}
	if len(allNodes) == 0 {
		return nil, fmt.Errorf("未生成开局节点")
	}
	return allNodes, nil
}

func (g *GameService) generateGameFirstNode(
	ctx context.Context,
	apiModel string,
	ch *model.Character,
	profile *model.Profile,
	characterID, versionID string,
	startYear int,
	gameCfgJSON string,
) ([]model.LifeNode, error) {
	promptName := "game_first_node_" + gamePromptSuffix(ch.Mode)
	system, err := g.ai.LoadPrompt(promptName)
	if err != nil {
		return nil, err
	}
	user := fmt.Sprintf("人物档案：\n%s\n起点年份：%d\n游戏配置：%s",
		store.ProfileJSONTimeline(profile), startYear, gameCfgJSON)
	raw, err := g.ai.ChatJSONModel(ctx, apiModel, system, user)
	if err != nil {
		return nil, err
	}
	return parseRegenerateTailNodes(raw, characterID, versionID, profile.DisplayName, 0, profile.BirthYear, 1)
}
