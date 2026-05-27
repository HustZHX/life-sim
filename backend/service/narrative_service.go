package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"life-sim/backend/ai"
	"life-sim/backend/config"
	"life-sim/backend/model"
	"life-sim/backend/store"
)

type NarrativeService struct {
	store *store.Store
	ai    *ai.Client
}

func NewNarrativeService(st *store.Store, aiClient *ai.Client) *NarrativeService {
	return &NarrativeService{store: st, ai: aiClient}
}

func (s *NarrativeService) ListSavedLightNovels(characterID string) ([]model.SavedLightNovelMeta, error) {
	dir := config.ResolveLightNovelsDir()
	return store.ListLightNovelFiles(dir, characterID)
}

func (s *NarrativeService) GetSavedLightNovel(id string) (*model.SavedLightNovel, error) {
	dir := config.ResolveLightNovelsDir()
	return store.GetLightNovelFile(dir, id)
}

func (s *NarrativeService) GetArtifact(characterID string, q model.NarrativeArtifactQuery) (*model.NarrativeArtifact, error) {
	a, err := s.store.GetNarrativeArtifact(q)
	if err != nil {
		return nil, err
	}
	if a.CharacterID != characterID {
		return nil, fmt.Errorf("缓存不属于该角色")
	}
	return a, nil
}

func (s *NarrativeService) StartLightNovelJob(ctx context.Context, characterID string, req model.LightNovelRequest) (*model.Job, *model.NarrativeArtifact, error) {
	if req.VersionID == "" {
		return nil, nil, fmt.Errorf("缺少 version_id")
	}
	if req.FromSequence > req.ToSequence {
		return nil, nil, fmt.Errorf("节点区间无效")
	}
	req.Person = model.NormalizeLightNovelPerson(req.Person)

	ch, err := s.store.GetCharacter(characterID)
	if err != nil {
		return nil, nil, err
	}

	nodes, err := s.loadNodesInRange(req.VersionID, req.FromSequence, req.ToSequence)
	if err != nil {
		return nil, nil, err
	}
	if len(nodes) == 0 {
		return nil, nil, fmt.Errorf("选定区间无节点")
	}

	q := model.NarrativeArtifactQuery{
		VersionID: req.VersionID, Kind: model.NarrativeKindLightNovel,
		NodeID: "", FromSequence: req.FromSequence, ToSequence: req.ToSequence,
		Person: req.Person,
	}
	if !req.Force {
		if cached, err := s.store.GetNarrativeArtifact(q); err == nil {
			return nil, cached, nil
		}
	} else {
		_ = s.store.DeleteNarrativeArtifact(q)
	}

	modelID := ai.NormalizeModelID(req.Model)
	if _, err := s.resolveAPIModel(modelID); err != nil {
		return nil, nil, err
	}

	reqJSON, _ := json.Marshal(req)
	job, err := s.store.CreateJobWithRequest(characterID, "narrative_light_novel", modelID, string(reqJSON))
	if err != nil {
		return nil, nil, err
	}

	go s.runLightNovelJob(context.Background(), job.ID, characterID, ch.Mode, req, nodes, q)

	return job, nil, nil
}

func (s *NarrativeService) ListLightNovelJobs(characterID string, limit int) ([]model.Job, error) {
	return s.store.ListJobsByCharacterAndType(characterID, "narrative_light_novel", limit)
}

func (s *NarrativeService) StartNodeNarrativeJob(ctx context.Context, characterID, nodeID, kind string, req model.NodeNarrativeRequest) (*model.Job, *model.NarrativeArtifact, error) {
	switch kind {
	case model.NarrativeKindDiary, model.NarrativeKindLetter, model.NarrativeKindArchive:
	default:
		return nil, nil, fmt.Errorf("不支持的叙事类型: %s", kind)
	}

	node, err := s.store.GetNode(nodeID)
	if err != nil {
		return nil, nil, err
	}
	if node.CharacterID != characterID {
		return nil, nil, fmt.Errorf("节点不属于该角色")
	}

	ch, err := s.store.GetCharacter(characterID)
	if err != nil {
		return nil, nil, err
	}

	q := model.NarrativeArtifactQuery{
		VersionID: node.VersionID, Kind: kind, NodeID: nodeID,
	}
	if !req.Force {
		if cached, err := s.store.GetNarrativeArtifact(q); err == nil {
			return nil, cached, nil
		}
	} else {
		_ = s.store.DeleteNarrativeArtifact(q)
	}

	modelID := ai.NormalizeModelID(req.Model)
	if _, err := s.resolveAPIModel(modelID); err != nil {
		return nil, nil, err
	}

	jobType := "narrative_" + kind
	job, err := s.store.CreateJob(characterID, jobType, modelID)
	if err != nil {
		return nil, nil, err
	}

	go s.runNodeNarrativeJob(context.Background(), job.ID, characterID, ch.Mode, kind, *node, q)

	return job, nil, nil
}

func (s *NarrativeService) resolveAPIModel(modelID string) (string, error) {
	return ai.ResolveAPIModel(modelID)
}

func (s *NarrativeService) loadNodesInRange(versionID string, fromSeq, toSeq int) ([]model.LifeNode, error) {
	all, err := s.store.GetNodesByVersion(versionID)
	if err != nil {
		return nil, err
	}
	out := make([]model.LifeNode, 0)
	for _, n := range all {
		if n.Sequence >= fromSeq && n.Sequence <= toSeq {
			out = append(out, n)
		}
	}
	return out, nil
}

func (s *NarrativeService) runLightNovelJob(ctx context.Context, jobID, characterID, characterMode string, req model.LightNovelRequest, nodes []model.LifeNode, q model.NarrativeArtifactQuery) {
	job, _ := s.store.GetJob(jobID)
	job.Status = model.JobRunning
	job.Progress = 5
	job.StageText = "正在准备轻小说…"
	_ = s.store.UpdateJob(job)

	profile, err := s.store.GetProfile(characterID)
	if err != nil {
		s.failNarrativeJob(job, err.Error())
		return
	}

	promptName := "generate_light_novel_random.txt"
	if characterMode == model.ModeFamous {
		promptName = "generate_light_novel_famous.txt"
	}
	system, err := s.ai.LoadPrompt(promptName)
	if err != nil {
		s.failNarrativeJob(job, err.Error())
		return
	}

	apiModel, err := s.resolveAPIModel(job.Model)
	if err != nil {
		s.failNarrativeJob(job, err.Error())
		return
	}

	profileJSON := store.ProfileJSONTimeline(profile)
	batches := store.BatchLifeNodeSlices(nodes, store.ExpandBatchSize())
	personDirective := lightNovelPersonDirective(req.Person)
	var parts []string
	prevTail := ""

	for i, batch := range batches {
		progress := 10 + (i * 80 / maxInt(len(batches), 1))
		setJobStage(s.store, jobID, progress, fmt.Sprintf("正在撰写第 %d/%d 章…", i+1, len(batches)))

		user := fmt.Sprintf("人物档案：\n%s\n\n本章节包含的人生节点：\n%s\n\n【叙述人称】%s",
			profileJSON, store.MarshalNodesForNarrative(batch), personDirective)
		if prevTail != "" {
			user += fmt.Sprintf("\n\n上一章末尾（须自然衔接）：\n%s", prevTail)
		}
		if i == 0 {
			user += fmt.Sprintf("\n\n这是全书第 1/%d 章，请从选定区间的起点写起。", len(batches))
		}
		if i == len(batches)-1 && len(batches) > 1 {
			user += "\n\n这是最后一章，须收束本区间人生。"
		}

		done := make(chan struct{})
		go tickJobProgress(s.store, jobID, progress+1, progress+75/maxInt(len(batches), 1), done)

		raw, err := s.ai.ChatJSONModel(ctx, apiModel, system, user)
		close(done)
		if err != nil {
			s.failNarrativeJob(job, fmt.Sprintf("第 %d 章生成失败: %v", i+1, err))
			return
		}

		chapter, err := store.ParseLightNovelChapter(raw)
		if err != nil {
			s.failNarrativeJob(job, err.Error())
			return
		}
		formatted := store.FormatLightNovelChapter(chapter.Title, chapter.Content)
		parts = append(parts, formatted)
		prevTail = store.TailRunes(chapter.Content, 300)
	}

	content := strings.Join(parts, "\n\n---\n\n")
	artifact := &model.NarrativeArtifact{
		CharacterID: characterID, VersionID: q.VersionID, Kind: q.Kind,
		FromSequence: q.FromSequence, ToSequence: q.ToSequence, Person: req.Person,
		Content: content, Model: job.Model,
	}
	if err := s.store.SaveNarrativeArtifact(artifact); err != nil {
		s.failNarrativeJob(job, err.Error())
		return
	}

	branch := config.CurrentGitBranch()
	savedID, saveErr := store.SaveLightNovelFile(config.ResolveLightNovelsDir(), model.SavedLightNovel{
		CharacterID:  characterID,
		DisplayName:  profile.DisplayName,
		VersionID:    q.VersionID,
		FromSequence: q.FromSequence,
		ToSequence:   q.ToSequence,
		Person:       req.Person,
		Model:        job.Model,
		Branch:       branch,
		Content:      content,
	})
	if saveErr != nil {
		log.Printf("轻小说落盘失败: %v", saveErr)
	}

	result, _ := json.Marshal(map[string]interface{}{
		"artifact_id": artifact.ID, "content": content, "cached": false, "saved_file_id": savedID,
	})
	job.Status = model.JobCompleted
	job.Progress = 100
	job.StageText = "撰写完成"
	job.Result = string(result)
	_ = s.store.UpdateJob(job)
}

func (s *NarrativeService) runNodeNarrativeJob(ctx context.Context, jobID, characterID, characterMode, kind string, node model.LifeNode, q model.NarrativeArtifactQuery) {
	job, _ := s.store.GetJob(jobID)
	job.Status = model.JobRunning
	job.Progress = 10
	job.StageText = "正在生成…"
	_ = s.store.UpdateJob(job)

	profile, err := s.store.GetProfile(characterID)
	if err != nil {
		s.failNarrativeJob(job, err.Error())
		return
	}

	promptMap := map[string]string{
		model.NarrativeKindDiary:   "generate_node_diary.txt",
		model.NarrativeKindLetter:  "generate_node_letter.txt",
		model.NarrativeKindArchive: "generate_node_archive.txt",
	}
	system, err := s.ai.LoadPrompt(promptMap[kind])
	if err != nil {
		s.failNarrativeJob(job, err.Error())
		return
	}

	allNodes, _ := s.store.GetNodesByVersion(node.VersionID)
	priorNodes := make([]model.LifeNode, 0, 1)
	for _, n := range allNodes {
		if n.Sequence < node.Sequence {
			priorNodes = append(priorNodes, n)
		}
	}
	priorJSON := store.LastNodesLite(priorNodes, 1)

	user := fmt.Sprintf("人物档案：\n%s\n\n当前节点：\n%s",
		store.ProfileJSONTimeline(profile), store.MarshalSingleNodeForNarrative(node))
	if priorJSON != "[]" {
		user += "\n\n前一节点摘要（衔接参考）：\n" + priorJSON
	}
	_ = characterMode

	setJobStage(s.store, jobID, 25, "AI 正在撰写…")
	apiModel, _ := s.resolveAPIModel(job.Model)

	done := make(chan struct{})
	go tickJobProgress(s.store, jobID, 28, 85, done)

	raw, err := s.ai.ChatJSONModel(ctx, apiModel, system, user)
	close(done)
	if err != nil {
		s.failNarrativeJob(job, err.Error())
		return
	}

	content, err := store.ParseNarrativeContent(raw)
	if err != nil {
		s.failNarrativeJob(job, err.Error())
		return
	}

	artifact := &model.NarrativeArtifact{
		CharacterID: characterID, VersionID: q.VersionID, NodeID: q.NodeID,
		Kind: kind, Content: content, Model: job.Model,
	}
	if err := s.store.SaveNarrativeArtifact(artifact); err != nil {
		s.failNarrativeJob(job, err.Error())
		return
	}

	result, _ := json.Marshal(map[string]interface{}{
		"artifact_id": artifact.ID, "content": content, "cached": false,
	})
	job.Status = model.JobCompleted
	job.Progress = 100
	job.StageText = "完成"
	job.Result = string(result)
	_ = s.store.UpdateJob(job)
}

func (s *NarrativeService) failNarrativeJob(job *model.Job, msg string) {
	_ = s.store.MarkJobFailed(job.ID, msg)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func lightNovelPersonDirective(person string) string {
	switch model.NormalizeLightNovelPerson(person) {
	case model.LightNovelPersonSecond:
		return "全文以第二人称「你」撰写，读者即主人公本人，不用「我」或第三人称。"
	case model.LightNovelPersonThird:
		return "全文以第三人称撰写，以主人公姓名或「他/她」指代，不用「我」或「你」。"
	default:
		return "全文以第一人称「我」撰写。"
	}
}
