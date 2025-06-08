package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go-distributed-hashcat/internal/domain"
	"go-distributed-hashcat/internal/infrastructure/cache"
	"go-distributed-hashcat/internal/infrastructure/database"

	"github.com/google/uuid"
)

type jobRepository struct {
	db                 *database.SQLiteDB
	cache              cache.Cache
	getByIDStmt        *sql.Stmt
	getAllStmt         *sql.Stmt
	getByStatusStmt    *sql.Stmt
	getByAgentIDStmt   *sql.Stmt
	updateStmt         *sql.Stmt
	deleteStmt         *sql.Stmt
	updateStatusStmt   *sql.Stmt
	updateProgressStmt *sql.Stmt
}

func NewJobRepository(db *database.SQLiteDB) domain.JobRepository {
	repo := &jobRepository{
		db:    db,
		cache: cache.NewMemoryCache(15 * time.Second), // 15 second cache for jobs (shorter due to frequent updates)
	}

	// Prepare frequently used statements
	repo.prepareStatements()

	return repo
}

func (r *jobRepository) prepareStatements() {
	var err error

	r.getByIDStmt, err = r.db.DB().Prepare(`
		SELECT id, name, status, hash_type, attack_mode, hash_file, hash_file_id, wordlist, rules,
		       agent_id, progress, speed, eta, result, created_at, updated_at, started_at, completed_at
		FROM jobs WHERE id = ? LIMIT 1
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare getByID statement: %v", err))
	}

	r.getAllStmt, err = r.db.DB().Prepare(`
		SELECT id, name, status, hash_type, attack_mode, hash_file, hash_file_id, wordlist, rules,
		       agent_id, progress, speed, eta, result, created_at, updated_at, started_at, completed_at
		FROM jobs ORDER BY created_at DESC LIMIT 100
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare getAll statement: %v", err))
	}

	r.getByStatusStmt, err = r.db.DB().Prepare(`
		SELECT id, name, status, hash_type, attack_mode, hash_file, hash_file_id, wordlist, rules,
		       agent_id, progress, speed, eta, result, created_at, updated_at, started_at, completed_at
		FROM jobs WHERE status = ? ORDER BY created_at DESC LIMIT 50
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare getByStatus statement: %v", err))
	}

	r.getByAgentIDStmt, err = r.db.DB().Prepare(`
		SELECT id, name, status, hash_type, attack_mode, hash_file, hash_file_id, wordlist, rules,
		       agent_id, progress, speed, eta, result, created_at, updated_at, started_at, completed_at
		FROM jobs WHERE agent_id = ? ORDER BY created_at DESC LIMIT 20
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare getByAgentID statement: %v", err))
	}

	r.updateStmt, err = r.db.DB().Prepare(`
		UPDATE jobs SET 
		name = ?, status = ?, hash_type = ?, attack_mode = ?, hash_file = ?, hash_file_id = ?, wordlist = ?, rules = ?,
		agent_id = ?, progress = ?, speed = ?, eta = ?, result = ?, updated_at = ?, started_at = ?, completed_at = ?
		WHERE id = ?
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare update statement: %v", err))
	}

	r.deleteStmt, err = r.db.DB().Prepare(`DELETE FROM jobs WHERE id = ?`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare delete statement: %v", err))
	}

	r.updateStatusStmt, err = r.db.DB().Prepare(`UPDATE jobs SET status = ?, updated_at = ? WHERE id = ?`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare updateStatus statement: %v", err))
	}

	r.updateProgressStmt, err = r.db.DB().Prepare(`UPDATE jobs SET progress = ?, speed = ?, updated_at = ? WHERE id = ?`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare updateProgress statement: %v", err))
	}
}

func (r *jobRepository) Create(ctx context.Context, job *domain.Job) error {
	query := `
		INSERT INTO jobs (id, name, status, hash_type, attack_mode, hash_file, hash_file_id, wordlist, rules, 
		                  agent_id, progress, speed, eta, result, created_at, updated_at, started_at, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	job.CreatedAt = now
	job.UpdatedAt = now

	var agentID *string
	if job.AgentID != nil {
		agentIDStr := job.AgentID.String()
		agentID = &agentIDStr
	}

	var hashFileID *string
	if job.HashFileID != nil {
		hashFileIDStr := job.HashFileID.String()
		hashFileID = &hashFileIDStr
	}

	var eta *time.Time
	if job.ETA != nil {
		eta = job.ETA
	}

	var startedAt *time.Time
	if job.StartedAt != nil {
		startedAt = job.StartedAt
	}

	var completedAt *time.Time
	if job.CompletedAt != nil {
		completedAt = job.CompletedAt
	}

	_, err := r.db.DB().ExecContext(ctx, query,
		job.ID.String(),
		job.Name,
		job.Status,
		job.HashType,
		job.AttackMode,
		job.HashFile,
		hashFileID,
		job.Wordlist,
		job.Rules,
		agentID,
		job.Progress,
		job.Speed,
		eta,
		job.Result,
		job.CreatedAt,
		job.UpdatedAt,
		startedAt,
		completedAt,
	)

	if err == nil {
		// Cache the new job
		r.cache.Set(ctx, "job:"+job.ID.String(), job)
		// Invalidate list caches
		r.invalidateListCaches(ctx)
	}

	return err
}

func (r *jobRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	cacheKey := "job:" + id.String()

	// Try cache first
	var job domain.Job
	if found, err := r.cache.Get(ctx, cacheKey, &job); err == nil && found {
		return &job, nil
	}

	// Fallback to database
	job, err := r.scanJob(r.getByIDStmt.QueryRowContext(ctx, id.String()))
	if err != nil {
		return nil, err
	}

	// Cache the result
	r.cache.Set(ctx, cacheKey, &job)

	return &job, nil
}

func (r *jobRepository) GetAll(ctx context.Context) ([]domain.Job, error) {
	cacheKey := "jobs:all"

	// Try cache first
	var jobs []domain.Job
	if found, err := r.cache.Get(ctx, cacheKey, &jobs); err == nil && found {
		return jobs, nil
	}

	// Fallback to database
	jobs, err := r.queryJobs(ctx, r.getAllStmt)
	if err != nil {
		return nil, err
	}

	// Cache the result
	r.cache.Set(ctx, cacheKey, jobs)

	return jobs, nil
}

func (r *jobRepository) GetByStatus(ctx context.Context, status string) ([]domain.Job, error) {
	cacheKey := "jobs:status:" + status

	// Try cache first
	var jobs []domain.Job
	if found, err := r.cache.Get(ctx, cacheKey, &jobs); err == nil && found {
		return jobs, nil
	}

	// Fallback to database
	jobs, err := r.queryJobsWithArgs(ctx, r.getByStatusStmt, status)
	if err != nil {
		return nil, err
	}

	// Cache the result
	r.cache.Set(ctx, cacheKey, jobs)

	return jobs, nil
}

func (r *jobRepository) GetByAgentID(ctx context.Context, agentID uuid.UUID) ([]domain.Job, error) {
	cacheKey := "jobs:agent:" + agentID.String()

	// Try cache first
	var jobs []domain.Job
	if found, err := r.cache.Get(ctx, cacheKey, &jobs); err == nil && found {
		return jobs, nil
	}

	// Fallback to database
	jobs, err := r.queryJobsWithArgs(ctx, r.getByAgentIDStmt, agentID.String())
	if err != nil {
		return nil, err
	}

	// Cache the result
	r.cache.Set(ctx, cacheKey, jobs)

	return jobs, nil
}

// GetAvailableJobForAgent gets the next available job assigned to the agent that is ready to run
func (r *jobRepository) GetAvailableJobForAgent(ctx context.Context, agentID uuid.UUID) (*domain.Job, error) {
	// Query for pending jobs assigned to this agent
	query := `
		SELECT id, name, status, hash_type, attack_mode, hash_file, hash_file_id, 
		       wordlist, rules, agent_id, progress, speed, eta, result, 
		       created_at, updated_at, started_at, completed_at
		FROM jobs 
		WHERE agent_id = ? AND status = 'pending'
		ORDER BY created_at ASC
		LIMIT 1
	`

	row := r.db.DB().QueryRowContext(ctx, query, agentID.String())
	job, err := r.scanJob(row)
	if err != nil {
		if err.Error() == "job not found" {
			return nil, fmt.Errorf("no available jobs for agent")
		}
		return nil, err
	}

	return &job, nil
}

func (r *jobRepository) Update(ctx context.Context, job *domain.Job) error {
	job.UpdatedAt = time.Now()

	var agentID *string
	if job.AgentID != nil {
		agentIDStr := job.AgentID.String()
		agentID = &agentIDStr
	}

	var hashFileID *string
	if job.HashFileID != nil {
		hashFileIDStr := job.HashFileID.String()
		hashFileID = &hashFileIDStr
	}

	_, err := r.updateStmt.ExecContext(ctx,
		job.Name,
		job.Status,
		job.HashType,
		job.AttackMode,
		job.HashFile,
		hashFileID,
		job.Wordlist,
		job.Rules,
		agentID,
		job.Progress,
		job.Speed,
		job.ETA,
		job.Result,
		job.UpdatedAt,
		job.StartedAt,
		job.CompletedAt,
		job.ID.String(),
	)

	if err == nil {
		// Update cache
		r.cache.Set(ctx, "job:"+job.ID.String(), job)
		// Invalidate list caches
		r.invalidateListCaches(ctx)
	}

	return err
}

func (r *jobRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.deleteStmt.ExecContext(ctx, id.String())

	if err == nil {
		// Remove from cache
		r.cache.Delete(ctx, "job:"+id.String())
		// Invalidate list caches
		r.invalidateListCaches(ctx)
	}

	return err
}

func (r *jobRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.updateStatusStmt.ExecContext(ctx, status, time.Now(), id.String())

	if err == nil {
		// Invalidate caches
		r.cache.Delete(ctx, "job:"+id.String())
		r.invalidateListCaches(ctx)
	}

	return err
}

func (r *jobRepository) UpdateProgress(ctx context.Context, id uuid.UUID, progress float64, speed int64) error {
	_, err := r.updateProgressStmt.ExecContext(ctx, progress, speed, time.Now(), id.String())

	if err == nil {
		// Invalidate caches (don't cache individual job for progress updates due to frequency)
		r.cache.Delete(ctx, "job:"+id.String())
	}

	return err
}

func (r *jobRepository) invalidateListCaches(ctx context.Context) {
	// Delete all list-related caches
	r.cache.Delete(ctx, "jobs:all")
	// We can't easily invalidate all status-specific caches, so we clear the entire cache periodically
	// This is acceptable given the 15-second TTL
}

func (r *jobRepository) queryJobs(ctx context.Context, stmt *sql.Stmt) ([]domain.Job, error) {
	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanJobs(rows)
}

func (r *jobRepository) queryJobsWithArgs(ctx context.Context, stmt *sql.Stmt, args ...interface{}) ([]domain.Job, error) {
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanJobs(rows)
}

// scanJob optimized for single row scanning
func (r *jobRepository) scanJob(row *sql.Row) (domain.Job, error) {
	var job domain.Job
	var idStr string
	var agentIDStr sql.NullString
	var hashFileIDStr sql.NullString
	var eta sql.NullTime
	var startedAt sql.NullTime
	var completedAt sql.NullTime

	err := row.Scan(
		&idStr,
		&job.Name,
		&job.Status,
		&job.HashType,
		&job.AttackMode,
		&job.HashFile,
		&hashFileIDStr,
		&job.Wordlist,
		&job.Rules,
		&agentIDStr,
		&job.Progress,
		&job.Speed,
		&eta,
		&job.Result,
		&job.CreatedAt,
		&job.UpdatedAt,
		&startedAt,
		&completedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return job, fmt.Errorf("job not found")
		}
		return job, err
	}

	job.ID = uuid.MustParse(idStr)

	if agentIDStr.Valid {
		agentID := uuid.MustParse(agentIDStr.String)
		job.AgentID = &agentID
	}

	if hashFileIDStr.Valid {
		hashFileID := uuid.MustParse(hashFileIDStr.String)
		job.HashFileID = &hashFileID
	}

	if eta.Valid {
		job.ETA = &eta.Time
	}

	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}

	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	return job, nil
}

func (r *jobRepository) scanJobs(rows *sql.Rows) ([]domain.Job, error) {
	jobs := make([]domain.Job, 0, 20) // Pre-allocate slice

	for rows.Next() {
		var job domain.Job
		var idStr string
		var agentIDStr sql.NullString
		var hashFileIDStr sql.NullString
		var eta sql.NullTime
		var startedAt sql.NullTime
		var completedAt sql.NullTime

		err := rows.Scan(
			&idStr,
			&job.Name,
			&job.Status,
			&job.HashType,
			&job.AttackMode,
			&job.HashFile,
			&hashFileIDStr,
			&job.Wordlist,
			&job.Rules,
			&agentIDStr,
			&job.Progress,
			&job.Speed,
			&eta,
			&job.Result,
			&job.CreatedAt,
			&job.UpdatedAt,
			&startedAt,
			&completedAt,
		)
		if err != nil {
			return nil, err
		}

		job.ID = uuid.MustParse(idStr)

		if agentIDStr.Valid {
			agentID := uuid.MustParse(agentIDStr.String)
			job.AgentID = &agentID
		}

		if hashFileIDStr.Valid {
			hashFileID := uuid.MustParse(hashFileIDStr.String)
			job.HashFileID = &hashFileID
		}

		if eta.Valid {
			job.ETA = &eta.Time
		}

		if startedAt.Valid {
			job.StartedAt = &startedAt.Time
		}

		if completedAt.Valid {
			job.CompletedAt = &completedAt.Time
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}
