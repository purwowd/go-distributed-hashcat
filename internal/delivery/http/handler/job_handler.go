package handler

import (
	"net/http"
	"os"
	"sort"
	"strings"

	"go-distributed-hashcat/internal/domain"
	"go-distributed-hashcat/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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

	c.JSON(http.StatusOK, gin.H{"data": enrichedJobs})
}

func (h *JobHandler) StartJob(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	if err := h.jobUsecase.StartJob(c.Request.Context(), id); err != nil {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.jobUsecase.UpdateJobProgress(c.Request.Context(), id, req.Progress, req.Speed); err != nil {
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

	if err := h.jobUsecase.CompleteJob(c.Request.Context(), id, req.Result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast job completion with result
	Hub.BroadcastJobStatus(id.String(), "completed", req.Result)

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

	if err := h.jobUsecase.FailJob(c.Request.Context(), id, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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

	// Baca konten wordlist
	wordlistLines, err := readWordlistFile(wordlist.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read wordlist"})
		return
	}

	// Analisis kecepatan agent berdasarkan resource
	type AgentSpeed struct {
		AgentID string
		Speed   int
	}
	var agentSpeeds []AgentSpeed
	for _, agent := range agents {
		speed := 1 // Default untuk CPU
		if strings.Contains(agent.Capabilities, "GPU") { // Ganti dengan field yang sesuai
			speed = 5
		}
		agentSpeeds = append(agentSpeeds, AgentSpeed{AgentID: agent.ID.String(), Speed: speed})
	}

	// Urutkan agent berdasarkan kecepatan (descending)
	sort.Slice(agentSpeeds, func(i, j int) bool {
		return agentSpeeds[i].Speed > agentSpeeds[j].Speed
	})

	// Hitung total bobot kecepatan
	totalSpeed := 0
	for _, agentSpeed := range agentSpeeds {
		totalSpeed += agentSpeed.Speed
	}

	// Bagi wordlist berdasarkan bobot kecepatan
	wordlistParts := make(map[string][]string)
	currentIndex := 0
	for _, agentSpeed := range agentSpeeds {
		partSize := len(wordlistLines) * agentSpeed.Speed / totalSpeed
		wordlistParts[agentSpeed.AgentID] = wordlistLines[currentIndex : currentIndex+partSize]
		currentIndex += partSize
	}

	// Buat job untuk setiap bagian
	for agentID, words := range wordlistParts {
		_, err := h.jobUsecase.CreateJob(c.Request.Context(), &domain.CreateJobRequest{
			HashFileID: request.HashFileID,
			Wordlist:   strings.Join(words, "\n"), // Gabungkan array menjadi string
			AgentID:    agentID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Jobs created successfully"})
}

// Fungsi pembantu untuk membaca file wordlist
func readWordlistFile(path string) ([]string, error) {
    content, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    return strings.Split(string(content), "\n"), nil
}
