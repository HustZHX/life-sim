package ai

import (
	"context"
	"time"
)

const sessionMaxMessages = 4
const sessionTTL = 30 * time.Minute

// Session 轻量多轮上下文，用于推荐→生成等短链路。
type Session struct {
	CharacterID  string
	ProfileHash  string
	SystemPrompt string
	Messages     []chatMessage
	CreatedAt    time.Time
}

func NewSession(characterID, profileHash, systemPrompt, userContent, assistantContent string) *Session {
	return &Session{
		CharacterID:  characterID,
		ProfileHash:  profileHash,
		SystemPrompt: systemPrompt,
		Messages: []chatMessage{
			{Role: "user", Content: userContent},
			{Role: "assistant", Content: assistantContent},
		},
		CreatedAt: time.Now(),
	}
}

func (s *Session) Valid(profileHash string) bool {
	if s == nil {
		return false
	}
	if time.Since(s.CreatedAt) > sessionTTL {
		return false
	}
	return s.ProfileHash == profileHash
}

func (s *Session) AppendTurn(userContent, assistantContent string) {
	s.Messages = append(s.Messages,
		chatMessage{Role: "user", Content: userContent},
		chatMessage{Role: "assistant", Content: assistantContent},
	)
	if len(s.Messages) > sessionMaxMessages {
		s.Messages = s.Messages[len(s.Messages)-sessionMaxMessages:]
	}
}

// ChatJSONSession 在已有 session 上追加 user 消息并调用 API。
func (c *Client) ChatJSONSession(ctx context.Context, apiModel string, session *Session, userContent string) (string, error) {
	if apiModel == "" {
		apiModel = c.cfg.ModelFast
	}
	msgs := make([]chatMessage, 0, len(session.Messages)+2)
	msgs = append(msgs, chatMessage{Role: "system", Content: session.SystemPrompt})
	msgs = append(msgs, session.Messages...)
	msgs = append(msgs, chatMessage{Role: "user", Content: userContent})

	reqBody := chatRequest{
		Model:          apiModel,
		Messages:       msgs,
		MaxTokens:      8192,
		Temperature:    0.7,
		ResponseFormat: &responseFormat{Type: "json_object"},
	}

	content, err := c.doChatWithRetry(ctx, reqBody)
	if err != nil {
		return "", err
	}
	return extractJSON(content), nil
}
