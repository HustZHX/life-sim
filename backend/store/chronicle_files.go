package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"life-sim/backend/model"
)

func SaveChronicleFile(dir string, entry model.SavedChronicle) (string, error) {
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, entry.ID+".json")
	raw, err := json.Marshal(entry)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, raw, 0644); err != nil {
		return "", err
	}
	return entry.ID, nil
}

func ListChronicleFiles(dir, characterID string) ([]model.SavedChronicleMeta, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	out := make([]model.SavedChronicleMeta, 0)
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".json") {
			continue
		}
		meta, err := readChronicleMeta(filepath.Join(dir, ent.Name()))
		if err != nil {
			continue
		}
		if characterID != "" && meta.CharacterID != characterID {
			continue
		}
		out = append(out, meta)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

func GetChronicleFile(dir, id string) (*model.SavedChronicle, error) {
	if id == "" || strings.Contains(id, "..") || strings.ContainsAny(id, `/\`) {
		return nil, fmt.Errorf("无效的史书 id")
	}
	path := filepath.Join(dir, id+".json")
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("史书不存在")
		}
		return nil, err
	}
	var entry model.SavedChronicle
	if err := json.Unmarshal(raw, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func readChronicleMeta(path string) (model.SavedChronicleMeta, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return model.SavedChronicleMeta{}, err
	}
	var entry model.SavedChronicle
	if err := json.Unmarshal(raw, &entry); err != nil {
		return model.SavedChronicleMeta{}, err
	}
	return model.SavedChronicleMeta{
		ID:           entry.ID,
		CharacterID:  entry.CharacterID,
		DisplayName:  entry.DisplayName,
		VersionID:    entry.VersionID,
		FromSequence: entry.FromSequence,
		ToSequence:   entry.ToSequence,
		Model:        entry.Model,
		Branch:       entry.Branch,
		CreatedAt:    entry.CreatedAt,
	}, nil
}
