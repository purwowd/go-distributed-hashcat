package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Agent represents a cracking agent
type Agent struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	Port         int       `json:"port" db:"port"`
	Status       string    `json:"status" db:"status"` // online, offline, busy
	Capabilities string    `json:"capabilities" db:"capabilities"`
	AgentKey     string    `json:"agent_key" db:"agent_key"`
	LastSeen     time.Time `json:"last_seen" db:"last_seen"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Job represents a cracking job
type Job struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Status      string     `json:"status" db:"status"` // pending, running, completed, failed, paused
	HashType    int        `json:"hash_type" db:"hash_type"`
	AttackMode  int        `json:"attack_mode" db:"attack_mode"`
	HashFile    string     `json:"hash_file" db:"hash_file"`
	HashFileID  *uuid.UUID `json:"hash_file_id" db:"hash_file_id"`
	Wordlist    string     `json:"wordlist" db:"wordlist"`
	WordlistID  *uuid.UUID `json:"wordlist_id" db:"wordlist_id"`
	Rules       string     `json:"rules" db:"rules"`
	AgentID     *uuid.UUID `json:"agent_id" db:"agent_id"`
	Progress    float64    `json:"progress" db:"progress"`
	Speed       int64      `json:"speed" db:"speed"`
	ETA         *time.Time `json:"eta" db:"eta"`
	Result      string     `json:"result" db:"result"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	StartedAt   *time.Time `json:"started_at" db:"started_at"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
}

// HashFile represents uploaded hash files
type HashFile struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	OrigName  string    `json:"orig_name" db:"orig_name"`
	Path      string    `json:"path" db:"path"`
	Size      int64     `json:"size" db:"size"`
	Type      string    `json:"type" db:"type"` // hccapx, hccap, hash
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Wordlist represents a wordlist file
type Wordlist struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	OrigName  string    `json:"orig_name" db:"orig_name"`
	Path      string    `json:"path" db:"path"`
	Size      int64     `json:"size" db:"size"`
	WordCount *int64    `json:"word_count,omitempty" db:"word_count"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// JobStatus represents different job statuses
type JobStatus struct {
	JobID     uuid.UUID `json:"job_id"`
	Status    string    `json:"status"`
	Progress  float64   `json:"progress"`
	Speed     int64     `json:"speed"`
	ETA       string    `json:"eta"`
	Result    string    `json:"result"`
	Timestamp time.Time `json:"timestamp"`
}

// CreateJobRequest represents the request to create a new job
type CreateJobRequest struct {
	Name       string `json:"name" binding:"required"`
	HashType   int    `json:"hash_type" binding:"gte=0"`
	AttackMode int    `json:"attack_mode" binding:"gte=0"`
	HashFileID string `json:"hash_file_id" binding:"required"`
	Wordlist   string `json:"wordlist" binding:"required"`
	WordlistID string `json:"wordlist_id,omitempty"`
	AgentID    string `json:"agent_id,omitempty"` // Optional manual agent assignment
	Rules      string `json:"rules,omitempty"`
}

// EnrichedJob extends Job with readable names for frontend display
type EnrichedJob struct {
	Job
	AgentName    string `json:"agent_name,omitempty"`
	WordlistName string `json:"wordlist_name,omitempty"`
	HashFileName string `json:"hash_file_name,omitempty"`
}

// CreateAgentRequest represents the request to register a new agent
type CreateAgentRequest struct {
	Name         string `json:"name" binding:"required"`
	IPAddress    string `json:"ip_address" binding:"omitempty"`
	Port         int    `json:"port,omitempty"` // Optional, will default to 8080
	Capabilities string `json:"capabilities,omitempty"`
	AgentKey     string `json:"agent_key,omitempty"` // Agent key for validation
	Status       string `json:"status,omitempty"`
}

// DuplicateAgentError represents an error when trying to create an agent that already exists
type DuplicateAgentError struct {
	Name      string
	IPAddress string
	Port      int
}

func (e *DuplicateAgentError) Error() string {
	return fmt.Sprintf("agent with name '%s' and IP address '%s:%d' already exists", e.Name, e.IPAddress, e.Port)
}

// AlreadyRegisteredAgentError represents an error when trying to register an agent that is already registered
type AlreadyRegisteredAgentError struct {
	Name         string
	IPAddress    string
	Port         int
	Capabilities string
}

func (e *AlreadyRegisteredAgentError) Error() string {
	return fmt.Sprintf("agent '%s' is already registered with IP address '%s', port '%d', and capabilities '%s'",
		e.Name, e.IPAddress, e.Port, e.Capabilities)
}
