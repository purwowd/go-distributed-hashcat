package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"errors"
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

// RegisterAgent hanya membuat agent baru dengan status default "offline"
func (h *AgentHandler) RegisterAgent(c *gin.Context) {
	type registerAgentDTO struct {
		Name         string `json:"name"` // Name is optional now, will be retrieved from database based on agent key
		IPAddress    string `json:"ip_address"`
		Port         int    `json:"port"`
		Capabilities string `json:"capabilities"`
		AgentKey     string `json:"agent_key" binding:"required"`
	}

	var dto registerAgentDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Status default selalu offline saat pertama kali daftar
	status := "offline"

	// Validasi required fields
	if dto.AgentKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "agent key is required",
			"code":    "AGENT_KEY_REQUIRED",
			"message": "Agent key is required to create an agent.",
		})
		return
	}

	// Get agent name from database based on agent key
	existingAgentByKey, err := h.agentUsecase.GetByAgentKey(c.Request.Context(), dto.AgentKey)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "agent key not found",
				"code":    "AGENT_KEY_NOT_FOUND",
				"message": "The provided agent key does not exist in the database. Please generate a valid agent key first.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Use agent name from database, or use provided name if it matches
	agentName := existingAgentByKey.Name
	if dto.Name != "" && dto.Name != agentName {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "agent name mismatch",
			"code":    "AGENT_NAME_MISMATCH",
			"message": "The provided agent name does not match the name associated with the provided agent key.",
		})
		return
	}

	req := domain.CreateAgentRequest{
		Name:         agentName, // Use name from database
		IPAddress:    dto.IPAddress,
		Port:         dto.Port,
		Capabilities: dto.Capabilities,
		AgentKey:     dto.AgentKey,
		Status:       status,
	}

	agent, err := h.agentUsecase.RegisterAgent(c.Request.Context(), &req)
	if err != nil {
		// Handle specific validation errors
		if strings.Contains(err.Error(), "agent key") {
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   err.Error(),
					"code":    "AGENT_KEY_NOT_FOUND",
					"message": "The provided agent key does not exist in the database. Please generate a valid agent key first.",
				})
				return
			}
			if strings.Contains(err.Error(), "does not match") {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   err.Error(),
					"code":    "AGENT_NAME_MISMATCH",
					"message": "The agent name does not match the name associated with the provided agent key.",
				})
				return
			}
		}

		// Handle IP address conflicts
		if strings.Contains(err.Error(), "IP address") && strings.Contains(err.Error(), "already used") {
			c.JSON(http.StatusConflict, gin.H{
				"error":   err.Error(),
				"code":    "IP_ADDRESS_CONFLICT",
				"message": "The IP address is already in use by another agent.",
			})
			return
		}

		if strings.Contains(err.Error(), "already exists") {
			if strings.Contains(err.Error(), "IP address") {
				c.JSON(http.StatusConflict, gin.H{
					"error":   err.Error(),
					"code":    "IP_ADDRESS_CONFLICT",
					"message": "The IP address is already in use by another agent.",
				})
				return
			}
			if strings.Contains(err.Error(), "agent name") {
				c.JSON(http.StatusConflict, gin.H{
					"error":   err.Error(),
					"code":    "AGENT_NAME_CONFLICT",
					"message": "An agent with this name already exists.",
				})
				return
			}
		}

		// Handle other errors
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

// UpdateAgentStatus updates the status of an agent
func (h *AgentHandler) UpdateAgentStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid agent ID",
			"code":    "INVALID_AGENT_ID",
			"message": "The provided agent ID is not valid.",
		})
		return
	}

	var req struct {
		Status  string `json:"status" binding:"required"`
		Message string `json:"message,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"code":    "INVALID_REQUEST",
			"message": "The request body is invalid.",
		})
		return
	}

	// Validate status
	validStatuses := []string{"online", "offline", "busy", "error"}
	isValidStatus := false
	for _, validStatus := range validStatuses {
		if req.Status == validStatus {
			isValidStatus = true
			break
		}
	}

	if !isValidStatus {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid status",
			"code":    "INVALID_STATUS",
			"message": "Status must be one of: online, offline, busy, error",
		})
		return
	}

	// Update agent status
	if err := h.agentUsecase.UpdateAgentStatus(c.Request.Context(), id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update agent status",
			"code":    "UPDATE_STATUS_FAILED",
			"message": "Failed to update agent status.",
		})
		return
	}

	// Update last seen
	// if err := h.agentUsecase.UpdateAgentLastSeen(c.Request.Context(), id); err != nil {
	// 	// Log error but don't fail the request
	// 	log.Printf("Failed to update agent last seen: %v", err)
	// }

	c.JSON(http.StatusOK, gin.H{
		"message": "Agent status updated successfully",
		"data": gin.H{
			"id":     id.String(),
			"status": req.Status,
		},
	})
}

// UpdateAgentHeartbeat updates the heartbeat of an agent
func (h *AgentHandler) UpdateAgentHeartbeat(c *gin.Context) {
	id := c.Param("id")
	agentID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agent ID"})
		return
	}

	if err := h.agentUsecase.UpdateAgentLastSeen(c.Request.Context(), agentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "heartbeat updated"})
}

// UpdateAgentSpeed updates the speed of an agent
func (h *AgentHandler) UpdateAgentSpeed(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid agent ID",
			"code":    "INVALID_AGENT_ID",
			"message": "The provided agent ID is not valid.",
		})
		return
	}

	var req struct {
		Speed   int64  `json:"speed" binding:"required,gte=0"`
		Message string `json:"message,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"code":    "INVALID_REQUEST",
			"message": "The request body is invalid.",
		})
		return
	}

	// Update agent speed
	if err := h.agentUsecase.UpdateAgentSpeed(c.Request.Context(), id, req.Speed); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update agent speed",
			"code":    "UPDATE_SPEED_FAILED",
			"message": "Failed to update agent speed.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Agent speed updated successfully",
		"data": gin.H{
			"id":    id.String(),
			"speed": req.Speed,
		},
	})
}

// UpdateAgentData updates only the data fields (ip_address, port, capabilities) without changing status
func (h *AgentHandler) UpdateAgentData(c *gin.Context) {
	type updateAgentDataDTO struct {
		AgentKey     string `json:"agent_key" binding:"required"`
		IPAddress    string `json:"ip_address"`
		Port         int    `json:"port"`
		Capabilities string `json:"capabilities"`
	}

	var dto updateAgentDataDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update only data fields, keep status unchanged (offline)
	if err := h.agentUsecase.UpdateAgentData(c.Request.Context(), dto.AgentKey, dto.IPAddress, dto.Port, dto.Capabilities); err != nil {
		// Handle specific validation errors
		if strings.Contains(err.Error(), "AGENT_KEY_NOT_FOUND:") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   fmt.Sprintf("Agent key %s not found in database", dto.AgentKey),
				"code":    "AGENT_KEY_NOT_FOUND",
				"message": fmt.Sprintf("The provided agent key '%s' does not exist in the database. Please generate a valid agent key first.", dto.AgentKey),
			})
			return
		}

		// Handle IP address conflicts
		if strings.Contains(err.Error(), "IP address") && strings.Contains(err.Error(), "already used") {
			c.JSON(http.StatusConflict, gin.H{
				"error":   fmt.Sprintf("IP address %s already in use", dto.IPAddress),
				"code":    "IP_ADDRESS_CONFLICT",
				"message": err.Error(),
			})
			return
		}

		// Handle other errors
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update agent data",
			"code":    "UPDATE_FAILED",
			"message": err.Error(),
		})
		return
	}

	// Success - no notification, just close modal
	c.JSON(http.StatusOK, gin.H{
		"message": "Agent data updated successfully",
		"code":    "UPDATE_SUCCESS",
	})
}

// GenerateAgentKey creates a new agent key entry
func (h *AgentHandler) GenerateAgentKey(c *gin.Context) {
	type generateAgentKeyDTO struct {
		Name string `json:"name" binding:"required"`
	}

	var dto generateAgentKeyDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agent, err := h.agentUsecase.GenerateAgentKey(c.Request.Context(), dto.Name)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{
				"error":   err.Error(),
				"code":    "AGENT_NAME_EXISTS",
				"message": "An agent with this name already exists.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": agent})
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

// DeleteAgent deletes an agent
func (h *AgentHandler) DeleteAgent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid agent ID",
			"code":    "INVALID_AGENT_ID",
			"message": "The provided agent ID is not valid.",
		})
		return
	}

	err = h.agentUsecase.DeleteAgent(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrAgentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Agent not found",
				"code":    "AGENT_NOT_FOUND",
				"message": "The specified agent was not found.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete agent",
			"code":    "DELETE_FAILED",
			"message": "Failed to delete the agent.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Agent deleted successfully",
		"data": gin.H{
			"id": id.String(),
		},
	})
}

// AgentStartup handles agent startup and validation
func (h *AgentHandler) AgentStartup(c *gin.Context) {
	var req struct {
		Name         string `json:"name" binding:"required"`
		IPAddress    string `json:"ip_address" binding:"required"`
		Port         int    `json:"port"`
		Capabilities string `json:"capabilities"`
		AgentKey     string `json:"agent_key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"code":    "INVALID_REQUEST",
			"message": "The request body is invalid.",
		})
		return
	}

	// Validate agent key exists
	existingAgentByKey, err := h.agentUsecase.GetByAgentKey(c.Request.Context(), req.AgentKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Agent key not found",
			"code":    "AGENT_KEY_NOT_FOUND",
			"message": "The provided agent key does not exist in the database.",
		})
		return
	}

	// Validate agent name matches
	if existingAgentByKey.Name != req.Name {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Agent name mismatch",
			"code":    "AGENT_NAME_MISMATCH",
			"message": "Agent name does not match the name associated with the provided agent key.",
		})
		return
	}

	// Check if agent already has IP, port, and capabilities
	hasExistingData := existingAgentByKey.IPAddress != "" &&
		existingAgentByKey.Port != 0 &&
		existingAgentByKey.Capabilities != ""

	if hasExistingData {
		// Agent already exists with data, just update status to online
		if err := h.agentUsecase.UpdateAgentStatus(c.Request.Context(), existingAgentByKey.ID, "online"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update agent status",
				"code":    "UPDATE_STATUS_FAILED",
				"message": "Failed to update agent status to online.",
			})
			return
		}

		// Update last seen
		if err := h.agentUsecase.UpdateAgentLastSeen(c.Request.Context(), existingAgentByKey.ID); err != nil {
			// Log error but don't fail the request
			log.Printf("Failed to update agent last seen: %v", err)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Agent already exists, status updated to online",
			"code":    "AGENT_ALREADY_EXISTS",
			"data": gin.H{
				"id":           existingAgentByKey.ID.String(),
				"name":         existingAgentByKey.Name,
				"ip_address":   existingAgentByKey.IPAddress,
				"port":         existingAgentByKey.Port,
				"capabilities": existingAgentByKey.Capabilities,
				"agent_key":    existingAgentByKey.AgentKey,
				"status":       "online",
			},
		})
		return
	}

	// Agent exists but no data, update with new data
	// Set default port if not provided
	if req.Port == 0 {
		req.Port = 8080
	}

	// Update agent with new data
	existingAgentByKey.IPAddress = req.IPAddress
	existingAgentByKey.Port = req.Port
	existingAgentByKey.Capabilities = req.Capabilities
	existingAgentByKey.Status = "online"
	existingAgentByKey.LastSeen = time.Now() // Update LastSeen to current time
	existingAgentByKey.UpdatedAt = time.Now()

	if err := h.agentUsecase.UpdateAgent(c.Request.Context(), existingAgentByKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update agent data",
			"code":    "UPDATE_AGENT_FAILED",
			"message": "Failed to update agent with new data.",
		})
		return
	}

	// Update last seen (this will also update the database)
	if err := h.agentUsecase.UpdateAgentLastSeen(c.Request.Context(), existingAgentByKey.ID); err != nil {
		// Log error but don't fail the request
		log.Printf("Failed to update agent last seen: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Agent data updated and status set to online",
		"code":    "AGENT_UPDATED",
		"data": gin.H{
			"id":           existingAgentByKey.ID.String(),
			"name":         existingAgentByKey.Name,
			"ip_address":   existingAgentByKey.IPAddress,
			"port":         existingAgentByKey.Port,
			"capabilities": existingAgentByKey.Capabilities,
			"agent_key":    existingAgentByKey.AgentKey,
			"status":       "online",
		},
	})
}

// AgentHeartbeat handles agent heartbeat using agent key
func (h *AgentHandler) AgentHeartbeat(c *gin.Context) {
	var req struct {
		AgentKey string `json:"agent_key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"code":    "INVALID_REQUEST",
			"message": "The request body is invalid.",
		})
		return
	}

	// Get agent by agent key
	agent, err := h.agentUsecase.GetByAgentKey(c.Request.Context(), req.AgentKey)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Agent not found",
			"code":    "AGENT_NOT_FOUND",
			"message": "The agent with the provided key was not found.",
		})
		return
	}

	// Update last seen
	if err := h.agentUsecase.UpdateAgentLastSeen(c.Request.Context(), agent.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update agent heartbeat",
			"code":    "UPDATE_HEARTBEAT_FAILED",
			"message": "Failed to update agent heartbeat.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Agent heartbeat updated successfully",
		"data": gin.H{
			"id":         agent.ID.String(),
			"name":       agent.Name,
			"status":     agent.Status,
			"updated_at": time.Now().Format(time.RFC3339),
		},
	})
}

type LocalFile struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Type    string `json:"type"`
	Hash    string `json:"hash"`
	ModTime string `json:"mod_time"`
}
