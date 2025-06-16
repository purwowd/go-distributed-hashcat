package handler

import (
	"net/http"

	"go-distributed-hashcat/internal/domain"
	"go-distributed-hashcat/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AgentHandler struct {
	agentUsecase usecase.AgentUsecase
}

func NewAgentHandler(agentUsecase usecase.AgentUsecase) *AgentHandler {
	return &AgentHandler{
		agentUsecase: agentUsecase,
	}
}

func (h *AgentHandler) RegisterAgent(c *gin.Context) {
	var req domain.CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get agent from middleware (already validated)
	agentFromKey, exists := c.Get("agent")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Agent key validation failed"})
		return
	}

	keyAgent := agentFromKey.(*domain.Agent)

	// Register agent with the validated key
	agent, err := h.agentUsecase.RegisterAgentWithKey(c.Request.Context(), keyAgent.AgentKey, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": agent})
}

func (h *AgentHandler) GetAgent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	agent, err := h.agentUsecase.GetAgent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": agent})
}

func (h *AgentHandler) GetAllAgents(c *gin.Context) {
	agents, err := h.agentUsecase.GetAllAgents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": agents})
}

func (h *AgentHandler) UpdateAgentStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.agentUsecase.UpdateAgentStatus(c.Request.Context(), id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agent status updated successfully"})
}

func (h *AgentHandler) DeleteAgent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	if err := h.agentUsecase.DeleteAgent(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agent deleted successfully"})
}

func (h *AgentHandler) Heartbeat(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	if err := h.agentUsecase.UpdateAgentHeartbeat(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Heartbeat updated"})
}

func (h *AgentHandler) RegisterAgentFiles(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	var req struct {
		AgentID uuid.UUID            `json:"agent_id"`
		Files   map[string]LocalFile `json:"files"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate agent ID matches URL parameter
	if req.AgentID != id {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Agent ID mismatch"})
		return
	}

	// For now, just acknowledge the registration
	// In future, we could store file metadata in database
	c.JSON(http.StatusOK, gin.H{
		"message":    "Agent files registered successfully",
		"agent_id":   req.AgentID,
		"file_count": len(req.Files),
	})
}

// LocalFile represents a local file on agent
type LocalFile struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Type    string `json:"type"`
	Hash    string `json:"hash"`
	ModTime string `json:"mod_time"`
}
