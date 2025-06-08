package repository_test

import (
	"context"
	"testing"
	"time"

	"go-distributed-hashcat/internal/domain"
	"go-distributed-hashcat/internal/infrastructure/database"
	"go-distributed-hashcat/internal/infrastructure/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AgentRepositoryTestSuite struct {
	suite.Suite
	db   *database.SQLiteDB
	repo domain.AgentRepository
}

func (suite *AgentRepositoryTestSuite) SetupTest() {
	// Create in-memory SQLite database for testing
	db, err := database.NewSQLiteDB(":memory:")
	suite.Require().NoError(err)

	suite.db = db
	suite.repo = repository.NewAgentRepository(db)
}

func (suite *AgentRepositoryTestSuite) TearDownTest() {
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *AgentRepositoryTestSuite) TestCreate() {
	agent := &domain.Agent{
		ID:           uuid.New(),
		Name:         "Test Agent",
		IPAddress:    "192.168.1.100",
		Port:         8080,
		Status:       "online",
		Capabilities: "NVIDIA RTX 4090",
		LastSeen:     time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := suite.repo.Create(context.Background(), agent)
	assert.NoError(suite.T(), err)

	// Verify agent was created
	retrievedAgent, err := suite.repo.GetByID(context.Background(), agent.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedAgent)
	assert.Equal(suite.T(), agent.Name, retrievedAgent.Name)
	assert.Equal(suite.T(), agent.IPAddress, retrievedAgent.IPAddress)
	assert.Equal(suite.T(), agent.Port, retrievedAgent.Port)
	assert.Equal(suite.T(), agent.Status, retrievedAgent.Status)
}

func (suite *AgentRepositoryTestSuite) TestGetByID() {
	// Create an agent first
	agent := &domain.Agent{
		ID:           uuid.New(),
		Name:         "Test Agent",
		IPAddress:    "192.168.1.100",
		Port:         8080,
		Status:       "online",
		Capabilities: "NVIDIA RTX 4090",
		LastSeen:     time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := suite.repo.Create(context.Background(), agent)
	suite.Require().NoError(err)

	// Test successful retrieval
	retrievedAgent, err := suite.repo.GetByID(context.Background(), agent.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedAgent)
	assert.Equal(suite.T(), agent.ID, retrievedAgent.ID)
	assert.Equal(suite.T(), agent.Name, retrievedAgent.Name)

	// Test non-existent agent
	nonExistentID := uuid.New()
	retrievedAgent, err = suite.repo.GetByID(context.Background(), nonExistentID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), retrievedAgent)
}

func (suite *AgentRepositoryTestSuite) TestGetAll() {
	// Initially should be empty
	agents, err := suite.repo.GetAll(context.Background())
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), agents)

	// Create multiple agents
	agent1 := &domain.Agent{
		ID:           uuid.New(),
		Name:         "Agent 1",
		IPAddress:    "192.168.1.100",
		Port:         8080,
		Status:       "online",
		Capabilities: "GPU 1",
		LastSeen:     time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	agent2 := &domain.Agent{
		ID:           uuid.New(),
		Name:         "Agent 2",
		IPAddress:    "192.168.1.101",
		Port:         8081,
		Status:       "offline",
		Capabilities: "GPU 2",
		LastSeen:     time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = suite.repo.Create(context.Background(), agent1)
	suite.Require().NoError(err)
	err = suite.repo.Create(context.Background(), agent2)
	suite.Require().NoError(err)

	// Test retrieval
	agents, err = suite.repo.GetAll(context.Background())
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), agents, 2)

	// Verify we got both agents
	agentNames := []string{agents[0].Name, agents[1].Name}
	assert.Contains(suite.T(), agentNames, "Agent 1")
	assert.Contains(suite.T(), agentNames, "Agent 2")
}

func (suite *AgentRepositoryTestSuite) TestUpdateStatus() {
	// Create an agent first
	agent := &domain.Agent{
		ID:           uuid.New(),
		Name:         "Test Agent",
		IPAddress:    "192.168.1.100",
		Port:         8080,
		Status:       "online",
		Capabilities: "NVIDIA RTX 4090",
		LastSeen:     time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := suite.repo.Create(context.Background(), agent)
	suite.Require().NoError(err)

	// Update status
	newStatus := "busy"
	err = suite.repo.UpdateStatus(context.Background(), agent.ID, newStatus)
	assert.NoError(suite.T(), err)

	// Verify status was updated
	retrievedAgent, err := suite.repo.GetByID(context.Background(), agent.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newStatus, retrievedAgent.Status)
}

func (suite *AgentRepositoryTestSuite) TestUpdateLastSeen() {
	// Create an agent first
	originalTime := time.Now().Add(-1 * time.Hour) // 1 hour ago
	agent := &domain.Agent{
		ID:           uuid.New(),
		Name:         "Test Agent",
		IPAddress:    "192.168.1.100",
		Port:         8080,
		Status:       "online",
		Capabilities: "NVIDIA RTX 4090",
		LastSeen:     originalTime,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := suite.repo.Create(context.Background(), agent)
	suite.Require().NoError(err)

	// Update last seen
	err = suite.repo.UpdateLastSeen(context.Background(), agent.ID)
	assert.NoError(suite.T(), err)

	// Verify last seen was updated (should be more recent)
	retrievedAgent, err := suite.repo.GetByID(context.Background(), agent.ID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), retrievedAgent.LastSeen.After(originalTime))
}

func (suite *AgentRepositoryTestSuite) TestDelete() {
	// Create an agent first
	agent := &domain.Agent{
		ID:           uuid.New(),
		Name:         "Test Agent",
		IPAddress:    "192.168.1.100",
		Port:         8080,
		Status:       "online",
		Capabilities: "NVIDIA RTX 4090",
		LastSeen:     time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := suite.repo.Create(context.Background(), agent)
	suite.Require().NoError(err)

	// Verify agent exists
	retrievedAgent, err := suite.repo.GetByID(context.Background(), agent.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedAgent)

	// Delete agent
	err = suite.repo.Delete(context.Background(), agent.ID)
	assert.NoError(suite.T(), err)

	// Verify agent no longer exists
	retrievedAgent, err = suite.repo.GetByID(context.Background(), agent.ID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), retrievedAgent)
}

func (suite *AgentRepositoryTestSuite) TestUpdate() {
	// Create an agent first
	agent := &domain.Agent{
		ID:           uuid.New(),
		Name:         "Test Agent",
		IPAddress:    "192.168.1.100",
		Port:         8080,
		Status:       "online",
		Capabilities: "NVIDIA RTX 4090",
		LastSeen:     time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := suite.repo.Create(context.Background(), agent)
	suite.Require().NoError(err)

	// Update agent
	agent.Name = "Updated Agent"
	agent.Capabilities = "NVIDIA RTX 4090 Ti"
	agent.Status = "busy"
	agent.UpdatedAt = time.Now()

	err = suite.repo.Update(context.Background(), agent)
	assert.NoError(suite.T(), err)

	// Verify updates
	retrievedAgent, err := suite.repo.GetByID(context.Background(), agent.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Agent", retrievedAgent.Name)
	assert.Equal(suite.T(), "NVIDIA RTX 4090 Ti", retrievedAgent.Capabilities)
	assert.Equal(suite.T(), "busy", retrievedAgent.Status)
}

func TestAgentRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(AgentRepositoryTestSuite))
}
