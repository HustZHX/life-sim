package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"life-sim/backend/model"
)

func (s *Store) CreateDialogueSession(sess *model.DialogueSession) error {
	now := time.Now()
	if sess.ID == "" {
		sess.ID = uuid.New().String()
	}
	if sess.CreatedAt.IsZero() {
		sess.CreatedAt = now
	}
	sess.UpdatedAt = now
	_, err := s.db.Exec(`
		INSERT INTO dialogue_sessions
		  (id, character_id, version_id, node_id, node_sequence, speaker_identity, model, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sess.ID, sess.CharacterID, sess.VersionID, sess.NodeID, sess.NodeSequence,
		sess.SpeakerIdentity, sess.Model, sess.CreatedAt, sess.UpdatedAt,
	)
	return err
}

func (s *Store) FindLatestDialogueSession(characterID, versionID, nodeID, identity string) (*model.DialogueSession, error) {
	row := s.db.QueryRow(`
		SELECT id, character_id, version_id, node_id, node_sequence, speaker_identity, model, created_at, updated_at
		FROM dialogue_sessions
		WHERE character_id=? AND version_id=? AND node_id=? AND speaker_identity=?
		ORDER BY updated_at DESC LIMIT 1`,
		characterID, versionID, nodeID, identity,
	)
	return scanDialogueSession(row)
}

func (s *Store) FindLatestDialogueSessionForNode(characterID, versionID, nodeID string) (*model.DialogueSession, error) {
	row := s.db.QueryRow(`
		SELECT id, character_id, version_id, node_id, node_sequence, speaker_identity, model, created_at, updated_at
		FROM dialogue_sessions
		WHERE character_id=? AND version_id=? AND node_id=?
		ORDER BY updated_at DESC LIMIT 1`,
		characterID, versionID, nodeID,
	)
	return scanDialogueSession(row)
}

func (s *Store) GetDialogueSession(id string) (*model.DialogueSession, error) {
	row := s.db.QueryRow(`
		SELECT id, character_id, version_id, node_id, node_sequence, speaker_identity, model, created_at, updated_at
		FROM dialogue_sessions WHERE id=?`, id)
	return scanDialogueSession(row)
}

func scanDialogueSession(row *sql.Row) (*model.DialogueSession, error) {
	var sess model.DialogueSession
	var created, updated string
	if err := row.Scan(
		&sess.ID, &sess.CharacterID, &sess.VersionID, &sess.NodeID, &sess.NodeSequence,
		&sess.SpeakerIdentity, &sess.Model, &created, &updated,
	); err != nil {
		return nil, err
	}
	sess.CreatedAt = parseDBTime(created)
	sess.UpdatedAt = parseDBTime(updated)
	return &sess, nil
}

func (s *Store) TouchDialogueSession(id string) error {
	now := time.Now()
	_, err := s.db.Exec(`UPDATE dialogue_sessions SET updated_at=? WHERE id=?`, now, id)
	return err
}

func (s *Store) UpdateDialogueSessionModel(id, model string) error {
	now := time.Now()
	_, err := s.db.Exec(`UPDATE dialogue_sessions SET model=?, updated_at=? WHERE id=?`, model, now, id)
	return err
}

func (s *Store) ListDialogueSessions(characterID, versionID string, limit int) ([]model.DialogueSessionSummary, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `
		SELECT s.id, s.character_id, s.version_id, s.node_id, s.node_sequence, s.speaker_identity, s.model,
		       s.created_at, s.updated_at,
		       COALESCE(n.year, 0), COALESCE(n.age, 0), COALESCE(n.title, ''),
		       (SELECT COUNT(*) FROM dialogue_messages m WHERE m.session_id = s.id),
		       COALESCE((SELECT m.content FROM dialogue_messages m WHERE m.session_id = s.id ORDER BY m.created_at DESC LIMIT 1), '')
		FROM dialogue_sessions s
		LEFT JOIN life_nodes n ON n.id = s.node_id
		WHERE s.character_id = ?`
	args := []any{characterID}
	if versionID != "" {
		query += ` AND s.version_id = ?`
		args = append(args, versionID)
	}
	query += `
		AND EXISTS (SELECT 1 FROM dialogue_messages m WHERE m.session_id = s.id)
		ORDER BY s.updated_at DESC
		LIMIT ?`
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.DialogueSessionSummary
	for rows.Next() {
		var item model.DialogueSessionSummary
		var created, updated string
		if err := rows.Scan(
			&item.ID, &item.CharacterID, &item.VersionID, &item.NodeID, &item.NodeSequence,
			&item.SpeakerIdentity, &item.Model, &created, &updated,
			&item.NodeYear, &item.NodeAge, &item.NodeTitle,
			&item.MessageCount, &item.LastMessage,
		); err != nil {
			return nil, err
		}
		item.CreatedAt = parseDBTime(created)
		item.UpdatedAt = parseDBTime(updated)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) ListDialogueMessages(sessionID string, limit int) ([]model.DialogueMessage, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(`
		SELECT id, session_id, role, content, created_at
		FROM dialogue_messages WHERE session_id=?
		ORDER BY created_at ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []model.DialogueMessage
	for rows.Next() {
		var m model.DialogueMessage
		var created string
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &created); err != nil {
			return nil, err
		}
		m.CreatedAt = parseDBTime(created)
		all = append(all, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(all) > limit {
		all = all[len(all)-limit:]
	}
	return all, nil
}

func (s *Store) SaveDialogueMessage(m *model.DialogueMessage) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	_, err := s.db.Exec(`
		INSERT INTO dialogue_messages (id, session_id, role, content, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		m.ID, m.SessionID, m.Role, m.Content, m.CreatedAt,
	)
	return err
}
