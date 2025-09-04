package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-distributed-hashcat/internal/domain"

	"github.com/google/uuid"
)

type JobUsecase interface {
	CreateJob(ctx context.Context, req *domain.CreateJobRequest) (*domain.Job, error)
	GetJob(ctx context.Context, id uuid.UUID) (*domain.Job, error)
	GetAllJobs(ctx context.Context) ([]domain.Job, error)
	GetJobsByStatus(ctx context.Context, status string) ([]domain.Job, error)
	GetJobsByAgentID(ctx context.Context, agentID uuid.UUID) ([]domain.Job, error)
	GetAvailableJobForAgent(ctx context.Context, agentID uuid.UUID) (*domain.Job, error)
	StartJob(ctx context.Context, id uuid.UUID) error
	UpdateJobProgress(ctx context.Context, id uuid.UUID, progress float64, speed int64) error
	UpdateJobData(ctx context.Context, job *domain.Job) error
	CompleteJob(ctx context.Context, id uuid.UUID, result string, speed int64) error
	FailJob(ctx context.Context, id uuid.UUID, reason string) error
	PauseJob(ctx context.Context, id uuid.UUID) error
	ResumeJob(ctx context.Context, id uuid.UUID) error
	DeleteJob(ctx context.Context, id uuid.UUID) error
	AssignJobsToAgents(ctx context.Context) error
}

type jobUsecase struct {
	jobRepo      domain.JobRepository
	agentRepo    domain.AgentRepository
	hashFileRepo domain.HashFileRepository
	wordlistRepo domain.WordlistRepository
}

func NewJobUsecase(jobRepo domain.JobRepository, agentRepo domain.AgentRepository, hashFileRepo domain.HashFileRepository, wordlistRepo domain.WordlistRepository) JobUsecase {
	return &jobUsecase{
		jobRepo:      jobRepo,
		agentRepo:    agentRepo,
		hashFileRepo: hashFileRepo,
		wordlistRepo: wordlistRepo,
	}
}

func (u *jobUsecase) CreateJob(ctx context.Context, req *domain.CreateJobRequest) (*domain.Job, error) {
	// Validate hash file exists
	hashFileID, err := uuid.Parse(req.HashFileID)
	if err != nil {
		return nil, fmt.Errorf("invalid hash file ID: %w", err)
	}

	hashFile, err := u.hashFileRepo.GetByID(ctx, hashFileID)
	if err != nil {
		return nil, fmt.Errorf("hash file not found: %w", err)
	}

	job := &domain.Job{
		ID:             uuid.New(),
		Name:           req.Name,
		Status:         "pending",
		HashType:       req.HashType,
		AttackMode:     req.AttackMode,
		HashFile:       hashFile.Path,
		HashFileID:     &hashFileID,
		Wordlist:       req.Wordlist,
		Rules:          req.Rules,
		Progress:       0,
		Speed:          0,
		TotalWords:     req.TotalWords,
		ProcessedWords: 0,
	}

	// Handle wordlist ID if provided
	var wordlistID *uuid.UUID
	if req.WordlistID != "" {
		parsedWordlistID, err := uuid.Parse(req.WordlistID)
		if err != nil {
			return nil, fmt.Errorf("invalid wordlist ID: %w", err)
		}
		wordlistID = &parsedWordlistID
		job.WordlistID = wordlistID
	}

	// Handle agent assignment (single or multiple)
	if len(req.AgentIDs) > 0 {
		// Multiple agent assignment for distributed jobs
		agentIDs := make([]uuid.UUID, 0, len(req.AgentIDs))

		for _, agentIDStr := range req.AgentIDs {
			agentID, err := uuid.Parse(agentIDStr)
			if err != nil {
				return nil, fmt.Errorf("invalid agent ID %s: %w", agentIDStr, err)
			}

			// Verify agent exists and is available
			agent, err := u.agentRepo.GetByID(ctx, agentID)
			if err != nil {
				return nil, fmt.Errorf("agent not found: %w", err)
			}

			if agent.Status != "online" {
				return nil, fmt.Errorf("agent %s is not available (status: %s)", agent.Name, agent.Status)
			}

			agentIDs = append(agentIDs, agentID)
		}

		// Create separate job for each agent (distributed job creation)
		if len(agentIDs) > 1 {
			// Get wordlist details for distribution
			var totalWords int64

			if req.WordlistID != "" {
				wordlistID, err := uuid.Parse(req.WordlistID)
				if err == nil {
					// Get wordlist from repository to get accurate word count
					wordlist, err := u.wordlistRepo.GetByID(ctx, wordlistID)
					if err == nil && wordlist.WordCount != nil {
						totalWords = *wordlist.WordCount
					} else {
						// Fallback: parse wordlist content from request if repository fails
						if req.Wordlist != "" {
							wordlistContent := strings.Split(req.Wordlist, "\n")
							// Filter empty lines
							var validWords []string
							for _, word := range wordlistContent {
								word = strings.TrimSpace(word)
								if word != "" {
									validWords = append(validWords, word)
								}
							}
							totalWords = int64(len(validWords))
						}
					}
				}
			}

			// Calculate agent performance scores (similar to frontend)
			type AgentPerformance struct {
				AgentID      uuid.UUID
				Name         string
				Capabilities string
				Speed        int
				Weight       float64
			}

			var agentPerformances []AgentPerformance
			totalSpeed := 0

			for _, agentID := range agentIDs {
				agent, err := u.agentRepo.GetByID(ctx, agentID)
				if err != nil {
					continue
				}

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

				totalSpeed += int(speed)

				agentPerformances = append(agentPerformances, AgentPerformance{
					AgentID:      agent.ID,
					Name:         agent.Name,
					Capabilities: agent.Capabilities,
					Speed:        int(speed),
					Weight:       0, // Will be calculated below
				})
			}

			// Sort agents by speed (highest first) for proper distribution
			for i := 0; i < len(agentPerformances)-1; i++ {
				for j := i + 1; j < len(agentPerformances); j++ {
					if agentPerformances[i].Speed < agentPerformances[j].Speed {
						agentPerformances[i], agentPerformances[j] = agentPerformances[j], agentPerformances[i]
					}
				}
			}

			// Calculate weights
			for i := range agentPerformances {
				agentPerformances[i].Weight = float64(agentPerformances[i].Speed) / float64(totalSpeed)
			}

			// Create sub-jobs for each agent with skip/limit parameters
			var subJobs []*domain.Job
			currentSkip := int64(0)

			for i, agentPerf := range agentPerformances {
				// Calculate word count for this agent
				wordCount := int64(float64(totalWords) * agentPerf.Weight)

				// Ensure last agent gets remaining words
				if i == len(agentPerformances)-1 {
					wordCount = totalWords - currentSkip
				}

				// Ensure minimum words per agent
				if wordCount < 1 {
					wordCount = 1
				}

				// Calculate skip and limit values for this agent
				skip := currentSkip
				limit := wordCount

				subJob := &domain.Job{
					ID:             uuid.New(),
					Name:           fmt.Sprintf("%s (%s)", req.Name, agentPerf.Name),
					Status:         "pending",
					HashType:       req.HashType,
					AttackMode:     req.AttackMode,
					HashFile:       hashFile.Path,
					HashFileID:     &hashFileID,
					Wordlist:       req.Wordlist, // Use original wordlist for all agents
					WordlistID:     wordlistID,   // Reference to original wordlist
					Rules:          req.Rules,
					Progress:       0,
					Speed:          0,
					TotalWords:     wordCount,
					ProcessedWords: 0,
					AgentID:        &agentPerf.AgentID,
					Skip:           &skip,  // Hashcat --skip parameter
					WordLimit:      &limit, // Hashcat --limit parameter
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
				}

				// Save sub-job
				if err := u.jobRepo.Create(ctx, subJob); err != nil {
					return nil, fmt.Errorf("failed to create sub-job %d: %w", i, err)
				}

				// Auto-start the job after creation
				if err := u.StartJob(ctx, subJob.ID); err != nil {
					fmt.Printf("Warning: failed to auto-start job %s: %v\n", subJob.Name, err)
				} else {
					fmt.Printf("✅ Auto-started job \"%s\" for agent %s\n", subJob.Name, agentPerf.Name)
				}

				subJobs = append(subJobs, subJob)

				// Log distribution info
				fmt.Printf("✅ Created job \"%s\" for agent %s with skip=%d, limit=%d words (%.1f%%)\n",
					subJob.Name, agentPerf.Name, skip, limit, agentPerf.Weight*100)

				// Update currentSkip for next agent
				currentSkip += wordCount
			}

			// Return the first sub-job as the primary result
			// Other sub-jobs are created but not returned
			return subJobs[0], nil
		} else {
			// Single agent assignment
			job.AgentID = &agentIDs[0]
		}

	} else if req.AgentID != "" {
		// Single agent assignment (legacy)
		agentID, err := uuid.Parse(req.AgentID)
		if err != nil {
			return nil, fmt.Errorf("invalid agent ID: %w", err)
		}

		// Verify agent exists and is available
		agent, err := u.agentRepo.GetByID(ctx, agentID)
		if err != nil {
			return nil, fmt.Errorf("agent not found: %w", err)
		}

		if agent.Status != "online" {
			return nil, fmt.Errorf("agent %s is not available (status: %s)", agent.Name, agent.Status)
		}

		job.AgentID = &agentID
	} else {
		// No agent assigned - job will be in "unassigned" state
		// This is valid for job queuing systems
	}

	if err := u.jobRepo.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	// Auto-start the job if it has an agent assigned
	if job.AgentID != nil {
		if err := u.StartJob(ctx, job.ID); err != nil {
			fmt.Printf("Warning: failed to auto-start job %s: %v\n", job.Name, err)
		} else {
			fmt.Printf("✅ Auto-started job \"%s\"\n", job.Name)
		}
	}

	return job, nil
}

func (u *jobUsecase) GetJob(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	job, err := u.jobRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	return job, nil
}

func (u *jobUsecase) GetAllJobs(ctx context.Context) ([]domain.Job, error) {
	jobs, err := u.jobRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs: %w", err)
	}
	return jobs, nil
}

func (u *jobUsecase) GetJobsByStatus(ctx context.Context, status string) ([]domain.Job, error) {
	jobs, err := u.jobRepo.GetByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by status: %w", err)
	}
	return jobs, nil
}

func (u *jobUsecase) GetJobsByAgentID(ctx context.Context, agentID uuid.UUID) ([]domain.Job, error) {
	jobs, err := u.jobRepo.GetByAgentID(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by agent ID: %w", err)
	}
	return jobs, nil
}

func (u *jobUsecase) GetAvailableJobForAgent(ctx context.Context, agentID uuid.UUID) (*domain.Job, error) {
	job, err := u.jobRepo.GetAvailableJobForAgent(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get available job for agent: %w", err)
	}
	return job, nil
}

func (u *jobUsecase) StartJob(ctx context.Context, id uuid.UUID) error {
	job, err := u.jobRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	if job.Status != "pending" && job.Status != "paused" {
		return fmt.Errorf("job cannot be started from status: %s", job.Status)
	}

	now := time.Now()
	job.Status = "running"
	job.StartedAt = &now

	if err := u.jobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	return nil
}

func (u *jobUsecase) UpdateJobProgress(ctx context.Context, id uuid.UUID, progress float64, speed int64) error {
	if err := u.jobRepo.UpdateProgress(ctx, id, progress, speed); err != nil {
		return fmt.Errorf("failed to update job progress: %w", err)
	}
	return nil
}

func (u *jobUsecase) UpdateJobData(ctx context.Context, job *domain.Job) error {
	if err := u.jobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update job data: %w", err)
	}
	return nil
}

func (u *jobUsecase) CompleteJob(ctx context.Context, id uuid.UUID, result string, speed int64) error {
	job, err := u.jobRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	now := time.Now()
	job.Speed = speed
	job.Progress = 100
	job.CompletedAt = &now

	// Check if password was found
	if result != "" && result != "Password not found - exhausted" {
		// Password found! Set status to failed and stop other related running jobs
		job.Status = "failed"
		job.Result = result

		// Stop other related running jobs
		if err := u.stopRelatedRunningJobs(ctx, job); err != nil {
			// Log error but don't fail the job completion
			fmt.Printf("Warning: failed to stop related running jobs: %v\n", err)
		}
	} else {
		// Password not found - set status to completed
		job.Status = "completed"
		job.Result = result
	}

	if err := u.jobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	return nil
}

func (u *jobUsecase) FailJob(ctx context.Context, id uuid.UUID, reason string) error {
	job, err := u.jobRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	now := time.Now()
	job.Status = "failed"
	job.Result = reason
	job.CompletedAt = &now

	// If the failure reason indicates password not found, set progress to 100%
	if reason == "Password not found" || strings.Contains(reason, "Password not found") {
		job.Progress = 100.0
	}

	if err := u.jobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	// Update agent status to online
	if job.AgentID != nil {
		if err := u.agentRepo.UpdateStatus(ctx, *job.AgentID, "online"); err != nil {
			// Log error but don't fail the job failure
			fmt.Printf("Warning: failed to update agent status: %v\n", err)
		}
	}

	return nil
}

func (u *jobUsecase) PauseJob(ctx context.Context, id uuid.UUID) error {
	if err := u.jobRepo.UpdateStatus(ctx, id, "paused"); err != nil {
		return fmt.Errorf("failed to pause job: %w", err)
	}
	return nil
}

func (u *jobUsecase) ResumeJob(ctx context.Context, id uuid.UUID) error {
	if err := u.jobRepo.UpdateStatus(ctx, id, "pending"); err != nil {
		return fmt.Errorf("failed to resume job: %w", err)
	}
	return nil
}

func (u *jobUsecase) DeleteJob(ctx context.Context, id uuid.UUID) error {
	if err := u.jobRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}
	return nil
}

func (u *jobUsecase) AssignJobsToAgents(ctx context.Context) error {
	// Get pending jobs
	pendingJobs, err := u.jobRepo.GetByStatus(ctx, "pending")
	if err != nil {
		return fmt.Errorf("failed to get pending jobs: %w", err)
	}

	if len(pendingJobs) == 0 {
		return nil // Nothing to assign
	}

	// Get available agents
	agents, err := u.agentRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get agents: %w", err)
	}

	var availableAgents []domain.Agent
	for _, agent := range agents {
		if agent.Status == "online" {
			availableAgents = append(availableAgents, agent)
		}
	}

	if len(availableAgents) == 0 {
		return nil // No available agents
	}

	// Filter jobs that need assignment (don't have AgentID yet)
	var jobsNeedingAssignment []domain.Job
	for _, job := range pendingJobs {
		if job.AgentID == nil {
			jobsNeedingAssignment = append(jobsNeedingAssignment, job)
		}
	}

	// Assign jobs to agents (round-robin)
	for i, job := range jobsNeedingAssignment {
		if i >= len(availableAgents) {
			break // More jobs than agents
		}

		agent := availableAgents[i%len(availableAgents)]
		job.AgentID = &agent.ID

		if err := u.jobRepo.Update(ctx, &job); err != nil {
			return fmt.Errorf("failed to assign job to agent: %w", err)
		}

		// Update agent status to busy
		if err := u.agentRepo.UpdateStatus(ctx, agent.ID, "busy"); err != nil {
			return fmt.Errorf("failed to update agent status: %w", err)
		}
	}

	return nil
}

// stopRelatedRunningJobs stops all running jobs that are related to the completed job
// This is used when a password is found to stop other agents from continuing
func (u *jobUsecase) stopRelatedRunningJobs(ctx context.Context, completedJob *domain.Job) error {
	// Extract base job name (remove agent-specific suffixes)
	baseJobName := u.extractBaseJobName(completedJob.Name)
	if baseJobName == "" {
		return nil // Not a distributed job
	}

	// Find all running jobs with similar names
	runningJobs, err := u.jobRepo.GetByStatus(ctx, "running")
	if err != nil {
		return fmt.Errorf("failed to get running jobs: %w", err)
	}

	var jobsToStop []*domain.Job
	for _, job := range runningJobs {
		// Skip the completed job itself
		if job.ID == completedJob.ID {
			continue
		}

		// Check if this job is related (same base name)
		if u.isRelatedJob(job.Name, baseJobName) {
			jobsToStop = append(jobsToStop, &job)
		}
	}

	// Stop all related running jobs
	for _, job := range jobsToStop {
		// Set progress to 100% and status to failed
		job.Progress = 100.0
		job.Status = "failed"
		job.Result = "Password found by another agent - stopping"
		now := time.Now()
		job.CompletedAt = &now

		if err := u.jobRepo.Update(ctx, job); err != nil {
			fmt.Printf("Warning: failed to stop related job %s: %v\n", job.Name, err)
			continue
		}

		// Update agent status to online
		if job.AgentID != nil {
			if err := u.agentRepo.UpdateStatus(ctx, *job.AgentID, "online"); err != nil {
				fmt.Printf("Warning: failed to update agent status for job %s: %v\n", job.Name, err)
			}
		}

		fmt.Printf("✅ Stopped related job %s because password was found by %s\n",
			job.Name, completedJob.Name)
	}

	return nil
}

// extractBaseJobName extracts the base job name from a distributed job name
// Examples:
// "test-cracking (Part 1 - test-agent-A)" -> "test-cracking"
// "test-cracking (Part 2 - test-agent-B)" -> "test-cracking"
// "test-cracking (test-agent-C)" -> "test-cracking"
func (u *jobUsecase) extractBaseJobName(jobName string) string {
	// Look for patterns like "(Part X - Agent)" or "(Agent)"
	if strings.Contains(jobName, " (Part ") {
		// Extract everything before " (Part "
		parts := strings.Split(jobName, " (Part ")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	} else if strings.Contains(jobName, " (") && strings.Contains(jobName, ")") {
		// Extract everything before " ("
		parts := strings.Split(jobName, " (")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	return "" // Not a distributed job
}

// isRelatedJob checks if a job name is related to a base job name
func (u *jobUsecase) isRelatedJob(jobName, baseJobName string) bool {
	if baseJobName == "" {
		return false
	}

	// Check if the job name starts with the base name and contains agent-specific info
	return strings.HasPrefix(jobName, baseJobName) &&
		(strings.Contains(jobName, " (Part ") || strings.Contains(jobName, " ("))
}
