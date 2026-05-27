package store

import (
	"strings"
	"testing"
)

func TestParseEraContextResult(t *testing.T) {
	raw := `{
	  "era_summary": "东汉末年战乱频仍",
	  "major_events": [
	    {"year": 184, "name": "黄巾起义", "impact": "郡县征发加剧"},
	    {"year": 208, "name": "赤壁之战", "impact": "荆襄流民增多"}
	  ],
	  "daily_life_context": "自耕农负担沉重"
	}`
	text, err := ParseEraContextResult(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "黄巾起义") || !strings.Contains(text, "184年") {
		t.Fatalf("missing events: %s", text)
	}
	if !strings.Contains(text, "日常生活背景") {
		t.Fatalf("missing daily life: %s", text)
	}
}
