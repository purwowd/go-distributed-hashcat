package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpDelivery "go-distributed-hashcat/internal/delivery/http"
	"go-distributed-hashcat/internal/infrastructure/database"
	"go-distributed-hashcat/internal/infrastructure/repository"
	"go-distributed-hashcat/internal/usecase"

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

func init() {
	// Add migration commands to root
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateGenerateCmd)
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateStatusCmd)

	// Flags
	migrateCmd.PersistentFlags().StringVar(&migrationsDir, "migrations-dir", migrationsDir, "Directory containing migration files")
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

	// Initialize HTTP router
	router := httpDelivery.NewRouter(agentUsecase, jobUsecase, hashFileUsecase, wordlistUsecase)

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Server.Port),
		Handler: router,
	}

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
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
