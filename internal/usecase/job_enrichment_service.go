package usecase

import (
	"context"
	"strings"
	"sync"
	"time"

	"go-distributed-hashcat/internal/domain"

	"github.com/google/uuid"
)

// JobEnrichmentService handles enriching jobs with readable names
type JobEnrichmentService interface {
	EnrichJobs(ctx context.Context, jobs []domain.Job) ([]domain.EnrichedJob, error)
	ClearCache()
	GetCacheStats() map[string]int
}

// Cache entry with TTL
type cacheEntry struct {
	data      interface{}
	timestamp time.Time
	ttl       time.Duration
}

func (c *cacheEntry) isExpired() bool {
	return time.Since(c.timestamp) > c.ttl
}

// In-memory cache implementation
type memoryCache struct {
	agents    sync.Map // uuid.UUID -> domain.Agent
	wordlists sync.Map // uuid.UUID -> domain.Wordlist
	hashFiles sync.Map // uuid.UUID -> domain.HashFile
	ttl       time.Duration
	mu        sync.RWMutex
}

func newMemoryCache(ttl time.Duration) *memoryCache {
	cache := &memoryCache{
		ttl: ttl,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

func (c *memoryCache) cleanup() {
	ticker := time.NewTicker(c.ttl / 2) // Cleanup every half TTL
	defer ticker.Stop()

	for range ticker.C {
		c.cleanupExpired()
	}
}

func (c *memoryCache) cleanupExpired() {
	// Cleanup agents
	c.agents.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*cacheEntry); ok && entry.isExpired() {
			c.agents.Delete(key)
		}
		return true
	})

	// Cleanup wordlists
	c.wordlists.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*cacheEntry); ok && entry.isExpired() {
			c.wordlists.Delete(key)
		}
		return true
	})

	// Cleanup hash files
	c.hashFiles.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*cacheEntry); ok && entry.isExpired() {
			c.hashFiles.Delete(key)
		}
		return true
	})
}

func (c *memoryCache) getAgent(id uuid.UUID) (*domain.Agent, bool) {
	if value, ok := c.agents.Load(id); ok {
		if entry, ok := value.(*cacheEntry); ok && !entry.isExpired() {
			if agent, ok := entry.data.(*domain.Agent); ok {
				return agent, true
			}
		}
		// Remove expired entry
		c.agents.Delete(id)
	}
	return nil, false
}

func (c *memoryCache) setAgent(id uuid.UUID, agent *domain.Agent) {
	entry := &cacheEntry{
		data:      agent,
		timestamp: time.Now(),
		ttl:       c.ttl,
	}
	c.agents.Store(id, entry)
}

func (c *memoryCache) getWordlist(id uuid.UUID) (*domain.Wordlist, bool) {
	if value, ok := c.wordlists.Load(id); ok {
		if entry, ok := value.(*cacheEntry); ok && !entry.isExpired() {
			if wordlist, ok := entry.data.(*domain.Wordlist); ok {
				return wordlist, true
			}
		}
		c.wordlists.Delete(id)
	}
	return nil, false
}

func (c *memoryCache) setWordlist(id uuid.UUID, wordlist *domain.Wordlist) {
	entry := &cacheEntry{
		data:      wordlist,
		timestamp: time.Now(),
		ttl:       c.ttl,
	}
	c.wordlists.Store(id, entry)
}

func (c *memoryCache) getHashFile(id uuid.UUID) (*domain.HashFile, bool) {
	if value, ok := c.hashFiles.Load(id); ok {
		if entry, ok := value.(*cacheEntry); ok && !entry.isExpired() {
			if hashFile, ok := entry.data.(*domain.HashFile); ok {
				return hashFile, true
			}
		}
		c.hashFiles.Delete(id)
	}
	return nil, false
}

func (c *memoryCache) setHashFile(id uuid.UUID, hashFile *domain.HashFile) {
	entry := &cacheEntry{
		data:      hashFile,
		timestamp: time.Now(),
		ttl:       c.ttl,
	}
	c.hashFiles.Store(id, entry)
}

func (c *memoryCache) clear() {
	c.agents = sync.Map{}
	c.wordlists = sync.Map{}
	c.hashFiles = sync.Map{}
}

func (c *memoryCache) getStats() map[string]int {
	stats := map[string]int{
		"agents":    0,
		"wordlists": 0,
		"hashFiles": 0,
	}

	c.agents.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*cacheEntry); ok && !entry.isExpired() {
			stats["agents"]++
		}
		return true
	})

	c.wordlists.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*cacheEntry); ok && !entry.isExpired() {
			stats["wordlists"]++
		}
		return true
	})

	c.hashFiles.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*cacheEntry); ok && !entry.isExpired() {
			stats["hashFiles"]++
		}
		return true
	})

	return stats
}

// jobEnrichmentService implementation
type jobEnrichmentService struct {
	agentRepo    domain.AgentRepository
	wordlistRepo domain.WordlistRepository
	hashFileRepo domain.HashFileRepository
	cache        *memoryCache
}

// NewJobEnrichmentService creates a new job enrichment service
func NewJobEnrichmentService(
	agentRepo domain.AgentRepository,
	wordlistRepo domain.WordlistRepository,
	hashFileRepo domain.HashFileRepository,
) JobEnrichmentService {
	return &jobEnrichmentService{
		agentRepo:    agentRepo,
		wordlistRepo: wordlistRepo,
		hashFileRepo: hashFileRepo,
		cache:        newMemoryCache(5 * time.Minute), // 5 minute TTL
	}
}

// EnrichJobs enriches jobs with readable names using batch loading and caching
func (s *jobEnrichmentService) EnrichJobs(ctx context.Context, jobs []domain.Job) ([]domain.EnrichedJob, error) {
	if len(jobs) == 0 {
		return []domain.EnrichedJob{}, nil
	}

	// Extract unique IDs that need to be fetched
	agentIDs := s.extractMissingAgentIDs(jobs)
	wordlistIDs := s.extractMissingWordlistIDs(jobs)
	hashFileIDs := s.extractMissingHashFileIDs(jobs)

	// Batch fetch missing entities
	if err := s.batchLoadAgents(ctx, agentIDs); err != nil {
		return nil, err
	}

	if err := s.batchLoadWordlists(ctx, wordlistIDs); err != nil {
		return nil, err
	}

	if err := s.batchLoadHashFiles(ctx, hashFileIDs); err != nil {
		return nil, err
	}

	// Enrich jobs using cached data
	enrichedJobs := make([]domain.EnrichedJob, len(jobs))
	for i, job := range jobs {
		enrichedJobs[i] = domain.EnrichedJob{
			Job:          job,
			AgentName:    s.getAgentName(job.AgentID),
			WordlistName: s.getWordlistName(job.Wordlist),
			HashFileName: s.getHashFileName(job.HashFileID, job.HashFile),
		}
	}

	return enrichedJobs, nil
}

func (s *jobEnrichmentService) extractMissingAgentIDs(jobs []domain.Job) []uuid.UUID {
	var missingIDs []uuid.UUID
	seen := make(map[uuid.UUID]bool)

	for _, job := range jobs {
		if job.AgentID != nil && !seen[*job.AgentID] {
			seen[*job.AgentID] = true
			if _, cached := s.cache.getAgent(*job.AgentID); !cached {
				missingIDs = append(missingIDs, *job.AgentID)
			}
		}
	}

	return missingIDs
}

func (s *jobEnrichmentService) extractMissingWordlistIDs(jobs []domain.Job) []uuid.UUID {
	var missingIDs []uuid.UUID
	seen := make(map[uuid.UUID]bool)

	for _, job := range jobs {
		if job.Wordlist != "" {
			if id, err := uuid.Parse(job.Wordlist); err == nil && !seen[id] {
				seen[id] = true
				if _, cached := s.cache.getWordlist(id); !cached {
					missingIDs = append(missingIDs, id)
				}
			}
		}
	}

	return missingIDs
}

func (s *jobEnrichmentService) extractMissingHashFileIDs(jobs []domain.Job) []uuid.UUID {
	var missingIDs []uuid.UUID
	seen := make(map[uuid.UUID]bool)

	for _, job := range jobs {
		if job.HashFileID != nil && !seen[*job.HashFileID] {
			seen[*job.HashFileID] = true
			if _, cached := s.cache.getHashFile(*job.HashFileID); !cached {
				missingIDs = append(missingIDs, *job.HashFileID)
			}
		}
	}

	return missingIDs
}

func (s *jobEnrichmentService) batchLoadAgents(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	// If repository supports batch loading, use it
	// For now, we'll load individually but could be optimized
	for _, id := range ids {
		agent, err := s.agentRepo.GetByID(ctx, id)
		if err == nil {
			s.cache.setAgent(id, agent)
		}
		// Don't return error for individual failures - use fallback
	}

	return nil
}

func (s *jobEnrichmentService) batchLoadWordlists(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	for _, id := range ids {
		wordlist, err := s.wordlistRepo.GetByID(ctx, id)
		if err == nil {
			s.cache.setWordlist(id, wordlist)
		}
	}

	return nil
}

func (s *jobEnrichmentService) batchLoadHashFiles(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	for _, id := range ids {
		hashFile, err := s.hashFileRepo.GetByID(ctx, id)
		if err == nil {
			s.cache.setHashFile(id, hashFile)
		}
	}

	return nil
}

func (s *jobEnrichmentService) getAgentName(agentID *uuid.UUID) string {
	if agentID == nil {
		return ""
	}

	if agent, cached := s.cache.getAgent(*agentID); cached {
		return agent.Name
	}

	// Fallback for cache miss
	return agentID.String()[:8] + "..."
}

func (s *jobEnrichmentService) getWordlistName(wordlist string) string {
	if wordlist == "" {
		return ""
	}

	// Try to parse as UUID
	if id, err := uuid.Parse(wordlist); err == nil {
		if wl, cached := s.cache.getWordlist(id); cached {
			if wl.OrigName != "" {
				return wl.OrigName
			}
			return wl.Name
		}
		// Fallback for cache miss
		return wordlist[:8] + "..."
	}

	// Direct filename
	return wordlist
}

func (s *jobEnrichmentService) getHashFileName(hashFileID *uuid.UUID, hashFilePath string) string {
	if hashFileID != nil {
		if hashFile, cached := s.cache.getHashFile(*hashFileID); cached {
			if hashFile.OrigName != "" {
				return hashFile.OrigName
			}
			return hashFile.Name
		}
		// Fallback for cache miss
		return hashFileID.String()[:8] + "..."
	}

	// Fallback to parsing path
	if hashFilePath != "" {
		parts := strings.Split(hashFilePath, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}

	return ""
}

func (s *jobEnrichmentService) ClearCache() {
	s.cache.clear()
}

func (s *jobEnrichmentService) GetCacheStats() map[string]int {
	return s.cache.getStats()
}
