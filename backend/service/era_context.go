package service

import (
	"context"
	"fmt"
	"strings"

	"life-sim/backend/model"
	"life-sim/backend/store"
)

func eraContextPromptName(mode string) string {
	if mode == model.ModeFamous {
		return "generate_era_context_famous.txt"
	}
	return "generate_era_context_random.txt"
}

func (s *CharacterService) generateEraContext(
	ctx context.Context,
	characterMode string,
	profileJSON string,
	apiModel string,
	cfg timelineJobConfig,
) (string, error) {
	system, err := s.ai.LoadPrompt(eraContextPromptName(characterMode))
	if err != nil {
		return "", err
	}
	user := fmt.Sprintf(
		"请整理 start_year=%d 至 end_year=%d 区间内，影响该人物生活的大事记。\n人物档案：\n%s",
		cfg.StartYear, cfg.EndYear, profileJSON,
	)
	if cfg.Instructions != "" {
		user += fmt.Sprintf("\n\n用户特殊要求/备注（须兼容）：\n%s", cfg.Instructions)
	}
	raw, err := s.ai.ChatJSONModel(ctx, apiModel, system, user)
	if err != nil {
		return "", err
	}
	return store.ParseEraContextResult(raw)
}

func resolveEraContext(
	ctx context.Context,
	s *CharacterService,
	characterMode string,
	profileJSON string,
	apiModel string,
	cfg timelineJobConfig,
	userOverride string,
) (string, error) {
	if text := strings.TrimSpace(userOverride); text != "" {
		if strings.Contains(text, "【时代背景与大事记】") {
			return text, nil
		}
		return "【时代背景与大事记】\n" + text, nil
	}
	return s.generateEraContext(ctx, characterMode, profileJSON, apiModel, cfg)
}
