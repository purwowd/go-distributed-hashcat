package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	httpDelivery "go-distributed-hashcat/internal/delivery/http"
	"go-distributed-hashcat/internal/delivery/http/handler"
	"go-distributed-hashcat/internal/infrastructure/database"
	"go-distributed-hashcat/internal/infrastructure/repository"
	"go-distributed-hashcat/internal/usecase"
	"go-distributed-hashcat/internal/domain"
	"sort"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var migrationsDir = "./internal/infrastructure/database/migrations"

func main() {
	Execute()
}

// Config struct with extended database support
type Config struct {
	Server struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"server"`
	Database struct {
		Type     string `mapstructure:"type"`     // sqlite, postgres, mysql
		Path     string `mapstructure:"path"`     // For SQLite
		Host     string `mapstructure:"host"`     // For PostgreSQL/MySQL
		Port     int    `mapstructure:"port"`     // For PostgreSQL/MySQL
		Name     string `mapstructure:"name"`     // Database name
		User     string `mapstructure:"user"`     // Username
		Password string `mapstructure:"password"` // Password
	} `mapstructure:"database"`
	Upload struct {
		Directory string `mapstructure:"directory"`
	} `mapstructure:"upload"`
}

// Load configuration with .env support
func loadConfig() *Config {
	// Configure environment variable bindings first
	viper.AutomaticEnv()

	// Map nested config to environment variables (works with all config types)
	viper.BindEnv("server.port", "HASHCAT_SERVER_PORT", "PORT")
	viper.BindEnv("database.path", "HASHCAT_DATABASE_PATH", "DB_PATH")
	viper.BindEnv("database.type", "HASHCAT_DATABASE_TYPE", "DB_TYPE")
	viper.BindEnv("database.host", "HASHCAT_DATABASE_HOST", "DB_HOST")
	viper.BindEnv("database.port", "HASHCAT_DATABASE_PORT", "DB_PORT")
	viper.BindEnv("database.name", "HASHCAT_DATABASE_NAME", "DB_NAME")
	viper.BindEnv("database.user", "HASHCAT_DATABASE_USER", "DB_USER")
	viper.BindEnv("database.password", "HASHCAT_DATABASE_PASSWORD", "DB_PASSWORD")
	viper.BindEnv("upload.directory", "HASHCAT_UPLOAD_DIRECTORY", "UPLOAD_DIR")

	// Set defaults
	viper.SetDefault("server.port", 1337)
	viper.SetDefault("database.path", "./data/hashcat.db")
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("upload.directory", "./uploads")

	// Try to load .env file first
	viper.SetConfigName(".env")
	viper.SetConfigType("dotenv") // Use dotenv instead of env
	viper.AddConfigPath(".")

	envErr := viper.ReadInConfig()

	// If .env not found, try YAML fallback
	if envErr != nil {
		log.Printf("No .env file found, trying YAML config: %v", envErr)

		// Reset viper for YAML
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./configs")

		if yamlErr := viper.ReadInConfig(); yamlErr != nil {
			if _, ok := yamlErr.(viper.ConfigFileNotFoundError); ok {
				log.Println("No config file found, using defaults and environment variables")
			} else {
				log.Fatalf("Error reading config file: %v", yamlErr)
			}
		} else {
			log.Printf("Using YAML config: %s", viper.ConfigFileUsed())
		}
	} else {
		log.Printf("Using .env config: %s", viper.ConfigFileUsed())
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode config: %v", err)
	}

	// Log config source for debugging with actual values
	log.Printf("Configuration loaded - Server: %d, Database: %s (%s), Upload: %s",
		config.Server.Port, config.Database.Path, config.Database.Type, config.Upload.Directory)

	return &config
}

// Root command
var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "Distributed Hashcat Server",
	Long:  `A distributed password cracking system using hashcat.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default: start the server
		startServer()
	},
}

// Migration commands
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration commands",
	Long:  `Manage database migrations - generate, run, rollback, and check status.`,
}

var migrateGenerateCmd = &cobra.Command{
	Use:   "generate [name]",
	Short: "Generate a new migration file",
	Long: `Generate a new migration file with timestamp-based version.

Example:
  ./server migrate generate "add job templates table"
  ./server migrate generate "add_index_to_jobs"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		migrationName := args[0]

		// Initialize database connection (just for migration runner)
		config := loadConfig()
		db, err := database.NewSQLiteDB(config.Database.Path)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		// Create migration runner
		runner := database.NewMigrationRunner(db.DB(), migrationsDir)

		// Generate migration
		if err := runner.GenerateMigration(migrationName); err != nil {
			log.Fatalf("Failed to generate migration: %v", err)
		}

		fmt.Printf("🎉 Migration generated successfully!\n")
		fmt.Printf("📝 Edit the file in: %s\n", migrationsDir)
		fmt.Printf("💡 Add your SQL to the UP and DOWN sections\n")
		fmt.Printf("🚀 Run 'migrate up' when ready to apply\n")
	},
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Run pending migrations",
	Long: `Run all pending database migrations.

Example:
  ./server migrate up`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize database connection
		config := loadConfig()
		db, err := database.NewSQLiteDB(config.Database.Path)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		// Create migration runner
		runner := database.NewMigrationRunner(db.DB(), migrationsDir)

		// Run migrations
		if err := runner.MigrateUp(); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback the last migration",
	Long: `Rollback the most recently applied migration.

Example:
  ./server migrate down`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize database connection
		config := loadConfig()
		db, err := database.NewSQLiteDB(config.Database.Path)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		// Create migration runner
		runner := database.NewMigrationRunner(db.DB(), migrationsDir)

		// Rollback migration
		if err := runner.MigrateDown(); err != nil {
			log.Fatalf("Failed to rollback migration: %v", err)
		}
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Long: `Show the status of all migrations - which are applied and which are pending.

Example:
  ./server migrate status`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize database connection
		config := loadConfig()
		db, err := database.NewSQLiteDB(config.Database.Path)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		// Create migration runner
		runner := database.NewMigrationRunner(db.DB(), migrationsDir)

		// Show status
		if err := runner.Status(); err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}
	},
}

// Wordlist upload commands
var wordlistCmd = &cobra.Command{
	Use:   "wordlist",
	Short: "Wordlist management commands",
	Long:  `Manage wordlist files - upload, list, and delete wordlists.`,
}

var wordlistUploadCmd = &cobra.Command{
	Use:   "upload [path]",
	Short: "Upload a wordlist file with optimized processing",
	Long: `Upload a wordlist file with optimized processing for large files (1M+ words).

This command provides fast processing for large wordlist files by:
- Using buffered I/O for efficient file reading
- Parallel word counting for files with 1M+ words
- Progress reporting during upload
- Automatic file validation and optimization

Parameters:
  path      Path to the wordlist file to upload
  --name    Custom name for the wordlist (optional)
  --count   Enable word counting (default: true)
  --chunk   Chunk size for processing in MB (default: 10)

Examples:
  ./server wordlist upload /path/to/rockyou.txt
  ./server wordlist upload /path/to/wordlist.txt --name "custom_name"
  ./server wordlist upload /path/to/large.txt --chunk 50 --count=false`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		customName, _ := cmd.Flags().GetString("name")
		enableCount, _ := cmd.Flags().GetBool("count")
		chunkSize, _ := cmd.Flags().GetInt("chunk")

		// Initialize database connection
		config := loadConfig()
		db, err := database.NewSQLiteDB(config.Database.Path)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		// Run migrations automatically
		runner := database.NewMigrationRunner(db.DB(), migrationsDir)
		if err := runner.MigrateUp(); err != nil {
			log.Printf("Warning: Failed to run migrations: %v", err)
		}

		// Initialize repositories and usecase
		wordlistRepo := repository.NewWordlistRepository(db)
		wordlistUsecase := usecase.NewWordlistUsecase(wordlistRepo, config.Upload.Directory)

		// Upload wordlist with CLI optimization
		if err := uploadWordlistCLI(wordlistUsecase, filePath, customName, enableCount, chunkSize); err != nil {
			if strings.Contains(err.Error(), "file already exists") {
				log.Printf("Error: %v", err)
				log.Printf("Use a different name with --name flag or delete the existing wordlist first")
				os.Exit(1)
			}
			log.Fatalf("Failed to upload wordlist: %v", err)
		}
	},
}

var wordlistListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all uploaded wordlists",
	Long: `List all uploaded wordlists with their details including size, word count, and creation date.

Examples:
  ./server wordlist list`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize database connection
		config := loadConfig()
		db, err := database.NewSQLiteDB(config.Database.Path)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		// Run migrations automatically
		runner := database.NewMigrationRunner(db.DB(), migrationsDir)
		if err := runner.MigrateUp(); err != nil {
			log.Printf("Warning: Failed to run migrations: %v", err)
		}

		// Initialize repositories and usecase
		wordlistRepo := repository.NewWordlistRepository(db)
		wordlistUsecase := usecase.NewWordlistUsecase(wordlistRepo, config.Upload.Directory)

		// List wordlists
		if err := listWordlistsCLI(wordlistUsecase); err != nil {
			log.Fatalf("Failed to list wordlists: %v", err)
		}
	},
}

var wordlistDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a wordlist by ID",
	Long: `Delete a wordlist by its UUID. This will remove both the database record and the physical file.

Examples:
  ./server wordlist delete 123e4567-e89b-12d3-a456-426614174000`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		wordlistID := args[0]

		// Initialize database connection
		config := loadConfig()
		db, err := database.NewSQLiteDB(config.Database.Path)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		// Run migrations automatically
		runner := database.NewMigrationRunner(db.DB(), migrationsDir)
		if err := runner.MigrateUp(); err != nil {
			log.Printf("Warning: Failed to run migrations: %v", err)
		}

		// Initialize repositories and usecase
		wordlistRepo := repository.NewWordlistRepository(db)
		wordlistUsecase := usecase.NewWordlistUsecase(wordlistRepo, config.Upload.Directory)

		// Delete wordlist
		if err := deleteWordlistCLI(wordlistUsecase, wordlistID); err != nil {
			log.Fatalf("Failed to delete wordlist: %v", err)
		}
	},
}

var wordlistCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up duplicate wordlist entries",
	Long: `Clean up duplicate wordlist entries by keeping only the most recent one for each original filename.

This command will:
- Find wordlists with duplicate orig_name
- Keep the most recent entry (by created_at)
- Delete older duplicate entries
- Clean up associated physical files

Examples:
  ./server wordlist cleanup`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize database connection
		config := loadConfig()
		db, err := database.NewSQLiteDB(config.Database.Path)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		// Run migrations automatically
		runner := database.NewMigrationRunner(db.DB(), migrationsDir)
		if err := runner.MigrateUp(); err != nil {
			log.Printf("Warning: Failed to run migrations: %v", err)
		}

		// Initialize repositories and usecase
		wordlistRepo := repository.NewWordlistRepository(db)
		wordlistUsecase := usecase.NewWordlistUsecase(wordlistRepo, config.Upload.Directory)

		// Cleanup duplicate wordlists
		if err := cleanupWordlistsCLI(wordlistUsecase); err != nil {
			log.Fatalf("Failed to cleanup wordlists: %v", err)
		}
	},
}

func init() {
	// Add migration commands to root
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateGenerateCmd)
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateStatusCmd)

	// Add wordlist commands to root
	rootCmd.AddCommand(wordlistCmd)
	wordlistCmd.AddCommand(wordlistUploadCmd)
	wordlistCmd.AddCommand(wordlistListCmd)
	wordlistCmd.AddCommand(wordlistDeleteCmd)
	wordlistCmd.AddCommand(wordlistCleanupCmd)

	// Flags
	migrateCmd.PersistentFlags().StringVar(&migrationsDir, "migrations-dir", migrationsDir, "Directory containing migration files")
	
	// Wordlist upload flags
	wordlistUploadCmd.Flags().String("name", "", "Custom name for the wordlist")
	wordlistUploadCmd.Flags().Bool("count", true, "Enable word counting")
	wordlistUploadCmd.Flags().Int("chunk", 10, "Chunk size for processing in MB")
}

// uploadWordlistCLI handles optimized wordlist upload with progress reporting
func uploadWordlistCLI(wordlistUsecase usecase.WordlistUsecase, filePath, customName string, enableCount bool, chunkSizeMB int) error {
	// Validate file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// Determine filename
	filename := customName
	if filename == "" {
		filename = filepath.Base(filePath)
	}

	// Convert chunk size to bytes
	chunkSize := int64(chunkSizeMB) * 1024 * 1024

	log.Printf("Starting wordlist upload: %s", filename)
	log.Printf("File size: %s", formatBytes(fileInfo.Size()))
	log.Printf("Chunk size: %s", formatBytes(chunkSize))
	log.Printf("Word counting: %t", enableCount)

	// Open file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create a custom reader that provides progress updates
	var reader io.Reader = file
	if enableCount {
		reader = &progressReader{
			file:       file,
			totalSize:  fileInfo.Size(),
			chunkSize:  chunkSize,
			onProgress: func(bytesRead, totalBytes int64) {
				percentage := float64(bytesRead) / float64(totalBytes) * 100
				log.Printf("Progress: %s / %s (%.1f%%)", 
					formatBytes(bytesRead), 
					formatBytes(totalBytes), 
					percentage)
			},
		}
	}

	// Upload wordlist
	wordlist, err := wordlistUsecase.UploadWordlist(
		context.Background(),
		filename,
		reader,
		fileInfo.Size(),
	)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	log.Printf("Wordlist uploaded successfully!")
	log.Printf("ID: %s", wordlist.ID)
	log.Printf("Name: %s", wordlist.Name)
	log.Printf("Size: %s", formatBytes(wordlist.Size))
	if wordlist.WordCount != nil {
		log.Printf("Word count: %s", formatNumber(*wordlist.WordCount))
	}
	log.Printf("Path: %s", wordlist.Path)

	return nil
}

// listWordlistsCLI lists all uploaded wordlists with their details
func listWordlistsCLI(wordlistUsecase usecase.WordlistUsecase) error {
	wordlists, err := wordlistUsecase.GetAllWordlists(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list wordlists: %w", err)
	}

	if len(wordlists) == 0 {
		log.Println("No wordlists found.")
		return nil
	}

	fmt.Println("Uploaded Wordlists:")
	fmt.Println("---------------------")
	for _, wl := range wordlists {
		fmt.Printf("ID: %s\n", wl.ID)
		fmt.Printf("Name: %s\n", wl.Name)
		fmt.Printf("Size: %s\n", formatBytes(wl.Size))
		if wl.WordCount != nil {
			fmt.Printf("Word Count: %s\n", formatNumber(*wl.WordCount))
		}
		fmt.Printf("Created At: %s\n", wl.CreatedAt.Format(time.RFC3339))
		fmt.Println("---------------------")
	}
	return nil
}

// deleteWordlistCLI deletes a wordlist by its ID
func deleteWordlistCLI(wordlistUsecase usecase.WordlistUsecase, wordlistID string) error {
	// Parse UUID
	id, err := uuid.Parse(wordlistID)
	if err != nil {
		return fmt.Errorf("invalid wordlist ID format: %w", err)
	}

	// Delete the wordlist
	if err := wordlistUsecase.DeleteWordlist(context.Background(), id); err != nil {
		return fmt.Errorf("failed to delete wordlist: %w", err)
	}

	log.Printf("Wordlist with ID %s deleted successfully.", wordlistID)
	return nil
}

// cleanupWordlistsCLI cleans up duplicate wordlist entries by keeping only the most recent one for each original filename.
func cleanupWordlistsCLI(wordlistUsecase usecase.WordlistUsecase) error {
	log.Println("Starting wordlist cleanup...")

	// Get all wordlists
	wordlists, err := wordlistUsecase.GetAllWordlists(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get all wordlists for cleanup: %w", err)
	}

	if len(wordlists) == 0 {
		log.Println("No wordlists to cleanup.")
		return nil
	}

	// Group wordlists by original filename
	wordlistGroups := make(map[string][]domain.Wordlist)
	for _, wl := range wordlists {
		originalName := wl.OrigName
		if originalName == "" {
			originalName = wl.Name // Fallback to name if original name is empty
		}
		wordlistGroups[originalName] = append(wordlistGroups[originalName], wl)
	}

	// Find duplicates and clean them up
	var totalDeleted int
	for originalName, group := range wordlistGroups {
		if len(group) > 1 {
			log.Printf("Found %d duplicates for '%s'", len(group), originalName)
			
			// Sort by created_at to keep the most recent one
			sort.Slice(group, func(i, j int) bool {
				return group[i].CreatedAt.Before(group[j].CreatedAt)
			})
			
			// Keep the first one (most recent) and delete the rest
			toKeep := group[0]
			toDelete := group[1:]
			
			log.Printf("Keeping: %s (ID: %s, Created: %s)", toKeep.Name, toKeep.ID, toKeep.CreatedAt.Format(time.RFC3339))
			
			for _, wl := range toDelete {
				log.Printf("Deleting duplicate: %s (ID: %s, Created: %s)", wl.Name, wl.ID, wl.CreatedAt.Format(time.RFC3339))
				
				if err := wordlistUsecase.DeleteWordlist(context.Background(), wl.ID); err != nil {
					log.Printf("Failed to delete duplicate %s (ID: %s): %v", wl.Name, wl.ID, err)
					continue
				}
				totalDeleted++
			}
		}
	}

	if totalDeleted == 0 {
		log.Println("No duplicate wordlists found.")
	} else {
		log.Printf("Cleanup completed. %d duplicate entries removed.", totalDeleted)
	}
	
	return nil
}

// progressReader wraps a file reader to provide progress updates
type progressReader struct {
	file       *os.File
	totalSize  int64
	chunkSize  int64
	bytesRead  int64
	onProgress func(bytesRead, totalBytes int64)
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.file.Read(p)
	pr.bytesRead += int64(n)
	
	// Report progress every chunk
	if pr.bytesRead%pr.chunkSize < int64(n) {
		pr.onProgress(pr.bytesRead, pr.totalSize)
	}
	
	return n, err
}

// formatBytes converts bytes to human readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatNumber formats large numbers with commas
func formatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%s", formatNumberWithCommas(n))
}

func formatNumberWithCommas(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%s,%03d", formatNumberWithCommas(n/1000), n%1000)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func startServer() {
	// Load configuration
	config := loadConfig()

	// Initialize database
	db, err := database.NewSQLiteDB(config.Database.Path)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations automatically on server start
	runner := database.NewMigrationRunner(db.DB(), migrationsDir)
	if err := runner.MigrateUp(); err != nil {
		log.Printf("Warning: Failed to run migrations: %v", err)
	}

	// Initialize repositories
	agentRepo := repository.NewAgentRepository(db)
	jobRepo := repository.NewJobRepository(db)
	hashFileRepo := repository.NewHashFileRepository(db)
	wordlistRepo := repository.NewWordlistRepository(db)

	// Initialize use cases
	agentUsecase := usecase.NewAgentUsecase(agentRepo)
	jobUsecase := usecase.NewJobUsecase(jobRepo, agentRepo, hashFileRepo)
	hashFileUsecase := usecase.NewHashFileUsecase(hashFileRepo, config.Upload.Directory)
	wordlistUsecase := usecase.NewWordlistUsecase(wordlistRepo, config.Upload.Directory)

	// Initialize enrichment service
	jobEnrichmentService := usecase.NewJobEnrichmentService(agentRepo, wordlistRepo, hashFileRepo)

	// ✅ Get WebSocket hub early for dependency injection
	wsHub := handler.GetHub() // Get the singleton hub

	// ✅ Set WebSocket hub to agent usecase for real-time broadcasts
	agentUsecase.SetWebSocketHub(wsHub)
	log.Printf("✅ WebSocket hub connected to agent usecase")

	// Initialize HTTP router
	router := httpDelivery.NewRouter(agentUsecase, jobUsecase, hashFileUsecase, wordlistUsecase, jobEnrichmentService)

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Server.Port),
		Handler: router,
	}

	// Initialize health monitoring with ultra-fast real-time intervals
	healthConfig := usecase.HealthConfig{
		CheckInterval:       1 * time.Second, // ✅ Ultra-fast: check every 1 second
		AgentTimeout:        5 * time.Second, // ✅ Ultra-fast timeout detection in 5 seconds
		HeartbeatGrace:      2 * time.Second, // ✅ Very short grace period
		MaxConcurrentChecks: 20,              // ✅ More concurrent checks
	}

	healthMonitor := usecase.NewAgentHealthMonitor(
		agentUsecase,
		wsHub,
		healthConfig,
	)

	// Start health monitor
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	healthMonitor.Start(ctx)
	defer healthMonitor.Stop()

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %d", config.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Stop health monitor
	healthMonitor.Stop()

	// Shutdown server with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited")
}
