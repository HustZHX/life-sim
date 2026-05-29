package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"life-sim/backend/model"
	"life-sim/backend/service"
)

type GameHandler struct {
	game *service.GameService
}

func NewGameHandler(game *service.GameService) *GameHandler {
	return &GameHandler{game: game}
}

func (h *GameHandler) CreateGameCharacter(c *gin.Context) {
	var req struct {
		Mode string `json:"mode" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}
	ch, err := h.game.CreateGameCharacter(req.Mode)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	OK(c, ch)
}

func (h *GameHandler) UpdateConfig(c *gin.Context) {
	id := c.Param("id")
	var req model.GameUpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	ch, err := h.game.UpdateGameConfig(id, req.GameConfig)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	OK(c, ch)
}

func (h *GameHandler) EraOptions(c *gin.Context) {
	id := c.Param("id")
	var req model.GameEraOptionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	opts, err := h.game.GenerateEraOptions(c.Request.Context(), id, req)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	OK(c, model.GameEraOptionsResponse{Options: opts})
}

func (h *GameHandler) GenerateProfile(c *gin.Context) {
	id := c.Param("id")
	var req model.GameProfileGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	p, err := h.game.GenerateGameProfile(c.Request.Context(), id, req)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	OK(c, p)
}

func (h *GameHandler) StartTimeline(c *gin.Context) {
	id := c.Param("id")
	var req model.GameTimelineStartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	job, err := h.game.StartTimelineJob(c.Request.Context(), id, req)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	OK(c, job)
}

func (h *GameHandler) GetChoices(c *gin.Context) {
	charID := c.Param("id")
	nodeID := c.Param("nodeId")
	regenerate := false
	if r := c.Query("regenerate"); r == "1" || r == "true" {
		regenerate = true
	}
	modelID := c.Query("model")
	resp, err := h.game.GetNodeChoices(c.Request.Context(), charID, nodeID, modelID, regenerate)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	OK(c, resp)
}

func (h *GameHandler) Choose(c *gin.Context) {
	charID := c.Param("id")
	nodeID := c.Param("nodeId")
	var req model.GameChooseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	job, err := h.game.ChooseJob(c.Request.Context(), charID, nodeID, req)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	OK(c, job)
}

func (h *GameHandler) PostChoices(c *gin.Context) {
	regenerate := true
	if r := c.Query("regenerate"); r == "0" || r == "false" {
		regenerate = false
	}
	charID := c.Param("id")
	nodeID := c.Param("nodeId")
	var body struct {
		Model string `json:"model"`
	}
	_ = c.ShouldBindJSON(&body)
	resp, err := h.game.GetNodeChoices(c.Request.Context(), charID, nodeID, body.Model, regenerate)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	OK(c, resp)
}
