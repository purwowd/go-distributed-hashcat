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
	GetCacheStats() map[string]interface{}
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

// Performance metrics tracking
type cacheMetrics struct {
	hits   int64
	misses int64
	total  int64
}

func (m *cacheMetrics) recordHit() {
	m.hits++
	m.total++
}

func (m *cacheMetrics) recordMiss() {
	m.misses++
	m.total++
}

func (m *cacheMetrics) getHitRate() float64 {
	if m.total == 0 {
		return 0
	}
	return float64(m.hits) / float64(m.total) * 100
}

func (m *cacheMetrics) getMissRate() float64 {
	if m.total == 0 {
		return 0
	}
	return float64(m.misses) / float64(m.total) * 100
}

func (m *cacheMetrics) reset() {
	m.hits = 0
	m.misses = 0
	m.total = 0
}

// In-memory cache implementation
type memoryCache struct {
	agents    sync.Map // uuid.UUID -> domain.Agent
	wordlists sync.Map // uuid.UUID -> domain.Wordlist
	hashFiles sync.Map // uuid.UUID -> domain.HashFile
	ttl       time.Duration
	metrics   *cacheMetrics
	createdAt time.Time
}

func newMemoryCache(ttl time.Duration) *memoryCache {
	cache := &memoryCache{
		ttl:       ttl,
		metrics:   &cacheMetrics{},
		createdAt: time.Now(),
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
				c.metrics.recordHit()
				return agent, true
			}
		}
		// Remove expired entry
		c.agents.Delete(id)
	}
	c.metrics.recordMiss()
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
				c.metrics.recordHit()
				return wordlist, true
			}
		}
		c.wordlists.Delete(id)
	}
	c.metrics.recordMiss()
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
				c.metrics.recordHit()
				return hashFile, true
			}
		}
		c.hashFiles.Delete(id)
	}
	c.metrics.recordMiss()
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
	c.metrics.reset()
	c.createdAt = time.Now()
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

	// Try to load directly from repository on cache miss
	if agent, err := s.agentRepo.GetByID(context.Background(), *agentID); err == nil {
		s.cache.setAgent(*agentID, agent)
		return agent.Name
	}

	// Final fallback for cache miss and repository failure
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
		// Try to load directly from repository on cache miss
		if wl, err := s.wordlistRepo.GetByID(context.Background(), id); err == nil {
			s.cache.setWordlist(id, wl)
			if wl.OrigName != "" {
				return wl.OrigName
			}
			return wl.Name
		}
		// Final fallback for cache miss and repository failure
		return wordlist[:8] + "..."
	}

	// Direct filename - return as is
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
		// Try to load directly from repository on cache miss
		if hashFile, err := s.hashFileRepo.GetByID(context.Background(), *hashFileID); err == nil {
			s.cache.setHashFile(*hashFileID, hashFile)
			if hashFile.OrigName != "" {
				return hashFile.OrigName
			}
			return hashFile.Name
		}
		// Final fallback for cache miss and repository failure
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

func (s *jobEnrichmentService) GetCacheStats() map[string]interface{} {
	agentCount := 0
	wordlistCount := 0
	hashFileCount := 0

	s.cache.agents.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*cacheEntry); ok && !entry.isExpired() {
			agentCount++
		}
		return true
	})

	s.cache.wordlists.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*cacheEntry); ok && !entry.isExpired() {
			wordlistCount++
		}
		return true
	})

	s.cache.hashFiles.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*cacheEntry); ok && !entry.isExpired() {
			hashFileCount++
		}
		return true
	})

	hitRate := s.cache.metrics.getHitRate()
	missRate := s.cache.metrics.getMissRate()
	uptime := time.Since(s.cache.createdAt).Seconds()

	// Calculate query reduction estimate (more cache hits = less DB queries)
	queryReduction := 0.0
	if hitRate > 0 {
		queryReduction = hitRate * 0.9 // Approximation: 90% of cache hits avoid DB queries
	}

	// Calculate response speed improvement
	responseSpeedImprovement := 0.0
	if hitRate > 0 {
		responseSpeedImprovement = hitRate * 0.95 // Cache is typically 95% faster than DB
	}

	return map[string]interface{}{
		"agents":                   agentCount,
		"wordlists":                wordlistCount,
		"hashFiles":                hashFileCount,
		"hitRate":                  hitRate,
		"missRate":                 missRate,
		"totalRequests":            s.cache.metrics.total,
		"cacheHits":                s.cache.metrics.hits,
		"cacheMisses":              s.cache.metrics.misses,
		"uptime":                   uptime,
		"queryReduction":           queryReduction,
		"responseSpeedImprovement": responseSpeedImprovement,
	}
}
