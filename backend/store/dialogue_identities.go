package store

import (
	"database/sql"
	"encoding/json"
	"time"

	"life-sim/backend/model"
)

func (s *Store) GetDialogueIdentityPreset(characterID, versionID, nodeID string) ([]model.DialogueIdentityOption, error) {
	var raw string
	err := s.db.QueryRow(`
		SELECT options_json FROM dialogue_identity_presets
		WHERE character_id=? AND version_id=? AND node_id=?`,
		characterID, versionID, nodeID,
	).Scan(&raw)
	if err != nil {
		return nil, err
	}
	var opts []model.DialogueIdentityOption
	if err := json.Unmarshal([]byte(raw), &opts); err != nil {
		return nil, err
	}
	return opts, nil
}

func (s *Store) SaveDialogueIdentityPreset(characterID, versionID, nodeID, modelID string, options []model.DialogueIdentityOption) error {
	raw, err := json.Marshal(options)
	if err != nil {
		return err
	}
	now := time.Now()
	_, err = s.db.Exec(`
		INSERT INTO dialogue_identity_presets
		  (character_id, version_id, node_id, options_json, model, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(character_id, version_id, node_id) DO UPDATE SET
		  options_json=excluded.options_json,
		  model=excluded.model,
		  updated_at=excluded.updated_at`,
		characterID, versionID, nodeID, string(raw), modelID, now, now,
	)
	return err
}

// CopyDialogueIdentityPresetsWithNewVersion 版本分叉时按 sequence 复制身份选项到新节点。
func (s *Store) CopyDialogueIdentityPresetsWithNewVersion(oldVersionID, newVersionID string) error {
	oldNodes, err := s.GetNodesByVersion(oldVersionID)
	if err != nil {
		return err
	}
	newNodes, err := s.GetNodesByVersion(newVersionID)
	if err != nil {
		return err
	}
	newBySeq := make(map[int]string, len(newNodes))
	for _, n := range newNodes {
		newBySeq[n.Sequence] = n.ID
	}
	for _, old := range oldNodes {
		newNodeID, ok := newBySeq[old.Sequence]
		if !ok {
			continue
		}
		var raw, modelID string
		err := s.db.QueryRow(`
			SELECT options_json, model FROM dialogue_identity_presets
			WHERE character_id=? AND version_id=? AND node_id=?`,
			old.CharacterID, oldVersionID, old.ID,
		).Scan(&raw, &modelID)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return err
		}
		now := time.Now()
		_, err = s.db.Exec(`
			INSERT INTO dialogue_identity_presets
			  (character_id, version_id, node_id, options_json, model, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(character_id, version_id, node_id) DO UPDATE SET
			  options_json=excluded.options_json,
			  model=excluded.model,
			  updated_at=excluded.updated_at`,
			old.CharacterID, newVersionID, newNodeID, raw, modelID, now, now,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
