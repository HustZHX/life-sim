package store

import (
	"strings"

	"life-sim/backend/model"
)

// traitFieldToCN 将 AI 可能输出的英文字段名规范为中文展示名。
var traitFieldToCN = map[string]string{
	"occupation":    "职业",
	"beliefs":       "信念",
	"personality":   "性格",
	"identity":      "身份",
	"status":        "地位",
	"social_class":  "阶层",
	"nationality":   "国籍",
	"worldview":     "世界观",
	"values":        "价值观",
	"ambition":      "志向",
	"temperament":   "气质",
	"relationship":  "人际关系",
	"title":         "头衔",
	"role":          "角色",
	"mindset":       "心态",
}

func NormalizeTraitChanges(changes []model.TraitChange) []model.TraitChange {
	if len(changes) == 0 {
		return changes
	}
	out := make([]model.TraitChange, len(changes))
	for i, c := range changes {
		out[i] = c
		key := strings.ToLower(strings.TrimSpace(c.Field))
		if cn, ok := traitFieldToCN[key]; ok {
			out[i].Field = cn
		}
	}
	return out
}
