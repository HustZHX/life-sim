package store

import "testing"

func TestParseRegenerateNodeFromTitle(t *testing.T) {
	raw := `{"events":"经历正文","thoughts":"我心想","personality_snapshot":"沉稳","trait_changes":[]}`
	r, err := ParseRegenerateNodeFromTitle(raw)
	if err != nil {
		t.Fatal(err)
	}
	if r.Events != "经历正文" || r.Thoughts != "我心想" {
		t.Fatalf("unexpected: %+v", r)
	}
}
