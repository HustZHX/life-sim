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

type CharacterHandler struct {
	svc  *service.CharacterService
	narr *service.NarrativeService
}

func NewCharacterHandler(svc *service.CharacterService, narr *service.NarrativeService) *CharacterHandler {
	return &CharacterHandler{svc: svc, narr: narr}
}

func (h *CharacterHandler) ListHistory(c *gin.Context) {
	limit := 100
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	items, err := h.svc.ListHistory(limit)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	OK(c, gin.H{"items": items, "total": len(items)})
}

func (h *CharacterHandler) CreateCharacter(c *gin.Context) {
	var req struct {
		Mode string `json:"mode" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误: "+err.Error())
		return
	}
	ch, err := h.svc.CreateCharacter(req.Mode)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	OK(c, ch)
}

func (h *CharacterHandler) Resolve(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Query string `json:"query" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	candidates, err := h.svc.ResolvePerson(c.Request.Context(), id, req.Query)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	OK(c, gin.H{"candidates": candidates})
}

func (h *CharacterHandler) Confirm(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		CandidateIndex int `json:"candidate_index"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	ch, err := h.svc.ConfirmPerson(id, req.CandidateIndex)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	OK(c, ch)
}

func (h *CharacterHandler) SuggestNames(c *gin.Context) {
	var req model.SuggestNamesRequest
	_ = c.ShouldBindJSON(&req)
	names, err := h.svc.SuggestRandomNames(c.Request.Context(), req)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	OK(c, gin.H{"names": names})
}

func (h *CharacterHandler) GenerateProfile(c *gin.Context) {
	id := c.Param("id")
	var req model.ProfileGenerateRequest
	_ = c.ShouldBindJSON(&req)
	profile, err := h.svc.GenerateProfile(c.Request.Context(), id, req)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	OK(c, profile)
}

func (h *CharacterHandler) UpdateProfile(c *gin.Context) {
	id := c.Param("id")
	var req model.Profile
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	profile, err := h.svc.UpdateProfile(id, &req)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	OK(c, profile)
}

func (h *CharacterHandler) RandomizeProfileField(c *gin.Context) {
	id := c.Param("id")
	var req model.ProfileRandomizeFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	profile, err := h.svc.RandomizeProfileField(c.Request.Context(), id, req)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	OK(c, profile)
}

func (h *CharacterHandler) ImportTemplate(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		FamousQuery string `json:"famous_query" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	profile, err := h.svc.ImportTemplate(c.Request.Context(), id, req.FamousQuery)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	OK(c, profile)
}

func (h *CharacterHandler) GetProfile(c *gin.Context) {
	id := c.Param("id")
	profile, err := h.svc.GetProfile(id)
	if err != nil {
		Fail(c, http.StatusNotFound, 404, "档案不存在")
		return
	}
	OK(c, profile)
}

func (h *CharacterHandler) RecommendTimeline(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Model string `json:"model"`
	}
	_ = c.ShouldBindJSON(&req)
	recs, err := h.svc.RecommendTimelineConfigs(c.Request.Context(), id, req.Model)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	OK(c, gin.H{"recommendations": recs})
}

func (h *CharacterHandler) GenerateTimeline(c *gin.Context) {
	id := c.Param("id")
	var req model.TimelineGenerateRequest
	_ = c.ShouldBindJSON(&req)
	job, err := h.svc.StartTimelineJob(c.Request.Context(), id, req)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	OK(c, job)
}

func (h *CharacterHandler) GetTimeline(c *gin.Context) {
	id := c.Param("id")
	timelineID := c.Query("timeline_id")
	version := c.Query("version")
	tl, nodes, ver, err := h.svc.GetTimeline(id, timelineID, version)
	if err != nil {
		Fail(c, http.StatusNotFound, 404, err.Error())
		return
	}
	OK(c, gin.H{"timeline": tl, "version": ver, "nodes": nodes})
}

func (h *CharacterHandler) ListTimelines(c *gin.Context) {
	id := c.Param("id")
	list, err := h.svc.ListTimelines(id)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	OK(c, gin.H{"timelines": list})
}

func (h *CharacterHandler) GetNode(c *gin.Context) {
	nodeID := c.Param("nodeId")
	node, err := h.svc.GetNode(nodeID)
	if err != nil {
		Fail(c, http.StatusNotFound, 404, "节点不存在")
		return
	}
	OK(c, node)
}

func (h *CharacterHandler) PatchNode(c *gin.Context) {
	charID := c.Param("id")
	nodeID := c.Param("nodeId")
	var req model.PatchNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	job, err := h.svc.PatchNodeAndRegenerate(c.Request.Context(), charID, nodeID, req)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	OK(c, job)
}

func (h *CharacterHandler) ApplyNarrativeChange(c *gin.Context) {
	charID := c.Param("id")
	var req model.NarrativeChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	job, err := h.svc.ApplyNarrativeChange(c.Request.Context(), charID, req)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	OK(c, job)
}

func (h *CharacterHandler) PreviewLifespan(c *gin.Context) {
	charID := c.Param("id")
	nodeID := c.Param("nodeId")
	var req model.LifespanPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	resp, err := h.svc.PreviewLifespan(c.Request.Context(), charID, nodeID, req)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	OK(c, resp)
}

func (h *CharacterHandler) RegenerateNodeEvents(c *gin.Context) {
	charID := c.Param("id")
	nodeID := c.Param("nodeId")
	var req model.RegenerateNodeEventsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, 400, "参数错误")
		return
	}
	resp, err := h.svc.RegenerateNodeEvents(c.Request.Context(), charID, nodeID, req)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	OK(c, resp)
}

func (h *CharacterHandler) ListVersions(c *gin.Context) {
	id := c.Param("id")
	timelineID := c.Query("timeline_id")
	if timelineID == "" {
		timelineID = c.Param("tid")
	}
	versions, err := h.svc.ListVersions(id, timelineID)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	OK(c, gin.H{"versions": versions})
}

func (h *CharacterHandler) GetVersionDiff(c *gin.Context) {
	charID := c.Param("id")
	vid := c.Param("vid")
	diff, err := h.svc.GetVersionDiff(charID, vid)
	if err != nil {
		Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	OK(c, diff)
}

func (h *CharacterHandler) Rollback(c *gin.Context) {
	charID := c.Param("id")
	vid := c.Param("vid")
	ch, err := h.svc.Rollback(charID, vid)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	OK(c, ch)
}

func (h *CharacterHandler) GetJob(c *gin.Context) {
	jobID := c.Param("jobId")
	job, err := h.svc.GetJob(jobID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			Fail(c, http.StatusNotFound, 404, "任务不存在（可能服务已重启，请重新提交）")
			return
		}
		Fail(c, http.StatusInternalServerError, 500, "查询任务失败: "+err.Error())
		return
	}
	OK(c, job)
}

func (h *CharacterHandler) GetCharacter(c *gin.Context) {
	id := c.Param("id")
	ch, err := h.svc.GetCharacter(id)
	if err != nil {
		Fail(c, http.StatusNotFound, 404, "角色不存在")
		return
	}
	OK(c, ch)
}

func (h *CharacterHandler) ListModels(c *gin.Context) {
	OK(c, gin.H{"models": h.svc.ListModels()})
}

func (h *CharacterHandler) Health(c *gin.Context) {
	OK(c, gin.H{"status": "ok"})
}

func (h *CharacterHandler) GetNarrative(c *gin.Context) {
	charID := c.Param("id")
	q := model.NarrativeArtifactQuery{
		VersionID:    c.Query("version_id"),
		Kind:         c.Query("kind"),
		NodeID:       c.Query("node_id"),
		FromSequence: queryInt(c, "from_sequence"),
		ToSequence:   queryInt(c, "to_sequence"),
	}
	if q.VersionID == "" || q.Kind == "" {
		Fail(c, http.StatusBadRequest, 400, "缺少 version_id 或 kind")
		return
	}
	artifact, err := h.narr.GetArtifact(charID, q)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			Fail(c, http.StatusNotFound, 404, "暂无缓存")
			return
		}
		Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	OK(c, artifact)
}

func queryInt(c *gin.Context, key string) int {
	if v := c.Query(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return 0
}

func (h *CharacterHandler) GenerateLightNovel(c *gin.Context) {
	charID := c.Param("id")
	var req model.LightNovelRequest
	_ = c.ShouldBindJSON(&req)
	job, cached, err := h.narr.StartLightNovelJob(c.Request.Context(), charID, req)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	if cached != nil {
		OK(c, gin.H{"artifact": cached, "cached": true})
		return
	}
	OK(c, job)
}

func (h *CharacterHandler) GenerateNodeNarrative(c *gin.Context) {
	charID := c.Param("id")
	nodeID := c.Param("nodeId")
	kind := c.Param("kind")
	var req model.NodeNarrativeRequest
	_ = c.ShouldBindJSON(&req)
	job, cached, err := h.narr.StartNodeNarrativeJob(c.Request.Context(), charID, nodeID, kind, req)
	if err != nil {
		Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	if cached != nil {
		OK(c, gin.H{"artifact": cached, "cached": true})
		return
	}
	OK(c, job)
}
