package store

import (
	"sort"

	"life-sim/backend/model"
)

func BuildBranchTree(versions []model.TimelineVersion, activeVersionID string) []model.BranchNode {
	if len(versions) == 0 {
		return nil
	}
	byID := make(map[string]model.TimelineVersion, len(versions))
	childrenOf := make(map[string][]string)
	var roots []string
	for _, v := range versions {
		byID[v.ID] = v
		pid := v.ParentVersionID
		if pid == "" {
			roots = append(roots, v.ID)
			continue
		}
		if _, ok := byID[pid]; !ok {
			// 父版本可能不在列表中，当作根
			roots = append(roots, v.ID)
			continue
		}
		childrenOf[pid] = append(childrenOf[pid], v.ID)
	}
	sort.Slice(roots, func(i, j int) bool {
		return byID[roots[i]].CreatedAt.Before(byID[roots[j]].CreatedAt)
	})
	for pid := range childrenOf {
		sort.Slice(childrenOf[pid], func(i, j int) bool {
			return byID[childrenOf[pid][i]].CreatedAt.Before(byID[childrenOf[pid][j]].CreatedAt)
		})
	}
	var build func(id string) model.BranchNode
	build = func(id string) model.BranchNode {
		v := byID[id]
		label := v.BranchLabel
		if label == "" {
			label = v.ChangeSummary
		}
		if label == "" {
			label = "分支"
		}
		node := model.BranchNode{
			ID:                 v.ID,
			ParentID:           v.ParentVersionID,
			Label:              label,
			ChangeSummary:      v.ChangeSummary,
			ForkSequence:       v.ForkSequence,
			ForkNodeID:         v.ForkNodeID,
			NodeCount:          v.NodeCount,
			DeathYearSnapshot:  v.DeathYearSnapshot,
			DeathCauseSnapshot: v.DeathCauseSnapshot,
			IsActive:           v.ID == activeVersionID,
			CreatesBranch:      VersionCreatesBranch(v),
			CreatedAt:          v.CreatedAt,
		}
		for _, cid := range childrenOf[id] {
			child := build(cid)
			node.Children = append(node.Children, child)
		}
		return node
	}
	out := make([]model.BranchNode, 0, len(roots))
	for _, rid := range roots {
		out = append(out, build(rid))
	}
	return out
}
