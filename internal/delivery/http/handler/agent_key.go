package handler

import (
	"net/http"
	"strings"

	"go-distributed-hashcat/internal/domain"
	"go-distributed-hashcat/internal/usecase"

	"github.com/gin-gonic/gin"
)

type AgentKeyHandler struct {
	agentKeyUsecase usecase.AgentKeyUsecase
}

func NewAgentKeyHandler(agentKeyUsecase usecase.AgentKeyUsecase) *AgentKeyHandler {
	return &AgentKeyHandler{
		agentKeyUsecase: agentKeyUsecase,
	}
}

// GenerateAgentKey generates a new agent key
// @Summary Generate new agent key
// @Description Generate a new agent key for agent authentication
// @Tags agent-keys
// @Accept json
// @Produce json
// @Param request body domain.GenerateAgentKeyRequest true "Generate agent key request"
// @Success 201 {object} domain.GenerateAgentKeyResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/agent-keys/generate [post]
func (h *AgentKeyHandler) GenerateAgentKey(c *gin.Context) {
	var req domain.GenerateAgentKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	response, err := h.agentKeyUsecase.GenerateAgentKey(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate agent key",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// ListAgentKeys lists all agent keys
// @Summary List agent keys
// @Description Get list of all agent keys with their status
// @Tags agent-keys
// @Produce json
// @Success 200 {array} domain.AgentKeyInfo
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/agent-keys [get]
func (h *AgentKeyHandler) ListAgentKeys(c *gin.Context) {
	keys, err := h.agentKeyUsecase.ListAgentKeys(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list agent keys",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"agent_keys": keys,
		"count":      len(keys),
	})
}

// RevokeAgentKey revokes an agent key
// @Summary Revoke agent key
// @Description Revoke an agent key to prevent further access
// @Tags agent-keys
// @Param key path string true "Agent key to revoke"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/agent-keys/{key}/revoke [delete]
func (h *AgentKeyHandler) RevokeAgentKey(c *gin.Context) {
	agentKey := c.Param("key")
	if agentKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Agent key is required",
		})
		return
	}

	err := h.agentKeyUsecase.RevokeAgentKey(c.Request.Context(), agentKey)
	if err != nil {
		if err.Error() == "agent key not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Agent key not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to revoke agent key",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Agent key revoked successfully",
		"agent_key": agentKey,
	})
}

// DeleteAgentKey permanently deletes an agent key
// @Summary Delete agent key
// @Description Permanently delete an agent key from the system
// @Tags agent-keys
// @Param key path string true "Agent key to delete"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/agent-keys/{key} [delete]
func (h *AgentKeyHandler) DeleteAgentKey(c *gin.Context) {
	agentKey := c.Param("key")
	if agentKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Agent key is required",
		})
		return
	}

	err := h.agentKeyUsecase.DeleteAgentKey(c.Request.Context(), agentKey)
	if err != nil {
		if err.Error() == "agent key not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Agent key not found",
			})
			return
		}

		if strings.Contains(err.Error(), "linked agent is currently online") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Cannot delete agent key: linked agent is currently online",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete agent key",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Agent key deleted successfully",
		"agent_key": agentKey,
	})
}
