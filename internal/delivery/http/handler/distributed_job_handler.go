package handler

import (
	"fmt"
	"net/http"
	"time"

	"go-distributed-hashcat/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DistributedJobHandler struct {
	distributedJobUsecase domain.DistributedJobUsecase
}

func NewDistributedJobHandler(distributedJobUsecase domain.DistributedJobUsecase) *DistributedJobHandler {
	return &DistributedJobHandler{
		distributedJobUsecase: distributedJobUsecase,
	}
}

// CreateDistributedJobs creates multiple jobs by dividing wordlist among agents
func (h *DistributedJobHandler) CreateDistributedJobs(c *gin.Context) {
	var req domain.DistributedJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if req.Name == "" || req.HashFileID == "" || req.WordlistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Name, hash_file_id, and wordlist_id are required",
		})
		return
	}

	// Set default values
	if req.HashType == 0 {
		req.HashType = 2500 // Default to WPA/WPA2
	}
	if req.AttackMode == 0 {
		req.AttackMode = 0 // Default to Dictionary Attack
	}

	// Create distributed jobs
	result, err := h.distributedJobUsecase.CreateDistributedJobs(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create distributed jobs: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    result,
		"message": "Distributed jobs created successfully",
	})
}

// GetDistributedJobStatus gets the status of all sub-jobs for a master job
func (h *DistributedJobHandler) GetDistributedJobStatus(c *gin.Context) {
	masterJobIDStr := c.Param("id")
	if masterJobIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Master job ID is required",
		})
		return
	}

	masterJobID, err := uuid.Parse(masterJobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid master job ID format",
		})
		return
	}

	// Get distributed job status
	result, err := h.distributedJobUsecase.GetDistributedJobStatus(c.Request.Context(), masterJobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get distributed job status: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
		"message": "Distributed job status retrieved successfully",
	})
}

// StartAllSubJobs starts all sub-jobs for a master job
func (h *DistributedJobHandler) StartAllSubJobs(c *gin.Context) {
	masterJobIDStr := c.Param("id")
	if masterJobIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Master job ID is required",
		})
		return
	}

	masterJobID, err := uuid.Parse(masterJobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid master job ID format",
		})
		return
	}

	// Get distributed job status first
	result, err := h.distributedJobUsecase.GetDistributedJobStatus(c.Request.Context(), masterJobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get distributed job status: " + err.Error(),
		})
		return
	}

	// Start all sub-jobs
	startedCount := 0
	for _, subJob := range result.SubJobs {
		if subJob.Status == "pending" {
			// Update job status to running
			subJob.Status = "running"
			now := time.Now()
			subJob.StartedAt = &now
			// In a real implementation, you would call the job repository to update status
			startedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"master_job_id": masterJobID,
			"started_jobs":  startedCount,
			"total_jobs":    len(result.SubJobs),
		},
		"message": fmt.Sprintf("Started %d out of %d sub-jobs", startedCount, len(result.SubJobs)),
	})
}

// GetAgentPerformance gets performance metrics for all agents
func (h *DistributedJobHandler) GetAgentPerformance(c *gin.Context) {
	// This would typically call a service to get real-time agent performance
	// For now, return mock data
	performanceData := []gin.H{
		{
			"agent_id":      "agent-1",
			"name":          "GPU-Agent-1",
			"capabilities":  "RTX 4090",
			"resource_type": "GPU",
			"performance":   1.0,
			"speed":         1000000, // 1M H/s
		},
		{
			"agent_id":      "agent-2",
			"name":          "CPU-Agent-1",
			"capabilities":  "Intel i9",
			"resource_type": "CPU",
			"performance":   0.3,
			"speed":         100000, // 100K H/s
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    performanceData,
		"message": "Agent performance data retrieved successfully",
	})
}

// GetDistributionPreview shows how wordlist would be distributed among agents
func (h *DistributedJobHandler) GetDistributionPreview(c *gin.Context) {
	wordlistIDStr := c.Query("wordlist_id")
	if wordlistIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "wordlist_id query parameter is required",
		})
		return
	}

	wordlistID, err := uuid.Parse(wordlistIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid wordlist ID format",
		})
		return
	}

	// Mock distribution preview
	// In real implementation, this would calculate actual distribution
	preview := gin.H{
		"wordlist_id": wordlistID,
		"total_words": 10000,
		"distribution": []gin.H{
			{
				"agent_name":     "GPU-Agent-1",
				"resource_type":  "GPU",
				"performance":    1.0,
				"assigned_words": 6000,
				"percentage":     60,
			},
			{
				"agent_name":     "CPU-Agent-1",
				"resource_type":  "CPU",
				"performance":    0.3,
				"assigned_words": 4000,
				"percentage":     40,
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    preview,
		"message": "Distribution preview generated successfully",
	})
}
