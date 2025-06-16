package middleware

import (
	"net/http"

	"go-distributed-hashcat/internal/usecase"

	"github.com/gin-gonic/gin"
)

// AgentKeyMiddleware validates X-Agent-Key header for agent endpoints
func AgentKeyMiddleware(agentKeyUsecase usecase.AgentKeyUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentKey := c.GetHeader("X-Agent-Key")

		// Check if key exists
		if agentKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing X-Agent-Key header",
				"code":  "MISSING_AGENT_KEY",
			})
			c.Abort()
			return
		}

		// Validate key
		agent, err := agentKeyUsecase.ValidateAgentKey(c.Request.Context(), agentKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
				"code":  "INVALID_AGENT_KEY",
			})
			c.Abort()
			return
		}

		// Store agent info in context for handlers to use
		c.Set("agent", agent)
		c.Set("agent_key", agentKey)

		c.Next()
	}
}
