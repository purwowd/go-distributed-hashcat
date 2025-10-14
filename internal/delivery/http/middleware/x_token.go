package middleware

import (
	"net/http"
	"strings"

	"go-distributed-hashcat/internal/infrastructure"
	"github.com/gin-gonic/gin"
)

func XTokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.GetHeader("X-Token"))
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing X-Token header"})
			c.Abort()
			return
		}

		global := infrastructure.GlobalXToken()
		if global == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "X-Token not configured"})
			c.Abort()
			return
		}

		if token != global {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid X-Token"})
			c.Abort()
			return
		}

		c.Next()
	}
}
