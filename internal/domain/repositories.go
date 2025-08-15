package domain

import (
	"context"

	"github.com/google/uuid"
)

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
