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
	"go-distributed-hashcat/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAgentUsecase is a mock implementation of domain.AgentUsecase
type MockAgentUsecase struct {
	mock.Mock
}

func (m *MockAgentUsecase) RegisterAgent(ctx context.Context, req *domain.CreateAgentRequest) (*domain.Agent, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Agent), args.Error(1)
}

func (m *MockAgentUsecase) GetAgent(ctx context.Context, id uuid.UUID) (*domain.Agent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Agent), args.Error(1)
}

func (m *MockAgentUsecase) GetAllAgents(ctx context.Context) ([]domain.Agent, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Agent), args.Error(1)
}

func (m *MockAgentUsecase) UpdateAgentStatus(ctx context.Context, id uuid.UUID, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockAgentUsecase) DeleteAgent(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAgentUsecase) GetAvailableAgent(ctx context.Context) (*domain.Agent, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Agent), args.Error(1)
}

func (m *MockAgentUsecase) UpdateAgentHeartbeat(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAgentUsecase) UpdateAgentLastSeen(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAgentUsecase) GetByAgentKey(ctx context.Context, agentKey string) (*domain.Agent, error) {
	args := m.Called(ctx, agentKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Agent), args.Error(1)
}

func (m *MockAgentUsecase) SetWebSocketHub(wsHub usecase.WebSocketHub) {
	m.Called(wsHub)
}

func (m *MockAgentUsecase) ValidateUniqueIPForAgentKey(ctx context.Context, agentKey, ipAddress, agentName string) error {
	args := m.Called(ctx, agentKey, ipAddress, agentName)
	return args.Error(0)
}

func (m *MockAgentUsecase) GetByNameAndIP(ctx context.Context, name, ip string, port int) (*domain.Agent, error) {
	args := m.Called(ctx, name, ip, port)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Agent), args.Error(1)
}

func (m *MockAgentUsecase) CreateAgent(ctx context.Context, agent *domain.Agent) error {
	args := m.Called(ctx, agent)
	return args.Error(0)
}

func (m *MockAgentUsecase) UpdateAgent(ctx context.Context, agent *domain.Agent) error {
	args := m.Called(ctx, agent)
	return args.Error(0)
}

func (m *MockAgentUsecase) GenerateAgentKey(ctx context.Context, name string) (*domain.Agent, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Agent), args.Error(1)
}

func (m *MockAgentUsecase) UpdateAgentData(ctx context.Context, agentKey string, ipAddress string, port int, capabilities string) error {
	args := m.Called(ctx, agentKey, ipAddress, port, capabilities)
	return args.Error(0)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestAgentHandler_CreateAgent(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAgentUsecase)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful agent creation",
			requestBody: map[string]interface{}{
				"name":         "test-agent",
				"ip_address":   "192.168.1.100",
				"port":         8080,
				"capabilities": "gpu,cpu",
				"agent_key":    "test_key_123",
			},
			mockSetup: func(mockUsecase *MockAgentUsecase) {
				// Setup GetByAgentKey mock for agent key validation
				existingAgent := &domain.Agent{
					ID:           uuid.New(),
					Name:         "test-agent",
					IPAddress:    "",
					Port:         0,
					Capabilities: "",
					Status:       "offline",
					AgentKey:     "test_key_123",
				}
				mockUsecase.On("GetByAgentKey", mock.Anything, "test_key_123").Return(existingAgent, nil)

				expectedAgent := &domain.Agent{
					ID:           uuid.New(),
					Name:         "test-agent",
					IPAddress:    "192.168.1.100",
					Port:         8080,
					Capabilities: "gpu,cpu",
					Status:       "offline", // Default status saat create
				}
				mockUsecase.On("RegisterAgent", mock.Anything, mock.AnythingOfType("*domain.CreateAgentRequest")).Return(expectedAgent, nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				data := response["data"].(map[string]interface{})
				assert.Equal(t, "test-agent", data["name"])
				assert.Equal(t, "192.168.1.100", data["ip_address"])
				assert.Equal(t, float64(8080), data["port"])
				assert.Equal(t, "offline", data["status"]) // Default status saat create
			},
		},
		{
			name:        "invalid JSON body",
			requestBody: "invalid-json",
			mockSetup: func(mockUsecase *MockAgentUsecase) {
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
				"name": "test-agent",
				// Missing agent_key (required), ip_address, port
			},
			mockSetup: func(mockUsecase *MockAgentUsecase) {
				// No mock calls expected - validation should fail before reaching usecase
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Field validation for 'AgentKey' failed on the 'required' tag")
			},
		},
		{
			name: "usecase error",
			requestBody: map[string]interface{}{
				"name":         "test-agent",
				"ip_address":   "192.168.1.100",
				"port":         8080,
				"capabilities": "gpu,cpu",
				"agent_key":    "test_key_123",
			},
			mockSetup: func(mockUsecase *MockAgentUsecase) {
				// Setup GetByAgentKey mock for agent key validation
				existingAgent := &domain.Agent{
					ID:           uuid.New(),
					Name:         "test-agent",
					IPAddress:    "",
					Port:         0,
					Capabilities: "",
					Status:       "offline",
					AgentKey:     "test_key_123",
				}
				mockUsecase.On("GetByAgentKey", mock.Anything, "test_key_123").Return(existingAgent, nil)

				mockUsecase.On("RegisterAgent", mock.Anything, mock.AnythingOfType("*domain.CreateAgentRequest")).Return(nil, errors.New("database error"))
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
			mockUsecase := new(MockAgentUsecase)
			tt.mockSetup(mockUsecase)

			handler := handler.NewAgentHandler(mockUsecase)
			router := setupTestRouter()
			router.POST("/agents", handler.RegisterAgent)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req, err := http.NewRequest("POST", "/agents", bytes.NewReader(body))
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

func TestAgentHandler_GetAgent(t *testing.T) {
	agentID := uuid.New()

	tests := []struct {
		name           string
		agentID        string
		mockSetup      func(*MockAgentUsecase)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:    "successful agent retrieval",
			agentID: agentID.String(),
			mockSetup: func(mockUsecase *MockAgentUsecase) {
				expectedAgent := &domain.Agent{
					ID:        agentID,
					Name:      "test-agent",
					IPAddress: "192.168.1.100",
					Port:      8080,
					Status:    "online",
				}
				mockUsecase.On("GetAgent", mock.Anything, agentID).Return(expectedAgent, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				data := response["data"].(map[string]interface{})
				assert.Equal(t, agentID.String(), data["id"])
				assert.Equal(t, "test-agent", data["name"])
			},
		},
		{
			name:    "invalid agent ID",
			agentID: "invalid-uuid",
			mockSetup: func(mockUsecase *MockAgentUsecase) {
				// No mock calls expected
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid agent ID")
			},
		},
		{
			name:    "agent not found",
			agentID: agentID.String(),
			mockSetup: func(mockUsecase *MockAgentUsecase) {
				mockUsecase.On("GetAgent", mock.Anything, agentID).Return(nil, errors.New("agent not found"))
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "agent not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := new(MockAgentUsecase)
			tt.mockSetup(mockUsecase)

			handler := handler.NewAgentHandler(mockUsecase)
			router := setupTestRouter()
			router.GET("/agents/:id", handler.GetAgent)

			req, err := http.NewRequest("GET", "/agents/"+tt.agentID, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestAgentHandler_GetAllAgents(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockAgentUsecase)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful agents retrieval",
			mockSetup: func(mockUsecase *MockAgentUsecase) {
				agents := []domain.Agent{
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
				mockUsecase.On("GetAllAgents", mock.Anything).Return(agents, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				data := response["data"].([]interface{})
				assert.Len(t, data, 2)
				agent1 := data[0].(map[string]interface{})
				agent2 := data[1].(map[string]interface{})
				assert.Equal(t, "agent-1", agent1["name"])
				assert.Equal(t, "agent-2", agent2["name"])
			},
		},
		{
			name: "no agents found",
			mockSetup: func(mockUsecase *MockAgentUsecase) {
				mockUsecase.On("GetAllAgents", mock.Anything).Return([]domain.Agent{}, nil)
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
			mockSetup: func(mockUsecase *MockAgentUsecase) {
				mockUsecase.On("GetAllAgents", mock.Anything).Return([]domain.Agent{}, errors.New("database error"))
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
			mockUsecase := new(MockAgentUsecase)
			tt.mockSetup(mockUsecase)

			handler := handler.NewAgentHandler(mockUsecase)
			router := setupTestRouter()
			router.GET("/agents", handler.GetAllAgents)

			req, err := http.NewRequest("GET", "/agents", nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestAgentHandler_UpdateAgentStatus(t *testing.T) {
	agentID := uuid.New()

	tests := []struct {
		name           string
		agentID        string
		requestBody    interface{}
		mockSetup      func(*MockAgentUsecase)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:    "successful status update",
			agentID: agentID.String(),
			requestBody: map[string]interface{}{
				"status": "offline",
			},
			mockSetup: func(mockUsecase *MockAgentUsecase) {
				mockUsecase.On("UpdateAgentStatus", mock.Anything, agentID, "offline").Return(nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Agent status updated successfully", response["message"])
			},
		},
		{
			name:    "invalid agent ID",
			agentID: "invalid-uuid",
			requestBody: map[string]interface{}{
				"status": "offline",
			},
			mockSetup: func(mockUsecase *MockAgentUsecase) {
				// No mock calls expected
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid agent ID")
			},
		},
		{
			name:        "missing status field",
			agentID:     agentID.String(),
			requestBody: map[string]interface{}{
				// Missing status
			},
			mockSetup: func(mockUsecase *MockAgentUsecase) {
				// No mock calls expected
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid request body")
			},
		},
		{
			name:    "usecase error",
			agentID: agentID.String(),
			requestBody: map[string]interface{}{
				"status": "offline",
			},
			mockSetup: func(mockUsecase *MockAgentUsecase) {
				mockUsecase.On("UpdateAgentStatus", mock.Anything, agentID, "offline").Return(errors.New("update failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Failed to update agent status")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := new(MockAgentUsecase)
			tt.mockSetup(mockUsecase)

			handler := handler.NewAgentHandler(mockUsecase)
			router := setupTestRouter()
			router.PUT("/agents/:id/status", handler.UpdateAgentStatus)

			body, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			req, err := http.NewRequest("PUT", "/agents/"+tt.agentID+"/status", bytes.NewReader(body))
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

func TestAgentHandler_DeleteAgent(t *testing.T) {
	agentID := uuid.New()

	tests := []struct {
		name           string
		agentID        string
		mockSetup      func(*MockAgentUsecase)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:    "successful agent deletion",
			agentID: agentID.String(),
			mockSetup: func(mockUsecase *MockAgentUsecase) {
				mockUsecase.On("DeleteAgent", mock.Anything, agentID).Return(nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Agent deleted successfully", response["message"])
			},
		},
		{
			name:    "invalid agent ID",
			agentID: "invalid-uuid",
			mockSetup: func(mockUsecase *MockAgentUsecase) {
				// No mock calls expected
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid agent ID")
			},
		},
		{
			name:    "agent not found",
			agentID: agentID.String(),
			mockSetup: func(mockUsecase *MockAgentUsecase) {
				mockUsecase.On("DeleteAgent", mock.Anything, agentID).Return(errors.New("agent not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Failed to delete agent")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := new(MockAgentUsecase)
			tt.mockSetup(mockUsecase)

			handler := handler.NewAgentHandler(mockUsecase)
			router := setupTestRouter()
			router.DELETE("/agents/:id", handler.DeleteAgent)

			req, err := http.NewRequest("DELETE", "/agents/"+tt.agentID, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestAgentHandler_HeartbeatAgent(t *testing.T) {
	agentID := uuid.New()

	tests := []struct {
		name           string
		agentID        string
		mockSetup      func(*MockAgentUsecase)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:    "successful heartbeat",
			agentID: agentID.String(),
			mockSetup: func(mockUsecase *MockAgentUsecase) {
				mockUsecase.On("UpdateAgentLastSeen", mock.Anything, agentID).Return(nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "heartbeat updated", response["message"])
			},
		},
		{
			name:    "invalid agent ID",
			agentID: "invalid-uuid",
			mockSetup: func(mockUsecase *MockAgentUsecase) {
				// No mock calls expected
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "invalid agent ID")
			},
		},
		{
			name:    "usecase error",
			agentID: agentID.String(),
			mockSetup: func(mockUsecase *MockAgentUsecase) {
				mockUsecase.On("UpdateAgentLastSeen", mock.Anything, agentID).Return(errors.New("heartbeat failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "heartbeat failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := new(MockAgentUsecase)
			tt.mockSetup(mockUsecase)

			handler := handler.NewAgentHandler(mockUsecase)
			router := setupTestRouter()
			router.POST("/agents/:id/heartbeat", handler.UpdateAgentHeartbeat)

			req, err := http.NewRequest("POST", "/agents/"+tt.agentID+"/heartbeat", nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
			mockUsecase.AssertExpectations(t)
		})
	}
}
