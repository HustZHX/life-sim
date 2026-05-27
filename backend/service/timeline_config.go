package service

import (
	"context"
	"fmt"

	"life-sim/backend/ai"
	"life-sim/backend/model"
	"life-sim/backend/store"
)

func (s *CharacterService) RecommendTimelineConfigs(ctx context.Context, characterID, modelID string) ([]model.TimelineRecommendation, error) {
	ch, err := s.store.GetCharacter(characterID)
	if err != nil {
		return nil, err
	}
	if ch.Status != model.StatusProfileReady && ch.Status != model.StatusTimelineReady {
		return nil, fmt.Errorf("请先生成人物档案")
	}
	profile, err := s.store.GetProfile(characterID)
	if err != nil {
		return nil, err
	}
	modelID = ai.NormalizeModelID(modelID)
	apiModel, err := s.resolveAPIModel(modelID)
	if err != nil {
		return nil, err
	}

	system, err := s.ai.LoadPrompt("recommend_timeline_configs.txt")
	if err != nil {
		return nil, err
	}
	profileJSON := store.ProfileJSONTimeline(profile)
	user := fmt.Sprintf("人物档案：\n%s", profileJSON)

	raw, err := s.ai.ChatJSONModel(ctx, apiModel, system, user)
	if err != nil {
		return nil, err
	}
	recs, err := store.ParseTimelineRecommendations(raw)
	if err != nil {
		return nil, err
	}
	for i := range recs {
		rangeCfg, err := store.ResolveTimelineRange(store.TimelineRangeConfig{
			TargetNodeCount: recs[i].TargetNodeCount,
			StartYear:       recs[i].StartYear,
			EndYear:         recs[i].EndYear,
		}, profile)
		if err != nil {
			continue
		}
		recs[i].TargetNodeCount = rangeCfg.TargetNodeCount
		recs[i].StepYears = rangeCfg.StepYears
		recs[i].StartYear, recs[i].EndYear = rangeCfg.StartYear, rangeCfg.EndYear
	}

	s.aiCache.putBundle(characterID, profile)
	s.aiCache.putSession(characterID, ai.NewSession(characterID, store.ProfileHash(profile), system, user, raw))

	return recs, nil
}

func (s *CharacterService) recalculateLifespan(ctx context.Context, profile *model.Profile, locked []model.LifeNode, edited *model.LifeNode, modelID string) (*model.LifespanRecalcResult, error) {
	system, err := s.ai.LoadPrompt("recalculate_lifespan.txt")
	if err != nil {
		return nil, err
	}
	user := fmt.Sprintf(
		"人物档案（含初始性格、信念、经历摘要等原形）：\n%s\n"+
			"已发生人生节点（含锚点）：\n%s\n"+
			"锚点节点：sequence=%d year=%d title=%s\nevents=%s",
		store.ProfileJSONInner(profile), store.MarshalNodesLocked(locked),
		edited.Sequence, edited.Year, edited.Title, edited.Events,
	)

	apiModel, err := s.resolveAPIModel(modelID)
	if err != nil {
		return nil, err
	}
	raw, err := s.ai.ChatJSONModel(ctx, apiModel, system, user)
	if err != nil {
		return nil, err
	}
	return store.ParseLifespanRecalc(raw)
}

func (s *CharacterService) PreviewLifespan(ctx context.Context, characterID, nodeID string, req model.LifespanPreviewRequest) (*model.LifespanPreviewResponse, error) {
	node, err := s.store.GetNode(nodeID)
	if err != nil {
		return nil, err
	}
	if node.CharacterID != characterID {
		return nil, fmt.Errorf("节点不属于该角色")
	}
	req.Model = ai.NormalizeModelID(req.Model)
	if _, err := s.resolveAPIModel(req.Model); err != nil {
		return nil, err
	}

	node.Title = req.Title
	node.Events = req.Events
	node.Thoughts = req.Thoughts
	node.PersonalitySnapshot = req.PersonalitySnapshot

	timeline, err := s.store.GetTimelineByVersionID(node.VersionID)
	if err != nil {
		return nil, fmt.Errorf("无法定位节点所属时间轴")
	}
	oldNodes, err := s.store.GetNodesByVersion(timeline.CurrentVersionID)
	if err != nil {
		return nil, err
	}
	locked := store.FilterNodesFromSequence(oldNodes, node.Sequence)
	for i := range locked {
		if locked[i].ID == node.ID {
			locked[i] = *node
			break
		}
	}

	profile, err := s.store.GetProfile(characterID)
	if err != nil {
		return nil, err
	}
	preview, err := s.recalculateLifespan(ctx, profile, locked, node, req.Model)
	if err != nil {
		return nil, err
	}
	contextHint := store.EstimateContextBytes(
		store.ProfileJSONInner(profile),
		store.MarshalNodesLocked(locked),
	)
	return &model.LifespanPreviewResponse{
		CurrentDeathYear: profile.DeathYear,
		AnchorYear:       node.Year,
		Preview:          *preview,
		ContextTokenHint: contextHint,
	}, nil
}
