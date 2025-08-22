package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-distributed-hashcat/internal/delivery/http/handler"
	"go-distributed-hashcat/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockJobUsecase is a mock implementation of usecase.JobUsecase
type MockJobUsecase struct {
	mock.Mock
}

// MockAgentRepository for testing
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

// MockWordlistRepository for testing
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

// MockHashFileRepository for testing
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

// MockJobEnrichmentService for testing
type MockJobEnrichmentService struct {
	mock.Mock
}

func (m *MockJobEnrichmentService) EnrichJobs(ctx context.Context, jobs []domain.Job) ([]domain.EnrichedJob, error) {
	args := m.Called(ctx, jobs)
	return args.Get(0).([]domain.EnrichedJob), args.Error(1)
}

func (m *MockJobEnrichmentService) GetCacheStats() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *MockJobEnrichmentService) ClearCache() {
	m.Called()
}

func (m *MockJobUsecase) CreateJob(ctx context.Context, req *domain.CreateJobRequest) (*domain.Job, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Job), args.Error(1)
}

func (m *MockJobUsecase) GetJob(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Job), args.Error(1)
}

func (m *MockJobUsecase) GetAllJobs(ctx context.Context) ([]domain.Job, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Job), args.Error(1)
}

func (m *MockJobUsecase) GetJobsByStatus(ctx context.Context, status string) ([]domain.Job, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]domain.Job), args.Error(1)
}

func (m *MockJobUsecase) StartJob(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockJobUsecase) UpdateJobProgress(ctx context.Context, id uuid.UUID, progress float64, speed int64) error {
	args := m.Called(ctx, id, progress, speed)
	return args.Error(0)
}

func (m *MockJobUsecase) CompleteJob(ctx context.Context, id uuid.UUID, result string) error {
	args := m.Called(ctx, id, result)
	return args.Error(0)
}

func (m *MockJobUsecase) FailJob(ctx context.Context, id uuid.UUID, reason string) error {
	args := m.Called(ctx, id, reason)
	return args.Error(0)
}

func (m *MockJobUsecase) PauseJob(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockJobUsecase) ResumeJob(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockJobUsecase) DeleteJob(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockJobUsecase) AssignJobsToAgents(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockJobUsecase) GetJobsByAgentID(ctx context.Context, agentID uuid.UUID) ([]domain.Job, error) {
	args := m.Called(ctx, agentID)
	return args.Get(0).([]domain.Job), args.Error(1)
}

func (m *MockJobUsecase) GetAvailableJobForAgent(ctx context.Context, agentID uuid.UUID) (*domain.Job, error) {
	args := m.Called(ctx, agentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Job), args.Error(1)
}

func (m *MockJobUsecase) UpdateJobData(ctx context.Context, job *domain.Job) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func TestJobHandler_CreateJob(t *testing.T) {
	hashFileID := uuid.New()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockJobUsecase)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful job creation",
			requestBody: map[string]interface{}{
				"name":         "test-job",
				"hash_type":    0,
				"attack_mode":  0,
				"hash_file_id": hashFileID.String(),
				"wordlist":     "rockyou.txt",
			},
			mockSetup: func(mockUsecase *MockJobUsecase) {
				expectedJob := &domain.Job{
					ID:         uuid.New(),
					Name:       "test-job",
					Status:     "pending",
					HashType:   0,
					AttackMode: 0,
					Wordlist:   "rockyou.txt",
				}
				mockUsecase.On("CreateJob", mock.Anything, mock.AnythingOfType("*domain.CreateJobRequest")).Return(expectedJob, nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				data := response["data"].(map[string]interface{})
				assert.Equal(t, "test-job", data["name"])
				assert.Equal(t, "pending", data["status"])
			},
		},
		{
			name:        "invalid JSON body",
			requestBody: "invalid-json",
			mockSetup: func(mockUsecase *MockJobUsecase) {
				// No mock calls expected
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "invalid character")
			},
		},
		{
			name: "missing required fields",
			requestBody: map[string]interface{}{
				"name": "test-job",
				// Missing hash_file_id, wordlist
			},
			mockSetup: func(mockUsecase *MockJobUsecase) {
				// No mock calls expected
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "required")
			},
		},
		{
			name: "usecase error",
			requestBody: map[string]interface{}{
				"name":         "test-job",
				"hash_type":    0,
				"attack_mode":  0,
				"hash_file_id": hashFileID.String(),
				"wordlist":     "rockyou.txt",
			},
			mockSetup: func(mockUsecase *MockJobUsecase) {
				mockUsecase.On("CreateJob", mock.Anything, mock.AnythingOfType("*domain.CreateJobRequest")).Return(nil, errors.New("hash file not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "hash file not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := new(MockJobUsecase)
			tt.mockSetup(mockUsecase)

			handler := handler.NewJobHandler(mockUsecase, nil, nil, nil)
			router := setupTestRouter()
			router.POST("/jobs", handler.CreateJob)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req, err := http.NewRequest("POST", "/jobs", bytes.NewReader(body))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestJobHandler_GetJob(t *testing.T) {
	jobID := uuid.New()

	tests := []struct {
		name           string
		jobID          string
		mockSetup      func(*MockJobUsecase)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:  "successful job retrieval",
			jobID: jobID.String(),
			mockSetup: func(mockUsecase *MockJobUsecase) {
				expectedJob := &domain.Job{
					ID:     jobID,
					Name:   "test-job",
					Status: "running",
				}
				mockUsecase.On("GetJob", mock.Anything, jobID).Return(expectedJob, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				data := response["data"].(map[string]interface{})
				assert.Equal(t, jobID.String(), data["id"])
				assert.Equal(t, "test-job", data["name"])
			},
		},
		{
			name:  "invalid job ID",
			jobID: "invalid-uuid",
			mockSetup: func(mockUsecase *MockJobUsecase) {
				// No mock calls expected
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid job ID")
			},
		},
		{
			name:  "job not found",
			jobID: jobID.String(),
			mockSetup: func(mockUsecase *MockJobUsecase) {
				mockUsecase.On("GetJob", mock.Anything, jobID).Return(nil, errors.New("job not found"))
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "job not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := new(MockJobUsecase)
			tt.mockSetup(mockUsecase)

			handler := handler.NewJobHandler(mockUsecase, nil, nil, nil)
			router := setupTestRouter()
			router.GET("/jobs/:id", handler.GetJob)

			req, err := http.NewRequest("GET", "/jobs/"+tt.jobID, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestJobHandler_GetAllJobs(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockJobUsecase, *MockJobEnrichmentService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful jobs retrieval",
			mockSetup: func(mockUsecase *MockJobUsecase, mockEnrichment *MockJobEnrichmentService) {
				jobs := []domain.Job{
					{
						ID:     uuid.New(),
						Name:   "job-1",
						Status: "pending",
					},
					{
						ID:     uuid.New(),
						Name:   "job-2",
						Status: "running",
					},
				}
				enrichedJobs := []domain.EnrichedJob{
					{
						Job: jobs[0],
					},
					{
						Job: jobs[1],
					},
				}
				mockUsecase.On("GetAllJobs", mock.Anything).Return(jobs, nil)
				mockEnrichment.On("EnrichJobs", mock.Anything, jobs).Return(enrichedJobs, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				data := response["data"].([]interface{})
				assert.Len(t, data, 2)
				job1 := data[0].(map[string]interface{})
				job2 := data[1].(map[string]interface{})
				assert.Equal(t, "job-1", job1["name"])
				assert.Equal(t, "job-2", job2["name"])
			},
		},
		{
			name: "no jobs found",
			mockSetup: func(mockUsecase *MockJobUsecase, mockEnrichment *MockJobEnrichmentService) {
				jobs := []domain.Job{}
				enrichedJobs := []domain.EnrichedJob{}
				mockUsecase.On("GetAllJobs", mock.Anything).Return(jobs, nil)
				mockEnrichment.On("EnrichJobs", mock.Anything, jobs).Return(enrichedJobs, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				data := response["data"].([]interface{})
				assert.Len(t, data, 0)
			},
		},
		{
			name: "usecase error",
			mockSetup: func(mockUsecase *MockJobUsecase, mockEnrichment *MockJobEnrichmentService) {
				mockUsecase.On("GetAllJobs", mock.Anything).Return([]domain.Job{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "database error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := new(MockJobUsecase)
			mockEnrichment := new(MockJobEnrichmentService)
			tt.mockSetup(mockUsecase, mockEnrichment)

			handler := handler.NewJobHandler(mockUsecase, mockEnrichment, nil, nil)
			router := setupTestRouter()
			router.GET("/jobs", handler.GetAllJobs)

			req, err := http.NewRequest("GET", "/jobs", nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
			mockUsecase.AssertExpectations(t)
			mockEnrichment.AssertExpectations(t)
		})
	}
}

func TestJobHandler_StartJob(t *testing.T) {
	jobID := uuid.New()

	tests := []struct {
		name           string
		jobID          string
		mockSetup      func(*MockJobUsecase)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:  "successful job start",
			jobID: jobID.String(),
			mockSetup: func(mockUsecase *MockJobUsecase) {
				mockUsecase.On("StartJob", mock.Anything, jobID).Return(nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Job started successfully", response["message"])
			},
		},
		{
			name:  "invalid job ID",
			jobID: "invalid-uuid",
			mockSetup: func(mockUsecase *MockJobUsecase) {
				// No mock calls expected
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid job ID")
			},
		},
		{
			name:  "usecase error",
			jobID: jobID.String(),
			mockSetup: func(mockUsecase *MockJobUsecase) {
				mockUsecase.On("StartJob", mock.Anything, jobID).Return(errors.New("job already running"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "job already running")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := new(MockJobUsecase)
			tt.mockSetup(mockUsecase)

			handler := handler.NewJobHandler(mockUsecase, nil, nil, nil)
			router := setupTestRouter()
			router.POST("/jobs/:id/start", handler.StartJob)

			req, err := http.NewRequest("POST", "/jobs/"+tt.jobID+"/start", nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestJobHandler_DeleteJob(t *testing.T) {
	jobID := uuid.New()

	tests := []struct {
		name           string
		jobID          string
		mockSetup      func(*MockJobUsecase)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:  "successful job deletion",
			jobID: jobID.String(),
			mockSetup: func(mockUsecase *MockJobUsecase) {
				mockUsecase.On("DeleteJob", mock.Anything, jobID).Return(nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Job deleted successfully", response["message"])
			},
		},
		{
			name:  "invalid job ID",
			jobID: "invalid-uuid",
			mockSetup: func(mockUsecase *MockJobUsecase) {
				// No mock calls expected
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid job ID")
			},
		},
		{
			name:  "job not found",
			jobID: jobID.String(),
			mockSetup: func(mockUsecase *MockJobUsecase) {
				mockUsecase.On("DeleteJob", mock.Anything, jobID).Return(errors.New("job not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "job not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := new(MockJobUsecase)
			tt.mockSetup(mockUsecase)

			handler := handler.NewJobHandler(mockUsecase, nil, nil, nil)
			router := setupTestRouter()
			router.DELETE("/jobs/:id", handler.DeleteJob)

			req, err := http.NewRequest("DELETE", "/jobs/"+tt.jobID, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
			mockUsecase.AssertExpectations(t)
		})
	}
}
