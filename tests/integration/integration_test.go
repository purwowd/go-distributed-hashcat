package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"go-distributed-hashcat/internal/delivery/http/handler"
	"go-distributed-hashcat/internal/domain"
	"go-distributed-hashcat/internal/infrastructure/database"
	"go-distributed-hashcat/internal/infrastructure/repository"
	"go-distributed-hashcat/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// APITestSuite contains integration tests for the entire API
type APITestSuite struct {
	suite.Suite
	router       *gin.Engine
	agentHandler *handler.AgentHandler
	jobHandler   *handler.JobHandler
	db           *database.SQLiteDB
	cleanup      func()
	// Add fields to store test data between methods
	agent1ID string
	agent2ID string
}

// SetupSuite initializes the test suite with real database and dependencies
func (suite *APITestSuite) SetupSuite() {
	// Setup is done per test for better isolation
}

// SetupTest creates a fresh database for each test
func (suite *APITestSuite) SetupTest() {
	// Create unique temporary database for each test
	tempDB := fmt.Sprintf("/tmp/test_hashcat_%d.db", time.Now().UnixNano())
	os.Remove(tempDB) // Clean up if exists

	// Initialize database
	db, err := database.NewSQLiteDB(tempDB)
	suite.Require().NoError(err)
	suite.db = db

	// Store cleanup function
	suite.cleanup = func() {
		if suite.db != nil {
			suite.db.Close()
		}
		os.Remove(tempDB)
	}

	// Initialize repositories
	agentRepo := repository.NewAgentRepository(db)
	jobRepo := repository.NewJobRepository(db)
	hashFileRepo := repository.NewHashFileRepository(db)
	wordlistRepo := repository.NewWordlistRepository(db)

	// Initialize use cases
	agentUsecase := usecase.NewAgentUsecase(agentRepo)
	jobUsecase := usecase.NewJobUsecase(jobRepo, agentRepo, hashFileRepo)

	// Initialize enrichment service for integration tests
	jobEnrichmentService := usecase.NewJobEnrichmentService(agentRepo, wordlistRepo, hashFileRepo)

	// Initialize handlers
	suite.agentHandler = handler.NewAgentHandler(agentUsecase)
	suite.jobHandler = handler.NewJobHandler(jobUsecase, jobEnrichmentService, nil, nil)

	// Setup router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	suite.setupRoutes()
}

// TearDownTest cleans up after each test
func (suite *APITestSuite) TearDownTest() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

// TearDownSuite cleans up after all tests
func (suite *APITestSuite) TearDownSuite() {
	// Nothing to do here since cleanup is per test
}

// setupRoutes configures all API routes
func (suite *APITestSuite) setupRoutes() {
	api := suite.router.Group("/api/v1")
	{
		// Agent routes
		agents := api.Group("/agents")
		{
			agents.POST("/", suite.agentHandler.RegisterAgent)
			agents.GET("/", suite.agentHandler.GetAllAgents)
			agents.GET("/:id", suite.agentHandler.GetAgent)
			agents.PUT("/:id/status", suite.agentHandler.UpdateAgentStatus)
			agents.DELETE("/:id", suite.agentHandler.DeleteAgent)
			agents.POST("/:id/heartbeat", suite.agentHandler.UpdateAgentHeartbeat)
		}

		// Job routes
		jobs := api.Group("/jobs")
		{
			jobs.POST("/", suite.jobHandler.CreateJob)
			jobs.GET("/", suite.jobHandler.GetAllJobs)
			jobs.GET("/:id", suite.jobHandler.GetJob)
			jobs.DELETE("/:id", suite.jobHandler.DeleteJob)
		}
	}

	// Health endpoint
	suite.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})
}

// makeRequest helper function to make HTTP requests
func (suite *APITestSuite) makeRequest(method, url string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)

	return recorder
}

// TestHealthEndpoint tests the health check endpoint
func (suite *APITestSuite) TestHealthEndpoint() {
	recorder := suite.makeRequest("GET", "/health", nil)

	assert.Equal(suite.T(), http.StatusOK, recorder.Code)

	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "healthy", response["status"])
	assert.Contains(suite.T(), response, "timestamp")
}

// TestAgentWorkflow tests the complete agent management workflow
func (suite *APITestSuite) TestAgentWorkflow() {
	t := suite.T()

	// 1. Initially no agents
	recorder := suite.makeRequest("GET", "/api/v1/agents/", nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var listResponse map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &listResponse)
	agents := listResponse["data"].([]interface{})
	assert.Len(t, agents, 0)

	// 2. Create agent keys first (this would normally be done by admin)
	agentKey1 := suite.createTestAgentKey("GPU-Agent-01")
	agentKey2 := suite.createTestAgentKey("CPU-Agent-02")

	// 3. Create first agent
	agentReq1 := domain.CreateAgentRequest{
		Name:         "GPU-Agent-01",
		IPAddress:    "192.168.1.100",
		Port:         8080,
		Capabilities: "NVIDIA RTX 4090, 24GB VRAM",
		AgentKey:     agentKey1,
	}

	recorder = suite.makeRequest("POST", "/api/v1/agents/", agentReq1)
	assert.Equal(t, http.StatusCreated, recorder.Code)

	var createResponse map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &createResponse)
	agent1Data := createResponse["data"].(map[string]interface{})
	suite.agent1ID = agent1Data["id"].(string)

	assert.Equal(t, agentReq1.Name, agent1Data["name"])
	assert.Equal(t, agentReq1.IPAddress, agent1Data["ip_address"])
	assert.Equal(t, float64(agentReq1.Port), agent1Data["port"])
	assert.Equal(t, "offline", agent1Data["status"]) // Default status is offline

	// 4. Create second agent
	agentReq2 := domain.CreateAgentRequest{
		Name:         "CPU-Agent-02",
		IPAddress:    "192.168.1.101",
		Port:         8081,
		Capabilities: "64-core AMD EPYC, 256GB RAM",
		AgentKey:     agentKey2,
	}

	recorder = suite.makeRequest("POST", "/api/v1/agents/", agentReq2)
	assert.Equal(t, http.StatusCreated, recorder.Code)

	json.Unmarshal(recorder.Body.Bytes(), &createResponse)
	agent2Data := createResponse["data"].(map[string]interface{})
	suite.agent2ID = agent2Data["id"].(string)

	// 5. List all agents (should have 2)
	recorder = suite.makeRequest("GET", "/api/v1/agents/", nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	json.Unmarshal(recorder.Body.Bytes(), &listResponse)
	agents = listResponse["data"].([]interface{})
	assert.Len(t, agents, 2)

	// 6. Get specific agent
	recorder = suite.makeRequest("GET", fmt.Sprintf("/api/v1/agents/%s", suite.agent1ID), nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var getResponse map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &getResponse)
	agentData := getResponse["data"].(map[string]interface{})
	assert.Equal(t, suite.agent1ID, agentData["id"])
	assert.Equal(t, agentReq1.Name, agentData["name"])

	// 7. Update agent status
	statusUpdate := map[string]string{"status": "busy"}
	recorder = suite.makeRequest("PUT", fmt.Sprintf("/api/v1/agents/%s/status", suite.agent1ID), statusUpdate)
	assert.Equal(t, http.StatusOK, recorder.Code)

	// 8. Send heartbeat
	recorder = suite.makeRequest("POST", fmt.Sprintf("/api/v1/agents/%s/heartbeat", suite.agent1ID), nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	// 9. Invalid agent ID should return 400
	recorder = suite.makeRequest("GET", "/api/v1/agents/invalid-uuid", nil)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// createTestAgentKey creates a test agent key in the database for integration tests
func (suite *APITestSuite) createTestAgentKey(name string) string {
	agentRepo := repository.NewAgentRepository(suite.db)

	// Create an agent key entry (this would normally be done by admin)
	agentKey := &domain.Agent{
		ID:           uuid.New(),
		Name:         name,
		IPAddress:    "",
		Port:         0,
		Status:       "offline",
		Capabilities: "",
		AgentKey:     fmt.Sprintf("test_key_%s_%d", name, time.Now().UnixNano()),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := agentRepo.Create(context.Background(), agentKey)
	suite.Require().NoError(err)

	return agentKey.AgentKey
}

// createTestHashFile creates a test hash file for integration tests
func (suite *APITestSuite) createTestHashFile() *domain.HashFile {
	hashFileRepo := repository.NewHashFileRepository(suite.db)

	hashFile := &domain.HashFile{
		ID:        uuid.New(),
		Name:      "test.hash",
		OrigName:  "test.hash",
		Path:      "/tmp/test.hash",
		Size:      1024,
		Type:      "hash",
		CreatedAt: time.Now(),
	}

	err := hashFileRepo.Create(context.Background(), hashFile)
	suite.Require().NoError(err)

	return hashFile
}

// TestJobWorkflow tests the complete job management workflow
func (suite *APITestSuite) TestJobWorkflow() {
	t := suite.T()

	// Prerequisites: Need agents first
	suite.TestAgentWorkflow()

	// Create test hash files
	hashFile1 := suite.createTestHashFile()
	hashFile2 := suite.createTestHashFile()

	// 1. Initially no jobs
	recorder := suite.makeRequest("GET", "/api/v1/jobs/", nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var listResponse map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &listResponse)
	jobs := listResponse["data"].([]interface{})
	assert.Len(t, jobs, 0)

	// 2. Create job with auto-assignment (no agent_id specified)
	jobReq1 := domain.CreateJobRequest{
		Name:       "Auto-Assign Hash Job",
		HashType:   2500,                  // WPA/WPA2
		AttackMode: 0,                     // Dictionary attack
		HashFileID: hashFile1.ID.String(), // Use real hash file ID
		Wordlist:   "rockyou.txt",
		// WordlistID is optional, so we'll omit it
	}

	recorder = suite.makeRequest("POST", "/api/v1/jobs/", jobReq1)
	assert.Equal(t, http.StatusCreated, recorder.Code)

	var createResponse map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &createResponse)
	job1Data := createResponse["data"].(map[string]interface{})
	job1ID := job1Data["id"].(string)

	assert.Equal(t, jobReq1.Name, job1Data["name"])
	assert.Equal(t, "pending", job1Data["status"])
	assert.Equal(t, float64(jobReq1.HashType), job1Data["hash_type"])
	assert.Equal(t, float64(jobReq1.AttackMode), job1Data["attack_mode"])

	// 3. Create job with manual agent assignment
	jobReq2 := domain.CreateJobRequest{
		Name:       "Manual-Assign Hash Job",
		HashType:   1000,                  // NTLM
		AttackMode: 3,                     // Brute force
		HashFileID: hashFile2.ID.String(), // Use real hash file ID
		Wordlist:   "custom.txt",
		// WordlistID is optional, so we'll omit it
		AgentID: suite.agent2ID, // Manual assignment to agent2 (not agent1)
	}

	recorder = suite.makeRequest("POST", "/api/v1/jobs/", jobReq2)
	assert.Equal(t, http.StatusCreated, recorder.Code)

	json.Unmarshal(recorder.Body.Bytes(), &createResponse)
	job2Data := createResponse["data"].(map[string]interface{})

	assert.Equal(t, jobReq2.Name, job2Data["name"])
	// Should be assigned to agent2
	if job2Data["agent_id"] != nil {
		assert.Equal(t, suite.agent2ID, job2Data["agent_id"])
	}

	// 4. List all jobs (should have 2)
	recorder = suite.makeRequest("GET", "/api/v1/jobs/", nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	json.Unmarshal(recorder.Body.Bytes(), &listResponse)
	jobs = listResponse["data"].([]interface{})
	assert.Len(t, jobs, 2)

	// 5. Get specific job
	recorder = suite.makeRequest("GET", fmt.Sprintf("/api/v1/jobs/%s", job1ID), nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var getResponse map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &getResponse)
	jobData := getResponse["data"].(map[string]interface{})
	assert.Equal(t, job1ID, jobData["id"])
	assert.Equal(t, jobReq1.Name, jobData["name"])

	// 6. Invalid job ID should return 400
	recorder = suite.makeRequest("GET", "/api/v1/jobs/invalid-uuid", nil)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestAgentSelectionFeature specifically tests the new agent selection feature
func (suite *APITestSuite) TestAgentSelectionFeature() {
	t := suite.T()

	// Setup: Create agents first
	suite.TestAgentWorkflow()

	// Create test hash files
	hashFile1 := suite.createTestHashFile()
	hashFile2 := suite.createTestHashFile()

	// Test 1: Job without agent_id (auto-assignment)
	autoJobReq := domain.CreateJobRequest{
		Name:       "Auto-Assignment Test",
		HashType:   2500,
		AttackMode: 0,
		HashFileID: hashFile1.ID.String(), // Use real hash file ID
		Wordlist:   "rockyou.txt",
		// No AgentID specified - should auto-assign
	}

	recorder := suite.makeRequest("POST", "/api/v1/jobs/", autoJobReq)
	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)
	jobData := response["data"].(map[string]interface{})

	assert.Equal(t, autoJobReq.Name, jobData["name"])
	assert.Equal(t, "pending", jobData["status"])
	// Auto-assignment logic will handle agent assignment

	// Test 2: Job with specific agent_id (manual assignment)
	manualJobReq := domain.CreateJobRequest{
		Name:       "Manual-Assignment Test",
		HashType:   1000,
		AttackMode: 3,
		HashFileID: hashFile2.ID.String(), // Use real hash file ID
		Wordlist:   "custom.txt",
		AgentID:    suite.agent2ID, // Specifically assign to agent2
	}

	recorder = suite.makeRequest("POST", "/api/v1/jobs/", manualJobReq)
	assert.Equal(t, http.StatusCreated, recorder.Code)

	json.Unmarshal(recorder.Body.Bytes(), &response)
	jobData = response["data"].(map[string]interface{})

	assert.Equal(t, manualJobReq.Name, jobData["name"])
	assert.Equal(t, "pending", jobData["status"])
	// Should be assigned to agent2 if available
	if jobData["agent_id"] != nil {
		assert.Equal(t, suite.agent2ID, jobData["agent_id"])
	}

	// Test 3: Job with invalid agent_id (should handle gracefully)
	invalidAgentJobReq := domain.CreateJobRequest{
		Name:       "Invalid-Agent Test",
		HashType:   2500,
		AttackMode: 0,
		HashFileID: hashFile2.ID.String(), // Use real hash file ID
		Wordlist:   "rockyou.txt",
		AgentID:    "invalid-uuid-format",
	}

	recorder = suite.makeRequest("POST", "/api/v1/jobs/", invalidAgentJobReq)
	// Should return error or handle gracefully
	assert.True(t, recorder.Code >= 400)
}

// TestErrorHandling tests various error scenarios
func (suite *APITestSuite) TestErrorHandling() {
	t := suite.T()

	// Test invalid JSON
	req := httptest.NewRequest("POST", "/api/v1/agents/", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	suite.router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	// Test missing required fields
	incompleteReq := map[string]interface{}{
		"name": "", // Empty name should fail validation
	}
	recorder = suite.makeRequest("POST", "/api/v1/agents/", incompleteReq)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	// Test non-existent resource
	recorder = suite.makeRequest("GET", fmt.Sprintf("/api/v1/agents/%s", uuid.New()), nil)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

// TestConcurrentRequests tests API under concurrent load
func (suite *APITestSuite) TestConcurrentRequests() {
	t := suite.T()

	// Create agent keys first for concurrent tests
	agentKeys := make([]string, 5)
	for i := 0; i < 5; i++ {
		agentKeys[i] = suite.createTestAgentKey(fmt.Sprintf("Concurrent-Agent-%d", i))
	}

	// Create multiple agents concurrently
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func(index int) {
			agentReq := domain.CreateAgentRequest{
				Name:         fmt.Sprintf("Concurrent-Agent-%d", index),
				IPAddress:    fmt.Sprintf("192.168.1.%d", 200+index),
				Port:         8080 + index,
				Capabilities: fmt.Sprintf("Test GPU %d", index),
				AgentKey:     agentKeys[index],
			}

			recorder := suite.makeRequest("POST", "/api/v1/agents/", agentReq)
			assert.Equal(t, http.StatusCreated, recorder.Code)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify all agents were created
	recorder := suite.makeRequest("GET", "/api/v1/agents/", nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)
	agents := response["data"].([]interface{})
	assert.GreaterOrEqual(t, len(agents), 5)
}

// TestAPIIntegration runs the integration test suite
func TestAPIIntegration(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
