package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"life-sim/backend/ai"
	"life-sim/backend/model"
	"life-sim/backend/store"
)

const dialogueContextMessages = 20

type DialogueService struct {
	store *store.Store
	ai    *ai.Client
	char  *CharacterService
}

func NewDialogueService(st *store.Store, aiClient *ai.Client, charSvc *CharacterService) *DialogueService {
	return &DialogueService{store: st, ai: aiClient, char: charSvc}
}

func (s *DialogueService) GetSavedIdentityOptions(characterID, nodeID string) ([]model.DialogueIdentityOption, error) {
	node, err := s.store.GetNode(nodeID)
	if err != nil {
		return nil, fmt.Errorf("节点不存在")
	}
	if node.CharacterID != characterID {
		return nil, fmt.Errorf("节点不属于该人物")
	}
	opts, err := s.store.GetDialogueIdentityPreset(characterID, node.VersionID, nodeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return []model.DialogueIdentityOption{}, nil
		}
		return nil, err
	}
	return opts, nil
}

func (s *DialogueService) ResolveIdentityOptions(ctx context.Context, characterID, nodeID, modelID string, regenerate bool) ([]model.DialogueIdentityOption, error) {
	node, err := s.store.GetNode(nodeID)
	if err != nil {
		return nil, fmt.Errorf("节点不存在")
	}
	if node.CharacterID != characterID {
		return nil, fmt.Errorf("节点不属于该人物")
	}
	if !regenerate {
		if saved, err := s.store.GetDialogueIdentityPreset(characterID, node.VersionID, nodeID); err == nil && len(saved) > 0 {
			return saved, nil
		}
	}
	opts, err := s.generateIdentityOptionsAI(ctx, characterID, nodeID, modelID)
	if err != nil {
		return nil, err
	}
	if len(opts) > 0 {
		if err := s.store.SaveDialogueIdentityPreset(characterID, node.VersionID, nodeID, ai.NormalizeModelID(modelID), opts); err != nil {
			return nil, err
		}
	}
	return opts, nil
}

func (s *DialogueService) generateIdentityOptionsAI(ctx context.Context, characterID, nodeID, modelID string) ([]model.DialogueIdentityOption, error) {
	node, err := s.store.GetNode(nodeID)
	if err != nil {
		return nil, fmt.Errorf("节点不存在")
	}
	if node.CharacterID != characterID {
		return nil, fmt.Errorf("节点不属于该人物")
	}
	profile, err := s.store.GetProfile(characterID)
	if err != nil {
		return nil, err
	}
	nodes, err := s.store.GetNodesByVersion(node.VersionID)
	if err != nil {
		return nil, err
	}

	system, err := s.ai.LoadPrompt("dialogue_identity_options.txt")
	if err != nil {
		return nil, err
	}
	nodeJSON, _ := json.Marshal(map[string]any{
		"sequence": node.Sequence, "year": node.Year, "age": node.Age,
		"title": node.Title, "events": node.Events, "scene": node.Scene, "entities": node.Entities,
	})
	user := fmt.Sprintf(
		"人物：%s\n档案摘要：\n%s\n\n前置节点：\n%s\n\n当前节点：\n%s",
		profile.DisplayName,
		store.ProfileJSONInner(profile),
		store.MarshalNodesLiteBeforeSequence(nodes, node.Sequence),
		string(nodeJSON),
	)

	apiModel, err := ai.ResolveAPIModel(modelID)
	if err != nil {
		return nil, err
	}
	raw, err := s.ai.ChatJSONModel(ctx, apiModel, system, user)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Options []model.DialogueIdentityOption `json:"options"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, fmt.Errorf("解析身份选项失败: %w", err)
	}
	return resp.Options, nil
}

func (s *DialogueService) CreateSession(ctx context.Context, characterID, nodeID string, req model.CreateDialogueSessionRequest) (*model.DialogueSession, error) {
	node, err := s.store.GetNode(nodeID)
	if err != nil {
		return nil, fmt.Errorf("节点不存在")
	}
	if node.CharacterID != characterID {
		return nil, fmt.Errorf("节点不属于该人物")
	}
	identity := strings.TrimSpace(req.Identity)
	if identity == "" {
		return nil, fmt.Errorf("身份不能为空")
	}
	modelID := ai.NormalizeModelID(req.Model)

	if req.Resume {
		if existing, err := s.store.FindLatestDialogueSession(characterID, node.VersionID, nodeID, identity); err == nil {
			msgs, _ := s.store.ListDialogueMessages(existing.ID, 100)
			existing.Messages = msgs
			return existing, nil
		}
	}

	sess := &model.DialogueSession{
		CharacterID:     characterID,
		VersionID:       node.VersionID,
		NodeID:          nodeID,
		NodeSequence:    node.Sequence,
		SpeakerIdentity: identity,
		Model:           modelID,
	}
	if err := s.store.CreateDialogueSession(sess); err != nil {
		return nil, err
	}
	sess.Messages = []model.DialogueMessage{}
	return sess, nil
}

func (s *DialogueService) GetSession(characterID, sessionID string) (*model.DialogueSession, error) {
	sess, err := s.store.GetDialogueSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("会话不存在")
	}
	if sess.CharacterID != characterID {
		return nil, fmt.Errorf("会话不属于该人物")
	}
	msgs, err := s.store.ListDialogueMessages(sess.ID, 100)
	if err != nil {
		return nil, err
	}
	sess.Messages = msgs
	return sess, nil
}

func (s *DialogueService) ListDialogueSessions(characterID, versionID string) ([]model.DialogueSessionSummary, error) {
	if _, err := s.store.GetCharacter(characterID); err != nil {
		return nil, fmt.Errorf("人物不存在")
	}
	return s.store.ListDialogueSessions(characterID, versionID, 100)
}

func (s *DialogueService) GetLatestSessionForNode(characterID, nodeID string) (*model.DialogueSession, error) {
	node, err := s.store.GetNode(nodeID)
	if err != nil {
		return nil, fmt.Errorf("节点不存在")
	}
	if node.CharacterID != characterID {
		return nil, fmt.Errorf("节点不属于该人物")
	}
	sess, err := s.store.FindLatestDialogueSessionForNode(characterID, node.VersionID, nodeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	msgs, err := s.store.ListDialogueMessages(sess.ID, 100)
	if err != nil {
		return nil, err
	}
	sess.Messages = msgs
	return sess, nil
}

func (s *DialogueService) SendMessage(ctx context.Context, characterID, sessionID string, req model.SendDialogueMessageRequest) (*model.SendDialogueMessageResponse, error) {
	sess, err := s.GetSession(characterID, sessionID)
	if err != nil {
		return nil, err
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, fmt.Errorf("消息不能为空")
	}
	if req.Model != "" {
		sess.Model = ai.NormalizeModelID(req.Model)
		_ = s.store.UpdateDialogueSessionModel(sess.ID, sess.Model)
	}

	node, err := s.store.GetNode(sess.NodeID)
	if err != nil {
		return nil, err
	}
	profile, err := s.store.GetProfile(characterID)
	if err != nil {
		return nil, err
	}
	nodes, err := s.store.GetNodesByVersion(node.VersionID)
	if err != nil {
		return nil, err
	}

	systemPrompt, err := s.buildDialogueSystemPrompt(profile, node, nodes, sess.SpeakerIdentity)
	if err != nil {
		return nil, err
	}

	apiModel, err := ai.ResolveAPIModel(sess.Model)
	if err != nil {
		return nil, err
	}

	msgs := s.buildChatMessages(systemPrompt, sess.Messages, content)
	reply, err := s.ai.ChatTextModel(ctx, apiModel, msgs)
	if err != nil {
		return nil, err
	}

	userMsg := &model.DialogueMessage{SessionID: sess.ID, Role: "user", Content: content}
	if err := s.store.SaveDialogueMessage(userMsg); err != nil {
		return nil, err
	}
	asstMsg := &model.DialogueMessage{SessionID: sess.ID, Role: "assistant", Content: reply}
	if err := s.store.SaveDialogueMessage(asstMsg); err != nil {
		return nil, err
	}
	_ = s.store.TouchDialogueSession(sess.ID)

	var impact *model.PersonalityImpact
	if detected, err := s.detectPersonalityImpact(ctx, apiModel, profile, node, content, reply); err == nil && detected != nil && detected.HasImpact {
		impact = detected
	}

	return &model.SendDialogueMessageResponse{
		UserMessage:      *userMsg,
		AssistantMessage: *asstMsg,
		Impact:           impact,
	}, nil
}

func (s *DialogueService) buildDialogueSystemPrompt(profile *model.Profile, node *model.LifeNode, nodes []model.LifeNode, identity string) (string, error) {
	base, err := s.ai.LoadPrompt("dialogue_chat_system.txt")
	if err != nil {
		return "", err
	}

	memories, _ := s.store.ListMemories(node.CharacterID, node.VersionID, node.Sequence)
	var memLines []string
	for _, m := range memories {
		memLines = append(memLines, fmt.Sprintf("- [%s于%d岁] %s", m.SpeakerIdentity, nodeAgeAtSequence(nodes, m.SourceSequence), m.Content))
	}
	memBlock := "（暂无）"
	if len(memLines) > 0 {
		memBlock = strings.Join(memLines, "\n")
	}

	identityDesc := identity
	if identity == model.DialogueIdentityReader {
		identityDesc = "命运读者：来自未来、通晓此人一生命运走向的读者，在此刻与他交谈；不要主动剧透未来，除非对方追问且符合对话逻辑。"
	}

	traitJSON, _ := json.Marshal(node.TraitChanges)
	sceneJSON, _ := json.Marshal(node.Scene)

	return fmt.Sprintf(`%s

【人物档案】
%s

【前置人生节点（sequence<%d）】
%s

【当前节点（你的此刻）】
sequence=%d year=%d age=%d title=%s
events=%s
thoughts=%s
personality_snapshot=%s
trait_changes=%s
scene=%s

【对谈者记忆（你在此时能想起的）】
%s

【用户身份】
%s
`, base,
		store.ProfileJSONInner(profile),
		node.Sequence,
		store.MarshalNodesLiteBeforeSequence(nodes, node.Sequence),
		node.Sequence, node.Year, node.Age, node.Title,
		node.Events, node.Thoughts, node.PersonalitySnapshot,
		string(traitJSON), string(sceneJSON),
		memBlock,
		identityDesc,
	), nil
}

func nodeAgeAtSequence(nodes []model.LifeNode, seq int) int {
	for _, n := range nodes {
		if n.Sequence == seq {
			return n.Age
		}
	}
	return 0
}

func (s *DialogueService) buildChatMessages(systemPrompt string, history []model.DialogueMessage, newUser string) []ai.ChatMessage {
	msgs := []ai.ChatMessage{{Role: "system", Content: systemPrompt}}
	start := 0
	if len(history) > dialogueContextMessages {
		start = len(history) - dialogueContextMessages
	}
	for _, m := range history[start:] {
		role := m.Role
		if role != "user" && role != "assistant" {
			continue
		}
		msgs = append(msgs, ai.ChatMessage{Role: role, Content: m.Content})
	}
	msgs = append(msgs, ai.ChatMessage{Role: "user", Content: newUser})
	return msgs
}

func (s *DialogueService) detectPersonalityImpact(ctx context.Context, apiModel string, profile *model.Profile, node *model.LifeNode, userMsg, assistantMsg string) (*model.PersonalityImpact, error) {
	system, err := s.ai.LoadPrompt("dialogue_personality_impact.txt")
	if err != nil {
		return nil, err
	}
	traitJSON, _ := json.Marshal(node.TraitChanges)
	user := fmt.Sprintf(
		"人物：%s\n当前节点 thoughts=%s\npersonality_snapshot=%s\ntrait_changes=%s\n\n本轮对话：\n用户：%s\n人物：%s",
		profile.DisplayName, node.Thoughts, node.PersonalitySnapshot, string(traitJSON),
		userMsg, assistantMsg,
	)
	raw, err := s.ai.ChatJSONModel(ctx, apiModel, system, user)
	if err != nil {
		return nil, err
	}
	return store.ParsePersonalityImpact(raw)
}

func (s *DialogueService) ListMemories(characterID, versionID string, upToSequence int) ([]model.CharacterMemory, error) {
	return s.store.ListMemories(characterID, versionID, upToSequence)
}

func (s *DialogueService) CreateMemory(characterID string, req model.CreateMemoryRequest) (*model.CharacterMemory, error) {
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, fmt.Errorf("记忆内容不能为空")
	}
	node, err := s.store.GetNode(req.SourceNodeID)
	if err != nil {
		return nil, fmt.Errorf("来源节点不存在")
	}
	if node.CharacterID != characterID {
		return nil, fmt.Errorf("节点不属于该人物")
	}
	seq := req.SourceSequence
	if seq <= 0 && node.Sequence >= 0 {
		seq = node.Sequence
	}
	m := &model.CharacterMemory{
		CharacterID:     characterID,
		VersionID:       req.VersionID,
		SourceNodeID:    req.SourceNodeID,
		SourceSequence:  seq,
		SpeakerIdentity: strings.TrimSpace(req.SpeakerIdentity),
		Content:         content,
	}
	if err := s.store.SaveMemory(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *DialogueService) SummarizeMemory(ctx context.Context, modelID, dialogueText, identity string) (string, error) {
	system, err := s.ai.LoadPrompt("dialogue_memory_summarize.txt")
	if err != nil {
		return "", err
	}
	user := fmt.Sprintf("对谈者身份：%s\n\n对话内容：\n%s", identity, dialogueText)
	apiModel, err := ai.ResolveAPIModel(modelID)
	if err != nil {
		return "", err
	}
	raw, err := s.ai.ChatJSONModel(ctx, apiModel, system, user)
	if err != nil {
		return "", err
	}
	var resp struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return "", err
	}
	return strings.TrimSpace(resp.Content), nil
}

func (s *DialogueService) UpdateMemory(characterID, memoryID string, req model.UpdateMemoryRequest) (*model.CharacterMemory, error) {
	m, err := s.store.GetMemory(memoryID)
	if err != nil {
		return nil, fmt.Errorf("记忆不存在")
	}
	if m.CharacterID != characterID {
		return nil, fmt.Errorf("记忆不属于该人物")
	}
	return s.store.UpdateMemory(memoryID, strings.TrimSpace(req.Content))
}

func (s *DialogueService) DeleteMemory(characterID, memoryID string) error {
	m, err := s.store.GetMemory(memoryID)
	if err != nil {
		return fmt.Errorf("记忆不存在")
	}
	if m.CharacterID != characterID {
		return fmt.Errorf("记忆不属于该人物")
	}
	return s.store.DeleteMemory(memoryID)
}

func (s *DialogueService) GetMemoryForUpdate(characterID, memoryID string) (*model.CharacterMemory, error) {
	m, err := s.store.GetMemory(memoryID)
	if err != nil {
		return nil, err
	}
	if m.CharacterID != characterID {
		return nil, fmt.Errorf("记忆不属于该人物")
	}
	return m, nil
}

func (s *DialogueService) ValidateActiveVersion(characterID, versionID string) error {
	ch, err := s.store.GetCharacter(characterID)
	if err != nil {
		return err
	}
	if ch.CurrentTimelineID == "" {
		return nil
	}
	timeline, err := s.store.GetTimeline(ch.CurrentTimelineID)
	if err != nil {
		return err
	}
	if timeline.CurrentVersionID != versionID {
		return fmt.Errorf("当前不在激活分支，请切换分支后再写入")
	}
	return nil
}
