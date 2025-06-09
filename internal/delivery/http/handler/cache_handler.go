package handler

import (
	"net/http"

	"go-distributed-hashcat/internal/usecase"

	"github.com/gin-gonic/gin"
)

type CacheHandler struct {
	enrichmentService usecase.JobEnrichmentService
}

func NewCacheHandler(enrichmentService usecase.JobEnrichmentService) *CacheHandler {
	return &CacheHandler{
		enrichmentService: enrichmentService,
	}
}

// GetCacheStats returns cache statistics
func (h *CacheHandler) GetCacheStats(c *gin.Context) {
	stats := h.enrichmentService.GetCacheStats()
	c.JSON(http.StatusOK, gin.H{
		"data":    stats,
		"message": "Cache statistics retrieved successfully",
	})
}

// ClearCache clears all cached data
func (h *CacheHandler) ClearCache(c *gin.Context) {
	h.enrichmentService.ClearCache()
	c.JSON(http.StatusOK, gin.H{
		"message": "Cache cleared successfully",
	})
}
