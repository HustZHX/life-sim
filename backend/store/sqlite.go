package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
	"life-sim/backend/model"
)

type Store struct {
	db *sql.DB
}

func New(databasePath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(databasePath), 0755); err != nil {
		return nil, err
	}
	// WAL + busy_timeout：允许前端轮询读任务与后台 goroutine 写进度并发，避免 SQLITE_BUSY。
	dsn := databasePath + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS characters (
  id TEXT PRIMARY KEY,
  mode TEXT NOT NULL,
  display_name TEXT DEFAULT '',
  status TEXT NOT NULL DEFAULT 'draft',
  resolve_query TEXT DEFAULT '',
  confirmed_identity TEXT DEFAULT '',
  current_version_id TEXT DEFAULT '',
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS profiles (
  character_id TEXT PRIMARY KEY REFERENCES characters(id),
  data_json TEXT NOT NULL,
  updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS timeline_versions (
  id TEXT PRIMARY KEY,
  character_id TEXT NOT NULL REFERENCES characters(id),
  parent_version_id TEXT DEFAULT '',
  trigger_node_id TEXT DEFAULT '',
  change_summary TEXT DEFAULT '',
  created_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS life_nodes (
  id TEXT PRIMARY KEY,
  character_id TEXT NOT NULL REFERENCES characters(id),
  version_id TEXT NOT NULL REFERENCES timeline_versions(id),
  sequence INTEGER NOT NULL,
  year INTEGER NOT NULL,
  age INTEGER NOT NULL,
  title TEXT NOT NULL,
  events TEXT NOT NULL,
  thoughts TEXT NOT NULL,
  personality_snapshot TEXT NOT NULL,
  trait_changes_json TEXT DEFAULT '[]'
);

CREATE TABLE IF NOT EXISTS jobs (
  id TEXT PRIMARY KEY,
  character_id TEXT NOT NULL REFERENCES characters(id),
  type TEXT NOT NULL,
  status TEXT NOT NULL,
  progress INTEGER DEFAULT 0,
  model TEXT DEFAULT '',
  result_json TEXT DEFAULT '',
  error TEXT DEFAULT '',
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_nodes_version ON life_nodes(version_id, sequence);
CREATE INDEX IF NOT EXISTS idx_versions_char ON timeline_versions(character_id);
`
	_, err := s.db.Exec(schema)
	if err != nil {
		return err
	}
	_, _ = s.db.Exec(`ALTER TABLE jobs ADD COLUMN model TEXT DEFAULT ''`)
	_, _ = s.db.Exec(`ALTER TABLE life_nodes ADD COLUMN entities_json TEXT DEFAULT '{}'`)
	_, _ = s.db.Exec(`ALTER TABLE life_nodes ADD COLUMN scene_json TEXT DEFAULT '{}'`)
	_, _ = s.db.Exec(`PRAGMA journal_mode=WAL`)
	_, _ = s.db.Exec(`PRAGMA busy_timeout=5000`)

	_, _ = s.db.Exec(`ALTER TABLE characters ADD COLUMN current_timeline_id TEXT DEFAULT ''`)
	_, _ = s.db.Exec(`ALTER TABLE timeline_versions ADD COLUMN timeline_id TEXT DEFAULT ''`)

	_, err = s.db.Exec(`
CREATE TABLE IF NOT EXISTS timelines (
  id TEXT PRIMARY KEY,
  character_id TEXT NOT NULL REFERENCES characters(id),
  title TEXT NOT NULL DEFAULT '',
  current_version_id TEXT DEFAULT '',
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_timelines_char ON timelines(character_id);
CREATE INDEX IF NOT EXISTS idx_versions_timeline ON timeline_versions(timeline_id);
`)
	if err != nil {
		return err
	}
	return s.migrateLegacyTimelines()
}

func (s *Store) CreateCharacter(mode string) (*model.Character, error) {
	now := time.Now()
	c := &model.Character{
		ID:        uuid.New().String(),
		Mode:      mode,
		Status:    model.StatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err := s.db.Exec(
		`INSERT INTO characters (id, mode, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		c.ID, c.Mode, c.Status, c.CreatedAt, c.UpdatedAt,
	)
	return c, err
}

func (s *Store) GetCharacter(id string) (*model.Character, error) {
	row := s.db.QueryRow(
		`SELECT id, mode, display_name, status, resolve_query, confirmed_identity, current_version_id, COALESCE(current_timeline_id,''), created_at, updated_at FROM characters WHERE id = ?`,
		id,
	)
	var c model.Character
	var created, updated string
	err := row.Scan(&c.ID, &c.Mode, &c.DisplayName, &c.Status, &c.ResolveQuery, &c.ConfirmedIdentity, &c.CurrentVersionID, &c.CurrentTimelineID, &created, &updated)
	if err != nil {
		return nil, err
	}
	c.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	if c.CreatedAt.IsZero() {
		c.CreatedAt, _ = time.Parse("2006-01-02 15:04:05.999999999-07:00", created)
	}
	c.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
	if c.UpdatedAt.IsZero() {
		c.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05.999999999-07:00", updated)
	}
	return &c, nil
}

func parseDBTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err == nil && !t.IsZero() {
		return t
	}
	t, _ = time.Parse("2006-01-02 15:04:05.999999999-07:00", s)
	if !t.IsZero() {
		return t
	}
	t, _ = time.Parse("2006-01-02 15:04:05", s)
	return t
}

func (s *Store) CountNodesByVersion(versionID string) (int, error) {
	if versionID == "" {
		return 0, nil
	}
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM life_nodes WHERE version_id = ?`, versionID).Scan(&n)
	return n, err
}

func (s *Store) ListCharacterHistory(limit int) ([]model.CharacterHistoryItem, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(`
		SELECT c.id, c.mode, c.display_name, c.status, c.resolve_query, c.current_version_id, c.created_at, c.updated_at,
		       COALESCE(json_extract(p.data_json, '$.era'), ''),
		       COALESCE(json_extract(p.data_json, '$.birth_year'), 0),
		       COALESCE(json_extract(p.data_json, '$.death_year'), 0),
		       COALESCE(json_extract(p.data_json, '$.display_name'), ''),
		       COALESCE((
		         SELECT COUNT(*) FROM life_nodes ln
		         WHERE ln.version_id = c.current_version_id AND c.current_version_id != ''
		       ), 0),
		       COALESCE((SELECT COUNT(*) FROM timelines t WHERE t.character_id = c.id), 0)
		FROM characters c
		LEFT JOIN profiles p ON p.character_id = c.id
		ORDER BY c.updated_at DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.CharacterHistoryItem
	for rows.Next() {
		var item model.CharacterHistoryItem
		var created, updated, profileName string
		if err := rows.Scan(
			&item.ID, &item.Mode, &item.DisplayName, &item.Status, &item.ResolveQuery,
			&item.CurrentVersionID, &created, &updated,
			&item.Era, &item.BirthYear, &item.DeathYear, &profileName,
			&item.NodeCount, &item.TimelineCount,
		); err != nil {
			return nil, err
		}
		if item.DisplayName == "" && profileName != "" {
			item.DisplayName = profileName
		}
		if item.DisplayName == "" && item.ResolveQuery != "" {
			item.DisplayName = item.ResolveQuery
		}
		if item.DisplayName == "" {
			item.DisplayName = "未命名"
		}
		item.CreatedAt = parseDBTime(created)
		item.UpdatedAt = parseDBTime(updated)
		list = append(list, item)
	}
	return list, rows.Err()
}

func (s *Store) UpdateCharacter(c *model.Character) error {
	c.UpdatedAt = time.Now()
	_, err := s.db.Exec(
		`UPDATE characters SET display_name=?, status=?, resolve_query=?, confirmed_identity=?, current_version_id=?, current_timeline_id=?, updated_at=? WHERE id=?`,
		c.DisplayName, c.Status, c.ResolveQuery, c.ConfirmedIdentity, c.CurrentVersionID, c.CurrentTimelineID, c.UpdatedAt, c.ID,
	)
	return err
}

func (s *Store) SaveProfile(p *model.Profile) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	p.RawJSON = string(data)
	now := time.Now()
	_, err = s.db.Exec(
		`INSERT INTO profiles (character_id, data_json, updated_at) VALUES (?, ?, ?)
		 ON CONFLICT(character_id) DO UPDATE SET data_json=excluded.data_json, updated_at=excluded.updated_at`,
		p.CharacterID, string(data), now,
	)
	return err
}

func (s *Store) GetProfile(characterID string) (*model.Profile, error) {
	var data string
	err := s.db.QueryRow(`SELECT data_json FROM profiles WHERE character_id = ?`, characterID).Scan(&data)
	if err != nil {
		return nil, err
	}
	var p model.Profile
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		return nil, err
	}
	p.CharacterID = characterID
	return &p, nil
}

func (s *Store) CreateVersion(v *model.TimelineVersion) error {
	_, err := s.db.Exec(
		`INSERT INTO timeline_versions (id, character_id, timeline_id, parent_version_id, trigger_node_id, change_summary, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		v.ID, v.CharacterID, v.TimelineID, v.ParentVersionID, v.TriggerNodeID, v.ChangeSummary, v.CreatedAt,
	)
	return err
}

func (s *Store) ListVersions(characterID, timelineID string) ([]model.TimelineVersion, error) {
	query := `SELECT id, character_id, COALESCE(timeline_id,''), parent_version_id, trigger_node_id, change_summary, created_at FROM timeline_versions WHERE character_id = ?`
	args := []interface{}{characterID}
	if timelineID != "" {
		query += ` AND timeline_id = ?`
		args = append(args, timelineID)
	}
	query += ` ORDER BY created_at DESC`
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.TimelineVersion
	for rows.Next() {
		var v model.TimelineVersion
		var created string
		if err := rows.Scan(&v.ID, &v.CharacterID, &v.TimelineID, &v.ParentVersionID, &v.TriggerNodeID, &v.ChangeSummary, &created); err != nil {
			return nil, err
		}
		v.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		if v.CreatedAt.IsZero() {
			v.CreatedAt, _ = time.Parse("2006-01-02 15:04:05.999999999-07:00", created)
		}
		list = append(list, v)
	}
	return list, nil
}

func (s *Store) GetVersion(id string) (*model.TimelineVersion, error) {
	row := s.db.QueryRow(
		`SELECT id, character_id, COALESCE(timeline_id,''), parent_version_id, trigger_node_id, change_summary, created_at FROM timeline_versions WHERE id = ?`,
		id,
	)
	var v model.TimelineVersion
	var created string
	if err := row.Scan(&v.ID, &v.CharacterID, &v.TimelineID, &v.ParentVersionID, &v.TriggerNodeID, &v.ChangeSummary, &created); err != nil {
		return nil, err
	}
	v.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	return &v, nil
}

func (s *Store) SaveNodes(nodes []model.LifeNode) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, n := range nodes {
		tc, _ := json.Marshal(n.TraitChanges)
		entitiesJSON := marshalEntities(n.Entities)
		sceneJSON := marshalScene(n.Scene)
		_, err := tx.Exec(
			`INSERT INTO life_nodes (id, character_id, version_id, sequence, year, age, title, events, thoughts, personality_snapshot, trait_changes_json, entities_json, scene_json) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			n.ID, n.CharacterID, n.VersionID, n.Sequence, n.Year, n.Age, n.Title, n.Events, n.Thoughts, n.PersonalitySnapshot, string(tc), entitiesJSON, sceneJSON,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) GetNodesByVersion(versionID string) ([]model.LifeNode, error) {
	rows, err := s.db.Query(
		`SELECT id, character_id, version_id, sequence, year, age, title, events, thoughts, personality_snapshot, trait_changes_json, COALESCE(entities_json, '{}'), COALESCE(scene_json, '{}') FROM life_nodes WHERE version_id = ? ORDER BY sequence`,
		versionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNodes(rows)
}

func scanNodes(rows *sql.Rows) ([]model.LifeNode, error) {
	var list []model.LifeNode
	for rows.Next() {
		var n model.LifeNode
		var tcJSON, entitiesJSON, sceneJSON string
		if err := rows.Scan(&n.ID, &n.CharacterID, &n.VersionID, &n.Sequence, &n.Year, &n.Age, &n.Title, &n.Events, &n.Thoughts, &n.PersonalitySnapshot, &tcJSON, &entitiesJSON, &sceneJSON); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(tcJSON), &n.TraitChanges)
		n.TraitChanges = NormalizeTraitChanges(n.TraitChanges)
		n.Entities = parseEntitiesJSON(entitiesJSON)
		n.Scene = parseSceneJSON(sceneJSON)
		list = append(list, n)
	}
	return list, nil
}

func (s *Store) GetNode(id string) (*model.LifeNode, error) {
	row := s.db.QueryRow(
		`SELECT id, character_id, version_id, sequence, year, age, title, events, thoughts, personality_snapshot, trait_changes_json, COALESCE(entities_json, '{}'), COALESCE(scene_json, '{}') FROM life_nodes WHERE id = ?`,
		id,
	)
	var n model.LifeNode
	var tcJSON, entitiesJSON, sceneJSON string
	if err := row.Scan(&n.ID, &n.CharacterID, &n.VersionID, &n.Sequence, &n.Year, &n.Age, &n.Title, &n.Events, &n.Thoughts, &n.PersonalitySnapshot, &tcJSON, &entitiesJSON, &sceneJSON); err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(tcJSON), &n.TraitChanges)
	n.TraitChanges = NormalizeTraitChanges(n.TraitChanges)
	n.Entities = parseEntitiesJSON(entitiesJSON)
	n.Scene = parseSceneJSON(sceneJSON)
	return &n, nil
}

func (s *Store) CreateJob(characterID, jobType, modelID string) (*model.Job, error) {
	now := time.Now()
	j := &model.Job{
		ID:          uuid.New().String(),
		CharacterID: characterID,
		Type:        jobType,
		Status:      model.JobPending,
		Progress:    0,
		Model:       modelID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	_, err := s.db.Exec(
		`INSERT INTO jobs (id, character_id, type, status, progress, model, result_json, error, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, '', '', ?, ?)`,
		j.ID, j.CharacterID, j.Type, j.Status, j.Progress, j.Model, j.CreatedAt, j.UpdatedAt,
	)
	return j, err
}

func (s *Store) UpdateJob(j *model.Job) error {
	j.UpdatedAt = time.Now()
	_, err := s.db.Exec(
		`UPDATE jobs SET status=?, progress=?, model=?, result_json=?, error=?, updated_at=? WHERE id=?`,
		j.Status, j.Progress, j.Model, j.Result, j.Error, j.UpdatedAt, j.ID,
	)
	return err
}

// BumpJobProgressIfRunning 仅更新 running 任务的进度，避免与 failJob 竞态覆盖状态。
func (s *Store) BumpJobProgressIfRunning(jobID string, progress int) error {
	_, err := s.db.Exec(
		`UPDATE jobs SET progress=?, updated_at=? WHERE id=? AND status=?`,
		progress, time.Now(), jobID, model.JobRunning,
	)
	return err
}

// MarkJobFailed 原子地将任务标为失败，不会被进度 tick 覆盖回 running。
func (s *Store) MarkJobFailed(jobID, msg string) error {
	_, err := s.db.Exec(
		`UPDATE jobs SET status=?, error=?, updated_at=? WHERE id=? AND status IN (?, ?)`,
		model.JobFailed, msg, time.Now(), jobID, model.JobPending, model.JobRunning,
	)
	return err
}

func (s *Store) GetJob(id string) (*model.Job, error) {
	row := s.db.QueryRow(
		`SELECT id, character_id, type, status, progress, COALESCE(model,''), result_json, error, created_at, updated_at FROM jobs WHERE id = ?`,
		id,
	)
	var j model.Job
	var created, updated string
	if err := row.Scan(&j.ID, &j.CharacterID, &j.Type, &j.Status, &j.Progress, &j.Model, &j.Result, &j.Error, &created, &updated); err != nil {
		return nil, err
	}
	j.CreatedAt = parseDBTime(created)
	j.UpdatedAt = parseDBTime(updated)
	return &j, nil
}

// RecoverStaleJobs 将服务重启前未完成的 running/pending 任务标为失败。
func (s *Store) RecoverStaleJobs() (int64, error) {
	res, err := s.db.Exec(
		`UPDATE jobs SET status = ?, error = ?, progress = 0, updated_at = ?
		 WHERE status IN (?, ?)`,
		model.JobFailed, "服务重启导致任务中断，请重新提交", time.Now(),
		model.JobRunning, model.JobPending,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s *Store) DeleteNodesByVersion(versionID string) error {
	_, err := s.db.Exec(`DELETE FROM life_nodes WHERE version_id = ?`, versionID)
	return err
}

func MergeProfile(dst *model.Profile, patch map[string]string) {
	if v, ok := patch["personality_initial"]; ok {
		dst.PersonalityInitial = v
	}
	if v, ok := patch["era_background"]; ok {
		dst.EraBackground = v
	}
	if v, ok := patch["beliefs_motto"]; ok {
		dst.BeliefsMotto = v
	}
	if v, ok := patch["beliefs_politics"]; ok {
		dst.BeliefsPolitics = v
	}
	if v, ok := patch["template_source"]; ok {
		dst.TemplateSource = v
	}
}

func DiffNodes(oldNodes, newNodes []model.LifeNode) []model.NodeFieldChange {
	return DiffNodeLists(oldNodes, newNodes)
}

func FilterNodesFromSequence(nodes []model.LifeNode, fromSeq int) []model.LifeNode {
	var out []model.LifeNode
	for _, n := range nodes {
		if n.Sequence <= fromSeq {
			out = append(out, n)
		}
	}
	return out
}

func NodesAfterSequence(nodes []model.LifeNode, seq int) []model.LifeNode {
	var out []model.LifeNode
	for _, n := range nodes {
		if n.Sequence > seq {
			out = append(out, n)
		}
	}
	return out
}

// MarshalNodesBeforeSequence 序列化 sequence 严格小于 beforeSeq 的节点，供重算时提供前置人生上下文。
func MarshalNodesBeforeSequence(nodes []model.LifeNode, beforeSeq int) string {
	prior := make([]model.LifeNode, 0)
	for _, n := range nodes {
		if n.Sequence < beforeSeq {
			prior = append(prior, n)
		}
	}
	b, _ := json.Marshal(prior)
	return string(b)
}

func EncodeCandidates(candidates []model.ResolveCandidate) (string, error) {
	b, err := json.Marshal(map[string]interface{}{"candidates": candidates})
	return string(b), err
}

func ParseResolveResult(raw string) ([]model.ResolveCandidate, error) {
	var resp struct {
		Candidates []model.ResolveCandidate `json:"candidates"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, err
	}
	return resp.Candidates, nil
}

func ParseProfile(raw string, characterID string) (*model.Profile, error) {
	var p model.Profile
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return nil, err
	}
	p.CharacterID = characterID
	return &p, nil
}

func ParseTimelineNodes(raw string, characterID, versionID, protagonistName string) ([]model.LifeNode, error) {
	var resp struct {
		Nodes []struct {
			Sequence            int                 `json:"sequence"`
			Year                int                 `json:"year"`
			Age                 int                 `json:"age"`
			Title               string              `json:"title"`
			Events              string              `json:"events"`
			Thoughts            string              `json:"thoughts"`
			PersonalitySnapshot string              `json:"personality_snapshot"`
			TraitChanges        json.RawMessage     `json:"trait_changes"`
			Entities            *model.NodeEntities `json:"entities"`
			Scene               *model.NodeScene    `json:"scene"`
		} `json:"nodes"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, err
	}
	nodes := make([]model.LifeNode, 0, len(resp.Nodes))
	for _, n := range resp.Nodes {
		traits, err := ParseTraitChangesJSON(n.TraitChanges)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, model.LifeNode{
			ID:                  uuid.New().String(),
			CharacterID:         characterID,
			VersionID:           versionID,
			Sequence:            n.Sequence,
			Year:                n.Year,
			Age:                 n.Age,
			Title:               n.Title,
			Events:              n.Events,
			Thoughts:            n.Thoughts,
			PersonalitySnapshot: n.PersonalitySnapshot,
			TraitChanges:        traits,
			Entities:            NormalizeNodeEntities(n.Entities, protagonistName),
			Scene:               NormalizeNodeScene(n.Scene),
		})
	}
	return nodes, nil
}

func MustProfileJSON(p *model.Profile) string {
	b, _ := json.Marshal(p)
	return string(b)
}

func ParseInnerCurrent(raw string, protagonistName string) (thoughts, personality string, traitChanges []model.TraitChange, entities *model.NodeEntities, scene *model.NodeScene, err error) {
	var resp struct {
		Thoughts            string              `json:"thoughts"`
		PersonalitySnapshot string              `json:"personality_snapshot"`
		TraitChanges        json.RawMessage     `json:"trait_changes"`
		Entities            *model.NodeEntities `json:"entities"`
		Scene               *model.NodeScene    `json:"scene"`
	}
	if err = json.Unmarshal([]byte(raw), &resp); err != nil {
		return
	}
	traits, err := ParseTraitChangesJSON(resp.TraitChanges)
	if err != nil {
		return
	}
	return resp.Thoughts, resp.PersonalitySnapshot, traits, NormalizeNodeEntities(resp.Entities, protagonistName), NormalizeNodeScene(resp.Scene), nil
}

type InnerSubsequentPatch struct {
	Thoughts            string
	PersonalitySnapshot string
	TraitChanges        []model.TraitChange
	Entities            *model.NodeEntities
	Scene               *model.NodeScene
}

func ParseInnerSubsequent(raw string, protagonistName string) (map[int]InnerSubsequentPatch, error) {
	var resp struct {
		Nodes []struct {
			Sequence            int                 `json:"sequence"`
			Thoughts            string              `json:"thoughts"`
			PersonalitySnapshot string              `json:"personality_snapshot"`
			TraitChanges        json.RawMessage     `json:"trait_changes"`
			Entities            *model.NodeEntities `json:"entities"`
			Scene               *model.NodeScene    `json:"scene"`
		} `json:"nodes"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, err
	}
	m := make(map[int]InnerSubsequentPatch, len(resp.Nodes))
	for _, n := range resp.Nodes {
		traits, err := ParseTraitChangesJSON(n.TraitChanges)
		if err != nil {
			return nil, err
		}
		m[n.Sequence] = InnerSubsequentPatch{
			Thoughts:            n.Thoughts,
			PersonalitySnapshot: n.PersonalitySnapshot,
			TraitChanges:        traits,
			Entities:            NormalizeNodeEntities(n.Entities, protagonistName),
			Scene:               NormalizeNodeScene(n.Scene),
		}
	}
	return m, nil
}

func CopyNodesWithNewVersion(nodes []model.LifeNode, versionID string) []model.LifeNode {
	out := make([]model.LifeNode, len(nodes))
	for i, n := range nodes {
		out[i] = n
		out[i].VersionID = versionID
		out[i].ID = uuid.New().String()
	}
	return out
}

func BuildUserContext(parts ...string) string {
	return fmt.Sprintf("%s", joinParts(parts))
}

func joinParts(parts []string) string {
	var b string
	for i, p := range parts {
		if i > 0 {
			b += "\n\n"
		}
		b += p
	}
	return b
}
