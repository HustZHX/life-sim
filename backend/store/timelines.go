package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"life-sim/backend/model"
)

func (s *Store) CreateTimeline(t *model.Timeline) error {
	now := time.Now()
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	t.UpdatedAt = now
	_, err := s.db.Exec(
		`INSERT INTO timelines (id, character_id, title, current_version_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		t.ID, t.CharacterID, t.Title, t.CurrentVersionID, t.CreatedAt, t.UpdatedAt,
	)
	return err
}

func (s *Store) GetTimeline(id string) (*model.Timeline, error) {
	row := s.db.QueryRow(
		`SELECT id, character_id, title, current_version_id, COALESCE(world_line_json,''), created_at, updated_at FROM timelines WHERE id = ?`, id,
	)
	var t model.Timeline
	var created, updated string
	var worldLineJSON sql.NullString
	if err := row.Scan(&t.ID, &t.CharacterID, &t.Title, &t.CurrentVersionID, &worldLineJSON, &created, &updated); err != nil {
		return nil, err
	}
	t.CreatedAt = parseDBTime(created)
	t.UpdatedAt = parseDBTime(updated)
	scanTimelineWorldLine(&t, worldLineJSON)
	if t.CurrentVersionID != "" {
		n, _ := s.CountNodesByVersion(t.CurrentVersionID)
		t.NodeCount = n
	}
	return &t, nil
}

func (s *Store) ListTimelines(characterID string) ([]model.Timeline, error) {
	rows, err := s.db.Query(
		`SELECT id, character_id, title, current_version_id, created_at, updated_at FROM timelines WHERE character_id = ? ORDER BY updated_at DESC`,
		characterID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.Timeline
	for rows.Next() {
		var t model.Timeline
		var created, updated string
		if err := rows.Scan(&t.ID, &t.CharacterID, &t.Title, &t.CurrentVersionID, &created, &updated); err != nil {
			return nil, err
		}
		t.CreatedAt = parseDBTime(created)
		t.UpdatedAt = parseDBTime(updated)
		if t.CurrentVersionID != "" {
			n, _ := s.CountNodesByVersion(t.CurrentVersionID)
			t.NodeCount = n
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (s *Store) CountVersionsByTimeline(timelineID string) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM timeline_versions WHERE timeline_id = ?`, timelineID).Scan(&n)
	return n, err
}

func (s *Store) ListActiveTimelineJobs(characterID string) ([]model.Job, error) {
	rows, err := s.db.Query(
		`SELECT id, character_id, type, status, progress, COALESCE(stage_text,''), COALESCE(model,''), COALESCE(request_json,''), result_json, error, created_at, updated_at
		 FROM jobs
		 WHERE character_id=? AND status IN (?, ?)
		   AND type IN ('timeline_generate','timeline_regenerate','node_inner_current','node_inner_subsequent','timeline_narrative_change','world_line_sync','world_line_refresh')
		 ORDER BY created_at DESC`,
		characterID, model.JobPending, model.JobRunning,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.Job, 0)
	for rows.Next() {
		var j model.Job
		var created, updated string
		if err := rows.Scan(&j.ID, &j.CharacterID, &j.Type, &j.Status, &j.Progress, &j.StageText, &j.Model, &j.RequestJSON, &j.Result, &j.Error, &created, &updated); err != nil {
			return nil, err
		}
		j.CreatedAt = parseDBTime(created)
		j.UpdatedAt = parseDBTime(updated)
		out = append(out, j)
	}
	return out, rows.Err()
}

// GetLatestTimelineGenerateJob 查找某时间轴最近一条 timeline_generate 任务。
func (s *Store) GetLatestTimelineGenerateJob(timelineID string) (*model.Job, error) {
	like := `%"timeline_id":"` + timelineID + `"%`
	row := s.db.QueryRow(
		`SELECT id, character_id, type, status, progress, COALESCE(stage_text,''), COALESCE(model,''), COALESCE(request_json,''), result_json, error, created_at, updated_at
		 FROM jobs
		 WHERE type = 'timeline_generate' AND (request_json LIKE ? OR result_json LIKE ?)
		 ORDER BY created_at DESC LIMIT 1`,
		like, like,
	)
	var j model.Job
	var created, updated string
	if err := row.Scan(&j.ID, &j.CharacterID, &j.Type, &j.Status, &j.Progress, &j.StageText, &j.Model, &j.RequestJSON, &j.Result, &j.Error, &created, &updated); err != nil {
		return nil, err
	}
	j.CreatedAt = parseDBTime(created)
	j.UpdatedAt = parseDBTime(updated)
	return &j, nil
}

func (s *Store) UpdateTimelineCurrentVersion(timelineID, versionID string) error {
	_, err := s.db.Exec(
		`UPDATE timelines SET current_version_id=?, updated_at=? WHERE id=?`,
		versionID, time.Now(), timelineID,
	)
	return err
}

func (s *Store) CountTimelinesByCharacter(characterID string) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM timelines WHERE character_id = ?`, characterID).Scan(&n)
	return n, err
}

func (s *Store) GetTimelineByVersionID(versionID string) (*model.Timeline, error) {
	var timelineID string
	err := s.db.QueryRow(`SELECT timeline_id FROM timeline_versions WHERE id = ?`, versionID).Scan(&timelineID)
	if err != nil {
		return nil, err
	}
	if timelineID == "" {
		return nil, sql.ErrNoRows
	}
	return s.GetTimeline(timelineID)
}

func (s *Store) migrateLegacyTimelines() error {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM timelines`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	rows, err := s.db.Query(`SELECT DISTINCT character_id FROM timeline_versions`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var charID string
		if err := rows.Scan(&charID); err != nil {
			return err
		}
		ch, err := s.GetCharacter(charID)
		if err != nil {
			continue
		}
		currentVersionID := ch.CurrentVersionID
		if currentVersionID == "" {
			_ = s.db.QueryRow(
				`SELECT id FROM timeline_versions WHERE character_id=? ORDER BY created_at DESC LIMIT 1`, charID,
			).Scan(&currentVersionID)
		}
		timelineID := uuid.New().String()
		title := "时间轴 1"
		if err := s.CreateTimeline(&model.Timeline{
			ID: timelineID, CharacterID: charID, Title: title, CurrentVersionID: currentVersionID,
		}); err != nil {
			return err
		}
		if _, err := s.db.Exec(`UPDATE timeline_versions SET timeline_id=? WHERE character_id=?`, timelineID, charID); err != nil {
			return err
		}
		ch.CurrentTimelineID = timelineID
		if err := s.UpdateCharacter(ch); err != nil {
			return err
		}
	}
	return rows.Err()
}
