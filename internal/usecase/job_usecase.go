package usecase

import (
	"context"
	"fmt"
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
	CompleteJob(ctx context.Context, id uuid.UUID, result string) error
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
}

func NewJobUsecase(jobRepo domain.JobRepository, agentRepo domain.AgentRepository, hashFileRepo domain.HashFileRepository) JobUsecase {
	return &jobUsecase{
		jobRepo:      jobRepo,
		agentRepo:    agentRepo,
		hashFileRepo: hashFileRepo,
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
	if req.WordlistID != "" {
		wordlistID, err := uuid.Parse(req.WordlistID)
		if err != nil {
			return nil, fmt.Errorf("invalid wordlist ID: %w", err)
		}
		job.WordlistID = &wordlistID
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
			// Create master job record
			masterJob := &domain.Job{
				ID:             uuid.New(),
				Name:           fmt.Sprintf("%s (Master)", req.Name),
				Status:         "distributed",
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
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}

			// Save master job
			if err := u.jobRepo.Create(ctx, masterJob); err != nil {
				return nil, fmt.Errorf("failed to create master job: %w", err)
			}

			// Create sub-jobs for each agent
			var subJobs []*domain.Job
			for i, agentID := range agentIDs {
				// Get agent details for naming
				agent, _ := u.agentRepo.GetByID(ctx, agentID)
				agentName := "Unknown"
				if agent != nil {
					agentName = agent.Name
				}

				subJob := &domain.Job{
					ID:             uuid.New(),
					Name:           fmt.Sprintf("%s (Part %d - %s)", req.Name, i+1, agentName),
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
					AgentID:        &agentID,
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
				}

				// Save sub-job
				if err := u.jobRepo.Create(ctx, subJob); err != nil {
					return nil, fmt.Errorf("failed to create sub-job %d: %w", i, err)
				}

				subJobs = append(subJobs, subJob)
			}

			// Return the first sub-job as the primary result
			// The master job and other sub-jobs are created but not returned
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

func (u *jobUsecase) CompleteJob(ctx context.Context, id uuid.UUID, result string) error {
	job, err := u.jobRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	now := time.Now()
	job.Status = "completed"
	job.Result = result
	job.Progress = 100.0
	job.CompletedAt = &now

	if err := u.jobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	// Update agent status to online
	if job.AgentID != nil {
		if err := u.agentRepo.UpdateStatus(ctx, *job.AgentID, "online"); err != nil {
			// Log error but don't fail the job completion
			fmt.Printf("Warning: failed to update agent status: %v\n", err)
		}
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
