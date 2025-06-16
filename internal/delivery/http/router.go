package http

import (
	"os"
	"time"

	"go-distributed-hashcat/internal/delivery/http/handler"
	"go-distributed-hashcat/internal/delivery/http/middleware"
	"go-distributed-hashcat/internal/usecase"

	"github.com/gin-gonic/gin"
)

func NewRouter(
	agentUsecase usecase.AgentUsecase,
	jobUsecase usecase.JobUsecase,
	hashFileUsecase usecase.HashFileUsecase,
	wordlistUsecase usecase.WordlistUsecase,
	jobEnrichmentService usecase.JobEnrichmentService,
	authUsecase usecase.AuthUsecase,
	agentKeyUsecase usecase.AgentKeyUsecase,
) *gin.Engine {
	// Set Gin to release mode for production performance
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// CORS middleware (must be first to handle preflight requests)
	// In development, allow frontend origin specifically
	if env := os.Getenv("GIN_MODE"); env == "debug" {
		router.Use(middleware.CORSWithSpecificOrigin("http://localhost:3000"))
	} else {
		router.Use(middleware.CORS())
	}

	// Performance middleware
	router.Use(middleware.Performance())
	router.Use(middleware.Gzip())
	router.Use(middleware.Cache())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.RequestTimeout(30 * time.Second))

	// Standard middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Initialize handlers
	agentHandler := handler.NewAgentHandler(agentUsecase)
	jobHandler := handler.NewJobHandler(jobUsecase, jobEnrichmentService)
	hashFileHandler := handler.NewHashFileHandler(hashFileUsecase)
	wordlistHandler := handler.NewWordlistHandler(wordlistUsecase)
	cacheHandler := handler.NewCacheHandler(jobEnrichmentService)
	wsHandler := handler.NewWebSocketHandler()
	authHandler := handler.NewAuthHandler(authUsecase)
	agentKeyHandler := handler.NewAgentKeyHandler(agentKeyUsecase)

	// Serve modern frontend (production build)
	router.Static("/assets", "./frontend/dist/assets")
	router.StaticFile("/", "./frontend/dist/index.html")
	router.StaticFile("/index.html", "./frontend/dist/index.html")

	// WebSocket endpoint
	router.GET("/ws", wsHandler.HandleWebSocket)

	// Health check (optimized) - Public
	router.OPTIONS("/health", func(c *gin.Context) {
		c.Status(204)
	})
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")

	// Handle all OPTIONS requests for API v1
	v1.OPTIONS("/*any", func(c *gin.Context) {
		c.Status(204)
	})

	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/me", middleware.AuthMiddleware(authUsecase), authHandler.GetCurrentUser)
			auth.POST("/change-password", middleware.AuthMiddleware(authUsecase), authHandler.ChangePassword)
		}

		// Agent key management routes (public for dashboard)
		agentKeys := v1.Group("/agent-keys")
		{
			agentKeys.POST("/generate", agentKeyHandler.GenerateAgentKey)
			agentKeys.GET("/", agentKeyHandler.ListAgentKeys)
			agentKeys.DELETE("/:key/revoke", agentKeyHandler.RevokeAgentKey)
			agentKeys.DELETE("/:key", agentKeyHandler.DeleteAgentKey)
		}

		// Agent routes (require agent key)
		agents := v1.Group("/agents")
		agents.Use(middleware.AgentKeyMiddleware(agentKeyUsecase))
		{
			agents.POST("/", agentHandler.RegisterAgent)
			agents.PUT("/:id/status", agentHandler.UpdateAgentStatus)
			agents.POST("/:id/heartbeat", agentHandler.Heartbeat)
			agents.POST("/:id/files", agentHandler.RegisterAgentFiles)
			agents.GET("/:id/jobs/next", jobHandler.GetAvailableJobForAgent)
		}

		// Agent management routes (public for dashboard)
		agentMgmt := v1.Group("/agents")
		{
			agentMgmt.GET("/", agentHandler.GetAllAgents)
			agentMgmt.GET("/:id", agentHandler.GetAgent)
			agentMgmt.GET("/:id/jobs", jobHandler.GetJobsByAgentID)
			agentMgmt.DELETE("/:id", agentHandler.DeleteAgent)
		}

		// Job routes
		jobs := v1.Group("/jobs")
		{
			jobs.POST("/", jobHandler.CreateJob)
			jobs.GET("/", jobHandler.GetAllJobs)
			jobs.GET("/:id", jobHandler.GetJob)
			jobs.POST("/:id/start", jobHandler.StartJob)
			jobs.PUT("/:id/progress", jobHandler.UpdateJobProgress)
			jobs.POST("/:id/complete", jobHandler.CompleteJob)
			jobs.POST("/:id/fail", jobHandler.FailJob)
			jobs.POST("/:id/pause", jobHandler.PauseJob)
			jobs.POST("/:id/resume", jobHandler.ResumeJob)
			jobs.POST("/:id/stop", jobHandler.StopJob)
			jobs.DELETE("/:id", jobHandler.DeleteJob)
			jobs.POST("/assign", jobHandler.AssignJobs)
		}

		// Hash file routes
		hashFiles := v1.Group("/hashfiles")
		{
			hashFiles.POST("/upload", hashFileHandler.UploadHashFile)
			hashFiles.GET("/", hashFileHandler.GetAllHashFiles)
			hashFiles.GET("/:id", hashFileHandler.GetHashFile)
			hashFiles.GET("/:id/download", hashFileHandler.DownloadHashFile)
			hashFiles.DELETE("/:id", hashFileHandler.DeleteHashFile)
		}

		// Wordlist routes
		wordlists := v1.Group("/wordlists")
		{
			wordlists.POST("/upload", wordlistHandler.UploadWordlist)
			wordlists.GET("/", wordlistHandler.GetAllWordlists)
			wordlists.GET("/:id", wordlistHandler.GetWordlist)
			wordlists.GET("/:id/download", wordlistHandler.DownloadWordlist)
			wordlists.DELETE("/:id", wordlistHandler.DeleteWordlist)
		}

		// Cache management routes
		cache := v1.Group("/cache")
		{
			cache.GET("/stats", cacheHandler.GetCacheStats)
			cache.DELETE("/clear", cacheHandler.ClearCache)
		}
	}

	// Legacy API routes for backward compatibility
	api := router.Group("/api")
	{
		// Legacy routes (without v1 prefix)
		api.GET("/agents", agentHandler.GetAllAgents)
		api.GET("/jobs", jobHandler.GetAllJobs)
		api.GET("/hash-files", hashFileHandler.GetAllHashFiles)
		api.GET("/wordlists", wordlistHandler.GetAllWordlists)

		// Legacy upload routes
		api.POST("/wordlists/upload", wordlistHandler.UploadWordlist)
		api.GET("/wordlists/:id/download", wordlistHandler.DownloadWordlist)
		api.DELETE("/wordlists/:id", wordlistHandler.DeleteWordlist)
	}

	return router
}
