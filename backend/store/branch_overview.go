package store

import "life-sim/backend/model"

// FlattenBranchTreeDFS 按深度优先展开分支树，便于总览列表展示。
func FlattenBranchTreeDFS(roots []model.BranchNode) []model.BranchNode {
	var out []model.BranchNode
	var walk func(b model.BranchNode)
	walk = func(b model.BranchNode) {
		out = append(out, b)
		for _, c := range b.Children {
			walk(c)
		}
	}
	for _, r := range roots {
		walk(r)
	}
	return out
}

func NodesToOverviewLite(nodes []model.LifeNode) []model.BranchOverviewNodeLite {
	out := make([]model.BranchOverviewNodeLite, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, model.BranchOverviewNodeLite{
			Sequence: n.Sequence,
			Year:     n.Year,
			Title:    n.Title,
		})
	}
	return out
}
