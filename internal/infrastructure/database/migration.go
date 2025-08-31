package database

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-distributed-hashcat/internal/infrastructure"
)

type Migration struct {
	Version   int
	Name      string
	UpSQL     string
	DownSQL   string
	Timestamp time.Time
}

type MigrationRunner struct {
	db            *sql.DB
	migrationsDir string
}

func NewMigrationRunner(db *sql.DB, migrationsDir string) *MigrationRunner {
	return &MigrationRunner{
		db:            db,
		migrationsDir: migrationsDir,
	}
}

// CreateMigrationsTable creates the migrations tracking table
func (mr *MigrationRunner) CreateMigrationsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		checksum TEXT NOT NULL
	)`

	_, err := mr.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	return nil
}

// GenerateMigration creates a new migration file
func (mr *MigrationRunner) GenerateMigration(name string) error {
	// Ensure migrations directory exists
	if err := os.MkdirAll(mr.migrationsDir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Generate timestamp-based version
	timestamp := time.Now()
	version := timestamp.Format("20060102150405") // YYYYMMDDHHMMSS

	// Clean name (replace spaces with underscores, lowercase)
	cleanName := strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	cleanName = strings.ReplaceAll(cleanName, "-", "_")

	filename := fmt.Sprintf("%s_%s.sql", version, cleanName)
	filepath := filepath.Join(mr.migrationsDir, filename)

	// Migration template
	template := fmt.Sprintf(`-- Migration: %s_%s.sql
-- Description: %s
-- Author: %s
-- Date: %s

-- +migrate Up
-- Write your UP migration here
-- Example:
-- CREATE TABLE example (
--     id TEXT PRIMARY KEY,
--     name TEXT NOT NULL,
--     created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
-- );

-- +migrate Down
-- Write your DOWN migration here
-- Example:
-- DROP TABLE IF EXISTS example;
`, version, cleanName, name, getUserName(), timestamp.Format("2006-01-02"))

	// Write file
	if err := os.WriteFile(filepath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to write migration file: %w", err)
	}

	infrastructure.ServerLogger.Info("Database migration system initialized")
	return nil
}

// LoadMigrations loads all migration files from directory
func (mr *MigrationRunner) LoadMigrations() ([]Migration, error) {
	var migrations []Migration

	err := filepath.WalkDir(mr.migrationsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		migration, err := mr.parseMigrationFile(path)
		if err != nil {
			infrastructure.ServerLogger.Warning("Skipping invalid migration file %s: %v", path, err)
			return nil // Continue with other files
		}

		migrations = append(migrations, migration)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// parseMigrationFile parses a single migration file
func (mr *MigrationRunner) parseMigrationFile(filePath string) (Migration, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return Migration{}, fmt.Errorf("failed to read file: %w", err)
	}

	filename := filepath.Base(filePath)

	// Extract version from filename (first part before underscore)
	parts := strings.Split(filename, "_")
	if len(parts) < 2 {
		return Migration{}, fmt.Errorf("invalid migration filename format: %s", filename)
	}

	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return Migration{}, fmt.Errorf("invalid version in filename: %s", parts[0])
	}

	// Extract name (everything except version and .sql extension)
	name := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".sql")

	// Parse UP and DOWN sections
	contentStr := string(content)
	upSQL, downSQL := mr.parseUpDown(contentStr)

	return Migration{
		Version: version,
		Name:    name,
		UpSQL:   upSQL,
		DownSQL: downSQL,
	}, nil
}

// parseUpDown extracts UP and DOWN SQL from migration content
func (mr *MigrationRunner) parseUpDown(content string) (string, string) {
	lines := strings.Split(content, "\n")
	var upSQL, downSQL strings.Builder
	var currentSection string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, "+migrate Up") {
			currentSection = "up"
			continue
		} else if strings.Contains(trimmed, "+migrate Down") {
			currentSection = "down"
			continue
		}

		// Skip comments and empty lines in SQL sections
		// if currentSection != "" && !strings.HasPrefix(trimmed, "--") && trimmed != "" {
		// 	if currentSection == "up" {
		// 		upSQL.WriteString(line + "\n")
		// 	} else if currentSection == "down" {
		// 		downSQL.WriteString(line + "\n")
		// 	}
		// }
		switch currentSection {
		case "up":
			upSQL.WriteString(line + "\n")
		case "down":
			downSQL.WriteString(line + "\n")
		}
	}

	return strings.TrimSpace(upSQL.String()), strings.TrimSpace(downSQL.String())
}

// GetAppliedMigrations returns list of applied migration versions
func (mr *MigrationRunner) GetAppliedMigrations() ([]int, error) {
	query := "SELECT version FROM schema_migrations ORDER BY version"
	rows, err := mr.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	var versions []int
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		versions = append(versions, version)
	}

	return versions, nil
}

// MigrateUp runs pending migrations
func (mr *MigrationRunner) MigrateUp() error {
	// Ensure migrations table exists
	if err := mr.CreateMigrationsTable(); err != nil {
		return err
	}

	// Load all migrations
	migrations, err := mr.LoadMigrations()
	if err != nil {
		return err
	}

	// Get applied migrations
	appliedVersions, err := mr.GetAppliedMigrations()
	if err != nil {
		return err
	}

	// Create map for quick lookup
	applied := make(map[int]bool)
	for _, version := range appliedVersions {
		applied[version] = true
	}

	// Run pending migrations
	var runCount int
	for _, migration := range migrations {
		if applied[migration.Version] {
			continue // Skip already applied
		}

		infrastructure.ServerLogger.Info("Running migration %d: %s", migration.Version, migration.Name)

		// Execute migration in transaction
		tx, err := mr.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		// Execute UP SQL
		if migration.UpSQL != "" {
			if _, err := tx.Exec(migration.UpSQL); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to execute migration %d: %w", migration.Version, err)
			}
		}

		// Record migration as applied
		checksum := generateChecksum(migration.UpSQL)
		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (version, name, checksum) VALUES (?, ?, ?)",
			migration.Version, migration.Name, checksum,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
		}

		infrastructure.ServerLogger.Success("Applied migration %d: %s", migration.Version, migration.Name)
		runCount++
	}

	if runCount == 0 {
		infrastructure.ServerLogger.Success("All migrations completed successfully")
	} else {
		infrastructure.ServerLogger.Success("Successfully applied %d migrations", runCount)
	}

	return nil
}

// MigrateDown rolls back the last migration
func (mr *MigrationRunner) MigrateDown() error {
	// Get applied migrations
	appliedVersions, err := mr.GetAppliedMigrations()
	if err != nil {
		return err
	}

	if len(appliedVersions) == 0 {
		infrastructure.ServerLogger.Info("No migrations to rollback")
		return nil
	}

	// Get the last applied migration
	lastVersion := appliedVersions[len(appliedVersions)-1]

	// Load migrations to find the one to rollback
	migrations, err := mr.LoadMigrations()
	if err != nil {
		return err
	}

	var targetMigration *Migration
	for _, migration := range migrations {
		if migration.Version == lastVersion {
			targetMigration = &migration
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration file not found for version %d", lastVersion)
	}

	infrastructure.ServerLogger.Info("Rolling back migration %d: %s", targetMigration.Version, targetMigration.Name)

	// Execute rollback in transaction
	tx, err := mr.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Execute DOWN SQL
	if targetMigration.DownSQL != "" {
		if _, err := tx.Exec(targetMigration.DownSQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to rollback migration %d: %w", targetMigration.Version, err)
		}
	}

	// Remove migration record
	if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = ?", targetMigration.Version); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove migration record %d: %w", targetMigration.Version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback %d: %w", targetMigration.Version, err)
	}

	infrastructure.ServerLogger.Success("Rolled back migration %d: %s", targetMigration.Version, targetMigration.Name)
	return nil
}

// Status shows migration status
func (mr *MigrationRunner) Status() error {
	// Load all migrations
	migrations, err := mr.LoadMigrations()
	if err != nil {
		return err
	}

	// Get applied migrations
	appliedVersions, err := mr.GetAppliedMigrations()
	if err != nil {
		return err
	}

	// Create applied lookup
	applied := make(map[int]bool)
	for _, version := range appliedVersions {
		applied[version] = true
	}

	infrastructure.ServerLogger.Info("Migration table created")
	infrastructure.ServerLogger.Info("=================")

	if len(migrations) == 0 {
		infrastructure.ServerLogger.Info("No migration files found")
		return nil
	}

	for _, migration := range migrations {
		status := "Pending"
		if applied[migration.Version] {
			status = "Applied"
		}
		infrastructure.ServerLogger.Info("Migration %d: %s - %s", migration.Version, migration.Name, status)
	}

	pendingCount := len(migrations) - len(appliedVersions)
	infrastructure.ServerLogger.Info("Summary: %d total, %d applied, %d pending",
		len(migrations), len(appliedVersions), pendingCount)

	return nil
}

// Helper functions
func getUserName() string {
	if name := os.Getenv("USER"); name != "" {
		return name
	}
	if name := os.Getenv("USERNAME"); name != "" {
		return name
	}
	return "system"
}

func generateChecksum(content string) string {
	// Simple checksum - in production you might want to use SHA256
	hash := 0
	for _, char := range content {
		hash = hash*31 + int(char)
	}
	return fmt.Sprintf("%x", hash)
}
