package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"go-distributed-hashcat/internal/domain"
	"go-distributed-hashcat/internal/usecase"

	"strconv"

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
	// Use a local DTO to relax validation rules for key generation flow
	type registerAgentDTO struct {
		Name         string `json:"name" binding:"required"`
		IPAddress    string `json:"ip_address"` // optional for key generation
		Port         int    `json:"port"`
		Capabilities string `json:"capabilities"`
		AgentKey     string `json:"agent_key"`
	}

	var dto registerAgentDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Map DTO to domain request
	req := domain.CreateAgentRequest{
		Name:         dto.Name,
		IPAddress:    dto.IPAddress,
		Port:         dto.Port,
		Capabilities: dto.Capabilities,
		AgentKey:     dto.AgentKey,
	}

	// If no IP address (key generation flow), set default port to 0
	if req.IPAddress == "" {
		req.Port = 0
	} else if req.Port == 0 {
		req.Port = 8080
	}

	agent, err := h.agentUsecase.RegisterAgent(c.Request.Context(), &req)
	if err != nil {
		// Check if it's a duplicate agent error
		var duplicateErr *domain.DuplicateAgentError
		if errors.As(err, &duplicateErr) {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("already exists %s", duplicateErr.Name)})
			return
		}
		// Check if it's an already registered agent error
		var alreadyRegisteredErr *domain.AlreadyRegisteredAgentError
		if errors.As(err, &alreadyRegisteredErr) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		// Check for agent not found error
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		// Check for invalid agent key error
		if strings.Contains(err.Error(), "invalid agent key") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		// Check for agent key already registered with different agent name
		if strings.Contains(err.Error(), "is already registered with a different agent name") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		// Check for agent key not registered in database
		if strings.Contains(err.Error(), "is not registered in database") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
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
	// Server-side pagination & search (in-memory for now)
	page := 1
	pageSize := 10
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if s := c.Query("page_size"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 && v <= 500 {
			pageSize = v
		}
	}
	search := strings.ToLower(strings.TrimSpace(c.Query("search")))

	agents, err := h.agentUsecase.GetAllAgents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Filter by search
	if search != "" {
		filtered := make([]domain.Agent, 0, len(agents))
		for _, a := range agents {
			if strings.Contains(strings.ToLower(a.Name), search) ||
				strings.Contains(strings.ToLower(a.IPAddress), search) ||
				strings.Contains(strings.ToLower(a.Status), search) ||
				strings.Contains(strings.ToLower(a.AgentKey), search) {
				filtered = append(filtered, a)
			}
		}
		agents = filtered
	}

	total := len(agents)
	// Pagination
	start := (page - 1) * pageSize
	if start < 0 {
		start = 0
	}
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	paginated := agents[start:end]

	c.JSON(http.StatusOK, gin.H{
		"data":      paginated,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
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

	if req.AgentID != id {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Agent ID mismatch"})
		return
	}

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
