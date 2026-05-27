package store

import (
	"encoding/json"
	"testing"
)

func TestParseTraitChangesJSON_objects(t *testing.T) {
	raw := json.RawMessage(`[{"field":"identity","before":"游侠","after":"官员","reason":"入仕"}]`)
	got, err := ParseTraitChangesJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Field != "身份" || got[0].After != "官员" {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestParseTraitChangesJSON_strings(t *testing.T) {
	raw := json.RawMessage(`["由浪漫转向现实","身份转为翰林供奉"]`)
	got, err := ParseTraitChangesJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Field != "变更" || got[1].After != "身份转为翰林供奉" {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestParseTraitChangesJSON_empty(t *testing.T) {
	got, err := ParseTraitChangesJSON(json.RawMessage(`[]`))
	if err != nil || len(got) != 0 {
		t.Fatalf("expected empty, got %+v err=%v", got, err)
	}
}
