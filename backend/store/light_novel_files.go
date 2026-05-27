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

// SaveLightNovelFile 将轻小说写入目录（每次生成新文件，不覆盖既有记录）。
func SaveLightNovelFile(dir string, entry model.SavedLightNovel) (string, error) {
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

// ListLightNovelFiles 列出目录下已保存轻小说（不含正文）；characterID 非空时按角色过滤。
func ListLightNovelFiles(dir, characterID string) ([]model.SavedLightNovelMeta, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	out := make([]model.SavedLightNovelMeta, 0)
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".json") {
			continue
		}
		meta, err := readLightNovelMeta(filepath.Join(dir, ent.Name()))
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

// GetLightNovelFile 读取单条已保存轻小说（含正文）。
func GetLightNovelFile(dir, id string) (*model.SavedLightNovel, error) {
	if id == "" || strings.Contains(id, "..") || strings.ContainsAny(id, `/\`) {
		return nil, fmt.Errorf("无效的轻小说 id")
	}
	path := filepath.Join(dir, id+".json")
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("轻小说不存在")
		}
		return nil, err
	}
	var entry model.SavedLightNovel
	if err := json.Unmarshal(raw, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func readLightNovelMeta(path string) (model.SavedLightNovelMeta, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return model.SavedLightNovelMeta{}, err
	}
	var entry model.SavedLightNovel
	if err := json.Unmarshal(raw, &entry); err != nil {
		return model.SavedLightNovelMeta{}, err
	}
	return model.SavedLightNovelMeta{
		ID:           entry.ID,
		CharacterID:  entry.CharacterID,
		DisplayName:  entry.DisplayName,
		VersionID:    entry.VersionID,
		FromSequence: entry.FromSequence,
		ToSequence:   entry.ToSequence,
		Person:       entry.Person,
		Model:        entry.Model,
		Branch:       entry.Branch,
		CreatedAt:    entry.CreatedAt,
	}, nil
}
