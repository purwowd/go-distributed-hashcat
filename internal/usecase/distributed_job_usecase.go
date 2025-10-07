package usecase

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"go-distributed-hashcat/internal/domain"

	"github.com/google/uuid"
)

type distributedJobUsecase struct {
	agentRepo    domain.AgentRepository
	jobRepo      domain.JobRepository
	wordlistRepo domain.WordlistRepository
	hashFileRepo domain.HashFileRepository
	uploadDir    string
}

func NewDistributedJobUsecase(
	agentRepo domain.AgentRepository,
	jobRepo domain.JobRepository,
	wordlistRepo domain.WordlistRepository,
	hashFileRepo domain.HashFileRepository,
	uploadDir string,
) domain.DistributedJobUsecase {
	return &distributedJobUsecase{
		agentRepo:    agentRepo,
		jobRepo:      jobRepo,
		wordlistRepo: wordlistRepo,
		hashFileRepo: hashFileRepo,
		uploadDir:    uploadDir,
	}
}

// CreateDistributedJobs creates multiple jobs by dividing wordlist among specified agents
func (u *distributedJobUsecase) CreateDistributedJobs(ctx context.Context, req *domain.DistributedJobRequest) (*domain.DistributedJobResult, error) {
	var agents []domain.Agent
	var err error

	// Determine which agents to use
	if req.AutoDistribute {
		// Use all available online agents
		agents, err = u.agentRepo.GetAll(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get agents: %w", err)
		}
		// Filter only online agents
		agents = u.filterOnlineAgents(agents)
	} else if len(req.AgentIDs) > 0 {
		// Use specific agents from request
		agents = make([]domain.Agent, 0, len(req.AgentIDs))
		for _, agentIDStr := range req.AgentIDs {
			agentID, err := uuid.Parse(agentIDStr)
			if err != nil {
				return nil, fmt.Errorf("invalid agent ID %s: %w", agentIDStr, err)
			}

			agent, err := u.agentRepo.GetByID(ctx, agentID)
			if err != nil {
				return nil, fmt.Errorf("agent not found: %w", err)
			}

			if agent.Status != "online" {
				return nil, fmt.Errorf("agent %s is not available (status: %s)", agent.Name, agent.Status)
			}

			agents = append(agents, *agent) // Dereference the pointer
		}
	} else {
		return nil, fmt.Errorf("no agents specified and auto-distribute is disabled")
	}

	if len(agents) == 0 {
		return nil, fmt.Errorf("no online agents available")
	}

	// Get wordlist details
	wordlistID, err := uuid.Parse(req.WordlistID)
	if err != nil {
		return nil, fmt.Errorf("invalid wordlist ID: %w", err)
	}

	wordlist, err := u.wordlistRepo.GetByID(ctx, wordlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wordlist: %w", err)
	}

	// Get hash file details
	hashFileID, err := uuid.Parse(req.HashFileID)
	if err != nil {
		return nil, fmt.Errorf("invalid hash file ID: %w", err)
	}

	hashFile, err := u.hashFileRepo.GetByID(ctx, hashFileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get hash file: %w", err)
	}

	// Calculate agent performance scores
	agentPerformances := u.calculateAgentPerformance(agents)

	// Divide wordlist based on agent performance
	wordlistSegments := u.divideWordlistByPerformance(wordlist, agentPerformances)

	// Create master job record (optional - only for coordination tracking)
	var masterJobID uuid.UUID
	var masterJob *domain.Job

	// Only create master job if explicitly requested or for complex coordination
	// For simple distributed jobs, we skip the master job to avoid clutter
	if req.CreateMasterJob {
		masterJobID = uuid.New()
		masterJob = &domain.Job{
			ID:         masterJobID,
			Name:       fmt.Sprintf("%s (Master)", req.Name),
			Status:     "distributed",
			HashType:   req.HashType,
			AttackMode: req.AttackMode,
			HashFile:   hashFile.OrigName,
			HashFileID: &hashFileID,
			Wordlist:   wordlist.OrigName,
			WordlistID: &wordlistID,
			Rules:      req.Rules,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
	}

	// Create sub-jobs for each agent using skip/limit ranges
	var subJobs []domain.Job
	var agentAssignments []domain.AgentPerformance
	var failedAgents []string

	for i, segment := range wordlistSegments {
		if i >= len(agentPerformances) {
			break
		}

		agent := agentPerformances[i]

		// Calculate skip and limit values for this segment
		skip := segment.StartIndex
		limit := segment.WordCount

		// Create sub-job with skip/limit parameters
		subJob := domain.Job{
			ID:         uuid.New(),
			Name:       fmt.Sprintf("%s (Part %d - %s)", req.Name, i+1, agent.Name),
			Status:     "pending",
			HashType:   req.HashType,
			AttackMode: req.AttackMode,
			HashFile:   hashFile.OrigName,
			HashFileID: &hashFileID,
			Wordlist:   wordlist.OrigName, // Use original wordlist, not segment file
			WordlistID: &wordlistID,
			Rules:      req.Rules,
			AgentID:    &agent.AgentID,
			Skip:       &skip,  // Hashcat --skip parameter
			WordLimit:  &limit, // Hashcat --limit parameter
			TotalWords: segment.WordCount,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		// Save sub-job to database - continue even if some fail
		if err := u.jobRepo.Create(ctx, &subJob); err != nil {
			// Log error but continue with other jobs
			failedAgents = append(failedAgents, agent.Name)
			continue
		}

		subJobs = append(subJobs, subJob)

		// Update agent assignment with word count
		agent.WordCount = segment.WordCount
		agentAssignments = append(agentAssignments, agent)
	}

	// Check if we have at least one successful job
	if len(subJobs) == 0 {
		return nil, fmt.Errorf("failed to create any sub-jobs - all agents failed")
	}

	// Save master job only if it was created
	if masterJob != nil {
		if err := u.jobRepo.Create(ctx, masterJob); err != nil {
			return nil, fmt.Errorf("failed to create master job: %w", err)
		}
	}

	// Calculate total distributed words
	totalDistributed := int64(0)
	for _, assignment := range agentAssignments {
		totalDistributed += assignment.WordCount
	}

	// Build success message with warnings if some agents failed
	message := fmt.Sprintf("Successfully created %d distributed jobs for %d agents", len(subJobs), len(agentAssignments))
	if len(failedAgents) > 0 {
		message += fmt.Sprintf(" (Warning: failed to create jobs for agents: %s)", strings.Join(failedAgents, ", "))
	}

	// Only return master job ID if master job was actually created
	var resultMasterJobID uuid.UUID
	if masterJob != nil {
		resultMasterJobID = masterJobID
	}

	return &domain.DistributedJobResult{
		MasterJobID:      resultMasterJobID,
		SubJobs:          subJobs,
		AgentAssignments: agentAssignments,
		TotalWords:       *wordlist.WordCount,
		DistributedWords: totalDistributed,
		Message:          message,
	}, nil
}

// filterOnlineAgents filters only online agents
func (u *distributedJobUsecase) filterOnlineAgents(agents []domain.Agent) []domain.Agent {
	var onlineAgents []domain.Agent
	for _, agent := range agents {
		if agent.Status == "online" {
			onlineAgents = append(onlineAgents, agent)
		}
	}
	return onlineAgents
}

// calculateAgentPerformance calculates performance scores for agents based on their actual speed data
func (u *distributedJobUsecase) calculateAgentPerformance(agents []domain.Agent) []domain.AgentPerformance {
	var performances []domain.AgentPerformance

	for _, agent := range agents {
		performance := domain.AgentPerformance{
			AgentID:      agent.ID,
			Name:         agent.Name,
			Capabilities: agent.Capabilities,
			ResourceType: "CPU",
			Performance:  0.3,         // Default CPU performance
			Speed:        agent.Speed, // Use actual speed from database
		}

		// Use actual speed from database, fallback to capability-based estimation if speed is 0
		if agent.Speed == 0 {
			// Enhanced performance detection based on capabilities as fallback
			capabilities := strings.ToLower(agent.Capabilities)

			if strings.Contains(capabilities, "rtx 4090") || strings.Contains(capabilities, "rtx 4080") {
				performance.ResourceType = "GPU"
				performance.Performance = 1.0
				performance.Speed = 5000000 // 5M H/s for high-end RTX
			} else if strings.Contains(capabilities, "rtx 4070") || strings.Contains(capabilities, "rtx 3060") {
				performance.ResourceType = "GPU"
				performance.Performance = 0.9
				performance.Speed = 4000000 // 4M H/s for mid-range RTX
			} else if strings.Contains(capabilities, "gtx 1660") || strings.Contains(capabilities, "gtx 1070") {
				performance.ResourceType = "GPU"
				performance.Performance = 0.7
				performance.Speed = 3000000 // 3M H/s for GTX series
			} else if strings.Contains(capabilities, "gpu") || strings.Contains(capabilities, "cuda") || strings.Contains(capabilities, "opencl") {
				performance.ResourceType = "GPU"
				performance.Performance = 0.8
				performance.Speed = 3500000 // 3.5M H/s for generic GPU
			} else if strings.Contains(capabilities, "ryzen 9") || strings.Contains(capabilities, "i9") {
				performance.ResourceType = "CPU"
				performance.Performance = 0.5
				performance.Speed = 200000 // 200K H/s for high-end CPU
			} else if strings.Contains(capabilities, "ryzen 7") || strings.Contains(capabilities, "i7") {
				performance.ResourceType = "CPU"
				performance.Performance = 0.4
				performance.Speed = 150000 // 150K H/s for mid-range CPU
			} else {
				// Default CPU performance
				performance.ResourceType = "CPU"
				performance.Performance = 0.3
				performance.Speed = 100000 // 100K H/s for standard CPU
			}
		} else {
			// Use actual speed data to determine resource type and performance
			if strings.Contains(strings.ToLower(agent.Capabilities), "gpu") ||
				strings.Contains(strings.ToLower(agent.Capabilities), "cuda") ||
				strings.Contains(strings.ToLower(agent.Capabilities), "opencl") {
				performance.ResourceType = "GPU"
			} else {
				performance.ResourceType = "CPU"
			}

			// Calculate performance score based on actual speed (normalize to 0-1 scale)
			// Assuming max speed of 10M H/s for normalization
			performance.Performance = float64(agent.Speed) / 10000000.0
			if performance.Performance > 1.0 {
				performance.Performance = 1.0
			}
		}

		performances = append(performances, performance)
	}

	// Sort by actual speed (highest first)
	sort.Slice(performances, func(i, j int) bool {
		return performances[i].Speed > performances[j].Speed
	})

	return performances
}

// divideWordlistByPerformance divides wordlist based on agent performance
func (u *distributedJobUsecase) divideWordlistByPerformance(wordlist *domain.Wordlist, agentPerformances []domain.AgentPerformance) []domain.WordlistSegment {
	var segments []domain.WordlistSegment

	if wordlist.WordCount == nil || *wordlist.WordCount == 0 {
		// If word count is unknown, create equal segments
		wordsPerAgent := int64(1000) // Default 1000 words per agent
		for i := range agentPerformances {
			segments = append(segments, domain.WordlistSegment{
				StartIndex: int64(i) * wordsPerAgent,
				EndIndex:   int64(i+1) * wordsPerAgent,
				WordCount:  wordsPerAgent,
			})
		}
		return segments
	}

	totalWords := *wordlist.WordCount
	totalPerformance := 0.0

	// Calculate total performance score
	for _, agent := range agentPerformances {
		totalPerformance += agent.Performance
	}

	// Distribute words based on performance ratio with better precision
	remainingWords := totalWords
	currentIndex := int64(0)

	for i, agent := range agentPerformances {
		if i == len(agentPerformances)-1 {
			// Last agent gets remaining words to ensure no words are lost
			segments = append(segments, domain.WordlistSegment{
				StartIndex: currentIndex,
				EndIndex:   totalWords,
				WordCount:  remainingWords,
			})
		} else {
			// Calculate words for this agent based on performance ratio
			performanceRatio := agent.Performance / totalPerformance
			wordsForAgent := int64(math.Floor(float64(totalWords) * performanceRatio))

			// Ensure minimum 2 words per agent
			if wordsForAgent < 2 {
				wordsForAgent = 2
			}

			// Ensure we don't exceed remaining words
			if wordsForAgent > remainingWords {
				wordsForAgent = remainingWords
			}

			// Create segment with proper start/end indices
			endIndex := currentIndex + wordsForAgent
			if endIndex > totalWords {
				endIndex = totalWords
			}

			segments = append(segments, domain.WordlistSegment{
				StartIndex: currentIndex,
				EndIndex:   endIndex,
				WordCount:  endIndex - currentIndex,
			})

			// Update indices for next iteration
			currentIndex = endIndex
			remainingWords -= wordsForAgent
		}
	}

	return segments
}

// NOTE: createWordlistSegment function removed - we now use hashcat's --skip and --limit
// parameters for distributed cracking instead of creating physical segment files

// HandleDistributedJobCompletion handles the completion of distributed jobs
// When one agent finds the password, all other running jobs are marked as failed
func (u *distributedJobUsecase) HandleDistributedJobCompletion(ctx context.Context, masterJobID uuid.UUID, successfulJobID uuid.UUID, password string) error {
	// Get master job
	masterJob, err := u.jobRepo.GetByID(ctx, masterJobID)
	if err != nil {
		return fmt.Errorf("failed to get master job: %w", err)
	}

	// Get all sub-jobs for this master job
	allJobs, err := u.jobRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all jobs: %w", err)
	}

	// Find sub-jobs based on master job name pattern
	masterJobName := strings.TrimSuffix(masterJob.Name, " (Master)")
	var subJobs []domain.Job
	for _, job := range allJobs {
		if strings.Contains(job.Name, masterJobName) && job.ID != masterJobID {
			subJobs = append(subJobs, job)
		}
	}

	// Mark all other running/pending sub-jobs as cancelled
	for _, subJob := range subJobs {
		if subJob.ID != successfulJobID && (subJob.Status == "running" || subJob.Status == "pending") {
			// Update job status to cancelled with 100% progress
			subJob.Status = "cancelled"
			subJob.Progress = 100
			subJob.Result = "Password found by another agent - job cancelled"
			subJob.CompletedAt = &time.Time{}

			if err := u.jobRepo.Update(ctx, &subJob); err != nil {
				log.Printf("Warning: failed to update sub-job %s: %v", subJob.ID, err)
			}
		}
	}

	// Update master job with success result
	masterJob.Status = "completed"
	masterJob.Progress = 100
	masterJob.Result = fmt.Sprintf("SUCCESS: Password found by agent - %s", password)
	masterJob.CompletedAt = &time.Time{}

	if err := u.jobRepo.Update(ctx, masterJob); err != nil {
		return fmt.Errorf("failed to update master job: %w", err)
	}

	log.Printf("✅ Distributed job %s completed successfully. Password: %s", masterJob.Name, password)
	return nil
}

// GetDistributedJobStatus gets the status of all sub-jobs for a master job
func (u *distributedJobUsecase) GetDistributedJobStatus(ctx context.Context, masterJobID uuid.UUID) (*domain.DistributedJobResult, error) {
	// Get master job
	masterJob, err := u.jobRepo.GetByID(ctx, masterJobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get master job: %w", err)
	}

	// Get all jobs with similar name pattern (sub-jobs)
	allJobs, err := u.jobRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all jobs: %w", err)
	}

	var subJobs []domain.Job
	var agentAssignments []domain.AgentPerformance

	// Filter sub-jobs based on master job name
	masterJobName := strings.TrimSuffix(masterJob.Name, " (Master)")
	for _, job := range allJobs {
		if strings.Contains(job.Name, masterJobName) && !strings.Contains(job.Name, "(Master)") {
			subJobs = append(subJobs, job)
		}
	}

	// Get agent assignments
	for _, subJob := range subJobs {
		if subJob.AgentID != nil {
			agent, err := u.agentRepo.GetByID(ctx, *subJob.AgentID)
			if err == nil {
				agentAssignments = append(agentAssignments, domain.AgentPerformance{
					AgentID:      agent.ID,
					Name:         agent.Name,
					Capabilities: agent.Capabilities,
					Speed:        0,
					ResourceType: "CPU",
					Performance:  0.5,
					WordCount:    0,
				})
			}
		}
	}

	return &domain.DistributedJobResult{
		MasterJobID:      masterJobID,
		SubJobs:          subJobs,
		AgentAssignments: agentAssignments,
		TotalWords:       0,
		DistributedWords: 0,
		Message:          fmt.Sprintf("Found %d sub-jobs for master job", len(subJobs)),
	}, nil
}
