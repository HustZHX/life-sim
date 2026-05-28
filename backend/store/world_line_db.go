package store

import (
	"database/sql"
	"time"

	"life-sim/backend/model"
)

func (s *Store) SaveWorldLine(timelineID string, wl *model.WorldLine) error {
	if wl == nil {
		return nil
	}
	wl.TimelineID = timelineID
	wl.UpdatedAt = time.Now()
	_, err := s.db.Exec(
		`UPDATE timelines SET world_line_json=?, updated_at=? WHERE id=?`,
		MustWorldLineJSON(wl), wl.UpdatedAt, timelineID,
	)
	return err
}

func (s *Store) GetWorldLine(timelineID string) (*model.WorldLine, error) {
	var raw sql.NullString
	err := s.db.QueryRow(`SELECT COALESCE(world_line_json,'') FROM timelines WHERE id=?`, timelineID).Scan(&raw)
	if err != nil {
		return nil, err
	}
	if !raw.Valid || raw.String == "" {
		return nil, sql.ErrNoRows
	}
	return ParseWorldLine(raw.String, timelineID)
}

func scanTimelineWorldLine(t *model.Timeline, worldLineJSON sql.NullString) {
	if worldLineJSON.Valid && worldLineJSON.String != "" {
		if wl, err := ParseWorldLine(worldLineJSON.String, t.ID); err == nil {
			t.WorldLine = wl
		}
	}
}
