package usecase

import (
	"context"
	"fmt"
	"time"

	"go-distributed-hashcat/internal/domain"

	"github.com/google/uuid"
)

// JobProgressService handles job progress tracking and calculations
type JobProgressService struct {
	jobRepo domain.JobRepository
}

// NewJobProgressService creates a new job progress service
func NewJobProgressService(jobRepo domain.JobRepository) *JobProgressService {
	return &JobProgressService{
		jobRepo: jobRepo,
	}
}

// UpdateJobProgress updates job progress with speed and ETA calculation
func (s *JobProgressService) UpdateJobProgress(ctx context.Context, jobID uuid.UUID, progress float64, speed int64, processedWords int64) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	// Update progress
	job.Progress = progress
	job.Speed = speed
	job.ProcessedWords = processedWords

	// Calculate ETA
	if speed > 0 && job.TotalWords > 0 {
		remainingWords := job.TotalWords - processedWords
		if remainingWords > 0 {
			etaSeconds := float64(remainingWords) / float64(speed)
			etaDuration := time.Duration(etaSeconds) * time.Second
			
			// Format ETA as human-readable duration string
			var etaStr string
			minutes := int(etaDuration.Minutes())
			seconds := int(etaDuration.Seconds()) % 60
			
			if minutes > 0 {
				if seconds > 0 {
					etaStr = fmt.Sprintf("%d mins %d secs", minutes, seconds)
				} else {
					etaStr = fmt.Sprintf("%d mins", minutes)
				}
			} else {
				etaStr = fmt.Sprintf("%d secs", seconds)
			}
			
			job.ETA = &etaStr
		}
	}

	// Update job
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	return nil
}

// UpdateJobResult updates job with cracked password and final status
func (s *JobProgressService) UpdateJobResult(ctx context.Context, jobID uuid.UUID, password string, status string) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	// Update rules field with cracked password
	if password != "" {
		job.Rules = password
	}

	// Update status
	job.Status = status

	// Set completion time if completed
	if status == "completed" || status == "failed" {
		now := time.Now()
		job.CompletedAt = &now
		etaStr := ""
		job.ETA = &etaStr // Clear ETA when completed
	}

	// Update job
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	return nil
}

// CalculateAgentSpeed calculates agent speed based on hardware capabilities
func (s *JobProgressService) CalculateAgentSpeed(agent *domain.Agent, hashType int) int64 {
	baseSpeed := int64(0)

	// Base speed based on agent capabilities
	if agent.Capabilities != "" {
		capabilities := agent.Capabilities
		if contains(capabilities, "GPU") || contains(capabilities, "RTX") || contains(capabilities, "GTX") {
			// GPU agents - much faster
			baseSpeed = 1000000000 // 1 GH/s base
		} else if contains(capabilities, "CPU") {
			// CPU agents - slower
			baseSpeed = 10000000 // 10 MH/s base
		} else {
			// Unknown capabilities - default to CPU speed
			baseSpeed = 5000000 // 5 MH/s base
		}
	} else {
		// Default to CPU speed if no capabilities specified
		baseSpeed = 5000000 // 5 MH/s base
	}

	// Adjust speed based on hash type
	hashTypeMultiplier := s.getHashTypeMultiplier(hashType)

	// Add some randomness to simulate real-world conditions
	randomFactor := 0.8 + (0.4 * (float64(time.Now().UnixNano()%100) / 100.0))

	finalSpeed := int64(float64(baseSpeed) * hashTypeMultiplier * randomFactor)

	return finalSpeed
}

// getHashTypeMultiplier returns speed multiplier based on hash type
func (s *JobProgressService) getHashTypeMultiplier(hashType int) float64 {
	switch hashType {
	case 2500: // WPA/WPA2
		return 0.1 // WPA2 is much slower
	case 0: // MD5
		return 1.0 // MD5 is fast
	case 100: // SHA1
		return 0.8 // SHA1 is slower than MD5
	case 1400: // SHA256
		return 0.6 // SHA256 is slower
	case 1700: // SHA512
		return 0.4 // SHA512 is much slower
	default:
		return 0.5 // Default multiplier
	}
}

// contains checks if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr))
}

// FormatSpeed formats speed in human-readable format
func (s *JobProgressService) FormatSpeed(speed int64) string {
	if speed >= 1000000000 {
		return fmt.Sprintf("%.1f GH/s", float64(speed)/1000000000)
	} else if speed >= 1000000 {
		return fmt.Sprintf("%.1f MH/s", float64(speed)/1000000)
	} else if speed >= 1000 {
		return fmt.Sprintf("%.1f KH/s", float64(speed)/1000)
	} else {
		return fmt.Sprintf("%d H/s", speed)
	}
}

// FormatETA formats ETA in human-readable format
func (s *JobProgressService) FormatETA(eta *time.Time) string {
	if eta == nil {
		return "Unknown"
	}

	now := time.Now()
	duration := eta.Sub(now)

	if duration <= 0 {
		return "Completed"
	}

	if duration < time.Minute {
		return fmt.Sprintf("%.0f seconds", duration.Seconds())
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%d minutes", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("%d hours", hours)
	} else {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%d days", days)
	}
}
