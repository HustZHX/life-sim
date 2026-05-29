package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"life-sim/backend/model"
	"life-sim/backend/store"
)

func gameLifeStageHint(age int) string {
	switch {
	case age <= 0:
		return "年龄未知：选项须贴合节点 events 所描述的生活阶段，勿赋予超出叙述的权力与职责。"
	case age <= 5:
		return fmt.Sprintf("婴幼儿（%d 岁）：选项须为人生方向级（启蒙去向、随亲迁居、过继改宗等），禁止日常琐事；禁止从政从戎。", age)
	case age <= 11:
		return fmt.Sprintf("童年（%d 岁）：选项须为人生大事（启蒙去向、定亲意向、随家迁徙、承嗣过继等），禁止玩耍闲聊类琐事；禁止率军、任职高官。", age)
	case age <= 17:
		return fmt.Sprintf("少年（%d 岁）：选项须为人生分岔（求学拜师、学艺从戎、婚配意向、出奔投靠等），禁止琐碎日常；禁止统帅大军、封侯拜相。", age)
	case age <= 39:
		return fmt.Sprintf("青年/壮年（%d 岁）：选项须为人生大事（出仕、从军、婚配、迁徙、创业、结党、守丧等），禁止无关细枝末节。", age)
	case age <= 59:
		return fmt.Sprintf("中年（%d 岁）：选项须为人生大事（仕途转折、归隐、继承、改业、联盟、避祸等），勿突然越级获权，禁止琐碎选项。", age)
	default:
		return fmt.Sprintf("老年（%d 岁）：选项须为人生大事（传嗣、著述、归隐、教导、致仕等），避免与体魄不符的征战；禁止日常琐事。", age)
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

func marshalGameChoiceHistory(nodes []model.LifeNode, rows []struct {
	NodeID       string
	NodeSequence int
	Record       model.GameNodeChoiceRecord
}, beforeSeq int) string {
	yearBySeq := make(map[int]int, len(nodes))
	for _, n := range nodes {
		yearBySeq[n.Sequence] = n.Year
	}
	type lite struct {
		NodeSequence int    `json:"node_sequence"`
		Year         int    `json:"year,omitempty"`
		Choice       string `json:"choice"`
	}
	out := make([]lite, 0)
	for _, row := range rows {
		if row.NodeSequence >= beforeSeq {
			continue
		}
		text := resolveGameChoiceDisplayText(&row.Record)
		if text == "" {
			continue
		}
		out = append(out, lite{
			NodeSequence: row.NodeSequence,
			Year:         yearBySeq[row.NodeSequence],
			Choice:       text,
		})
	}
	if len(out) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(out)
	return string(b)
}

func buildGameNodeChoicesUser(
	ch *model.Character,
	profile *model.Profile,
	node *model.LifeNode,
	journeyJSON, worldLineJSON, choiceHistoryJSON string,
) string {
	age := resolveNodeAge(profile, node)
	personalityBlock := fmt.Sprintf(
		"性格初态：%s\n信念/座右铭：%s\n经历摘要：%s\n社会阶层：%s\n职业：%s",
		strings.TrimSpace(profile.PersonalityInitial),
		strings.TrimSpace(profile.BeliefsMotto),
		store.TruncateRunes(strings.TrimSpace(profile.Experiences), 400),
		strings.TrimSpace(profile.SocialClass),
		strings.TrimSpace(profile.Occupation),
	)
	return fmt.Sprintf(
		`【模式】%s
【人物】%s
【角色状态】%s

【处境约束（必须遵守）】
%s

【人物性格与处境（选项须与此一致）】
%s

【世界线——天下大势与时代背景（选项应呼应，勿凭空捏造素材外重大事件）】
%s

【已历人生节点——主角亲身经历（选项须承接，勿与已发生事实矛盾）】
%s

【既往人生抉择——玩家已选分岔（勿重复同一方向，可推进或合理转折）】
%s

人物档案（完整）：
%s

游戏配置（所选时期/开局阶段）：
%s

【当前节点——请为此刻生成 3～5 个人生大事级抉择】
%s`,
		gameModeLabel(ch.Mode),
		profile.DisplayName,
		ch.Status,
		gameLifeStageHint(age),
		personalityBlock,
		worldLineJSON,
		journeyJSON,
		choiceHistoryJSON,
		store.ProfileJSONTimeline(profile),
		store.MarshalGameConfigForPrompt(ch.GameConfig),
		store.MarshalGameNodeForChoices(node),
	)
}

func buildGameNodeChoicesFollowUp(node *model.LifeNode, regenerate bool) string {
	if regenerate {
		return fmt.Sprintf(
			"请重新生成 3～5 个人生大事级抉择（须与上文世界线、已历经历、性格与既往抉择一致，勿重复已走过的分岔）。当前节点 sequence=%d，year=%d，《%s》。",
			node.Sequence, node.Year, node.Title,
		)
	}
	return fmt.Sprintf(
		"请为下列「当前节点」生成 3～5 个人生大事级抉择（须承接世界线、已历经历、性格与既往抉择）。当前节点 sequence=%d，year=%d，《%s》。",
		node.Sequence, node.Year, node.Title,
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
		"人物抉择：%s\n已有世界线最后事件年份：%d\n须在本时段（年份 > %d，至 next_node.year）补充客观时代大事/历史背景；即使抉择不改变天下大势亦不可留空。\n\n【模式】%s\n【人物】%s\n\n【处境约束（推演下一节点时必须遵守）】\n%s\n\n世界线：\n%s\n\n人物档案：\n%s\n\n游戏配置：\n%s\n\n锚点节点：\n%s",
		choiceLabel,
		lastWLYear,
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
