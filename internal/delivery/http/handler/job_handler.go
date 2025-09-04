package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"go-distributed-hashcat/internal/domain"
	"go-distributed-hashcat/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// defaultString returns fallback if s is empty
func defaultString(s string, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

type JobHandler struct {
	jobUsecase        usecase.JobUsecase
	enrichmentService usecase.JobEnrichmentService
	agentUsecase      usecase.AgentUsecase
	wordlistUsecase   usecase.WordlistUsecase
}

func NewJobHandler(jobUsecase usecase.JobUsecase, enrichmentService usecase.JobEnrichmentService, agentUsecase usecase.AgentUsecase, wordlistUsecase usecase.WordlistUsecase) *JobHandler {
	return &JobHandler{
		jobUsecase:        jobUsecase,
		enrichmentService: enrichmentService,
		agentUsecase:      agentUsecase,
		wordlistUsecase:   wordlistUsecase,
	}
}

func (h *JobHandler) CreateJob(c *gin.Context) {
	var req domain.CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job, err := h.jobUsecase.CreateJob(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": job})
}

func (h *JobHandler) GetJob(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	job, err := h.jobUsecase.GetJob(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": job})
}

func (h *JobHandler) GetAllJobs(c *gin.Context) {
	status := c.Query("status")

	var jobs []domain.Job
	var err error

	if status != "" {
		jobs, err = h.jobUsecase.GetJobsByStatus(c.Request.Context(), status)
	} else {
		jobs, err = h.jobUsecase.GetAllJobs(c.Request.Context())
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Enrich jobs with readable names using service
	enrichedJobs, err := h.enrichmentService.EnrichJobs(c.Request.Context(), jobs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Normalize response to ensure all fields are populated with sensible defaults
	normalized := make([]gin.H, 0, len(enrichedJobs))
	for _, ej := range enrichedJobs {
		var (
			hashFileID     string
			wordlistID     string
			agentID        string
			etaStr         string
			startedAtStr   string
			completedAtStr string
		)

		if ej.HashFileID != nil {
			hashFileID = ej.HashFileID.String()
		} else {
			hashFileID = ""
		}
		if ej.WordlistID != nil {
			wordlistID = ej.WordlistID.String()
		} else {
			wordlistID = ""
		}
		if ej.AgentID != nil {
			agentID = ej.AgentID.String()
		} else {
			agentID = ""
		}
		if ej.ETA != nil {
			etaStr = ej.ETA.Format(time.RFC3339)
		} else {
			etaStr = ""
		}
		if ej.StartedAt != nil {
			startedAtStr = ej.StartedAt.Format(time.RFC3339)
		} else {
			startedAtStr = ""
		}
		if ej.CompletedAt != nil {
			completedAtStr = ej.CompletedAt.Format(time.RFC3339)
		} else {
			completedAtStr = ""
		}

		normalized = append(normalized, gin.H{
			"id":             ej.ID.String(),
			"name":           defaultString(ej.Name, "-"),
			"status":         defaultString(ej.Status, "pending"),
			"hash_type":      ej.HashType,
			"attack_mode":    ej.AttackMode,
			"hash_file":      defaultString(ej.HashFile, ""),
			"hash_file_id":   hashFileID,
			"hash_file_name": defaultString(ej.HashFileName, "-"),
			"wordlist":       defaultString(ej.Wordlist, ""),
			"wordlist_id":    wordlistID,
			"wordlist_name":  defaultString(ej.WordlistName, "-"),
			"rules":          defaultString(ej.Rules, "-"),
			"agent_id":       agentID,
			"agent_name":     defaultString(ej.AgentName, "Unassigned"),
			"progress":       ej.Progress,
			"speed":          ej.Speed,
			"eta":            etaStr,
			"result":         defaultString(ej.Result, "-"),
			"created_at":     ej.CreatedAt.Format(time.RFC3339),
			"updated_at":     ej.UpdatedAt.Format(time.RFC3339),
			"started_at":     startedAtStr,
			"completed_at":   completedAtStr,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": normalized})
}

func (h *JobHandler) StartJob(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	if err := h.jobUsecase.StartJob(c.Request.Context(), id); err != nil {
		// Add detailed error logging
		log.Printf("❌ Failed to start job %s: %v", id.String(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast job status change
	Hub.BroadcastJobStatus(id.String(), "running", "")

	c.JSON(http.StatusOK, gin.H{"message": "Job started successfully"})
}

func (h *JobHandler) UpdateJobProgress(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	var req struct {
		Progress float64 `json:"progress" binding:"required"`
		Speed    int64   `json:"speed" binding:"required"`
		ETA      *string `json:"eta,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid request body for job progress update %s: %v", id.String(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.jobUsecase.UpdateJobProgress(c.Request.Context(), id, req.Progress, req.Speed); err != nil {
		log.Printf("Failed to update job progress %s: %v", id.String(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast real-time progress update
	eta := ""
	if req.ETA != nil {
		eta = *req.ETA
	}
	Hub.BroadcastJobProgress(id.String(), req.Progress, req.Speed, eta, "running")

	c.JSON(http.StatusOK, gin.H{"message": "Job progress updated successfully"})
}

func (h *JobHandler) CompleteJob(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	var req struct {
		Result string `json:"result"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get job details before completion
	job, err := h.jobUsecase.GetJob(c.Request.Context(), id)
	if err != nil {
		log.Printf("Failed to get job %s for completion logging: %v", id.String(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get agent details
	var agentName string
	if job.AgentID != nil {
		agent, err := h.agentUsecase.GetAgent(c.Request.Context(), *job.AgentID)
		if err != nil {
			log.Printf("Failed to get agent details for job %s: %v", id.String(), err)
			agentName = "Unknown Agent"
		} else {
			agentName = agent.Name
		}
	} else {
		agentName = "Unassigned Agent"
	}

	// Log job completion with agent details
	if req.Result != "" && req.Result != "Password not found - exhausted" {
		log.Printf("🎯 PASSWORD FOUND: Agent %s found password for job %s (Status: FAILED)", agentName, job.Name)
		log.Printf("   Result: %s", req.Result)
		log.Printf("   Job ID: %s", job.ID.String())
		log.Printf("   ⚡ Speed: %d H/s", job.Speed)
		log.Printf("   Progress: %.2f%%", job.Progress)

		// Check if this is a distributed job and log coordination info
		if strings.Contains(job.Name, " (Part ") || strings.Contains(job.Name, " (") {
			log.Printf("COORDINATION: This is a distributed job - stopping other agents...")
		}
	} else {
		log.Printf("✅ COMPLETED: Agent %s completed job %s (no password found)", agentName, job.Name)
		log.Printf("   🔍 Job ID: %s", job.ID.String())
		log.Printf("   ⚡ Speed: %d H/s", job.Speed)
		log.Printf("   📊 Progress: %.2f%%", job.Progress)
		if req.Result != "" {
			log.Printf("   📝 Reason: %s", req.Result)
		}
	}

	if err := h.jobUsecase.CompleteJob(c.Request.Context(), id, req.Result, job.Speed); err != nil {
		log.Printf("Failed to complete job %s: %v", id.String(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update agent status to online
	if job.AgentID != nil {
		if err := h.agentUsecase.UpdateAgentStatus(c.Request.Context(), *job.AgentID, "online"); err != nil {
			log.Printf("⚠️ Failed to update agent status to online for agent %s: %v", agentName, err)
		} else {
			log.Printf("✅ Successfully updated agent %s status to online", agentName)
		}
	}

	// Broadcast job completion with result - status will be determined by the usecase
	// Get the updated job to get the correct status
	updatedJob, err := h.jobUsecase.GetJob(c.Request.Context(), id)
	if err != nil {
		log.Printf("Failed to get updated job status for broadcasting: %v", err)
		// Fallback to completed status
		Hub.BroadcastJobStatus(id.String(), "completed", req.Result)
	} else {
		Hub.BroadcastJobStatus(id.String(), updatedJob.Status, req.Result)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job completed successfully"})
}

func (h *JobHandler) FailJob(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get job details before failure
	job, err := h.jobUsecase.GetJob(c.Request.Context(), id)
	if err != nil {
		log.Printf("❌ Failed to get job %s for failure logging: %v", id.String(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get agent details
	var agentName string
	if job.AgentID != nil {
		agent, err := h.agentUsecase.GetAgent(c.Request.Context(), *job.AgentID)
		if err != nil {
			log.Printf("⚠️ Failed to get agent details for job %s: %v", id.String(), err)
			agentName = "Unknown Agent"
		} else {
			agentName = agent.Name
		}
	} else {
		agentName = "Unassigned Agent"
	}

	// Log job failure with agent details
	log.Printf("💥 FAILED: Agent %s failed job %s", agentName, job.Name)
	log.Printf("   🔍 Job ID: %s", job.ID.String())
	log.Printf("   ⚡ Speed: %d H/s", job.Speed)
	log.Printf("   📊 Progress: %.2f%%", job.Progress)
	log.Printf("   📝 Reason: %s", req.Reason)

	if err := h.jobUsecase.FailJob(c.Request.Context(), id, req.Reason); err != nil {
		log.Printf("❌ Failed to mark job %s as failed: %v", id.String(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update agent status to online
	if job.AgentID != nil {
		if err := h.agentUsecase.UpdateAgentStatus(c.Request.Context(), *job.AgentID, "online"); err != nil {
			log.Printf("⚠️ Failed to update agent status to online for agent %s: %v", agentName, err)
		} else {
			log.Printf("✅ Successfully updated agent %s status to online", agentName)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job failed successfully"})
}

func (h *JobHandler) PauseJob(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	if err := h.jobUsecase.PauseJob(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast job status change
	Hub.BroadcastJobStatus(id.String(), "paused", "")

	c.JSON(http.StatusOK, gin.H{"message": "Job paused successfully"})
}

func (h *JobHandler) ResumeJob(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	if err := h.jobUsecase.ResumeJob(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast job status change
	Hub.BroadcastJobStatus(id.String(), "pending", "")

	c.JSON(http.StatusOK, gin.H{"message": "Job resumed successfully"})
}

func (h *JobHandler) StopJob(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	// Stop job by setting status to failed with stopped reason
	if err := h.jobUsecase.FailJob(c.Request.Context(), id, "Job stopped by user"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast job status change
	Hub.BroadcastJobStatus(id.String(), "failed", "Job stopped by user")

	c.JSON(http.StatusOK, gin.H{"message": "Job stopped successfully"})
}

func (h *JobHandler) DeleteJob(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	if err := h.jobUsecase.DeleteJob(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job deleted successfully"})
}

func (h *JobHandler) AssignJobs(c *gin.Context) {
	if err := h.jobUsecase.AssignJobsToAgents(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Jobs assigned to agents successfully"})
}

// GetJobsByAgentID gets all jobs assigned to a specific agent
func (h *JobHandler) GetJobsByAgentID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	jobs, err := h.jobUsecase.GetJobsByAgentID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": jobs})
}

// GetAvailableJobForAgent gets the next available job for an agent to execute
func (h *JobHandler) GetAvailableJobForAgent(c *gin.Context) {
	idStr := c.Param("id")
	agentID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	job, err := h.jobUsecase.GetAvailableJobForAgent(c.Request.Context(), agentID)
	if err != nil {
		// No available job is not an error
		c.JSON(http.StatusOK, gin.H{"data": nil, "message": "No available jobs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": job})
}

func (h *JobHandler) CreateParallelJobs(c *gin.Context) {
	// Ambil hashfile dan wordlist dari request
	var request struct {
		HashFileID string `json:"hash_file_id"`
		WordlistID string `json:"wordlist_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Ambil daftar agent
	agents, err := h.agentUsecase.GetAllAgents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agents"})
		return
	}

	// Filter hanya agent yang online
	var onlineAgents []domain.Agent
	for _, agent := range agents {
		if agent.Status == "online" {
			onlineAgents = append(onlineAgents, agent)
		}
	}

	if len(onlineAgents) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No online agents available"})
		return
	}

	log.Printf("Starting parallel job creation with %d online agents", len(onlineAgents))

	// Ambil detail wordlist
	wordlistID, err := uuid.Parse(request.WordlistID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wordlist ID"})
		return
	}
	wordlist, err := h.wordlistUsecase.GetWordlist(c.Request.Context(), wordlistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch wordlist"})
		return
	}

	// Get word count from repository
	var totalWords int64
	if wordlist.WordCount != nil {
		totalWords = *wordlist.WordCount
		log.Printf("📝 Wordlist contains %d words (from repository)", totalWords)
	} else {
		// Fallback: read file content if word count not available
		wordlistLines, err := readWordlistFile(wordlist.Path)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read wordlist"})
			return
		}

		// Filter non-empty words
		var validWords []string
		for _, word := range wordlistLines {
			word = strings.TrimSpace(word)
			if word != "" {
				validWords = append(validWords, word)
			}
		}
		totalWords = int64(len(validWords))
		log.Printf("📝 Wordlist contains %d valid words (from file)", totalWords)
	}

	// Analisis kecepatan agent berdasarkan capabilities
	type AgentSpeed struct {
		Agent  domain.Agent
		Speed  int
		Weight float64
	}
	var agentSpeeds []AgentSpeed

	for _, agent := range onlineAgents {
		// Use actual speed from database, fallback to capability-based estimation if speed is 0
		speed := agent.Speed
		if speed == 0 {
			// Fallback to capability-based estimation for agents without speed data
			speed = 1 // Default for CPU
			if strings.Contains(strings.ToLower(agent.Capabilities), "gpu") {
				speed = 5 // GPU is faster
			} else if strings.Contains(strings.ToLower(agent.Capabilities), "rtx") {
				speed = 8 // RTX lebih cepat lagi
			} else if strings.Contains(strings.ToLower(agent.Capabilities), "gtx") {
				speed = 6 // GTX lebih cepat
			}
		}

		agentSpeeds = append(agentSpeeds, AgentSpeed{
			Agent:  agent,
			Speed:  int(speed),
			Weight: 0, // Akan dihitung nanti
		})
	}

	// Hitung total bobot kecepatan
	totalSpeed := 0
	for _, agentSpeed := range agentSpeeds {
		totalSpeed += agentSpeed.Speed
	}

	// Sort agents by speed (highest first) for proper distribution
	for i := 0; i < len(agentSpeeds)-1; i++ {
		for j := i + 1; j < len(agentSpeeds); j++ {
			if agentSpeeds[i].Speed < agentSpeeds[j].Speed {
				agentSpeeds[i], agentSpeeds[j] = agentSpeeds[j], agentSpeeds[i]
			}
		}
	}

	// Calculate weight for each agent
	for i := range agentSpeeds {
		agentSpeeds[i].Weight = float64(agentSpeeds[i].Speed) / float64(totalSpeed)
	}

	// Log distribusi agent
	log.Printf("🤖 Agent distribution plan:")
	for _, agentSpeed := range agentSpeeds {
		wordCount := int64(float64(totalWords) * agentSpeed.Weight)
		log.Printf("   - %s (%s): Speed=%d, Weight=%.2f, Words=%d",
			agentSpeed.Agent.Name,
			agentSpeed.Agent.Capabilities,
			agentSpeed.Speed,
			agentSpeed.Weight,
			wordCount)
	}

	// Create jobs with skip/limit parameters based on agent performance
	var createdJobs []domain.Job
	currentSkip := int64(0)

	for i, agentSpeed := range agentSpeeds {
		// Calculate word count for this agent
		wordCount := int64(float64(totalWords) * agentSpeed.Weight)

		// Ensure last agent gets remaining words
		if i == len(agentSpeeds)-1 {
			wordCount = totalWords - currentSkip
		}

		// Ensure minimum words per agent
		if wordCount < 1 {
			wordCount = 1
		}

		// Calculate skip and limit values for this agent
		skip := currentSkip
		limit := wordCount

		// Create job with skip/limit parameters
		job, err := h.jobUsecase.CreateJob(c.Request.Context(), &domain.CreateJobRequest{
			HashFileID: request.HashFileID,
			WordlistID: request.WordlistID,
			Wordlist:   wordlist.OrigName,                      // Use original wordlist name
			AgentIDs:   []string{agentSpeed.Agent.ID.String()}, // Use AgentIDs for distributed job
			Name:       fmt.Sprintf("Parallel Job - %s (%s)", wordlist.Name, agentSpeed.Agent.Name),
		})
		if err != nil {
			log.Printf("Failed to create job for agent %s: %v", agentSpeed.Agent.ID.String(), err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job"})
			return
		}

		createdJobs = append(createdJobs, *job)
		log.Printf("📦 Created job %s for agent %s with skip=%d, limit=%d words (%.1f%%)",
			job.ID.String(),
			agentSpeed.Agent.Name,
			skip,
			limit,
			agentSpeed.Weight*100)

		// Update currentSkip for next agent
		currentSkip += wordCount
	}

	log.Printf("Successfully created %d parallel jobs", len(createdJobs))

	c.JSON(http.StatusOK, gin.H{
		"message": "Parallel jobs created successfully",
		"data": gin.H{
			"total_jobs":  len(createdJobs),
			"total_words": totalWords,
			"agents_used": len(agentSpeeds),
			"jobs":        createdJobs,
		},
	})
}

// UpdateJobDataFromAgent receives complete job data from agent and updates database immediately
func (h *JobHandler) UpdateJobDataFromAgent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	var req struct {
		AgentID    string  `json:"agent_id" binding:"required"`
		AttackMode int     `json:"attack_mode"`
		Rules      string  `json:"rules"`
		Speed      int64   `json:"speed"`
		ETA        *string `json:"eta,omitempty"`
		Progress   float64 `json:"progress"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse agent ID
	agentID, err := uuid.Parse(req.AgentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	// Get the job first
	job, err := h.jobUsecase.GetJob(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Update job with data from agent (only update if provided)
	if req.AttackMode > 0 {
		job.AttackMode = req.AttackMode
	}
	if req.Rules != "" {
		job.Rules = req.Rules
	}
	job.Speed = req.Speed
	job.Progress = req.Progress

	// Update agent ID if not already set
	if job.AgentID == nil {
		job.AgentID = &agentID
	}

	// Parse ETA if provided
	if req.ETA != nil && *req.ETA != "" {
		if etaTime, err := time.Parse(time.RFC3339, *req.ETA); err == nil {
			job.ETA = &etaTime
		}
	}

	// Update the job in database immediately
	if err := h.jobUsecase.UpdateJobData(c.Request.Context(), job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast real-time update
	eta := ""
	if req.ETA != nil {
		eta = *req.ETA
	}
	Hub.BroadcastJobProgress(id.String(), req.Progress, req.Speed, eta, job.Status)

	c.JSON(http.StatusOK, gin.H{"message": "Job data updated successfully"})
}

// Helper function to read wordlist file
func readWordlistFile(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(content), "\n"), nil
}

// GetParallelJobsSummary returns summary of parallel jobs with agent results
func (h *JobHandler) GetParallelJobsSummary(c *gin.Context) {
	// Get all jobs with status completed or failed
	completedJobs, err := h.jobUsecase.GetJobsByStatus(c.Request.Context(), "completed")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get completed jobs"})
		return
	}

	failedJobs, err := h.jobUsecase.GetJobsByStatus(c.Request.Context(), "failed")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get failed jobs"})
		return
	}

	allJobs := append(completedJobs, failedJobs...)

	// Group jobs by name (parallel jobs have similar names)
	jobGroups := make(map[string][]domain.Job)
	for _, job := range allJobs {
		if strings.Contains(job.Name, "Parallel Job") {
			baseName := strings.Replace(job.Name, "Parallel Job - ", "", 1)
			jobGroups[baseName] = append(jobGroups[baseName], job)
		}
	}

	var summaries []gin.H
	for baseName, jobs := range jobGroups {
		if len(jobs) > 1 { // Only show parallel jobs
			var agentResults []gin.H
			var successCount, failureCount int
			var foundPassword string

			for _, job := range jobs {
				// Get agent details
				var agentName string
				if job.AgentID != nil {
					agent, err := h.agentUsecase.GetAgent(c.Request.Context(), *job.AgentID)
					if err != nil {
						agentName = "Unknown Agent"
					} else {
						agentName = agent.Name
					}
				} else {
					agentName = "Unassigned Agent"
				}

				// Determine result
				var result string
				var status string
				if job.Status == "completed" && job.Result != "" && job.Result != "Password not found - exhausted" {
					result = fmt.Sprintf("SUCCESS: Found password (%s)", job.Result)
					status = "success"
					successCount++
					foundPassword = job.Result
				} else if job.Status == "completed" {
					result = "FAILED: No password found"
					status = "failed"
					failureCount++
				} else {
					result = fmt.Sprintf("FAILED: %s", job.Result)
					status = "failed"
					failureCount++
				}

				agentResults = append(agentResults, gin.H{
					"agent_name":   agentName,
					"agent_id":     job.AgentID,
					"job_id":       job.ID.String(),
					"status":       status,
					"result":       result,
					"speed":        job.Speed,
					"progress":     job.Progress,
					"started_at":   job.StartedAt,
					"completed_at": job.CompletedAt,
				})
			}

			// Overall summary
			var overallResult string
			if successCount > 0 {
				overallResult = fmt.Sprintf("SUCCESS: Password found by %d agent(s) - %s", successCount, foundPassword)
			} else {
				overallResult = fmt.Sprintf("FAILED: No password found by any agent (%d agents tried)", failureCount)
			}

			summaries = append(summaries, gin.H{
				"wordlist_name":  baseName,
				"total_agents":   len(jobs),
				"success_count":  successCount,
				"failure_count":  failureCount,
				"overall_result": overallResult,
				"agent_results":  agentResults,
				"created_at":     jobs[0].CreatedAt,
			})
		}
	}

	// Log summary
	log.Printf("📊 Parallel Jobs Summary:")
	for _, summary := range summaries {
		log.Printf("   📋 Wordlist: %s", summary["wordlist_name"])
		log.Printf("   🎯 Overall: %s", summary["overall_result"])
		log.Printf("   🤖 Agents: %d total (%d success, %d failed)",
			summary["total_agents"],
			summary["success_count"],
			summary["failure_count"])

		agentResults := summary["agent_results"].([]gin.H)
		for _, agentResult := range agentResults {
			log.Printf("      - %s: %s",
				agentResult["agent_name"],
				agentResult["result"])
		}
		log.Printf("")
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Parallel jobs summary retrieved successfully",
		"data": gin.H{
			"total_parallel_jobs": len(summaries),
			"summaries":           summaries,
		},
	})
}
