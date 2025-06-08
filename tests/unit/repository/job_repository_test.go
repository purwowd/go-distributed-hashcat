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

type JobRepositoryTestSuite struct {
	suite.Suite
	db   *database.SQLiteDB
	repo domain.JobRepository
}

func (suite *JobRepositoryTestSuite) SetupTest() {
	// Create in-memory SQLite database for testing
	db, err := database.NewSQLiteDB(":memory:")
	suite.Require().NoError(err)

	suite.db = db
	suite.repo = repository.NewJobRepository(db)
}

func (suite *JobRepositoryTestSuite) TearDownTest() {
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *JobRepositoryTestSuite) TestCreate() {
	hashFileID := uuid.New()
	agentID := uuid.New()

	job := &domain.Job{
		ID:         uuid.New(),
		Name:       "Test Job",
		Status:     "pending",
		HashType:   2500,
		AttackMode: 0,
		HashFile:   "/tmp/test.hash",
		HashFileID: &hashFileID,
		Wordlist:   "rockyou.txt",
		AgentID:    &agentID,
		Progress:   0.0,
		Speed:      0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := suite.repo.Create(context.Background(), job)
	assert.NoError(suite.T(), err)

	// Verify job was created
	retrievedJob, err := suite.repo.GetByID(context.Background(), job.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedJob)
	assert.Equal(suite.T(), job.Name, retrievedJob.Name)
	assert.Equal(suite.T(), job.Status, retrievedJob.Status)
	assert.Equal(suite.T(), job.HashType, retrievedJob.HashType)
	assert.Equal(suite.T(), job.AttackMode, retrievedJob.AttackMode)
}

func (suite *JobRepositoryTestSuite) TestGetByID() {
	// Create a job first
	job := &domain.Job{
		ID:         uuid.New(),
		Name:       "Test Job",
		Status:     "pending",
		HashType:   2500,
		AttackMode: 0,
		HashFile:   "/tmp/test.hash",
		Wordlist:   "rockyou.txt",
		Progress:   0.0,
		Speed:      0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := suite.repo.Create(context.Background(), job)
	suite.Require().NoError(err)

	// Test successful retrieval
	retrievedJob, err := suite.repo.GetByID(context.Background(), job.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedJob)
	assert.Equal(suite.T(), job.ID, retrievedJob.ID)
	assert.Equal(suite.T(), job.Name, retrievedJob.Name)

	// Test non-existent job
	nonExistentID := uuid.New()
	retrievedJob, err = suite.repo.GetByID(context.Background(), nonExistentID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), retrievedJob)
}

func (suite *JobRepositoryTestSuite) TestGetAll() {
	// Initially should be empty
	jobs, err := suite.repo.GetAll(context.Background())
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), jobs)

	// Create multiple jobs
	job1 := &domain.Job{
		ID:         uuid.New(),
		Name:       "Job 1",
		Status:     "pending",
		HashType:   2500,
		AttackMode: 0,
		HashFile:   "/tmp/test1.hash",
		Wordlist:   "rockyou.txt",
		Progress:   0.0,
		Speed:      0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	job2 := &domain.Job{
		ID:         uuid.New(),
		Name:       "Job 2",
		Status:     "running",
		HashType:   1000,
		AttackMode: 3,
		HashFile:   "/tmp/test2.hash",
		Wordlist:   "custom.txt",
		Progress:   25.5,
		Speed:      1000000,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err = suite.repo.Create(context.Background(), job1)
	suite.Require().NoError(err)
	err = suite.repo.Create(context.Background(), job2)
	suite.Require().NoError(err)

	// Test retrieval
	jobs, err = suite.repo.GetAll(context.Background())
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), jobs, 2)

	// Verify we got both jobs
	jobNames := []string{jobs[0].Name, jobs[1].Name}
	assert.Contains(suite.T(), jobNames, "Job 1")
	assert.Contains(suite.T(), jobNames, "Job 2")
}

func (suite *JobRepositoryTestSuite) TestGetByStatus() {
	// Create jobs with different statuses
	pendingJob := &domain.Job{
		ID:         uuid.New(),
		Name:       "Pending Job",
		Status:     "pending",
		HashType:   2500,
		AttackMode: 0,
		HashFile:   "/tmp/pending.hash",
		Wordlist:   "rockyou.txt",
		Progress:   0.0,
		Speed:      0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	runningJob := &domain.Job{
		ID:         uuid.New(),
		Name:       "Running Job",
		Status:     "running",
		HashType:   1000,
		AttackMode: 3,
		HashFile:   "/tmp/running.hash",
		Wordlist:   "custom.txt",
		Progress:   50.0,
		Speed:      1500000,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := suite.repo.Create(context.Background(), pendingJob)
	suite.Require().NoError(err)
	err = suite.repo.Create(context.Background(), runningJob)
	suite.Require().NoError(err)

	// Test getting pending jobs
	pendingJobs, err := suite.repo.GetByStatus(context.Background(), "pending")
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), pendingJobs, 1)
	assert.Equal(suite.T(), "Pending Job", pendingJobs[0].Name)

	// Test getting running jobs
	runningJobs, err := suite.repo.GetByStatus(context.Background(), "running")
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), runningJobs, 1)
	assert.Equal(suite.T(), "Running Job", runningJobs[0].Name)

	// Test getting non-existent status
	completedJobs, err := suite.repo.GetByStatus(context.Background(), "completed")
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), completedJobs)
}

func (suite *JobRepositoryTestSuite) TestUpdateStatus() {
	// Create a job first
	job := &domain.Job{
		ID:         uuid.New(),
		Name:       "Test Job",
		Status:     "pending",
		HashType:   2500,
		AttackMode: 0,
		HashFile:   "/tmp/test.hash",
		Wordlist:   "rockyou.txt",
		Progress:   0.0,
		Speed:      0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := suite.repo.Create(context.Background(), job)
	suite.Require().NoError(err)

	// Update status
	newStatus := "running"
	err = suite.repo.UpdateStatus(context.Background(), job.ID, newStatus)
	assert.NoError(suite.T(), err)

	// Verify status was updated
	retrievedJob, err := suite.repo.GetByID(context.Background(), job.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newStatus, retrievedJob.Status)
}

func (suite *JobRepositoryTestSuite) TestUpdateProgress() {
	// Create a job first
	job := &domain.Job{
		ID:         uuid.New(),
		Name:       "Test Job",
		Status:     "running",
		HashType:   2500,
		AttackMode: 0,
		HashFile:   "/tmp/test.hash",
		Wordlist:   "rockyou.txt",
		Progress:   0.0,
		Speed:      0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := suite.repo.Create(context.Background(), job)
	suite.Require().NoError(err)

	// Update progress
	newProgress := 75.5
	newSpeed := int64(2000000)
	err = suite.repo.UpdateProgress(context.Background(), job.ID, newProgress, newSpeed)
	assert.NoError(suite.T(), err)

	// Verify progress was updated
	retrievedJob, err := suite.repo.GetByID(context.Background(), job.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newProgress, retrievedJob.Progress)
	assert.Equal(suite.T(), newSpeed, retrievedJob.Speed)
}

func (suite *JobRepositoryTestSuite) TestDelete() {
	// Create a job first
	job := &domain.Job{
		ID:         uuid.New(),
		Name:       "Test Job",
		Status:     "pending",
		HashType:   2500,
		AttackMode: 0,
		HashFile:   "/tmp/test.hash",
		Wordlist:   "rockyou.txt",
		Progress:   0.0,
		Speed:      0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := suite.repo.Create(context.Background(), job)
	suite.Require().NoError(err)

	// Verify job exists
	retrievedJob, err := suite.repo.GetByID(context.Background(), job.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedJob)

	// Delete job
	err = suite.repo.Delete(context.Background(), job.ID)
	assert.NoError(suite.T(), err)

	// Verify job no longer exists
	retrievedJob, err = suite.repo.GetByID(context.Background(), job.ID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), retrievedJob)
}

func (suite *JobRepositoryTestSuite) TestGetByAgentID() {
	agentID1 := uuid.New()
	agentID2 := uuid.New()

	// Create jobs assigned to different agents
	job1 := &domain.Job{
		ID:         uuid.New(),
		Name:       "Agent 1 Job",
		Status:     "running",
		HashType:   2500,
		AttackMode: 0,
		HashFile:   "/tmp/agent1.hash",
		Wordlist:   "rockyou.txt",
		AgentID:    &agentID1,
		Progress:   30.0,
		Speed:      1000000,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	job2 := &domain.Job{
		ID:         uuid.New(),
		Name:       "Agent 2 Job",
		Status:     "pending",
		HashType:   1000,
		AttackMode: 3,
		HashFile:   "/tmp/agent2.hash",
		Wordlist:   "custom.txt",
		AgentID:    &agentID2,
		Progress:   0.0,
		Speed:      0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Create a job with no agent assigned
	job3 := &domain.Job{
		ID:         uuid.New(),
		Name:       "Unassigned Job",
		Status:     "pending",
		HashType:   2500,
		AttackMode: 0,
		HashFile:   "/tmp/unassigned.hash",
		Wordlist:   "rockyou.txt",
		AgentID:    nil,
		Progress:   0.0,
		Speed:      0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := suite.repo.Create(context.Background(), job1)
	suite.Require().NoError(err)
	err = suite.repo.Create(context.Background(), job2)
	suite.Require().NoError(err)
	err = suite.repo.Create(context.Background(), job3)
	suite.Require().NoError(err)

	// Test getting jobs for agent 1
	agent1Jobs, err := suite.repo.GetByAgentID(context.Background(), agentID1)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), agent1Jobs, 1)
	assert.Equal(suite.T(), "Agent 1 Job", agent1Jobs[0].Name)

	// Test getting jobs for agent 2
	agent2Jobs, err := suite.repo.GetByAgentID(context.Background(), agentID2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), agent2Jobs, 1)
	assert.Equal(suite.T(), "Agent 2 Job", agent2Jobs[0].Name)

	// Test getting jobs for non-existent agent
	nonExistentAgentID := uuid.New()
	noJobs, err := suite.repo.GetByAgentID(context.Background(), nonExistentAgentID)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), noJobs)
}

func TestJobRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(JobRepositoryTestSuite))
}
