package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"life-sim/backend/model"
)

func scanNarrativeArtifact(row *sql.Row) (*model.NarrativeArtifact, error) {
	var a model.NarrativeArtifact
	var created string
	err := row.Scan(
		&a.ID, &a.CharacterID, &a.VersionID, &a.NodeID, &a.Kind,
		&a.FromSequence, &a.ToSequence, &a.Person, &a.ContentKey, &a.Content, &a.Model, &created,
	)
	if err != nil {
		return nil, err
	}
	a.CreatedAt = parseDBTime(created)
	return &a, nil
}

const narrativeArtifactSelect = `
	SELECT id, character_id, version_id, COALESCE(node_id,''), kind,
	       from_sequence, to_sequence, COALESCE(person,'first'), COALESCE(content_key,''), content, COALESCE(model,''), created_at
	FROM narrative_artifacts`

func (s *Store) GetNarrativeArtifact(q model.NarrativeArtifactQuery) (*model.NarrativeArtifact, error) {
	person := model.NormalizeLightNovelPerson(q.Person)
	row := s.db.QueryRow(narrativeArtifactSelect+`
		WHERE version_id=? AND kind=? AND node_id=? AND from_sequence=? AND to_sequence=? AND COALESCE(person,'first')=?`,
		q.VersionID, q.Kind, q.NodeID, q.FromSequence, q.ToSequence, person,
	)
	return scanNarrativeArtifact(row)
}

func (s *Store) GetNarrativeArtifactByContentKey(contentKey, kind string) (*model.NarrativeArtifact, error) {
	if contentKey == "" {
		return nil, sql.ErrNoRows
	}
	row := s.db.QueryRow(narrativeArtifactSelect+`
		WHERE content_key=? AND kind=?
		ORDER BY created_at DESC LIMIT 1`, contentKey, kind)
	return scanNarrativeArtifact(row)
}

func (s *Store) SaveNarrativeArtifact(a *model.NarrativeArtifact) error {
	now := time.Now()
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = now
	}
	a.Person = model.NormalizeLightNovelPerson(a.Person)
	_, err := s.db.Exec(`
		INSERT INTO narrative_artifacts
		  (id, character_id, version_id, node_id, kind, from_sequence, to_sequence, person, content_key, content, model, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(version_id, kind, node_id, from_sequence, to_sequence, person)
		DO UPDATE SET content=excluded.content, content_key=excluded.content_key, model=excluded.model, created_at=excluded.created_at`,
		a.ID, a.CharacterID, a.VersionID, a.NodeID, a.Kind,
		a.FromSequence, a.ToSequence, a.Person, a.ContentKey, a.Content, a.Model, a.CreatedAt,
	)
	return err
}

func (s *Store) DeleteNarrativeArtifact(q model.NarrativeArtifactQuery) error {
	person := model.NormalizeLightNovelPerson(q.Person)
	res, err := s.db.Exec(`
		DELETE FROM narrative_artifacts
		WHERE version_id=? AND kind=? AND node_id=? AND from_sequence=? AND to_sequence=? AND COALESCE(person,'first')=?`,
		q.VersionID, q.Kind, q.NodeID, q.FromSequence, q.ToSequence, person,
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

func (s *Store) DeleteNarrativeArtifactByContentKey(contentKey, kind string) error {
	if contentKey == "" {
		return sql.ErrNoRows
	}
	res, err := s.db.Exec(`DELETE FROM narrative_artifacts WHERE content_key=? AND kind=?`, contentKey, kind)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
