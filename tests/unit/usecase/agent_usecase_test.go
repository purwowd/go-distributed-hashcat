package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-distributed-hashcat/internal/domain"
	"go-distributed-hashcat/internal/usecase"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func TestAgentUsecase_RegisterAgent(t *testing.T) {
	tests := []struct {
		name           string
		request        *domain.CreateAgentRequest
		mockSetup      func(*MockAgentRepository)
		expectedError  bool
		expectedStatus string
	}{
		{
			name: "successful agent registration",
			request: &domain.CreateAgentRequest{
				Name:         "test-agent",
				IPAddress:    "192.168.1.100",
				Port:         8080,
				Capabilities: "gpu,cpu",
			},
			mockSetup: func(repo *MockAgentRepository) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Agent")).Return(nil)
			},
			expectedError:  false,
			expectedStatus: "online",
		},
		{
			name: "repository error during creation",
			request: &domain.CreateAgentRequest{
				Name:         "test-agent",
				IPAddress:    "192.168.1.100",
				Port:         8080,
				Capabilities: "gpu,cpu",
			},
			mockSetup: func(repo *MockAgentRepository) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Agent")).Return(errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAgentRepository)
			tt.mockSetup(mockRepo)

			usecase := usecase.NewAgentUsecase(mockRepo)
			ctx := context.Background()

			agent, err := usecase.RegisterAgent(ctx, tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, agent)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, agent)
				assert.Equal(t, tt.request.Name, agent.Name)
				assert.Equal(t, tt.request.IPAddress, agent.IPAddress)
				assert.Equal(t, tt.request.Port, agent.Port)
				assert.Equal(t, tt.expectedStatus, agent.Status)
				assert.NotEqual(t, uuid.Nil, agent.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAgentUsecase_GetAgent(t *testing.T) {
	agentID := uuid.New()
	expectedAgent := &domain.Agent{
		ID:        agentID,
		Name:      "test-agent",
		IPAddress: "192.168.1.100",
		Port:      8080,
		Status:    "online",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name          string
		agentID       uuid.UUID
		mockSetup     func(*MockAgentRepository)
		expectedError bool
	}{
		{
			name:    "successful agent retrieval",
			agentID: agentID,
			mockSetup: func(repo *MockAgentRepository) {
				repo.On("GetByID", mock.Anything, agentID).Return(expectedAgent, nil)
			},
			expectedError: false,
		},
		{
			name:    "agent not found",
			agentID: agentID,
			mockSetup: func(repo *MockAgentRepository) {
				repo.On("GetByID", mock.Anything, agentID).Return(nil, errors.New("agent not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAgentRepository)
			tt.mockSetup(mockRepo)

			usecase := usecase.NewAgentUsecase(mockRepo)
			ctx := context.Background()

			agent, err := usecase.GetAgent(ctx, tt.agentID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, agent)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, agent)
				assert.Equal(t, expectedAgent.ID, agent.ID)
				assert.Equal(t, expectedAgent.Name, agent.Name)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAgentUsecase_GetAllAgents(t *testing.T) {
	expectedAgents := []domain.Agent{
		{
			ID:        uuid.New(),
			Name:      "agent-1",
			IPAddress: "192.168.1.100",
			Status:    "online",
		},
		{
			ID:        uuid.New(),
			Name:      "agent-2",
			IPAddress: "192.168.1.101",
			Status:    "offline",
		},
	}

	tests := []struct {
		name          string
		mockSetup     func(*MockAgentRepository)
		expectedError bool
		expectedCount int
	}{
		{
			name: "successful agents retrieval",
			mockSetup: func(repo *MockAgentRepository) {
				repo.On("GetAll", mock.Anything).Return(expectedAgents, nil)
			},
			expectedError: false,
			expectedCount: 2,
		},
		{
			name: "repository error",
			mockSetup: func(repo *MockAgentRepository) {
				repo.On("GetAll", mock.Anything).Return([]domain.Agent{}, errors.New("database error"))
			},
			expectedError: true,
			expectedCount: 0,
		},
		{
			name: "no agents found",
			mockSetup: func(repo *MockAgentRepository) {
				repo.On("GetAll", mock.Anything).Return([]domain.Agent{}, nil)
			},
			expectedError: false,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAgentRepository)
			tt.mockSetup(mockRepo)

			usecase := usecase.NewAgentUsecase(mockRepo)
			ctx := context.Background()

			agents, err := usecase.GetAllAgents(ctx)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, agents, tt.expectedCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAgentUsecase_UpdateAgentStatus(t *testing.T) {
	agentID := uuid.New()

	tests := []struct {
		name          string
		agentID       uuid.UUID
		status        string
		mockSetup     func(*MockAgentRepository)
		expectedError bool
	}{
		{
			name:    "successful status update",
			agentID: agentID,
			status:  "offline",
			mockSetup: func(repo *MockAgentRepository) {
				repo.On("UpdateStatus", mock.Anything, agentID, "offline").Return(nil)
			},
			expectedError: false,
		},
		{
			name:    "repository error during update",
			agentID: agentID,
			status:  "offline",
			mockSetup: func(repo *MockAgentRepository) {
				repo.On("UpdateStatus", mock.Anything, agentID, "offline").Return(errors.New("update failed"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAgentRepository)
			tt.mockSetup(mockRepo)

			usecase := usecase.NewAgentUsecase(mockRepo)
			ctx := context.Background()

			err := usecase.UpdateAgentStatus(ctx, tt.agentID, tt.status)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAgentUsecase_DeleteAgent(t *testing.T) {
	agentID := uuid.New()

	tests := []struct {
		name          string
		agentID       uuid.UUID
		mockSetup     func(*MockAgentRepository)
		expectedError bool
	}{
		{
			name:    "successful agent deletion",
			agentID: agentID,
			mockSetup: func(repo *MockAgentRepository) {
				repo.On("Delete", mock.Anything, agentID).Return(nil)
			},
			expectedError: false,
		},
		{
			name:    "repository error during deletion",
			agentID: agentID,
			mockSetup: func(repo *MockAgentRepository) {
				repo.On("Delete", mock.Anything, agentID).Return(errors.New("deletion failed"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAgentRepository)
			tt.mockSetup(mockRepo)

			usecase := usecase.NewAgentUsecase(mockRepo)
			ctx := context.Background()

			err := usecase.DeleteAgent(ctx, tt.agentID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAgentUsecase_GetAvailableAgent(t *testing.T) {
	tests := []struct {
		name          string
		agents        []domain.Agent
		mockSetup     func(*MockAgentRepository)
		expectedError bool
		hasAvailable  bool
	}{
		{
			name: "available agent found",
			agents: []domain.Agent{
				{ID: uuid.New(), Status: "offline"},
				{ID: uuid.New(), Status: "online"},
				{ID: uuid.New(), Status: "busy"},
			},
			mockSetup: func(repo *MockAgentRepository) {
				agents := []domain.Agent{
					{ID: uuid.New(), Status: "offline"},
					{ID: uuid.New(), Status: "online"},
					{ID: uuid.New(), Status: "busy"},
				}
				repo.On("GetAll", mock.Anything).Return(agents, nil)
			},
			expectedError: false,
			hasAvailable:  true,
		},
		{
			name: "no available agents",
			agents: []domain.Agent{
				{ID: uuid.New(), Status: "offline"},
				{ID: uuid.New(), Status: "busy"},
			},
			mockSetup: func(repo *MockAgentRepository) {
				agents := []domain.Agent{
					{ID: uuid.New(), Status: "offline"},
					{ID: uuid.New(), Status: "busy"},
				}
				repo.On("GetAll", mock.Anything).Return(agents, nil)
			},
			expectedError: true,
			hasAvailable:  false,
		},
		{
			name: "repository error",
			mockSetup: func(repo *MockAgentRepository) {
				repo.On("GetAll", mock.Anything).Return([]domain.Agent{}, errors.New("database error"))
			},
			expectedError: true,
			hasAvailable:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAgentRepository)
			tt.mockSetup(mockRepo)

			usecase := usecase.NewAgentUsecase(mockRepo)
			ctx := context.Background()

			agent, err := usecase.GetAvailableAgent(ctx)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, agent)
			} else {
				assert.NoError(t, err)
				if tt.hasAvailable {
					assert.NotNil(t, agent)
					assert.Equal(t, "online", agent.Status)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAgentUsecase_UpdateAgentHeartbeat(t *testing.T) {
	agentID := uuid.New()

	tests := []struct {
		name          string
		agentID       uuid.UUID
		mockSetup     func(*MockAgentRepository)
		expectedError bool
	}{
		{
			name:    "successful heartbeat update",
			agentID: agentID,
			mockSetup: func(repo *MockAgentRepository) {
				repo.On("UpdateLastSeen", mock.Anything, agentID).Return(nil)
			},
			expectedError: false,
		},
		{
			name:    "repository error during heartbeat update",
			agentID: agentID,
			mockSetup: func(repo *MockAgentRepository) {
				repo.On("UpdateLastSeen", mock.Anything, agentID).Return(errors.New("update failed"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAgentRepository)
			tt.mockSetup(mockRepo)

			usecase := usecase.NewAgentUsecase(mockRepo)
			ctx := context.Background()

			err := usecase.UpdateAgentHeartbeat(ctx, tt.agentID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
