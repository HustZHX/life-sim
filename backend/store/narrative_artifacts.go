package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"life-sim/backend/model"
)

func (s *Store) GetNarrativeArtifact(q model.NarrativeArtifactQuery) (*model.NarrativeArtifact, error) {
	row := s.db.QueryRow(`
		SELECT id, character_id, version_id, COALESCE(node_id,''), kind,
		       from_sequence, to_sequence, content, COALESCE(model,''), created_at
		FROM narrative_artifacts
		WHERE version_id=? AND kind=? AND node_id=? AND from_sequence=? AND to_sequence=?`,
		q.VersionID, q.Kind, q.NodeID, q.FromSequence, q.ToSequence,
	)
	var a model.NarrativeArtifact
	var created string
	err := row.Scan(
		&a.ID, &a.CharacterID, &a.VersionID, &a.NodeID, &a.Kind,
		&a.FromSequence, &a.ToSequence, &a.Content, &a.Model, &created,
	)
	if err != nil {
		return nil, err
	}
	a.CreatedAt = parseDBTime(created)
	return &a, nil
}

func (s *Store) SaveNarrativeArtifact(a *model.NarrativeArtifact) error {
	now := time.Now()
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = now
	}
	_, err := s.db.Exec(`
		INSERT INTO narrative_artifacts
		  (id, character_id, version_id, node_id, kind, from_sequence, to_sequence, content, model, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(version_id, kind, node_id, from_sequence, to_sequence)
		DO UPDATE SET content=excluded.content, model=excluded.model, created_at=excluded.created_at`,
		a.ID, a.CharacterID, a.VersionID, a.NodeID, a.Kind,
		a.FromSequence, a.ToSequence, a.Content, a.Model, a.CreatedAt,
	)
	return err
}

func (s *Store) DeleteNarrativeArtifact(q model.NarrativeArtifactQuery) error {
	res, err := s.db.Exec(`
		DELETE FROM narrative_artifacts
		WHERE version_id=? AND kind=? AND node_id=? AND from_sequence=? AND to_sequence=?`,
		q.VersionID, q.Kind, q.NodeID, q.FromSequence, q.ToSequence,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
