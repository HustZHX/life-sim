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

func tailSkeletonPromptName(mode string) string {
	if mode == model.ModeFamous {
		return "regenerate_tail_skeleton_famous.txt"
	}
	return "regenerate_tail_skeleton_random.txt"
}

func gameTimelineToStagePromptName(mode string) string {
	if mode == model.ModeFamous {
		return "game_timeline_to_stage_famous.txt"
	}
	return "game_first_node_random.txt"
}

func regenerateEventsPromptName(mode string) string {
	if mode == model.ModeFamous {
		return "regenerate_node_events_famous.txt"
	}
	return "regenerate_node_events_random.txt"
}

func planNarrativeChangePromptName(mode string) string {
	if mode == model.ModeFamous {
		return "plan_narrative_change_famous.txt"
	}
	return "plan_narrative_change_random.txt"
}
