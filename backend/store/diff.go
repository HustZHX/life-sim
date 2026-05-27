package store

import (
	"encoding/json"

	"life-sim/backend/model"
)

// NodesFromSequence 返回 sequence >= fromSeq 的节点（含锚点）。
func NodesFromSequence(nodes []model.LifeNode, fromSeq int) []model.LifeNode {
	out := make([]model.LifeNode, 0)
	for _, n := range nodes {
		if n.Sequence >= fromSeq {
			out = append(out, n)
		}
	}
	return out
}

func traitChangesJSON(n model.LifeNode) string {
	if len(n.TraitChanges) == 0 {
		return "[]"
	}
	b, err := json.Marshal(n.TraitChanges)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// DiffNodeLists 对比两组节点（通常已按锚点对齐），返回非 nil 切片。
func DiffNodeLists(oldNodes, newNodes []model.LifeNode) []model.NodeFieldChange {
	oldBySeq := make(map[int]model.LifeNode, len(oldNodes))
	for _, n := range oldNodes {
		oldBySeq[n.Sequence] = n
	}
	newBySeq := make(map[int]model.LifeNode, len(newNodes))
	for _, n := range newNodes {
		newBySeq[n.Sequence] = n
	}

	changes := make([]model.NodeFieldChange, 0)
	fields := []struct {
		name string
		get  func(model.LifeNode) string
	}{
		{"title", func(n model.LifeNode) string { return n.Title }},
		{"events", func(n model.LifeNode) string { return n.Events }},
		{"thoughts", func(n model.LifeNode) string { return n.Thoughts }},
		{"personality_snapshot", func(n model.LifeNode) string { return n.PersonalitySnapshot }},
		{"trait_changes", traitChangesJSON},
	}

	appendChange := func(nn model.LifeNode, field, before, after string) {
		changes = append(changes, model.NodeFieldChange{
			NodeID: nn.ID, Sequence: nn.Sequence, Year: nn.Year,
			Field: field, Before: before, After: after,
		})
	}

	for _, nn := range newNodes {
		on, ok := oldBySeq[nn.Sequence]
		if !ok {
			for _, f := range fields {
				appendChange(nn, f.name, "", f.get(nn))
			}
			continue
		}
		for _, f := range fields {
			b, a := f.get(on), f.get(nn)
			if b != a {
				appendChange(nn, f.name, b, a)
			}
		}
	}

	for _, on := range oldNodes {
		if _, ok := newBySeq[on.Sequence]; ok {
			continue
		}
		for _, f := range fields {
			if v := f.get(on); v != "" && v != "[]" {
				appendChange(on, f.name, v, "")
			}
		}
	}

	return changes
}

// ComputeVersionDiff 对比 parent 与 child 版本，从 trigger 节点起（含）计算字段差异。
func ComputeVersionDiff(oldNodes, newNodes []model.LifeNode, triggerNodeID string) []model.NodeFieldChange {
	fromSeq := 0
	if triggerNodeID != "" {
		for _, n := range oldNodes {
			if n.ID == triggerNodeID {
				fromSeq = n.Sequence
				break
			}
		}
	}
	return DiffNodeLists(NodesFromSequence(oldNodes, fromSeq), NodesFromSequence(newNodes, fromSeq))
}
