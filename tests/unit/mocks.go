package unit

import (
	"context"
	"io"

	"go-distributed-hashcat/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockWordlistUsecase is a mock implementation of usecase.WordlistUsecase
type MockWordlistUsecase struct {
	mock.Mock
}

func (m *MockWordlistUsecase) UploadWordlist(ctx context.Context, name string, content io.Reader, size int64) (*domain.Wordlist, error) {
	args := m.Called(ctx, name, content, size)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Wordlist), args.Error(1)
}

func (m *MockWordlistUsecase) GetWordlist(ctx context.Context, id uuid.UUID) (*domain.Wordlist, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Wordlist), args.Error(1)
}

func (m *MockWordlistUsecase) GetByOrigName(ctx context.Context, origName string) (*domain.Wordlist, error) {
	args := m.Called(ctx, origName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Wordlist), args.Error(1)
}

func (m *MockWordlistUsecase) GetAllWordlists(ctx context.Context) ([]domain.Wordlist, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Wordlist), args.Error(1)
}

func (m *MockWordlistUsecase) DeleteWordlist(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockHashFileRepository is a mock implementation of domain.HashFileRepository
type MockHashFileRepository struct {
	mock.Mock
}

func (m *MockHashFileRepository) Create(ctx context.Context, hashFile *domain.HashFile) error {
	args := m.Called(ctx, hashFile)
	return args.Error(0)
}

func (m *MockHashFileRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.HashFile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.HashFile), args.Error(1)
}

func (m *MockHashFileRepository) GetByOrigName(ctx context.Context, origName string) (*domain.HashFile, error) {
	args := m.Called(ctx, origName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.HashFile), args.Error(1)
}

func (m *MockHashFileRepository) GetAll(ctx context.Context) ([]domain.HashFile, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.HashFile), args.Error(1)
}

func (m *MockHashFileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockJobRepository is a mock implementation of domain.JobRepository
type MockJobRepository struct {
	mock.Mock
}

func (m *MockJobRepository) Create(ctx context.Context, job *domain.Job) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *MockJobRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Job), args.Error(1)
}

func (m *MockJobRepository) GetAll(ctx context.Context) ([]domain.Job, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Job), args.Error(1)
}

func (m *MockJobRepository) GetByStatus(ctx context.Context, status string) ([]domain.Job, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]domain.Job), args.Error(1)
}

func (m *MockJobRepository) GetByAgentID(ctx context.Context, agentID uuid.UUID) ([]domain.Job, error) {
	args := m.Called(ctx, agentID)
	return args.Get(0).([]domain.Job), args.Error(1)
}

func (m *MockJobRepository) GetAvailableJobForAgent(ctx context.Context, agentID uuid.UUID) (*domain.Job, error) {
	args := m.Called(ctx, agentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Job), args.Error(1)
}

func (m *MockJobRepository) Update(ctx context.Context, job *domain.Job) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *MockJobRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockJobRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockJobRepository) UpdateProgress(ctx context.Context, id uuid.UUID, progress float64, speed int64) error {
	args := m.Called(ctx, id, progress, speed)
	return args.Error(0)
}

// MockAgentRepository is a mock implementation of domain.AgentRepository
type MockAgentRepository struct {
	mock.Mock
}

func (m *MockAgentRepository) Create(ctx context.Context, agent *domain.Agent) error {
	args := m.Called(ctx, agent)
	return args.Error(0)
}

func (m *MockAgentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Agent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Agent), args.Error(1)
}

func (m *MockAgentRepository) GetByName(ctx context.Context, name string) (*domain.Agent, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Agent), args.Error(1)
}

func (m *MockAgentRepository) GetByNameAndIP(ctx context.Context, name, ip string, port int) (*domain.Agent, error) {
	args := m.Called(ctx, name, ip, port)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Agent), args.Error(1)
}

func (m *MockAgentRepository) GetAll(ctx context.Context) ([]domain.Agent, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Agent), args.Error(1)
}

func (m *MockAgentRepository) Update(ctx context.Context, agent *domain.Agent) error {
	args := m.Called(ctx, agent)
	return args.Error(0)
}

func (m *MockAgentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAgentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockAgentRepository) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAgentRepository) GetByIPAddress(ctx context.Context, ip string) (*domain.Agent, error) {
	args := m.Called(ctx, ip)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Agent), args.Error(1)
}

func (m *MockAgentRepository) GetByAgentKey(ctx context.Context, agentKey string) (*domain.Agent, error) {
	args := m.Called(ctx, agentKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Agent), args.Error(1)
}

func (m *MockAgentRepository) CreateAgent(ctx context.Context, agent *domain.Agent) error {
	args := m.Called(ctx, agent)
	return args.Error(0)
}

func (m *MockAgentRepository) UpdateAgent(ctx context.Context, agent *domain.Agent) error {
	args := m.Called(ctx, agent)
	return args.Error(0)
}

func (m *MockAgentRepository) GetByNameAndIPForStartup(ctx context.Context, name, ip string, port int) (*domain.Agent, error) {
	args := m.Called(ctx, name, ip, port)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Agent), args.Error(1)
}

// MockWordlistRepository is a mock implementation of domain.WordlistRepository
type MockWordlistRepository struct {
	mock.Mock
}

func (m *MockWordlistRepository) Create(ctx context.Context, wordlist *domain.Wordlist) error {
	args := m.Called(ctx, wordlist)
	return args.Error(0)
}

func (m *MockWordlistRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Wordlist, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Wordlist), args.Error(1)
}

func (m *MockWordlistRepository) GetByOrigName(ctx context.Context, origName string) (*domain.Wordlist, error) {
	args := m.Called(ctx, origName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Wordlist), args.Error(1)
}

func (m *MockWordlistRepository) GetAll(ctx context.Context) ([]domain.Wordlist, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Wordlist), args.Error(1)
}

func (m *MockWordlistRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
