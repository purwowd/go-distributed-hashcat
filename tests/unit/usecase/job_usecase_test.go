package usecase_test

import (
	"context"
	"errors"
	"testing"

	"go-distributed-hashcat/internal/domain"
	"go-distributed-hashcat/internal/usecase"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

// MockHashFileRepository for job tests
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

func (m *MockHashFileRepository) GetAll(ctx context.Context) ([]domain.HashFile, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.HashFile), args.Error(1)
}

func (m *MockHashFileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestJobUsecase_CreateJob(t *testing.T) {
	hashFileID := uuid.New()
	agentID := uuid.New()

	tests := []struct {
		name          string
		request       *domain.CreateJobRequest
		mockSetup     func(*MockJobRepository, *MockAgentRepository, *MockHashFileRepository)
		expectedError bool
	}{
		{
			name: "successful job creation without agent assignment",
			request: &domain.CreateJobRequest{
				Name:       "test-job",
				HashType:   0,
				AttackMode: 0,
				HashFileID: hashFileID.String(),
				Wordlist:   "rockyou.txt",
			},
			mockSetup: func(jobRepo *MockJobRepository, agentRepo *MockAgentRepository, hashFileRepo *MockHashFileRepository) {
				// Mock hash file exists
				hashFile := &domain.HashFile{
					ID:   hashFileID,
					Name: "test.hash",
					Path: "/uploads/test.hash",
				}
				hashFileRepo.On("GetByID", mock.Anything, hashFileID).Return(hashFile, nil)

				// Mock job creation - no agent assignment in CreateJob
				jobRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Job")).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "job creation with manual agent assignment",
			request: &domain.CreateJobRequest{
				Name:       "test-job",
				HashType:   0,
				AttackMode: 0,
				HashFileID: hashFileID.String(),
				Wordlist:   "rockyou.txt",
				AgentID:    agentID.String(),
			},
			mockSetup: func(jobRepo *MockJobRepository, agentRepo *MockAgentRepository, hashFileRepo *MockHashFileRepository) {
				// Mock hash file exists
				hashFile := &domain.HashFile{
					ID:   hashFileID,
					Name: "test.hash",
					Path: "/uploads/test.hash",
				}
				hashFileRepo.On("GetByID", mock.Anything, hashFileID).Return(hashFile, nil)

				// Mock specific agent exists
				agent := &domain.Agent{
					ID:     agentID,
					Status: "online",
				}
				agentRepo.On("GetByID", mock.Anything, agentID).Return(agent, nil)

				// Mock job creation
				jobRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Job")).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "hash file not found",
			request: &domain.CreateJobRequest{
				Name:       "test-job",
				HashType:   0,
				AttackMode: 0,
				HashFileID: hashFileID.String(),
				Wordlist:   "rockyou.txt",
			},
			mockSetup: func(jobRepo *MockJobRepository, agentRepo *MockAgentRepository, hashFileRepo *MockHashFileRepository) {
				hashFileRepo.On("GetByID", mock.Anything, hashFileID).Return(nil, errors.New("hash file not found"))
			},
			expectedError: true,
		},
		{
			name: "invalid agent assignment",
			request: &domain.CreateJobRequest{
				Name:       "test-job",
				HashType:   0,
				AttackMode: 0,
				HashFileID: hashFileID.String(),
				Wordlist:   "rockyou.txt",
				AgentID:    agentID.String(),
			},
			mockSetup: func(jobRepo *MockJobRepository, agentRepo *MockAgentRepository, hashFileRepo *MockHashFileRepository) {
				// Mock hash file exists
				hashFile := &domain.HashFile{
					ID:   hashFileID,
					Name: "test.hash",
					Path: "/uploads/test.hash",
				}
				hashFileRepo.On("GetByID", mock.Anything, hashFileID).Return(hashFile, nil)

				// Mock agent not found
				agentRepo.On("GetByID", mock.Anything, agentID).Return(nil, errors.New("agent not found"))
			},
			expectedError: true,
		},
		{
			name: "invalid hash file ID format",
			request: &domain.CreateJobRequest{
				Name:       "test-job",
				HashType:   0,
				AttackMode: 0,
				HashFileID: "invalid-uuid",
				Wordlist:   "rockyou.txt",
			},
			mockSetup: func(jobRepo *MockJobRepository, agentRepo *MockAgentRepository, hashFileRepo *MockHashFileRepository) {
				// No mocks needed - should fail on UUID parsing
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobRepo := new(MockJobRepository)
			agentRepo := new(MockAgentRepository)
			hashFileRepo := new(MockHashFileRepository)

			tt.mockSetup(jobRepo, agentRepo, hashFileRepo)

			usecase := usecase.NewJobUsecase(jobRepo, agentRepo, hashFileRepo)
			ctx := context.Background()

			job, err := usecase.CreateJob(ctx, tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, job)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, job)
				assert.Equal(t, tt.request.Name, job.Name)
				assert.Equal(t, "pending", job.Status)
				assert.NotEqual(t, uuid.Nil, job.ID)
			}

			jobRepo.AssertExpectations(t)
			agentRepo.AssertExpectations(t)
			hashFileRepo.AssertExpectations(t)
		})
	}
}

func TestJobUsecase_StartJob(t *testing.T) {
	jobID := uuid.New()
	agentID := uuid.New()

	tests := []struct {
		name          string
		jobID         uuid.UUID
		mockSetup     func(*MockJobRepository, *MockAgentRepository)
		expectedError bool
	}{
		{
			name:  "successful job start",
			jobID: jobID,
			mockSetup: func(jobRepo *MockJobRepository, agentRepo *MockAgentRepository) {
				job := &domain.Job{
					ID:      jobID,
					Status:  "pending",
					AgentID: &agentID,
				}
				jobRepo.On("GetByID", mock.Anything, jobID).Return(job, nil)
				jobRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Job")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:  "job not found",
			jobID: jobID,
			mockSetup: func(jobRepo *MockJobRepository, agentRepo *MockAgentRepository) {
				jobRepo.On("GetByID", mock.Anything, jobID).Return(nil, errors.New("job not found"))
			},
			expectedError: true,
		},
		{
			name:  "job already running",
			jobID: jobID,
			mockSetup: func(jobRepo *MockJobRepository, agentRepo *MockAgentRepository) {
				job := &domain.Job{
					ID:     jobID,
					Status: "running",
				}
				jobRepo.On("GetByID", mock.Anything, jobID).Return(job, nil)
			},
			expectedError: true,
		},
		{
			name:  "repository update error",
			jobID: jobID,
			mockSetup: func(jobRepo *MockJobRepository, agentRepo *MockAgentRepository) {
				job := &domain.Job{
					ID:      jobID,
					Status:  "pending",
					AgentID: &agentID,
				}
				jobRepo.On("GetByID", mock.Anything, jobID).Return(job, nil)
				jobRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Job")).Return(errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobRepo := new(MockJobRepository)
			agentRepo := new(MockAgentRepository)
			hashFileRepo := new(MockHashFileRepository)

			tt.mockSetup(jobRepo, agentRepo)

			usecase := usecase.NewJobUsecase(jobRepo, agentRepo, hashFileRepo)
			ctx := context.Background()

			err := usecase.StartJob(ctx, tt.jobID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			jobRepo.AssertExpectations(t)
			agentRepo.AssertExpectations(t)
		})
	}
}

func TestJobUsecase_UpdateJobProgress(t *testing.T) {
	jobID := uuid.New()

	tests := []struct {
		name          string
		jobID         uuid.UUID
		progress      float64
		speed         int64
		mockSetup     func(*MockJobRepository)
		expectedError bool
	}{
		{
			name:     "successful progress update",
			jobID:    jobID,
			progress: 50.5,
			speed:    1000000,
			mockSetup: func(jobRepo *MockJobRepository) {
				jobRepo.On("UpdateProgress", mock.Anything, jobID, 50.5, int64(1000000)).Return(nil)
			},
			expectedError: false,
		},
		{
			name:     "repository update error",
			jobID:    jobID,
			progress: 75.0,
			speed:    500000,
			mockSetup: func(jobRepo *MockJobRepository) {
				jobRepo.On("UpdateProgress", mock.Anything, jobID, 75.0, int64(500000)).Return(errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobRepo := new(MockJobRepository)
			agentRepo := new(MockAgentRepository)
			hashFileRepo := new(MockHashFileRepository)

			tt.mockSetup(jobRepo)

			usecase := usecase.NewJobUsecase(jobRepo, agentRepo, hashFileRepo)
			ctx := context.Background()

			err := usecase.UpdateJobProgress(ctx, tt.jobID, tt.progress, tt.speed)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			jobRepo.AssertExpectations(t)
		})
	}
}

func TestJobUsecase_CompleteJob(t *testing.T) {
	jobID := uuid.New()
	agentID := uuid.New()

	tests := []struct {
		name          string
		jobID         uuid.UUID
		result        string
		mockSetup     func(*MockJobRepository, *MockAgentRepository)
		expectedError bool
	}{
		{
			name:   "successful job completion",
			jobID:  jobID,
			result: "password123",
			mockSetup: func(jobRepo *MockJobRepository, agentRepo *MockAgentRepository) {
				job := &domain.Job{
					ID:      jobID,
					Status:  "running",
					AgentID: &agentID,
				}
				jobRepo.On("GetByID", mock.Anything, jobID).Return(job, nil)
				jobRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Job")).Return(nil)
				agentRepo.On("UpdateStatus", mock.Anything, agentID, "online").Return(nil)
			},
			expectedError: false,
		},
		{
			name:   "job not found",
			jobID:  jobID,
			result: "password123",
			mockSetup: func(jobRepo *MockJobRepository, agentRepo *MockAgentRepository) {
				jobRepo.On("GetByID", mock.Anything, jobID).Return(nil, errors.New("job not found"))
			},
			expectedError: true,
		},
		{
			name:   "completing already completed job",
			jobID:  jobID,
			result: "password123",
			mockSetup: func(jobRepo *MockJobRepository, agentRepo *MockAgentRepository) {
				job := &domain.Job{
					ID:     jobID,
					Status: "completed",
				}
				jobRepo.On("GetByID", mock.Anything, jobID).Return(job, nil)
				jobRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Job")).Return(nil)
			},
			expectedError: false, // Implementation allows completing any job
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobRepo := new(MockJobRepository)
			agentRepo := new(MockAgentRepository)
			hashFileRepo := new(MockHashFileRepository)

			tt.mockSetup(jobRepo, agentRepo)

			usecase := usecase.NewJobUsecase(jobRepo, agentRepo, hashFileRepo)
			ctx := context.Background()

			err := usecase.CompleteJob(ctx, tt.jobID, tt.result, 1000000) // Add speed parameter

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			jobRepo.AssertExpectations(t)
			agentRepo.AssertExpectations(t)
		})
	}
}

func TestJobUsecase_AssignJobsToAgents(t *testing.T) {
	agentID := uuid.New()
	jobID := uuid.New()

	tests := []struct {
		name          string
		mockSetup     func(*MockJobRepository, *MockAgentRepository)
		expectedError bool
	}{
		{
			name: "successful job assignment",
			mockSetup: func(jobRepo *MockJobRepository, agentRepo *MockAgentRepository) {
				// Mock pending jobs
				pendingJob := &domain.Job{
					ID:      jobID,
					Status:  "pending",
					AgentID: nil, // No agent assigned yet
				}
				jobRepo.On("GetByStatus", mock.Anything, "pending").Return([]domain.Job{*pendingJob}, nil)

				// Mock available agents
				agent := &domain.Agent{
					ID:     agentID,
					Status: "online",
				}
				agentRepo.On("GetAll", mock.Anything).Return([]domain.Agent{*agent}, nil)
				agentRepo.On("UpdateStatus", mock.Anything, agentID, "busy").Return(nil)

				// Mock job update
				jobRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Job")).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "no pending jobs",
			mockSetup: func(jobRepo *MockJobRepository, agentRepo *MockAgentRepository) {
				jobRepo.On("GetByStatus", mock.Anything, "pending").Return([]domain.Job{}, nil)
			},
			expectedError: false,
		},
		{
			name: "no available agents",
			mockSetup: func(jobRepo *MockJobRepository, agentRepo *MockAgentRepository) {
				pendingJob := &domain.Job{
					ID:      jobID,
					Status:  "pending",
					AgentID: nil,
				}
				jobRepo.On("GetByStatus", mock.Anything, "pending").Return([]domain.Job{*pendingJob}, nil)

				// Mock no available agents
				agentRepo.On("GetAll", mock.Anything).Return([]domain.Agent{}, nil)
			},
			expectedError: false, // Not an error, just no assignment happens
		},
		{
			name: "repository error getting pending jobs",
			mockSetup: func(jobRepo *MockJobRepository, agentRepo *MockAgentRepository) {
				jobRepo.On("GetByStatus", mock.Anything, "pending").Return([]domain.Job{}, errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobRepo := new(MockJobRepository)
			agentRepo := new(MockAgentRepository)
			hashFileRepo := new(MockHashFileRepository)

			tt.mockSetup(jobRepo, agentRepo)

			usecase := usecase.NewJobUsecase(jobRepo, agentRepo, hashFileRepo)
			ctx := context.Background()

			err := usecase.AssignJobsToAgents(ctx)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			jobRepo.AssertExpectations(t)
			agentRepo.AssertExpectations(t)
		})
	}
}

func TestJobUsecase_GetJobsByStatus(t *testing.T) {
	tests := []struct {
		name          string
		status        string
		mockSetup     func(*MockJobRepository)
		expectedCount int
		expectedError bool
	}{
		{
			name:   "get pending jobs",
			status: "pending",
			mockSetup: func(jobRepo *MockJobRepository) {
				jobs := []domain.Job{
					{ID: uuid.New(), Status: "pending"},
					{ID: uuid.New(), Status: "pending"},
				}
				jobRepo.On("GetByStatus", mock.Anything, "pending").Return(jobs, nil)
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name:   "no jobs with status",
			status: "completed",
			mockSetup: func(jobRepo *MockJobRepository) {
				jobRepo.On("GetByStatus", mock.Anything, "completed").Return([]domain.Job{}, nil)
			},
			expectedCount: 0,
			expectedError: false,
		},
		{
			name:   "repository error",
			status: "running",
			mockSetup: func(jobRepo *MockJobRepository) {
				jobRepo.On("GetByStatus", mock.Anything, "running").Return([]domain.Job{}, errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobRepo := new(MockJobRepository)
			agentRepo := new(MockAgentRepository)
			hashFileRepo := new(MockHashFileRepository)

			tt.mockSetup(jobRepo)

			usecase := usecase.NewJobUsecase(jobRepo, agentRepo, hashFileRepo)
			ctx := context.Background()

			jobs, err := usecase.GetJobsByStatus(ctx, tt.status)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, jobs, tt.expectedCount)
			}

			jobRepo.AssertExpectations(t)
		})
	}
}
