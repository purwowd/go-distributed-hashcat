package domain

import (
	"context"

	"github.com/google/uuid"
)

// AgentSpeedUpdate represents the data structure for updating agent speed
// This struct is used for request body and data transfer when updating agent speed
//
// Example usage:
//   - HTTP request body for PUT /api/v1/agents/{id}/speed
//   - Data transfer between layers (handler -> usecase -> repository)
//   - JSON serialization/deserialization
//
// JSON format:
//
//	{
//	  "agent_id": "uuid-string",
//	  "speed": 5000
//	}
type AgentSpeedUpdate struct {
	AgentID string `json:"agent_id" binding:"required"`
	Speed   int64  `json:"speed" binding:"required,min=0"`
}

// AgentRepository defines the interface for agent data operations
type AgentRepository interface {
	Create(ctx context.Context, agent *Agent) error
	GetByID(ctx context.Context, id uuid.UUID) (*Agent, error)
	GetByName(ctx context.Context, name string) (*Agent, error)
	GetByNameAndIP(ctx context.Context, name, ip string, port int) (*Agent, error)
	GetAll(ctx context.Context) ([]Agent, error)
	Update(ctx context.Context, agent *Agent) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateLastSeen(ctx context.Context, id uuid.UUID) error
	UpdateSpeed(ctx context.Context, id uuid.UUID, speed int64) error
	UpdateSpeedWithStatus(ctx context.Context, id uuid.UUID, speed int64, status string) error
	GetByIPAddress(ctx context.Context, ip string) (*Agent, error)
	GetByAgentKey(ctx context.Context, agentKey string) (*Agent, error)
	CreateAgent(ctx context.Context, agent *Agent) error // bisa panggil Create
	UpdateAgent(ctx context.Context, agent *Agent) error // bisa panggil Update
	GetByNameAndIPForStartup(ctx context.Context, name, ip string, port int) (*Agent, error)
}

// JobRepository defines the interface for job data operations
type JobRepository interface {
	Create(ctx context.Context, job *Job) error
	GetByID(ctx context.Context, id uuid.UUID) (*Job, error)
	GetAll(ctx context.Context) ([]Job, error)
	GetByStatus(ctx context.Context, status string) ([]Job, error)
	GetByAgentID(ctx context.Context, agentID uuid.UUID) ([]Job, error)
	GetAvailableJobForAgent(ctx context.Context, agentID uuid.UUID) (*Job, error)
	Update(ctx context.Context, job *Job) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateProgress(ctx context.Context, id uuid.UUID, progress float64, speed int64) error
}

// JobUsecase defines the interface for job business logic operations
type JobUsecase interface {
	CreateJob(ctx context.Context, req *CreateJobRequest) (*Job, error)
	GetJob(ctx context.Context, id uuid.UUID) (*Job, error)
	GetAllJobs(ctx context.Context) ([]Job, error)
	GetJobsByStatus(ctx context.Context, status string) ([]Job, error)
	GetJobsByAgentID(ctx context.Context, agentID uuid.UUID) ([]Job, error)
	StartJob(ctx context.Context, id uuid.UUID) error
	UpdateJobProgress(ctx context.Context, id uuid.UUID, progress float64, speed int64) error
	UpdateJobData(ctx context.Context, job *Job) error
	CompleteJob(ctx context.Context, id uuid.UUID, result string, speed int64) error
	FailJob(ctx context.Context, id uuid.UUID, reason string) error
	PauseJob(ctx context.Context, id uuid.UUID) error
	ResumeJob(ctx context.Context, id uuid.UUID) error
	DeleteJob(ctx context.Context, id uuid.UUID) error
	AssignJobsToAgents(ctx context.Context) error
}

// HashFileRepository defines the interface for hash file data operations
type HashFileRepository interface {
	Create(ctx context.Context, hashFile *HashFile) error
	GetByID(ctx context.Context, id uuid.UUID) (*HashFile, error)
	GetAll(ctx context.Context) ([]HashFile, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// WordlistRepository defines the interface for wordlist data operations
type WordlistRepository interface {
	Create(ctx context.Context, wordlist *Wordlist) error
	GetByID(ctx context.Context, id uuid.UUID) (*Wordlist, error)
	GetAll(ctx context.Context) ([]Wordlist, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// DistributedJobUsecase defines the interface for distributed job operations
type DistributedJobUsecase interface {
	CreateDistributedJobs(ctx context.Context, req *DistributedJobRequest) (*DistributedJobResult, error)
	GetDistributedJobStatus(ctx context.Context, masterJobID uuid.UUID) (*DistributedJobResult, error)
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetAll(ctx context.Context) ([]User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
}

// AuthUsecase defines the interface for authentication business logic operations
type AuthUsecase interface {
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	Logout(ctx context.Context, token string) error
	ValidateToken(ctx context.Context, token string) (*JWTClaims, error)
	RefreshToken(ctx context.Context, token string) (*LoginResponse, error)
	CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	GetAllUsers(ctx context.Context) ([]User, error)
	CheckUsernameExists(ctx context.Context, username string) (bool, error)
}
