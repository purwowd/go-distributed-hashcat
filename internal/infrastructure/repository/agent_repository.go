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

type agentRepository struct {
	db              *database.SQLiteDB
	cache           cache.Cache
	getByIDStmt     *sql.Stmt
	getByNameIPStmt *sql.Stmt
	getAllStmt      *sql.Stmt
	updateStmt      *sql.Stmt
	deleteStmt      *sql.Stmt
}

func NewAgentRepository(db *database.SQLiteDB) domain.AgentRepository {
	repo := &agentRepository{
		db:    db,
		cache: cache.NewMemoryCache(30 * time.Second), // 30 second cache for agents
	}

	// Prepare frequently used statements
	repo.prepareStatements()

	return repo
}

func (r *agentRepository) prepareStatements() {
	var err error

	// Prepare optimized queries
	r.getByIDStmt, err = r.db.DB().Prepare(`
		SELECT id, name, ip_address, port, status, capabilities, last_seen, agent_key, created_at, updated_at
		FROM agents WHERE id = ? LIMIT 1
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare getByID statement: %v", err))
	}

	r.getByNameIPStmt, err = r.db.DB().Prepare(`
		SELECT id, name, ip_address, port, status, capabilities, last_seen, agent_key, created_at, updated_at
		FROM agents WHERE name = ? AND ip_address = ? LIMIT 1
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare getByNameIP statement: %v", err))
	}

	r.getAllStmt, err = r.db.DB().Prepare(`
		SELECT id, name, ip_address, port, status, capabilities, last_seen, agent_key, created_at, updated_at
		FROM agents ORDER BY status DESC, updated_at DESC
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare getAll statement: %v", err))
	}

	r.updateStmt, err = r.db.DB().Prepare(`
		UPDATE agents SET 
		name = ?, ip_address = ?, port = ?, status = ?, capabilities = ?, updated_at = ?
		WHERE id = ?
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare update statement: %v", err))
	}

	r.deleteStmt, err = r.db.DB().Prepare(`DELETE FROM agents WHERE id = ?`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare delete statement: %v", err))
	}
}

func (r *agentRepository) Create(ctx context.Context, agent *domain.Agent) error {
	query := `
		INSERT INTO agents (id, name, ip_address, port, status, capabilities, last_seen, agent_key, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	agent.CreatedAt = now
	agent.UpdatedAt = now
	agent.LastSeen = now

	// Handle nullable agent_key
	var agentKey sql.NullString
	if agent.AgentKey != "" {
		agentKey = sql.NullString{String: agent.AgentKey, Valid: true}
	}

	_, err := r.db.DB().ExecContext(ctx, query,
		agent.ID.String(),
		agent.Name,
		agent.IPAddress,
		agent.Port,
		agent.Status,
		agent.Capabilities,
		agent.LastSeen,
		agentKey,
		agent.CreatedAt,
		agent.UpdatedAt,
	)

	if err == nil {
		// Cache the new agent
		r.cache.Set(ctx, "agent:"+agent.ID.String(), agent)
		// Invalidate list cache
		r.cache.Delete(ctx, "agents:all")
	}

	return err
}

func (r *agentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Agent, error) {
	cacheKey := "agent:" + id.String()

	// Try cache first
	var agent domain.Agent
	if found, err := r.cache.Get(ctx, cacheKey, &agent); err == nil && found {
		return &agent, nil
	}

	// Fallback to database with prepared statement
	var idStr string
	var agentKey sql.NullString

	err := r.getByIDStmt.QueryRowContext(ctx, id.String()).Scan(
		&idStr,
		&agent.Name,
		&agent.IPAddress,
		&agent.Port,
		&agent.Status,
		&agent.Capabilities,
		&agent.LastSeen,
		&agentKey,
		&agent.CreatedAt,
		&agent.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("agent not found")
		}
		return nil, err
	}

	agent.ID = uuid.MustParse(idStr)

	// Handle nullable fields
	if agentKey.Valid {
		agent.AgentKey = agentKey.String
	}

	// Cache the result
	r.cache.Set(ctx, cacheKey, &agent)

	return &agent, nil
}

func (r *agentRepository) GetByNameAndIP(ctx context.Context, name, ip string, port int) (*domain.Agent, error) {
	cacheKey := "agent:name_ip:" + name + ":" + ip

	// Try cache first
	var agent domain.Agent
	if found, err := r.cache.Get(ctx, cacheKey, &agent); err == nil && found {
		return &agent, nil
	}

	// Fallback to database with prepared statement
	var idStr string
	var agentKey sql.NullString

	err := r.getByNameIPStmt.QueryRowContext(ctx, name, ip).Scan(
		&idStr,
		&agent.Name,
		&agent.IPAddress,
		&agent.Port,
		&agent.Status,
		&agent.Capabilities,
		&agent.LastSeen,
		&agentKey,
		&agent.CreatedAt,
		&agent.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("agent not found")
		}
		return nil, err
	}

	agent.ID = uuid.MustParse(idStr)

	// Handle nullable fields
	if agentKey.Valid {
		agent.AgentKey = agentKey.String
	}

	// Cache the result
	r.cache.Set(ctx, cacheKey, &agent)

	return &agent, nil
}

func (r *agentRepository) GetAll(ctx context.Context) ([]domain.Agent, error) {
	cacheKey := "agents:all"

	// Try cache first
	var agents []domain.Agent
	if found, err := r.cache.Get(ctx, cacheKey, &agents); err == nil && found {
		return agents, nil
	}

	// Fallback to database with prepared statement
	rows, err := r.getAllStmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	agents = make([]domain.Agent, 0, 10) // Pre-allocate slice
	for rows.Next() {
		var agent domain.Agent
		var idStr string
		var agentKey sql.NullString

		err := rows.Scan(
			&idStr,
			&agent.Name,
			&agent.IPAddress,
			&agent.Port,
			&agent.Status,
			&agent.Capabilities,
			&agent.LastSeen,
			&agentKey,
			&agent.CreatedAt,
			&agent.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		agent.ID = uuid.MustParse(idStr)

		// Handle nullable fields
		if agentKey.Valid {
			agent.AgentKey = agentKey.String
		}

		agents = append(agents, agent)
	}

	// Cache the result
	r.cache.Set(ctx, cacheKey, agents)

	return agents, nil
}

func (r *agentRepository) Update(ctx context.Context, agent *domain.Agent) error {
	agent.UpdatedAt = time.Now()

	_, err := r.updateStmt.ExecContext(ctx,
		agent.Name,
		agent.IPAddress,
		agent.Port,
		agent.Status,
		agent.Capabilities,
		agent.UpdatedAt,
		agent.ID.String(),
	)

	if err == nil {
		// Update cache
		r.cache.Set(ctx, "agent:"+agent.ID.String(), agent)
		// Update name+IP cache
		r.cache.Set(ctx, "agent:name_ip:"+agent.Name+":"+agent.IPAddress, agent)
		// Invalidate list cache
		r.cache.Delete(ctx, "agents:all")
	}

	return err
}

func (r *agentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.deleteStmt.ExecContext(ctx, id.String())

	if err == nil {
		// Remove from cache
		r.cache.Delete(ctx, "agent:"+id.String())
		// Invalidate list cache
		r.cache.Delete(ctx, "agents:all")
	}

	return err
}

func (r *agentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE agents SET status = ?, updated_at = ? WHERE id = ?`
	now := time.Now()
	_, err := r.db.DB().ExecContext(ctx, query, status, now, id.String())

	if err == nil {
		// Invalidate caches
		r.cache.Delete(ctx, "agent:"+id.String())
		r.cache.Delete(ctx, "agents:all")
	}

	return err
}

func (r *agentRepository) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE agents SET last_seen = ?, updated_at = ? WHERE id = ?`
	now := time.Now()
	_, err := r.db.DB().ExecContext(ctx, query, now, now, id.String())

	if err == nil {
		// Invalidate caches
		r.cache.Delete(ctx, "agent:"+id.String())
		r.cache.Delete(ctx, "agents:all")
	}

	return err
}

func (r *agentRepository) GetByAgentKey(ctx context.Context, agentKey string) (*domain.Agent, error) {
	cacheKey := "agent:key:" + agentKey

	// Try cache first
	var agent domain.Agent
	if found, err := r.cache.Get(ctx, cacheKey, &agent); err == nil && found {
		return &agent, nil
	}

	// Fallback to database
	query := `
		SELECT id, name, ip_address, port, status, capabilities, last_seen, agent_key, created_at, updated_at
		FROM agents WHERE agent_key = ? LIMIT 1
	`

	var idStr string
	var agentKeyDB sql.NullString

	err := r.db.DB().QueryRowContext(ctx, query, agentKey).Scan(
		&idStr,
		&agent.Name,
		&agent.IPAddress,
		&agent.Port,
		&agent.Status,
		&agent.Capabilities,
		&agent.LastSeen,
		&agentKeyDB,
		&agent.CreatedAt,
		&agent.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("agent not found")
		}
		return nil, err
	}

	agent.ID = uuid.MustParse(idStr)

	// Handle nullable fields
	if agentKeyDB.Valid {
		agent.AgentKey = agentKeyDB.String
	}

	// Cache the result
	r.cache.Set(ctx, cacheKey, &agent)

	return &agent, nil
}

func (r *agentRepository) UpdateAgentKey(ctx context.Context, id uuid.UUID, agentKey string) error {
	query := `UPDATE agents SET agent_key = ?, updated_at = ? WHERE id = ?`
	now := time.Now()
	_, err := r.db.DB().ExecContext(ctx, query, agentKey, now, id.String())

	if err == nil {
		// Invalidate caches
		r.cache.Delete(ctx, "agent:"+id.String())
		r.cache.Delete(ctx, "agents:all")
		if agentKey != "" {
			r.cache.Delete(ctx, "agent:key:"+agentKey)
		}
	}

	return err
}

func (r *agentRepository) RevokeAgentKey(ctx context.Context, agentKey string) error {
	query := `UPDATE agents SET status = 'banned', updated_at = ? WHERE agent_key = ?`
	now := time.Now()
	_, err := r.db.DB().ExecContext(ctx, query, now, agentKey)

	if err == nil {
		// Invalidate caches
		r.cache.Delete(ctx, "agent:key:"+agentKey)
		r.cache.Delete(ctx, "agents:all")
	}

	return err
}
