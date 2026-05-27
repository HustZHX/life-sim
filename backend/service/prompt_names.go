package service

import "life-sim/backend/model"

func innerCurrentPromptName(mode string) string {
	if mode == model.ModeFamous {
		return "regenerate_inner_current_famous.txt"
	}
	return "regenerate_inner_current_random.txt"
}

func innerSubsequentPromptName(mode string) string {
	if mode == model.ModeFamous {
		return "regenerate_inner_subsequent_famous.txt"
	}
	return "regenerate_inner_subsequent_random.txt"
}

func fullTailPromptName(mode string) string {
	if mode == model.ModeFamous {
		return "regenerate_full_tail_famous.txt"
	}
	return "regenerate_full_tail_random.txt"
}
