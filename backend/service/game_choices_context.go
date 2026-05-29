package service

import (
	"encoding/json"
	"fmt"

	"life-sim/backend/model"
	"life-sim/backend/store"
)

func gameLifeStageHint(age int) string {
	switch {
	case age <= 0:
		return "年龄未知：选项须贴合节点 events 所描述的生活阶段，勿赋予超出叙述的权力与职责。"
	case age <= 5:
		return fmt.Sprintf("婴幼儿（%d 岁）：仅限家庭照料、启蒙玩耍、随长辈迁居等；禁止从政从戎、独立重大决策。", age)
	case age <= 11:
		return fmt.Sprintf("童年（%d 岁）：学塾启蒙、家务帮衬、邻里孩童交往、随家迁徙；禁止率军、任职高官、主导国策或独自远行征战。", age)
	case age <= 17:
		return fmt.Sprintf("少年（%d 岁）：可求学拜师、学艺帮工、局部冒险；禁止统帅大军、封侯拜相、称帝登基或承担一方牧守。", age)
	case age <= 39:
		return fmt.Sprintf("青年/壮年（%d 岁）：可按社会角色承担相应职责；仍须符合阶级、时代与当前处境。", age)
	case age <= 59:
		return fmt.Sprintf("中年（%d 岁）：选项须符合既有身份与时代；勿突然越级获得与履历不符的权力。", age)
	default:
		return fmt.Sprintf("老年（%d 岁）：宜侧重传承、休养、著述、教导后辈；避免与体魄、身份不符的激烈征战或新政。", age)
	}
}

func resolveNodeAge(profile *model.Profile, node *model.LifeNode) int {
	if node.Age > 0 {
		return node.Age
	}
	if profile != nil && profile.BirthYear > 0 && node.Year > 0 {
		return node.Year - profile.BirthYear
	}
	return 0
}

func gameModeLabel(mode string) string {
	if mode == model.ModeFamous {
		return "历史名人"
	}
	return "虚构人生"
}

func buildGameNodeChoicesUser(ch *model.Character, profile *model.Profile, node *model.LifeNode) string {
	age := resolveNodeAge(profile, node)
	return fmt.Sprintf(
		"【模式】%s\n【人物】%s\n【角色状态】%s\n\n【处境约束（必须遵守）】\n%s\n\n人物档案：\n%s\n\n游戏配置（所选时期/开局阶段）：\n%s\n\n当前节点：\n%s",
		gameModeLabel(ch.Mode),
		profile.DisplayName,
		ch.Status,
		gameLifeStageHint(age),
		store.ProfileJSONTimeline(profile),
		store.MarshalGameConfigForPrompt(ch.GameConfig),
		store.MarshalGameNodeForChoices(node),
	)
}

func buildGameApplyChoiceUser(
	choiceLabel string,
	lastWLYear int,
	wlJSON string,
	ch *model.Character,
	profile *model.Profile,
	anchor *model.LifeNode,
) string {
	age := resolveNodeAge(profile, anchor)
	anchorJSON, _ := json.Marshal(map[string]any{
		"sequence": anchor.Sequence,
		"year":     anchor.Year,
		"age":      age,
		"title":    anchor.Title,
		"events":   store.TruncateRunes(anchor.Events, 480),
	})
	return fmt.Sprintf(
		"人物抉择：%s\n已有世界线最后事件年份：%d\n\n【模式】%s\n【人物】%s\n\n【处境约束（推演下一节点时必须遵守）】\n%s\n\n世界线：\n%s\n\n人物档案：\n%s\n\n游戏配置：\n%s\n\n锚点节点：\n%s",
		choiceLabel,
		lastWLYear,
		gameModeLabel(ch.Mode),
		profile.DisplayName,
		gameLifeStageHint(age),
		wlJSON,
		store.ProfileJSONTimeline(profile),
		store.MarshalGameConfigForPrompt(ch.GameConfig),
		string(anchorJSON),
	)
}
