package store

import (
	"sort"

	"life-sim/backend/model"
)

// TrimSubsequentTail 保留 AI 返回的后续节点至多 target 条，并从 anchorSeq+1 重排 sequence。
func TrimSubsequentTail(nodes []model.LifeNode, anchorSeq, birthYear, target int) []model.LifeNode {
	if target <= 0 || len(nodes) == 0 {
		return nil
	}
	sorted := append([]model.LifeNode(nil), nodes...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Year != sorted[j].Year {
			return sorted[i].Year < sorted[j].Year
		}
		return sorted[i].Sequence < sorted[j].Sequence
	})
	if len(sorted) > target {
		sorted = sorted[:target]
	}
	out := make([]model.LifeNode, len(sorted))
	for i := range sorted {
		out[i] = sorted[i]
		out[i].Sequence = anchorSeq + 1 + i
		if birthYear > 0 && out[i].Year >= birthYear {
			out[i].Age = out[i].Year - birthYear
		}
	}
	return out
}

// MaxNodeYear 返回节点列表中最大 year。
func MaxNodeYear(nodes []model.LifeNode) int {
	maxY := 0
	for _, n := range nodes {
		if n.Year > maxY {
			maxY = n.Year
		}
	}
	return maxY
}
