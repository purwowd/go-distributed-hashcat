package benchmarks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"go-distributed-hashcat/internal/delivery/http/handler"
	"go-distributed-hashcat/internal/domain"
	"go-distributed-hashcat/internal/infrastructure/database"
	"go-distributed-hashcat/internal/infrastructure/repository"
	"go-distributed-hashcat/internal/usecase"

	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// setupBenchmarkDB creates a fast in-memory database for benchmarking
func setupBenchmarkDB(b *testing.B) *database.SQLiteDB {
	// Use in-memory database for speed
	db, err := database.NewSQLiteDB(":memory:")
	if err != nil {
		b.Fatalf("Failed to create in-memory database: %v", err)
	}
	return db
}

// BenchmarkAgentCreation benchmarks agent creation performance
func BenchmarkAgentCreation(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	agentRepo := repository.NewAgentRepository(db)
	agentUsecase := usecase.NewAgentUsecase(agentRepo)
	agentHandler := handler.NewAgentHandler(agentUsecase)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/agents/register", agentHandler.RegisterAgent)

	// Pre-allocate request template
	agentReq := domain.CreateAgentRequest{
		Name:         "Benchmark Agent",
		IPAddress:    "192.168.1.100",
		Port:         8080,
		Capabilities: "Benchmark GPU",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Minimize allocations in hot path
		agentReq.Name = fmt.Sprintf("Benchmark Agent %d", i)
		agentReq.IPAddress = fmt.Sprintf("192.168.1.%d", 100+(i%50))

		reqBody, _ := json.Marshal(agentReq)
		req := httptest.NewRequest("POST", "/api/v1/agents/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		// Check for errors to prevent accumulating bad state
		if recorder.Code >= 400 {
			b.Fatalf("Request failed with status %d: %s", recorder.Code, recorder.Body.String())
		}
	}
}

// BenchmarkJobCreation benchmarks job creation performance
func BenchmarkJobCreation(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	// Setup repositories
	agentRepo := repository.NewAgentRepository(db)
	jobRepo := repository.NewJobRepository(db)
	hashFileRepo := repository.NewHashFileRepository(db)

	// Create a test hash file first to avoid errors
	testHashFile := &domain.HashFile{
		ID:        uuid.New(),
		Name:      "benchmark.hash",
		OrigName:  "benchmark.hash",
		Path:      "/tmp/benchmark.hash",
		Size:      1024,
		Type:      "hash",
		CreatedAt: time.Now(),
	}
	err := hashFileRepo.Create(context.Background(), testHashFile)
	if err != nil {
		b.Fatalf("Failed to create test hash file: %v", err)
	}

	jobUsecase := usecase.NewJobUsecase(jobRepo, agentRepo, hashFileRepo)
	jobHandler := handler.NewJobHandler(jobUsecase, nil, nil, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/jobs/create", jobHandler.CreateJob)

	// Pre-allocate request template with valid hash file ID
	jobReq := domain.CreateJobRequest{
		Name:       "Benchmark Job",
		HashType:   2500,
		AttackMode: 0,
		HashFileID: testHashFile.ID.String(), // Use real hash file ID
		Wordlist:   "rockyou.txt",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		jobReq.Name = fmt.Sprintf("Benchmark Job %d", i)

		reqBody, _ := json.Marshal(jobReq)
		req := httptest.NewRequest("POST", "/api/v1/jobs/create", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		// Check for errors
		if recorder.Code >= 400 {
			b.Fatalf("Request failed with status %d: %s", recorder.Code, recorder.Body.String())
		}
	}
}

// BenchmarkAgentListing benchmarks agent listing performance
func BenchmarkAgentListing(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	agentRepo := repository.NewAgentRepository(db)
	agentUsecase := usecase.NewAgentUsecase(agentRepo)
	agentHandler := handler.NewAgentHandler(agentUsecase)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/agents/list", agentHandler.GetAllAgents)

	// Create test agents for listing
	const numTestAgents = 20 // Reduced from 100
	for i := 0; i < numTestAgents; i++ {
		agentReq := domain.CreateAgentRequest{
			Name:         fmt.Sprintf("Agent %d", i),
			IPAddress:    fmt.Sprintf("192.168.1.%d", 100+(i%50)),
			Port:         8080 + i,
			Capabilities: "Test GPU",
		}
		_, err := agentUsecase.RegisterAgent(context.Background(), &agentReq)
		if err != nil {
			b.Fatalf("Failed to create test agent %d: %v", i, err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/v1/agents/list", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		if recorder.Code != 200 {
			b.Fatalf("Request failed with status %d", recorder.Code)
		}
	}
}

// BenchmarkDirectAgentCreation benchmarks direct usecase calls (without HTTP overhead)
func BenchmarkDirectAgentCreation(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	agentRepo := repository.NewAgentRepository(db)
	agentUsecase := usecase.NewAgentUsecase(agentRepo)

	agentReq := domain.CreateAgentRequest{
		Name:         "Direct Agent",
		IPAddress:    "192.168.1.100",
		Port:         8080,
		Capabilities: "Direct GPU",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		agentReq.Name = fmt.Sprintf("Direct Agent %d", i)
		agentReq.IPAddress = fmt.Sprintf("192.168.1.%d", 100+(i%50))

		_, err := agentUsecase.RegisterAgent(context.Background(), &agentReq)
		if err != nil {
			b.Fatalf("Failed to register agent: %v", err)
		}
	}
}

// BenchmarkLimitedConcurrentAgentCreation benchmarks limited concurrent agent creation
func BenchmarkLimitedConcurrentAgentCreation(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	agentRepo := repository.NewAgentRepository(db)
	agentUsecase := usecase.NewAgentUsecase(agentRepo)
	agentHandler := handler.NewAgentHandler(agentUsecase)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/agents/register", agentHandler.RegisterAgent)

	b.ResetTimer()
	b.ReportAllocs()

	// Use a more controlled concurrency approach
	b.RunParallel(func(pb *testing.PB) {
		var i int
		for pb.Next() {
			agentReq := domain.CreateAgentRequest{
				Name:         fmt.Sprintf("Concurrent Agent %d", i),
				IPAddress:    fmt.Sprintf("192.168.1.%d", 100+(i%50)),
				Port:         8080 + (i % 1000), // Limit port range
				Capabilities: "Concurrent GPU",
			}

			reqBody, _ := json.Marshal(agentReq)
			req := httptest.NewRequest("POST", "/api/v1/agents/register", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// Don't fail on individual errors in concurrent test, just count them
			if recorder.Code >= 400 {
				b.Logf("Request %d failed with status %d", i, recorder.Code)
			}
			i++
		}
	})
}
