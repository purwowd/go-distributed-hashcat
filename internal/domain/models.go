package domain

import (
	"time"

	"github.com/google/uuid"
)

// Agent represents a cracking agent
type Agent struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	Port         int       `json:"port" db:"port"`
	Status       string    `json:"status" db:"status"` // online, offline, busy, banned, key_only
	Capabilities string    `json:"capabilities" db:"capabilities"`
	LastSeen     time.Time `json:"last_seen" db:"last_seen"`
	AgentKey     string    `json:"agent_key,omitempty" db:"agent_key"`
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
	IPAddress    string `json:"ip_address" binding:"required"`
	Port         int    `json:"port" binding:"required"`
	Capabilities string `json:"capabilities,omitempty"`
}

// User represents a system user
type User struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Username  string     `json:"username" db:"username"`
	Email     string     `json:"email" db:"email"`
	Password  string     `json:"-" db:"password"` // Hidden in JSON responses
	Role      string     `json:"role" db:"role"`  // admin, user
	IsActive  bool       `json:"is_active" db:"is_active"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	LastLogin *time.Time `json:"last_login" db:"last_login"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token     string    `json:"token"`
	User      User      `json:"user"`
	ExpiresAt time.Time `json:"expires_at"`
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"oneof=admin user"`
}

// UpdateUserRequest represents the request to update user
type UpdateUserRequest struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Role     string `json:"role,omitempty"`
	IsActive *bool  `json:"is_active,omitempty"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// GenerateAgentKeyRequest represents request to generate new agent key
type GenerateAgentKeyRequest struct {
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// GenerateAgentKeyResponse represents response with new agent key
type GenerateAgentKeyResponse struct {
	AgentKey    string     `json:"agent_key"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// AgentKey represents an agent authentication key (separate from agents)
type AgentKey struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	AgentKey    string     `json:"agent_key" db:"agent_key"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description,omitempty" db:"description"`
	Status      string     `json:"status" db:"status"` // active, expired, revoked
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	AgentID     *uuid.UUID `json:"agent_id,omitempty" db:"agent_id"` // Reference to agent when key is used
}

// AgentKeyInfo represents agent key information for listing
type AgentKeyInfo struct {
	AgentKey    string     `json:"agent_key"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Status      string     `json:"status"` // active, expired, revoked
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsed    *time.Time `json:"last_used,omitempty"`
	AgentID     *uuid.UUID `json:"agent_id,omitempty"` // If key is already used by an agent
}
