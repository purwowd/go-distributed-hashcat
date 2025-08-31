package usecase

import (
	"context"
	"fmt"
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

// CreateDistributedJobs creates multiple jobs by dividing wordlist among available agents
func (u *distributedJobUsecase) CreateDistributedJobs(ctx context.Context, req *domain.DistributedJobRequest) (*domain.DistributedJobResult, error) {
	// Get available agents
	agents, err := u.agentRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get agents: %w", err)
	}

	// Filter only online agents
	onlineAgents := u.filterOnlineAgents(agents)
	if len(onlineAgents) == 0 {
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
	agentPerformances := u.calculateAgentPerformance(onlineAgents)

	// Divide wordlist based on agent performance
	wordlistSegments := u.divideWordlistByPerformance(wordlist, agentPerformances)

	// Create master job record
	masterJobID := uuid.New()
	masterJob := &domain.Job{
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

	// Save master job
	if err := u.jobRepo.Create(ctx, masterJob); err != nil {
		return nil, fmt.Errorf("failed to create master job: %w", err)
	}

	// Calculate total distributed words
	totalDistributed := int64(0)
	for _, assignment := range agentAssignments {
		totalDistributed += assignment.WordCount
	}

	// Build success message with warnings if some agents failed
	message := fmt.Sprintf("Successfully created %d distributed jobs for %d agents", len(subJobs), len(agentAssignments))
	if len(failedAgents) > 0 {
		message += fmt.Sprintf(" (Warning: %d agents failed: %v)", len(failedAgents), failedAgents)
	}

	return &domain.DistributedJobResult{
		MasterJobID:      masterJobID,
		SubJobs:          subJobs,
		AgentAssignments: agentAssignments,
		TotalWords: func() int64 {
			if wordlist.WordCount != nil {
				return *wordlist.WordCount
			}
			return 0
		}(),
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

// calculateAgentPerformance calculates performance scores for agents
func (u *distributedJobUsecase) calculateAgentPerformance(agents []domain.Agent) []domain.AgentPerformance {
	var performances []domain.AgentPerformance

	for _, agent := range agents {
		performance := domain.AgentPerformance{
			AgentID:      agent.ID,
			Name:         agent.Name,
			Capabilities: agent.Capabilities,
			Speed:        0, // Will be updated based on capabilities
			ResourceType: "CPU",
			Performance:  0.5, // Default performance
		}

		// Determine resource type and performance based on capabilities
		capabilities := strings.ToLower(agent.Capabilities)
		if strings.Contains(capabilities, "gpu") || strings.Contains(capabilities, "cuda") || strings.Contains(capabilities, "opencl") {
			performance.ResourceType = "GPU"
			performance.Performance = 1.0 // GPU gets full performance
			performance.Speed = 1000000   // 1M H/s for GPU
		} else if strings.Contains(capabilities, "cpu") {
			performance.ResourceType = "CPU"
			performance.Performance = 0.3 // CPU gets 30% performance
			performance.Speed = 100000    // 100K H/s for CPU
		} else {
			// Auto-detect based on capabilities string
			if strings.Contains(capabilities, "rtx") || strings.Contains(capabilities, "gtx") || strings.Contains(capabilities, "radeon") {
				performance.ResourceType = "GPU"
				performance.Performance = 1.0
				performance.Speed = 1000000
			} else {
				performance.ResourceType = "CPU"
				performance.Performance = 0.3
				performance.Speed = 100000
			}
		}

		performances = append(performances, performance)
	}

	// Sort by performance (highest first)
	sort.Slice(performances, func(i, j int) bool {
		return performances[i].Performance > performances[j].Performance
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

	// Distribute words based on performance ratio
	remainingWords := totalWords
	for i, agent := range agentPerformances {
		if i == len(agentPerformances)-1 {
			// Last agent gets remaining words
			segments = append(segments, domain.WordlistSegment{
				StartIndex: totalWords - remainingWords,
				EndIndex:   totalWords,
				WordCount:  remainingWords,
			})
		} else {
			// Calculate words for this agent based on performance ratio
			performanceRatio := agent.Performance / totalPerformance
			wordsForAgent := int64(math.Ceil(float64(totalWords) * performanceRatio))

			// Ensure minimum words per agent
			if wordsForAgent < 100 {
				wordsForAgent = 100
			}

			// Ensure we don't exceed remaining words
			if wordsForAgent > remainingWords {
				wordsForAgent = remainingWords
			}

			segments = append(segments, domain.WordlistSegment{
				StartIndex: totalWords - remainingWords,
				EndIndex:   totalWords - remainingWords + wordsForAgent,
				WordCount:  wordsForAgent,
			})

			remainingWords -= wordsForAgent
		}
	}

	return segments
}

// NOTE: createWordlistSegment function removed - we now use hashcat's --skip and --limit
// parameters for distributed cracking instead of creating physical segment files

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
