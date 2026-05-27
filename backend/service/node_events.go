package service

import (
	"context"
	"fmt"

	"life-sim/backend/ai"
	"life-sim/backend/model"
	"life-sim/backend/store"
)

func (s *CharacterService) RegenerateNodeEvents(ctx context.Context, characterID, nodeID string, req model.RegenerateNodeEventsRequest) (*model.RegenerateNodeEventsResponse, error) {
	node, err := s.store.GetNode(nodeID)
	if err != nil {
		return nil, err
	}
	if node.CharacterID != characterID {
		return nil, fmt.Errorf("节点不属于该角色")
	}
	title := req.Title
	if title == "" {
		title = node.Title
	}
	if title == "" {
		return nil, fmt.Errorf("请提供节点标题")
	}

	ch, err := s.store.GetCharacter(characterID)
	if err != nil {
		return nil, err
	}
	profile, err := s.store.GetProfile(characterID)
	if err != nil {
		return nil, err
	}

	timeline, err := s.store.GetTimelineByVersionID(node.VersionID)
	if err != nil {
		return nil, fmt.Errorf("无法定位节点所属时间轴")
	}
	allNodes, err := s.store.GetNodesByVersion(timeline.CurrentVersionID)
	if err != nil {
		return nil, err
	}

	modelID := ai.NormalizeModelID(req.Model)
	apiModel, err := s.resolveAPIModel(modelID)
	if err != nil {
		return nil, err
	}

	system, err := s.ai.LoadPrompt(regenerateEventsPromptName(ch.Mode))
	if err != nil {
		return nil, err
	}

	user := fmt.Sprintf(
		"人物档案：\n%s\n"+
			"前置人生节点（sequence<%d）：\n%s\n"+
			"本节点：sequence=%d year=%d age=%d title=%s\n"+
			"请根据标题「%s」重新撰写本节点经历、内心想法与性格快照（一并输出）。",
		store.ProfileJSONInner(profile),
		node.Sequence, store.MarshalNodesLiteBeforeSequence(allNodes, node.Sequence),
		node.Sequence, node.Year, node.Age, title, title,
	)

	raw, err := s.ai.ChatJSONModel(ctx, apiModel, system, user)
	if err != nil {
		return nil, err
	}
	parsed, err := store.ParseRegenerateNodeFromTitle(raw)
	if err != nil {
		return nil, err
	}
	return &model.RegenerateNodeEventsResponse{
		Events:              parsed.Events,
		Thoughts:            parsed.Thoughts,
		PersonalitySnapshot: parsed.PersonalitySnapshot,
		TraitChanges:        parsed.TraitChanges,
	}, nil
}
