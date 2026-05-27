package service

import (
	"context"
	"fmt"

	"life-sim/backend/model"
	"life-sim/backend/store"
)

func (s *CharacterService) generateStandardTimelineNodes(
	ctx context.Context,
	jobID, characterID string,
	profile *model.Profile,
	profileJSON, profileHash, apiModel string,
	cfg timelineJobConfig,
	versionID string,
) ([]model.LifeNode, error) {
	span := cfg.EndYear - cfg.StartYear

	systemPrompt, err := s.ai.LoadPrompt("generate_timeline.txt")
	if err != nil {
		return nil, err
	}
	chunkPrompt, err := s.ai.LoadPrompt("generate_timeline_chunk.txt")
	if err != nil {
		return nil, err
	}

	chunks := splitTimelineRange(cfg.StartYear, cfg.EndYear, cfg.TargetNodeCount)
	var allNodes []model.LifeNode

	for i, chunk := range chunks {
		progress := 10 + (i * 75 / len(chunks))
		setJobStage(s.store, jobID, progress, store.FormatChunkStageText(chunk.ChunkIndex, chunk.TotalChunks, chunk.StartYear, chunk.EndYear))

		var prompt, user string
		switch {
		case chunk.TotalChunks == 1:
			prompt = systemPrompt
			user = buildTimelineUserContent(profileJSON, cfg, span)
		case i == 0:
			prompt = systemPrompt
			user = buildTimelineChunkUser(profileJSON, cfg, chunk, span)
		default:
			prompt = chunkPrompt
			user = buildTimelineContinuationUser(profileJSON, cfg, chunk, store.LastNodesLite(allNodes, 2))
		}

		done := make(chan struct{})
		go tickJobProgress(s.store, jobID, progress+2, progress+70/len(chunks), done)

		var raw string
		if i == 0 && chunk.TotalChunks == 1 {
			if sess := s.aiCache.getSession(characterID, profileHash); sess != nil {
				user = buildTimelineUserWithoutProfile(cfg, span)
				raw, err = s.ai.ChatJSONSession(ctx, apiModel, sess, user)
			} else {
				raw, err = s.ai.ChatJSONModel(ctx, apiModel, prompt, user)
			}
		} else {
			raw, err = s.ai.ChatJSONModel(ctx, apiModel, prompt, user)
		}
		close(done)

		if err != nil {
			return nil, fmt.Errorf("第 %d/%d 段生成失败: %w", i+1, len(chunks), err)
		}

		chunkNodes, err := store.ParseTimelineNodes(raw, characterID, versionID, profile.DisplayName)
		if err != nil {
			return nil, fmt.Errorf("第 %d/%d 段解析失败: %w", i+1, len(chunks), err)
		}
		store.RenumberTimelineNodes(chunkNodes, chunk.SequenceStart)
		allNodes = append(allNodes, chunkNodes...)
	}
	return allNodes, nil
}
