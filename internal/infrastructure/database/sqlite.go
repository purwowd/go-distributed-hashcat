package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDB struct {
	db *sql.DB
}

func NewSQLiteDB(dbPath string) (*SQLiteDB, error) {
	// Enhanced connection string with performance optimizations
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_temp_store=memory&_mmap_size=268435456", dbPath)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for better performance
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	sqliteDB := &SQLiteDB{db: db}

	if err := sqliteDB.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Apply performance optimizations after migration
	if err := sqliteDB.optimizePerformance(); err != nil {
		return nil, fmt.Errorf("failed to optimize database: %w", err)
	}

	return sqliteDB, nil
}

func (s *SQLiteDB) Close() error {
	return s.db.Close()
}

func (s *SQLiteDB) DB() *sql.DB {
	return s.db
}

func (s *SQLiteDB) optimizePerformance() error {
	optimizations := []string{
		"PRAGMA synchronous = NORMAL",
		"PRAGMA cache_size = 10000",
		"PRAGMA temp_store = memory",
		"PRAGMA mmap_size = 268435456",
		"PRAGMA journal_mode = WAL",
		"PRAGMA wal_autocheckpoint = 1000",
		"PRAGMA auto_vacuum = INCREMENTAL",
		"PRAGMA incremental_vacuum(1000)",
		"ANALYZE",
	}

	for _, optimization := range optimizations {
		if _, err := s.db.Exec(optimization); err != nil {
			return fmt.Errorf("failed to execute optimization '%s': %w", optimization, err)
		}
	}

	return nil
}

func (s *SQLiteDB) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS agents (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			ip_address TEXT NOT NULL,
			port INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'offline',
			capabilities TEXT,
			last_seen DATETIME,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			agent_key TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS jobs (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			hash_type INTEGER NOT NULL,
			attack_mode INTEGER NOT NULL,
			hash_file TEXT NOT NULL,
			hash_file_id TEXT,
			wordlist TEXT NOT NULL,
			wordlist_id TEXT,
			rules TEXT,
			agent_id TEXT,
			progress REAL DEFAULT 0,
			speed INTEGER DEFAULT 0,
			eta DATETIME,
			result TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			started_at DATETIME,
			completed_at DATETIME,
			FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE SET NULL,
			FOREIGN KEY (hash_file_id) REFERENCES hash_files(id) ON DELETE SET NULL,
			FOREIGN KEY (wordlist_id) REFERENCES wordlists(id) ON DELETE SET NULL
		)`,
		`CREATE TABLE IF NOT EXISTS hash_files (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			orig_name TEXT NOT NULL,
			path TEXT NOT NULL,
			size INTEGER NOT NULL,
			type TEXT NOT NULL,
			created_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS wordlists (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			orig_name TEXT NOT NULL,
			path TEXT NOT NULL,
			size INTEGER NOT NULL,
			word_count INTEGER,
			created_at DATETIME NOT NULL
		)`,
		// Comprehensive indexes for performance
		`CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status)`,
		`CREATE INDEX IF NOT EXISTS idx_agents_last_seen ON agents(last_seen DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_agents_created_at ON agents(created_at DESC)`,

		// Agent Keys table
		`CREATE TABLE IF NOT EXISTS agent_keys (
			id TEXT PRIMARY KEY,
			agent_key TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME NOT NULL,
			expires_at DATETIME,
			last_used_at DATETIME,
			agent_id TEXT,
			FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE SET NULL
		)`,

		// Agent Keys indexes
		`CREATE INDEX IF NOT EXISTS idx_agent_keys_key ON agent_keys(agent_key)`,
		`CREATE INDEX IF NOT EXISTS idx_agent_keys_status ON agent_keys(status)`,
		`CREATE INDEX IF NOT EXISTS idx_agent_keys_agent_id ON agent_keys(agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status)`,
		`CREATE INDEX IF NOT EXISTS idx_jobs_agent_id ON jobs(agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_jobs_updated_at ON jobs(updated_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_jobs_status_agent ON jobs(status, agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_jobs_status_created ON jobs(status, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_hash_files_type ON hash_files(type)`,
		`CREATE INDEX IF NOT EXISTS idx_hash_files_created_at ON hash_files(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_hash_files_size ON hash_files(size)`,
		`CREATE INDEX IF NOT EXISTS idx_wordlists_size ON wordlists(size)`,
		`CREATE INDEX IF NOT EXISTS idx_wordlists_created_at ON wordlists(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_wordlists_word_count ON wordlists(word_count)`,
		// Composite indexes for common query patterns
		`CREATE INDEX IF NOT EXISTS idx_agents_status_updated ON agents(status, updated_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_jobs_agent_status ON jobs(agent_id, status)`,
		`ALTER TABLE jobs ADD COLUMN hash_file_id TEXT REFERENCES hash_files(id)`,
		`ALTER TABLE jobs ADD COLUMN wordlist_id TEXT REFERENCES wordlists(id)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			// Ignore "duplicate column" errors for ALTER TABLE
			errMsg := err.Error()
			if errMsg != "duplicate column name: hash_file_id" && errMsg != "duplicate column name: wordlist_id" {
				return fmt.Errorf("failed to execute migration query: %s, error: %w", query, err)
			}
		}
	}

	return nil
}
