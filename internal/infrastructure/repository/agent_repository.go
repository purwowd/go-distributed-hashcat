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
	db                 *database.SQLiteDB
	cache              cache.Cache
	getByIDStmt        *sql.Stmt
	getByNameStmt      *sql.Stmt
	getByNameIPStmt    *sql.Stmt
	getByIPAddressStmt *sql.Stmt
	getAllStmt         *sql.Stmt
	updateStmt         *sql.Stmt
	deleteStmt         *sql.Stmt
	getByAgentKeyStmt  *sql.Stmt
}

func NewAgentRepository(db *database.SQLiteDB) domain.AgentRepository {
	repo := &agentRepository{
		db:    db,
		cache: cache.NewMemoryCache(30 * time.Second),
	}

	repo.prepareStatements()
	return repo
}

func (r *agentRepository) prepareStatements() {
	var err error

	r.getByIDStmt, err = r.db.DB().Prepare(`
		SELECT id, name, ip_address, port, status, capabilities, agent_key, speed, last_seen, created_at, updated_at
		FROM agents WHERE id = ? LIMIT 1
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare getByID statement: %v", err))
	}

	r.getByNameStmt, err = r.db.DB().Prepare(`
		SELECT id, name, ip_address, port, status, capabilities, agent_key, speed, last_seen, created_at, updated_at
		FROM agents WHERE name = ? LIMIT 1
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare getByName statement: %v", err))
	}

	r.getByNameIPStmt, err = r.db.DB().Prepare(`
		SELECT id, name, ip_address, port, status, capabilities, agent_key, speed, last_seen, created_at, updated_at
		FROM agents WHERE name = ? AND ip_address = ? AND port = ? LIMIT 1
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare getByNameIP statement: %v", err))
	}

	r.getByIPAddressStmt, err = r.db.DB().Prepare(`
		SELECT id, name, ip_address, port, status, capabilities, agent_key, speed, last_seen, created_at, updated_at
		FROM agents WHERE ip_address = ? LIMIT 1
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare getByIPAddress statement: %v", err))
	}

	r.getAllStmt, err = r.db.DB().Prepare(`
		SELECT id, name, ip_address, port, status, capabilities, agent_key, speed, last_seen, created_at, updated_at
		FROM agents ORDER BY created_at DESC, id ASC
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare getAll statement: %v", err))
	}

	r.updateStmt, err = r.db.DB().Prepare(`
		UPDATE agents SET
		name = ?, ip_address = ?, port = ?, status = ?, capabilities = ?, agent_key = ?, speed = ?, last_seen = ?, updated_at = ?
		WHERE id = ?
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare update statement: %v", err))
	}

	r.deleteStmt, err = r.db.DB().Prepare(`
		DELETE FROM agents WHERE id = ?
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare delete statement: %v", err))
	}

	r.getByAgentKeyStmt, err = r.db.DB().Prepare(`
		SELECT id, name, ip_address, port, status, capabilities, agent_key, speed, last_seen, created_at, updated_at
		FROM agents WHERE agent_key = ? LIMIT 1
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare getByAgentKey statement: %v", err))
	}
}

func (r *agentRepository) Create(ctx context.Context, agent *domain.Agent) error {
	// Invalidate cache for IP address if it exists
	if agent.IPAddress != "" {
		cacheKey := "agent:ip:" + agent.IPAddress
		r.cache.Delete(ctx, cacheKey)
	}

	// Invalidate cache for agent key if it exists
	if agent.AgentKey != "" {
		cacheKey := "agent:key:" + agent.AgentKey
		r.cache.Delete(ctx, cacheKey)
	}

	// Invalidate cache for all agents list
	r.cache.Delete(ctx, "agents:all")

	query := `
        INSERT INTO agents (id, name, ip_address, port, status, capabilities, agent_key, speed, last_seen, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

	_, err := r.db.DB().ExecContext(ctx, query,
		agent.ID.String(),
		agent.Name,
		agent.IPAddress,
		agent.Port,
		agent.Status,
		agent.Capabilities,
		agent.AgentKey,
		agent.Speed,
		agent.LastSeen,
		agent.CreatedAt,
		agent.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	return nil
}

func (r *agentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Agent, error) {
	cacheKey := "agent:" + id.String()

	var agent domain.Agent
	if found, err := r.cache.Get(ctx, cacheKey, &agent); err == nil && found {
		return &agent, nil
	}

	var idStr string
	err := r.getByIDStmt.QueryRowContext(ctx, id.String()).Scan(
		&idStr,
		&agent.Name,
		&agent.IPAddress,
		&agent.Port,
		&agent.Status,
		&agent.Capabilities,
		&agent.AgentKey,
		&agent.Speed,
		&agent.LastSeen,
		&agent.CreatedAt,
		&agent.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrAgentNotFound // <-- use custom error
		}
		return nil, err
	}

	agent.ID = uuid.MustParse(idStr)
	r.cache.Set(ctx, cacheKey, &agent)

	return &agent, nil
}

func (r *agentRepository) GetByName(ctx context.Context, name string) (*domain.Agent, error) {
	cacheKey := "agent:name:" + name

	var agent domain.Agent
	if found, err := r.cache.Get(ctx, cacheKey, &agent); err == nil && found {
		return &agent, nil
	}

	var idStr string
	err := r.getByNameStmt.QueryRowContext(ctx, name).Scan(
		&idStr,
		&agent.Name,
		&agent.IPAddress,
		&agent.Port,
		&agent.Status,
		&agent.Capabilities,
		&agent.AgentKey,
		&agent.Speed,
		&agent.LastSeen,
		&agent.CreatedAt,
		&agent.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrAgentNotFound // <-- use custom error
		}
		return nil, err
	}

	agent.ID = uuid.MustParse(idStr)
	r.cache.Set(ctx, cacheKey, &agent)

	return &agent, nil
}

func (r *agentRepository) GetByNameAndIP(ctx context.Context, name, ip string, port int) (*domain.Agent, error) {
	cacheKey := "agent:name_ip:" + name + ":" + ip

	var agent domain.Agent
	if found, err := r.cache.Get(ctx, cacheKey, &agent); err == nil && found {
		return &agent, nil
	}

	var idStr string
	err := r.getByNameIPStmt.QueryRowContext(ctx, name, ip, port).Scan(
		&idStr,
		&agent.Name,
		&agent.IPAddress,
		&agent.Port,
		&agent.Status,
		&agent.Capabilities,
		&agent.AgentKey,
		&agent.Speed,
		&agent.LastSeen,
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
	r.cache.Set(ctx, cacheKey, &agent)

	return &agent, nil
}

func (r *agentRepository) GetAll(ctx context.Context) ([]domain.Agent, error) {
	cacheKey := "agents:all"

	var agents []domain.Agent
	if found, err := r.cache.Get(ctx, cacheKey, &agents); err == nil && found {
		return agents, nil
	}

	rows, err := r.getAllStmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	agents = make([]domain.Agent, 0, 10)
	for rows.Next() {
		var agent domain.Agent
		var idStr string

		err := rows.Scan(
			&idStr,
			&agent.Name,
			&agent.IPAddress,
			&agent.Port,
			&agent.Status,
			&agent.Capabilities,
			&agent.AgentKey,
			&agent.Speed,
			&agent.LastSeen,
			&agent.CreatedAt,
			&agent.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		agent.ID = uuid.MustParse(idStr)
		agents = append(agents, agent)
	}

	r.cache.Set(ctx, cacheKey, agents)

	return agents, nil
}

func (r *agentRepository) Update(ctx context.Context, agent *domain.Agent) error {
	// Invalidate cache before update
	if agent.IPAddress != "" {
		cacheKey := "agent:ip:" + agent.IPAddress
		r.cache.Delete(ctx, cacheKey)
	}

	// Invalidate cache for agent key
	if agent.AgentKey != "" {
		cacheKey := "agent:key:" + agent.AgentKey
		r.cache.Delete(ctx, cacheKey)
	}

	// Invalidate cache for agent ID
	cacheKey := "agent:id:" + agent.ID.String()
	r.cache.Delete(ctx, cacheKey)

	if r.updateStmt == nil {
		return fmt.Errorf("updateStmt is not prepared")
	}

	_, err := r.updateStmt.ExecContext(ctx,
		agent.Name,
		agent.IPAddress,
		agent.Port,
		agent.Status,
		agent.Capabilities,
		agent.AgentKey,
		agent.Speed,
		agent.LastSeen,
		agent.UpdatedAt,
		agent.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update agent: %w", err)
	}

	return nil
}

func (r *agentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.deleteStmt.ExecContext(ctx, id.String())

	if err == nil {
		r.cache.Delete(ctx, "agent:"+id.String())
		r.cache.Delete(ctx, "agents:all")
	}

	return err
}

func (r *agentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE agents SET status = ?, updated_at = ? WHERE id = ?
	`
	now := time.Now()
	_, err := r.db.DB().ExecContext(ctx, query, status, now, id.String())

	if err == nil {
		r.cache.Delete(ctx, "agent:"+id.String())
		r.cache.Delete(ctx, "agents:all")
	}

	return err
}

func (r *agentRepository) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE agents SET last_seen = ?, updated_at = ? WHERE id = ?
	`
	now := time.Now()
	_, err := r.db.DB().ExecContext(ctx, query, now, now, id.String())

	if err == nil {
		r.cache.Delete(ctx, "agent:"+id.String())
		r.cache.Delete(ctx, "agents:all")
	}

	return err
}

func (r *agentRepository) UpdateSpeed(ctx context.Context, id uuid.UUID, speed int64) error {
	query := `
		UPDATE agents SET speed = ?, updated_at = ? WHERE id = ?
	`
	now := time.Now()
	_, err := r.db.DB().ExecContext(ctx, query, speed, now, id.String())

	if err == nil {
		r.cache.Delete(ctx, "agent:"+id.String())
		r.cache.Delete(ctx, "agents:all")
	}

	return err
}

func (r *agentRepository) GetByIPAddress(ctx context.Context, ip string) (*domain.Agent, error) {
	cacheKey := "agent:ip:" + ip

	var agent domain.Agent
	if found, err := r.cache.Get(ctx, cacheKey, &agent); err == nil && found {
		return &agent, nil
	}

	if r.getByIPAddressStmt == nil {
		return nil, fmt.Errorf("getByIPAddressStmt is not prepared")
	}

	var idStr string
	err := r.getByIPAddressStmt.QueryRowContext(ctx, ip).Scan(
		&idStr,
		&agent.Name,
		&agent.IPAddress,
		&agent.Port,
		&agent.Status,
		&agent.Capabilities,
		&agent.AgentKey,
		&agent.Speed,
		&agent.LastSeen,
		&agent.CreatedAt,
		&agent.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrAgentNotFound // <-- use custom error
		}
		return nil, err
	}

	agent.ID = uuid.MustParse(idStr)
	r.cache.Set(ctx, cacheKey, &agent)

	return &agent, nil
}

func (r *agentRepository) GetByAgentKey(ctx context.Context, agentKey string) (*domain.Agent, error) {
	cacheKey := "agent:key:" + agentKey

	var agent domain.Agent
	if found, err := r.cache.Get(ctx, cacheKey, &agent); err == nil && found {
		return &agent, nil
	}

	if r.getByAgentKeyStmt == nil {
		return nil, fmt.Errorf("getByAgentKeyStmt is not prepared")
	}

	var idStr string
	err := r.getByAgentKeyStmt.QueryRowContext(ctx, agentKey).Scan(
		&idStr,
		&agent.Name,
		&agent.IPAddress,
		&agent.Port,
		&agent.Status,
		&agent.Capabilities,
		&agent.AgentKey,
		&agent.Speed,
		&agent.LastSeen,
		&agent.CreatedAt,
		&agent.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrAgentNotFound
		}
		return nil, err
	}

	agent.ID = uuid.MustParse(idStr)
	r.cache.Set(ctx, cacheKey, &agent)

	return &agent, nil
}

func (r *agentRepository) CreateAgent(ctx context.Context, agent *domain.Agent) error {
	return r.Create(ctx, agent)
}

func (r *agentRepository) UpdateAgent(ctx context.Context, agent *domain.Agent) error {
	return r.Update(ctx, agent)
}

func (r *agentRepository) GetByNameAndIPForStartup(ctx context.Context, name, ip string, port int) (*domain.Agent, error) {
	agent, err := r.GetByNameAndIP(ctx, name, ip, port)
	if err != nil {
		if err == domain.ErrAgentNotFound {
			return nil, domain.ErrAgentNotFound
		}
		return nil, err
	}
	return agent, nil
}
