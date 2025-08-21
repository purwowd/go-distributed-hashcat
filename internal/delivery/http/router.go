package http

import (
	"net/http"
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
) *gin.Engine {
	// Set Gin to release mode for production performance
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// CORS middleware (must be first to handle preflight requests)
	// In development, allow frontend origin specifically
	if env := os.Getenv("GIN_MODE"); env == "debug" {
		router.Use(middleware.CORSWithSpecificOrigin("http://30.30.30.39:3000"))
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
	jobHandler := handler.NewJobHandler(jobUsecase, jobEnrichmentService, agentUsecase, wordlistUsecase)
	hashFileHandler := handler.NewHashFileHandler(hashFileUsecase)
	wordlistHandler := handler.NewWordlistHandler(wordlistUsecase)
	cacheHandler := handler.NewCacheHandler(jobEnrichmentService)
	wsHandler := handler.NewWebSocketHandler()

	// Initialize distributed job handler (placeholder for now)
	// distributedJobHandler := handler.NewDistributedJobHandler(distributedJobUsecase)

	// Serve modern frontend (production build)
	router.Static("/assets", "./frontend/dist/assets")
	router.StaticFile("/", "./frontend/dist/index.html")
	router.StaticFile("/index.html", "./frontend/dist/index.html")

	// WebSocket endpoint
	router.GET("/ws", wsHandler.HandleWebSocket)

	// Health check (optimized)
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
		// Agent routes
		agents := v1.Group("/agents")
		{
			agents.POST("/generate-key", agentHandler.GenerateAgentKey) // New route for generating agent keys
			agents.POST("/startup", agentHandler.AgentStartup)          // New route for agent startup
			agents.POST("/heartbeat", agentHandler.AgentHeartbeat)      // New route for agent heartbeat
			agents.POST("/update-data", agentHandler.UpdateAgentData)   // New route for updating agent data (no status change)
			agents.POST("/", agentHandler.RegisterAgent)
			agents.GET("/", agentHandler.GetAllAgents)
			agents.GET("/:id", agentHandler.GetAgent)
			agents.PUT("/:id/status", agentHandler.UpdateAgentStatus)
			agents.PUT("/:id/heartbeat", agentHandler.UpdateAgentHeartbeat)
			agents.POST("/:id/files", agentHandler.RegisterAgentFiles)
			agents.GET("/:id/jobs", jobHandler.GetJobsByAgentID)
			agents.GET("/:id/jobs/next", jobHandler.GetAvailableJobForAgent)
			agents.DELETE("/:id", agentHandler.DeleteAgent)
		}

		// Job routes
		jobs := v1.Group("/jobs")
		{
			jobs.POST("/", jobHandler.CreateJob)
			jobs.GET("/", jobHandler.GetAllJobs)
			jobs.GET("/parallel/summary", jobHandler.GetParallelJobsSummary)
			jobs.POST("/assign", jobHandler.AssignJobs)
			jobs.POST("/auto", jobHandler.CreateParallelJobs)
			jobs.GET("/agent/:id", jobHandler.GetAvailableJobForAgent)
			jobs.GET("/:id", jobHandler.GetJob)
			jobs.POST("/:id/start", jobHandler.StartJob)
			jobs.PUT("/:id/progress", jobHandler.UpdateJobProgress)
			jobs.PUT("/:id/data", jobHandler.UpdateJobDataFromAgent)
			jobs.POST("/:id/complete", jobHandler.CompleteJob)
			jobs.POST("/:id/fail", jobHandler.FailJob)
			jobs.POST("/:id/pause", jobHandler.PauseJob)
			jobs.POST("/:id/resume", jobHandler.ResumeJob)
			jobs.POST("/:id/stop", jobHandler.StopJob)
			jobs.DELETE("/:id", jobHandler.DeleteJob)
		}

		// Distributed Job routes
		distributedJobs := v1.Group("/distributed-jobs")
		{
			distributedJobs.POST("/", func(c *gin.Context) {
				c.JSON(http.StatusNotImplemented, gin.H{
					"success": false,
					"error":   "Distributed jobs feature not yet implemented",
				})
			})
			distributedJobs.GET("/:id/status", func(c *gin.Context) {
				c.JSON(http.StatusNotImplemented, gin.H{
					"success": false,
					"error":   "Distributed jobs feature not yet implemented",
				})
			})
			distributedJobs.POST("/:id/start-all", func(c *gin.Context) {
				c.JSON(http.StatusNotImplemented, gin.H{
					"success": false,
					"error":   "Distributed jobs feature not yet implemented",
				})
			})
			distributedJobs.GET("/performance", func(c *gin.Context) {
				c.JSON(http.StatusNotImplemented, gin.H{
					"success": false,
					"error":   "Distributed jobs feature not yet implemented",
				})
			})
			distributedJobs.GET("/preview", func(c *gin.Context) {
				c.JSON(http.StatusNotImplemented, gin.H{
					"success": false,
					"error":   "Distributed jobs feature not yet implemented",
				})
			})
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
			wordlists.GET("/:id/content", wordlistHandler.GetWordlistContent)
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
