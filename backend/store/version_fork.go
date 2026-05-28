package store

import (
	"strings"

	"life-sim/backend/model"
)

// VersionCreatesBranch 判断该版本是否代表一次「推演分叉」（修改节点并推演后续、叙述变更等）。
// 仅更新本节点内心/性格、对话影响等不算分叉。
func VersionCreatesBranch(v model.TimelineVersion) bool {
	if v.ParentVersionID == "" {
		return false
	}
	label := strings.TrimSpace(v.BranchLabel)
	summary := strings.TrimSpace(v.ChangeSummary)
	combined := label + " " + summary

	if strings.Contains(label, "推演后续") {
		return true
	}
	if strings.Contains(label, "世界线") {
		return true
	}
	if strings.Contains(combined, "叙述变更") {
		return true
	}
	if strings.Contains(label, "更新节点") {
		return false
	}
	if strings.Contains(label, "对话影响") {
		return false
	}
	if strings.Contains(summary, "重算节点") && strings.Contains(summary, "内心") {
		return false
	}
	if strings.Contains(summary, "编辑节点") && strings.Contains(summary, "推演后续") {
		return true
	}
	return false
}

// PrepareBranchTreeVersions 折叠「非分叉」中间版本，使分支树只展示真正的推演分叉。
func PrepareBranchTreeVersions(versions []model.TimelineVersion) []model.TimelineVersion {
	if len(versions) == 0 {
		return nil
	}
	byID := make(map[string]model.TimelineVersion, len(versions))
	for _, v := range versions {
		byID[v.ID] = v
	}

	resolveForkParent := func(v model.TimelineVersion) string {
		pid := v.ParentVersionID
		for pid != "" {
			p, ok := byID[pid]
			if !ok {
				return ""
			}
			if p.ParentVersionID == "" || VersionCreatesBranch(p) {
				return pid
			}
			pid = p.ParentVersionID
		}
		return ""
	}

	out := make([]model.TimelineVersion, 0, len(versions))
	for _, v := range versions {
		if v.ParentVersionID == "" {
			out = append(out, v)
			continue
		}
		if !VersionCreatesBranch(v) {
			continue
		}
		nv := v
		nv.ParentVersionID = resolveForkParent(v)
		out = append(out, nv)
	}
	return out
}

// ResolveOverviewVersionID 取某分支在总览中应展示的节点版本（沿非分叉链取最新叶版本）。
func ResolveOverviewVersionID(versions []model.TimelineVersion, branchVersionID, activeVersionID string) string {
	if len(versions) == 0 {
		return branchVersionID
	}
	if branchVersionID == activeVersionID {
		return activeVersionID
	}
	byID := make(map[string]model.TimelineVersion, len(versions))
	children := make(map[string][]string)
	for _, v := range versions {
		byID[v.ID] = v
		if v.ParentVersionID != "" {
			children[v.ParentVersionID] = append(children[v.ParentVersionID], v.ID)
		}
	}
	if !descendsTo(byID, children, branchVersionID, activeVersionID) {
		return branchVersionID
	}

	cur := branchVersionID
	for {
		kids := children[cur]
		if len(kids) == 0 {
			return cur
		}
		next := kids[len(kids)-1]
		for _, kid := range kids {
			if kid == activeVersionID || descendsTo(byID, children, kid, activeVersionID) {
				next = kid
				break
			}
		}
		if VersionCreatesBranch(byID[next]) {
			return cur
		}
		cur = next
	}
}

// ResolveActiveBranchInTree 将当前版本映射到折叠后分支树上的节点（用于高亮「当前分支」）。
func ResolveActiveBranchInTree(versions []model.TimelineVersion, activeVersionID string) string {
	if activeVersionID == "" || len(versions) == 0 {
		return activeVersionID
	}
	byID := make(map[string]model.TimelineVersion, len(versions))
	for _, v := range versions {
		byID[v.ID] = v
	}
	cur := activeVersionID
	nearestFork := activeVersionID
	for cur != "" {
		v, ok := byID[cur]
		if !ok {
			break
		}
		if VersionCreatesBranch(v) {
			nearestFork = cur
		}
		if v.ParentVersionID == "" {
			if nearestFork != "" && nearestFork != activeVersionID && VersionCreatesBranch(byID[nearestFork]) {
				return nearestFork
			}
			return cur
		}
		cur = v.ParentVersionID
	}
	return nearestFork
}

func descendsTo(byID map[string]model.TimelineVersion, children map[string][]string, from, target string) bool {
	if from == target {
		return true
	}
	for _, kid := range children[from] {
		if descendsTo(byID, children, kid, target) {
			return true
		}
	}
	return false
}
