package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"life-sim/backend/model"
	"life-sim/backend/service"
)

type DialogueHandler struct {
	dialogue *service.DialogueService
	char     *service.CharacterService
}

func NewDialogueHandler(dialogue *service.DialogueService, char *service.CharacterService) *DialogueHandler {
	return &DialogueHandler{dialogue: dialogue, char: char}
}

func (h *DialogueHandler) GetSavedIdentityOptions(c *gin.Context) {
	charID := c.Param("id")
	nodeID := c.Param("nodeId")
	opts, err := h.dialogue.GetSavedIdentityOptions(charID, nodeID)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	OK(c, gin.H{"options": opts, "has_saved": len(opts) > 0})
}

func (h *DialogueHandler) GenerateIdentityOptions(c *gin.Context) {
	charID := c.Param("id")
	nodeID := c.Param("nodeId")
	var req struct {
		Model      string `json:"model"`
		Regenerate bool   `json:"regenerate"`
	}
	_ = c.ShouldBindJSON(&req)
	opts, err := h.dialogue.ResolveIdentityOptions(c.Request.Context(), charID, nodeID, req.Model, req.Regenerate)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	OK(c, gin.H{"options": opts})
}

func (h *DialogueHandler) ListSessions(c *gin.Context) {
	charID := c.Param("id")
	versionID := c.Query("version_id")
	items, err := h.dialogue.ListDialogueSessions(charID, versionID)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	OK(c, gin.H{"items": items})
}

func (h *DialogueHandler) GetLatestSessionForNode(c *gin.Context) {
	charID := c.Param("id")
	nodeID := c.Param("nodeId")
	sess, err := h.dialogue.GetLatestSessionForNode(charID, nodeID)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	if sess == nil {
		OK(c, gin.H{"session": nil})
		return
	}
	OK(c, gin.H{"session": sess})
}

func (h *DialogueHandler) CreateSession(c *gin.Context) {
	charID := c.Param("id")
	nodeID := c.Param("nodeId")
	var req model.CreateDialogueSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}
	sess, err := h.dialogue.CreateSession(c.Request.Context(), charID, nodeID, req)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	OK(c, sess)
}

func (h *DialogueHandler) GetSession(c *gin.Context) {
	charID := c.Param("id")
	sessionID := c.Param("sessionId")
	sess, err := h.dialogue.GetSession(charID, sessionID)
	if err != nil {
		Fail(c, http.StatusNotFound, 404, err.Error())
		return
	}
	OK(c, sess)
}

func (h *DialogueHandler) SendMessage(c *gin.Context) {
	charID := c.Param("id")
	sessionID := c.Param("sessionId")
	var req model.SendDialogueMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}
	resp, err := h.dialogue.SendMessage(c.Request.Context(), charID, sessionID, req)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	OK(c, resp)
}

func (h *DialogueHandler) ListMemories(c *gin.Context) {
	charID := c.Param("id")
	versionID := c.Query("version_id")
	if versionID == "" {
		Fail(c, http.StatusBadRequest, 400, "缺少 version_id")
		return
	}
	upTo := -1
	if v := c.Query("up_to_sequence"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			upTo = n
		}
	}
	items, err := h.dialogue.ListMemories(charID, versionID, upTo)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	OK(c, gin.H{"items": items})
}

func (h *DialogueHandler) CreateMemory(c *gin.Context) {
	charID := c.Param("id")
	var req model.CreateMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}
	if err := h.dialogue.ValidateActiveVersion(charID, req.VersionID); err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	m, err := h.dialogue.CreateMemory(charID, req)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	OK(c, m)
}

func (h *DialogueHandler) SummarizeMemory(c *gin.Context) {
	var req struct {
		Model    string `json:"model"`
		Text     string `json:"text" binding:"required"`
		Identity string `json:"identity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}
	content, err := h.dialogue.SummarizeMemory(c.Request.Context(), req.Model, req.Text, req.Identity)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	OK(c, gin.H{"content": content})
}

func (h *DialogueHandler) UpdateMemory(c *gin.Context) {
	charID := c.Param("id")
	memoryID := c.Param("memoryId")
	var req model.UpdateMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}
	m, err := h.dialogue.GetMemoryForUpdate(charID, memoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			Fail(c, http.StatusNotFound, 404, "记忆不存在")
			return
		}
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	if err := h.dialogue.ValidateActiveVersion(charID, m.VersionID); err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	updated, err := h.dialogue.UpdateMemory(charID, memoryID, req)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	OK(c, updated)
}

func (h *DialogueHandler) DeleteMemory(c *gin.Context) {
	charID := c.Param("id")
	memoryID := c.Param("memoryId")
	m, err := h.dialogue.GetMemoryForUpdate(charID, memoryID)
	if err != nil {
		Fail(c, http.StatusNotFound, 404, "记忆不存在")
		return
	}
	if err := h.dialogue.ValidateActiveVersion(charID, m.VersionID); err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	if err := h.dialogue.DeleteMemory(charID, memoryID); err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	OK(c, gin.H{"deleted": true})
}

func (h *DialogueHandler) ApplyImpact(c *gin.Context) {
	charID := c.Param("id")
	nodeID := c.Param("nodeId")
	var req model.ApplyDialogueImpactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}
	resp, err := h.char.ApplyDialogueImpact(c.Request.Context(), charID, nodeID, req)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	OK(c, resp)
}
