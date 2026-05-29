package store

import (
	"fmt"
	"strings"

	"life-sim/backend/model"
)

// powerKeywordsYoung blocks obvious high-power actions for children (age < 14).
var powerKeywordsYoung = []string{
	"率军", "统帅", "挂帅", "领兵", "出征", "征战", "拜将", "拜相",
	"称帝", "登基", "监国", "开府", "牧守", "太守", "刺史", "丞相",
	"将军", "大将", "元帅", "统领三军", "号令天下",
}

// ValidateGameChoicesForAge rejects obviously age-inappropriate option text.
func ValidateGameChoicesForAge(age int, opts []model.GameChoiceOption) error {
	if age <= 0 || age >= 14 {
		return nil
	}
	for _, o := range opts {
		text := o.Label + o.Description
		for _, kw := range powerKeywordsYoung {
			if strings.Contains(text, kw) {
				return fmt.Errorf("选项「%s」含「%s」，不适合 %d 岁", o.Label, kw, age)
			}
		}
	}
	return nil
}
