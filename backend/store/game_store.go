package store

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"life-sim/backend/model"
)

func (s *Store) SaveGameProfileSnapshot(characterID, timelineID string, maxSequence int, profile *model.Profile) error {
	data, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	now := time.Now()
	_, err = s.db.Exec(
		`INSERT INTO game_profile_snapshots (character_id, timeline_id, max_sequence, profile_json, created_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(timeline_id, max_sequence) DO UPDATE SET profile_json=excluded.profile_json, created_at=excluded.created_at`,
		characterID, timelineID, maxSequence, string(data), now,
	)
	return err
}

func (s *Store) GetGameProfileSnapshot(timelineID string, maxSequence int) (*model.Profile, error) {
	var data string
	err := s.db.QueryRow(
		`SELECT profile_json FROM game_profile_snapshots WHERE timeline_id = ? AND max_sequence = ?`,
		timelineID, maxSequence,
	).Scan(&data)
	if err != nil {
		return nil, err
	}
	var p model.Profile
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) DeleteGameProfileSnapshotsAfter(timelineID string, maxSequence int) error {
	_, err := s.db.Exec(
		`DELETE FROM game_profile_snapshots WHERE timeline_id = ? AND max_sequence > ?`,
		timelineID, maxSequence,
	)
	return err
}

func (s *Store) SaveGameNodeChoices(characterID, nodeID, versionID string, nodeSequence int, record *model.GameNodeChoiceRecord) error {
	data, err := json.Marshal(record.Options)
	if err != nil {
		return err
	}
	now := time.Now()
	_, err = s.db.Exec(
		`INSERT INTO game_node_choices (node_id, version_id, character_id, node_sequence, choices_json, chosen_id, custom_text, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(node_id, version_id) DO UPDATE SET
		   choices_json=excluded.choices_json,
		   chosen_id=excluded.chosen_id,
		   custom_text=excluded.custom_text,
		   node_sequence=excluded.node_sequence,
		   updated_at=excluded.updated_at`,
		nodeID, versionID, characterID, nodeSequence, string(data), record.ChosenID, record.CustomText, now, now,
	)
	return err
}

func (s *Store) GetGameNodeChoices(nodeID, versionID string) (*model.GameNodeChoiceRecord, int, error) {
	var choicesJSON, chosenID, customText string
	var nodeSeq int
	err := s.db.QueryRow(
		`SELECT node_sequence, choices_json, chosen_id, custom_text FROM game_node_choices WHERE node_id = ? AND version_id = ?`,
		nodeID, versionID,
	).Scan(&nodeSeq, &choicesJSON, &chosenID, &customText)
	if err == sql.ErrNoRows {
		return nil, 0, sql.ErrNoRows
	}
	if err != nil {
		return nil, 0, err
	}
	var opts []model.GameChoiceOption
	_ = json.Unmarshal([]byte(choicesJSON), &opts)
	return &model.GameNodeChoiceRecord{
		Options:    opts,
		ChosenID:   chosenID,
		CustomText: customText,
	}, nodeSeq, nil
}

func (s *Store) ListGameNodeChoicesByVersion(versionID string) ([]struct {
	NodeID       string
	NodeSequence int
	Record       model.GameNodeChoiceRecord
}, error) {
	rows, err := s.db.Query(
		`SELECT node_id, node_sequence, choices_json, chosen_id, custom_text FROM game_node_choices WHERE version_id = ? ORDER BY node_sequence`,
		versionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		NodeID       string
		NodeSequence int
		Record       model.GameNodeChoiceRecord
	}
	for rows.Next() {
		var nodeID, choicesJSON, chosenID, customText string
		var nodeSeq int
		if err := rows.Scan(&nodeID, &nodeSeq, &choicesJSON, &chosenID, &customText); err != nil {
			return nil, err
		}
		var opts []model.GameChoiceOption
		_ = json.Unmarshal([]byte(choicesJSON), &opts)
		out = append(out, struct {
			NodeID       string
			NodeSequence int
			Record       model.GameNodeChoiceRecord
		}{
			NodeID:       nodeID,
			NodeSequence: nodeSeq,
			Record: model.GameNodeChoiceRecord{
				Options: opts, ChosenID: chosenID, CustomText: customText,
			},
		})
	}
	return out, rows.Err()
}

func (s *Store) DeleteGameNodeChoicesAfter(characterID string, afterSequence int) error {
	_, err := s.db.Exec(
		`DELETE FROM game_node_choices WHERE character_id = ? AND node_sequence > ?`,
		characterID, afterSequence,
	)
	return err
}

func TruncateWorldLineEvents(wl *model.WorldLine, maxYear int) {
	if wl == nil {
		return
	}
	out := wl.Events[:0]
	for _, ev := range wl.Events {
		if ev.Year <= maxYear {
			out = append(out, ev)
		}
	}
	wl.Events = out
	if maxYear > 0 {
		wl.EndYear = maxYear
	}
}

func AppendWorldLineDelta(wl *model.WorldLine, delta *model.WorldLine, nextNodeSeq int) {
	if wl == nil || delta == nil {
		return
	}
	if delta.EraSummary != "" {
		wl.EraSummary = delta.EraSummary
	}
	if delta.HistoricalTrend != "" {
		wl.HistoricalTrend = delta.HistoricalTrend
	}
	if delta.DailyLifeContext != "" {
		wl.DailyLifeContext = delta.DailyLifeContext
	}
	lastYear := 0
	for _, ev := range wl.Events {
		if ev.Year > lastYear {
			lastYear = ev.Year
		}
	}
	for _, ev := range delta.Events {
		if ev.Year <= lastYear {
			continue
		}
		if ev.CausedByNodeSeq == nil && nextNodeSeq > 0 && strings.TrimSpace(ev.DivergenceNote) != "" {
			seq := nextNodeSeq
			ev.CausedByNodeSeq = &seq
		}
		wl.Events = append(wl.Events, ev)
		if ev.Year > wl.EndYear {
			wl.EndYear = ev.Year
		}
	}
}
