package store

import (
	"encoding/json"
	"fmt"

	"life-sim/backend/model"

	"github.com/google/uuid"
)

const expandBatchSize = 5

func ExpandBatchSize() int { return expandBatchSize }

// TimelineSkeletonNode 两阶段生成：史实/人生骨架。
type TimelineSkeletonNode struct {
	Sequence         int      `json:"sequence"`
	Year             int      `json:"year"`
	Age              int      `json:"age"`
	Title            string   `json:"title"`
	HistoricalAnchor string   `json:"historical_anchor"`
	EmotionalTone    string   `json:"emotional_tone"`
	KeyFigures       []string `json:"key_figures"`
	KeyPlaces        []string `json:"key_places"`
}

// TimelineExpandNode 两阶段生成：叙事扩写结果。
type TimelineExpandNode struct {
	Sequence            int                 `json:"sequence"`
	Events              string              `json:"events"`
	Thoughts            string              `json:"thoughts"`
	PersonalitySnapshot string              `json:"personality_snapshot"`
	TraitChanges json.RawMessage `json:"trait_changes"`
	Entities     json.RawMessage `json:"entities"`
	Scene        json.RawMessage `json:"scene"`
}

func ParseTimelineSkeleton(raw string) ([]TimelineSkeletonNode, error) {
	var resp struct {
		Nodes []TimelineSkeletonNode `json:"nodes"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, err
	}
	if len(resp.Nodes) == 0 {
		return nil, fmt.Errorf("骨架节点为空")
	}
	return resp.Nodes, nil
}

func ParseTimelineExpandBatch(raw string, _ string) (map[int]TimelineExpandNode, error) {
	var resp struct {
		Nodes []TimelineExpandNode `json:"nodes"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, err
	}
	out := make(map[int]TimelineExpandNode, len(resp.Nodes))
	for _, n := range resp.Nodes {
		out[n.Sequence] = n
	}
	return out, nil
}

func MergeSkeletonAndExpand(
	skeleton []TimelineSkeletonNode,
	expanded map[int]TimelineExpandNode,
	characterID, versionID, protagonistName string,
) ([]model.LifeNode, error) {
	nodes := make([]model.LifeNode, 0, len(skeleton))
	for _, sk := range skeleton {
		ex, ok := expanded[sk.Sequence]
		if !ok {
			return nil, fmt.Errorf("缺少 sequence=%d 的扩写结果", sk.Sequence)
		}
		traits, err := ParseTraitChangesJSON(ex.TraitChanges)
		if err != nil {
			return nil, err
		}
		entities, err := ParseEntitiesJSON(ex.Entities, protagonistName)
		if err != nil {
			return nil, err
		}
		scene, err := ParseSceneJSON(ex.Scene)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, model.LifeNode{
			ID:                  uuid.New().String(),
			CharacterID:         characterID,
			VersionID:           versionID,
			Sequence:            sk.Sequence,
			Year:                sk.Year,
			Age:                 sk.Age,
			Title:               sk.Title,
			Events:              ex.Events,
			Thoughts:            ex.Thoughts,
			PersonalitySnapshot: ex.PersonalitySnapshot,
			TraitChanges:        traits,
			Entities:            entities,
			Scene:               scene,
		})
	}
	return nodes, nil
}

func MarshalSkeletonBatch(batch []TimelineSkeletonNode) string {
	b, _ := json.Marshal(batch)
	return string(b)
}

func MarshalSkeletonContextPrior(nodes []model.LifeNode, count int) string {
	return LastNodesLite(nodes, count)
}

// BatchSkeletonSlices 将骨架按固定大小分批。
func BatchSkeletonSlices(all []TimelineSkeletonNode, size int) [][]TimelineSkeletonNode {
	if size <= 0 {
		size = expandBatchSize
	}
	var batches [][]TimelineSkeletonNode
	for i := 0; i < len(all); i += size {
		end := i + size
		if end > len(all) {
			end = len(all)
		}
		batches = append(batches, all[i:end])
	}
	return batches
}
