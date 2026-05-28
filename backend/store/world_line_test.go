package store

import "testing"

func TestParseWorldLine_causedByNodeSequenceArray(t *testing.T) {
	raw := `{
		"historical_trend": "test",
		"events": [
			{
				"year": 1900,
				"name": "事件A",
				"caused_by_node_sequence": [2, 5]
			},
			{
				"year": 1901,
				"name": "事件B",
				"caused_by_node_sequence": 3
			}
		]
	}`
	wl, err := ParseWorldLine(raw, "tid-1")
	if err != nil {
		t.Fatalf("ParseWorldLine: %v", err)
	}
	if len(wl.Events) != 2 {
		t.Fatalf("events len = %d", len(wl.Events))
	}
	if wl.Events[0].CausedByNodeSeq == nil || *wl.Events[0].CausedByNodeSeq != 2 {
		t.Fatalf("event[0] seq = %v", wl.Events[0].CausedByNodeSeq)
	}
	if wl.Events[1].CausedByNodeSeq == nil || *wl.Events[1].CausedByNodeSeq != 3 {
		t.Fatalf("event[1] seq = %v", wl.Events[1].CausedByNodeSeq)
	}
}
