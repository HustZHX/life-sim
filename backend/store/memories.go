package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"life-sim/backend/model"
)

func (s *Store) ListMemories(characterID, versionID string, upToSequence int) ([]model.CharacterMemory, error) {
	query := `
		SELECT id, character_id, version_id, source_node_id, source_sequence,
		       speaker_identity, content, created_at, updated_at
		FROM character_memories
		WHERE character_id=? AND version_id=?`
	args := []any{characterID, versionID}
	if upToSequence >= 0 {
		query += ` AND source_sequence <= ?`
		args = append(args, upToSequence)
	}
	query += ` ORDER BY source_sequence ASC, created_at ASC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.CharacterMemory
	for rows.Next() {
		var m model.CharacterMemory
		var created, updated string
		if err := rows.Scan(
			&m.ID, &m.CharacterID, &m.VersionID, &m.SourceNodeID, &m.SourceSequence,
			&m.SpeakerIdentity, &m.Content, &created, &updated,
		); err != nil {
			return nil, err
		}
		m.CreatedAt = parseDBTime(created)
		m.UpdatedAt = parseDBTime(updated)
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) GetMemory(id string) (*model.CharacterMemory, error) {
	row := s.db.QueryRow(`
		SELECT id, character_id, version_id, source_node_id, source_sequence,
		       speaker_identity, content, created_at, updated_at
		FROM character_memories WHERE id=?`, id)
	var m model.CharacterMemory
	var created, updated string
	if err := row.Scan(
		&m.ID, &m.CharacterID, &m.VersionID, &m.SourceNodeID, &m.SourceSequence,
		&m.SpeakerIdentity, &m.Content, &created, &updated,
	); err != nil {
		return nil, err
	}
	m.CreatedAt = parseDBTime(created)
	m.UpdatedAt = parseDBTime(updated)
	return &m, nil
}

func (s *Store) SaveMemory(m *model.CharacterMemory) error {
	now := time.Now()
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
	_, err := s.db.Exec(`
		INSERT INTO character_memories
		  (id, character_id, version_id, source_node_id, source_sequence, speaker_identity, content, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.CharacterID, m.VersionID, m.SourceNodeID, m.SourceSequence,
		m.SpeakerIdentity, m.Content, m.CreatedAt, m.UpdatedAt,
	)
	return err
}

func (s *Store) UpdateMemory(id, content string) (*model.CharacterMemory, error) {
	now := time.Now()
	res, err := s.db.Exec(
		`UPDATE character_memories SET content=?, updated_at=? WHERE id=?`,
		content, now, id,
	)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, sql.ErrNoRows
	}
	return s.GetMemory(id)
}

func (s *Store) DeleteMemory(id string) error {
	res, err := s.db.Exec(`DELETE FROM character_memories WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// CopyMemoriesWithNewVersion 版本分叉时复制记忆。
func (s *Store) CopyMemoriesWithNewVersion(oldVersionID, newVersionID string) error {
	rows, err := s.db.Query(`
		SELECT character_id, source_node_id, source_sequence, speaker_identity, content, created_at
		FROM character_memories WHERE version_id=?`, oldVersionID)
	if err != nil {
		return err
	}
	defer rows.Close()

	now := time.Now()
	for rows.Next() {
		var characterID, sourceNodeID, speakerIdentity, content, created string
		var sourceSequence int
		if err := rows.Scan(&characterID, &sourceNodeID, &sourceSequence, &speakerIdentity, &content, &created); err != nil {
			return err
		}
		createdAt := parseDBTime(created)
		_, err := s.db.Exec(`
			INSERT INTO character_memories
			  (id, character_id, version_id, source_node_id, source_sequence, speaker_identity, content, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			uuid.New().String(), characterID, newVersionID, sourceNodeID, sourceSequence,
			speakerIdentity, content, createdAt, now,
		)
		if err != nil {
			return err
		}
	}
	return rows.Err()
}
