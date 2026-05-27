package service

import (
	"context"
	"encoding/json"
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
	user := fmt.Sprintf("人物档案：\n%s\nbirth_year=%d, death_year=%d",
		store.MustProfileJSON(profile), profile.BirthYear, profile.DeathYear)

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
	return recs, nil
}

func (s *CharacterService) recalculateLifespan(ctx context.Context, profile *model.Profile, locked []model.LifeNode, edited *model.LifeNode, modelID string) (*model.LifespanRecalcResult, error) {
	system, err := s.ai.LoadPrompt("recalculate_lifespan.txt")
	if err != nil {
		return nil, err
	}
	lockedJSON, _ := json.Marshal(locked)
	user := fmt.Sprintf(
		"人物档案（含初始性格、信念、经历摘要等原形）：\n%s\n"+
			"已发生人生节点（含锚点）：\n%s\n"+
			"锚点节点：sequence=%d year=%d title=%s\nevents=%s\n原 death_year=%d",
		store.MustProfileJSON(profile), string(lockedJSON),
		edited.Sequence, edited.Year, edited.Title, edited.Events, profile.DeathYear,
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
