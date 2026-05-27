package store

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"life-sim/backend/model"
)

// NarrativeChangeOperation 单条节点变更操作
type NarrativeChangeOperation struct {
	Op                  string              `json:"op"`
	Sequence            int                 `json:"sequence,omitempty"`
	AfterSequence       int                 `json:"after_sequence,omitempty"`
	Year                int                 `json:"year,omitempty"`
	Age                 int                 `json:"age,omitempty"`
	Title               string              `json:"title,omitempty"`
	Events              string              `json:"events,omitempty"`
	Thoughts            string              `json:"thoughts,omitempty"`
	PersonalitySnapshot string              `json:"personality_snapshot,omitempty"`
	TraitChanges        json.RawMessage     `json:"trait_changes,omitempty"`
	Entities            *model.NodeEntities `json:"entities,omitempty"`
	Scene               *model.NodeScene    `json:"scene,omitempty"`
}

// NarrativeChangePlan AI 规划的叙述变更
type NarrativeChangePlan struct {
	ChangeSummary  string                     `json:"change_summary"`
	AnchorSequence int                        `json:"anchor_sequence"`
	Operations     []NarrativeChangeOperation `json:"operations"`
	DeathYear      int                        `json:"death_year,omitempty"`
	DeathCause     string                     `json:"death_cause,omitempty"`
	DeathReasoning string                     `json:"death_reasoning,omitempty"`
}

func ParseNarrativeChangePlan(raw string) (*NarrativeChangePlan, error) {
	var plan NarrativeChangePlan
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		return nil, err
	}
	if plan.ChangeSummary == "" {
		return nil, fmt.Errorf("缺少 change_summary")
	}
	if plan.AnchorSequence <= 0 {
		return nil, fmt.Errorf("anchor_sequence 无效")
	}
	for i, op := range plan.Operations {
		switch op.Op {
		case "modify":
			if op.Sequence <= 0 {
				return nil, fmt.Errorf("operations[%d]: modify 缺少 sequence", i)
			}
		case "insert_after":
			if op.AfterSequence <= 0 {
				return nil, fmt.Errorf("operations[%d]: insert_after 缺少 after_sequence", i)
			}
			if op.Title == "" || op.Events == "" {
				return nil, fmt.Errorf("operations[%d]: insert_after 缺少 title/events", i)
			}
		default:
			return nil, fmt.Errorf("operations[%d]: 未知 op %q", i, op.Op)
		}
	}
	return &plan, nil
}

// ApplyNarrativeChangePlan 按规划修改/插入节点，截断锚点之后并重新编号
func ApplyNarrativeChangePlan(
	orig []model.LifeNode,
	plan *NarrativeChangePlan,
	characterID, protagonistName string,
) ([]model.LifeNode, error) {
	nodes := append([]model.LifeNode{}, orig...)
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Sequence < nodes[j].Sequence })

	for i, op := range plan.Operations {
		switch op.Op {
		case "modify":
			if err := applyModifyOp(&nodes, op, protagonistName); err != nil {
				return nil, fmt.Errorf("operations[%d]: %w", i, err)
			}
		case "insert_after":
			var err error
			nodes, err = applyInsertAfterOp(nodes, op, characterID, protagonistName)
			if err != nil {
				return nil, fmt.Errorf("operations[%d]: %w", i, err)
			}
		}
	}

	locked := make([]model.LifeNode, 0)
	for _, n := range nodes {
		if n.Sequence <= plan.AnchorSequence {
			locked = append(locked, n)
		}
	}
	if len(locked) == 0 {
		return nil, fmt.Errorf("锚点 sequence=%d 无对应节点", plan.AnchorSequence)
	}
	sort.Slice(locked, func(i, j int) bool { return locked[i].Sequence < locked[j].Sequence })
	resequenceNodes(locked)
	return locked, nil
}

func applyModifyOp(nodes *[]model.LifeNode, op NarrativeChangeOperation, protagonistName string) error {
	for i := range *nodes {
		if (*nodes)[i].Sequence != op.Sequence {
			continue
		}
		n := &(*nodes)[i]
		if op.Title != "" {
			n.Title = op.Title
		}
		if op.Events != "" {
			n.Events = op.Events
		}
		if op.Thoughts != "" {
			n.Thoughts = op.Thoughts
		}
		if op.PersonalitySnapshot != "" {
			n.PersonalitySnapshot = op.PersonalitySnapshot
		}
		if op.Year > 0 {
			n.Year = op.Year
		}
		if op.Age > 0 {
			n.Age = op.Age
		}
		if len(op.TraitChanges) > 0 {
			traits, err := ParseTraitChangesJSON(op.TraitChanges)
			if err != nil {
				return err
			}
			n.TraitChanges = traits
		}
		if op.Entities != nil {
			n.Entities = NormalizeNodeEntities(op.Entities, protagonistName)
		}
		if op.Scene != nil {
			n.Scene = NormalizeNodeScene(op.Scene)
		}
		return nil
	}
	return fmt.Errorf("modify 目标 sequence=%d 不存在", op.Sequence)
}

func applyInsertAfterOp(
	nodes []model.LifeNode,
	op NarrativeChangeOperation,
	characterID, protagonistName string,
) ([]model.LifeNode, error) {
	found := false
	for _, n := range nodes {
		if n.Sequence == op.AfterSequence {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("insert_after 锚点 sequence=%d 不存在", op.AfterSequence)
	}

	traits, err := ParseTraitChangesJSON(op.TraitChanges)
	if err != nil {
		return nil, err
	}
	newNode := model.LifeNode{
		ID:                  uuid.New().String(),
		CharacterID:         characterID,
		Sequence:            op.AfterSequence + 1,
		Year:                op.Year,
		Age:                 op.Age,
		Title:               op.Title,
		Events:              op.Events,
		Thoughts:            op.Thoughts,
		PersonalitySnapshot: op.PersonalitySnapshot,
		TraitChanges:        traits,
		Entities:            NormalizeNodeEntities(op.Entities, protagonistName),
		Scene:               NormalizeNodeScene(op.Scene),
	}

	out := make([]model.LifeNode, 0, len(nodes)+1)
	for _, n := range nodes {
		if n.Sequence <= op.AfterSequence {
			out = append(out, n)
		}
	}
	out = append(out, newNode)
	for _, n := range nodes {
		if n.Sequence > op.AfterSequence {
			nn := n
			nn.Sequence = n.Sequence + 1
			out = append(out, nn)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Sequence < out[j].Sequence })
	return out, nil
}

func resequenceNodes(nodes []model.LifeNode) {
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Sequence < nodes[j].Sequence })
	for i := range nodes {
		nodes[i].Sequence = i + 1
	}
}
