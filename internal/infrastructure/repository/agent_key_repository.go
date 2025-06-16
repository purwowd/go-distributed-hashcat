package repository

import (
	"context"
	"database/sql"
	"time"

	"go-distributed-hashcat/internal/domain"
	"go-distributed-hashcat/internal/infrastructure/database"

	"github.com/google/uuid"
)

type agentKeyRepository struct {
	db *database.SQLiteDB
}

func NewAgentKeyRepository(db *database.SQLiteDB) domain.AgentKeyRepository {
	return &agentKeyRepository{
		db: db,
	}
}

func (r *agentKeyRepository) Create(ctx context.Context, agentKey *domain.AgentKey) error {
	query := `
		INSERT INTO agent_keys (id, agent_key, name, description, status, created_at, expires_at, last_used_at, agent_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	agentKey.CreatedAt = now

	var expiresAt interface{}
	if agentKey.ExpiresAt != nil {
		expiresAt = *agentKey.ExpiresAt
	}

	var lastUsedAt interface{}
	if agentKey.LastUsedAt != nil {
		lastUsedAt = *agentKey.LastUsedAt
	}

	var agentID interface{}
	if agentKey.AgentID != nil {
		agentID = agentKey.AgentID.String()
	}

	_, err := r.db.DB().ExecContext(ctx, query,
		agentKey.ID.String(),
		agentKey.AgentKey,
		agentKey.Name,
		agentKey.Description,
		agentKey.Status,
		agentKey.CreatedAt,
		expiresAt,
		lastUsedAt,
		agentID,
	)

	return err
}

func (r *agentKeyRepository) GetByKey(ctx context.Context, key string) (*domain.AgentKey, error) {
	query := `
		SELECT id, agent_key, name, description, status, created_at, expires_at, last_used_at, agent_id
		FROM agent_keys
		WHERE agent_key = ?
	`

	var agentKey domain.AgentKey
	var expiresAt sql.NullTime
	var lastUsedAt sql.NullTime
	var agentID sql.NullString

	err := r.db.DB().QueryRowContext(ctx, query, key).Scan(
		&agentKey.ID,
		&agentKey.AgentKey,
		&agentKey.Name,
		&agentKey.Description,
		&agentKey.Status,
		&agentKey.CreatedAt,
		&expiresAt,
		&lastUsedAt,
		&agentID,
	)

	if err != nil {
		return nil, err
	}

	if expiresAt.Valid {
		agentKey.ExpiresAt = &expiresAt.Time
	}

	if lastUsedAt.Valid {
		agentKey.LastUsedAt = &lastUsedAt.Time
	}

	if agentID.Valid {
		id, err := uuid.Parse(agentID.String)
		if err == nil {
			agentKey.AgentID = &id
		}
	}

	return &agentKey, nil
}

func (r *agentKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.AgentKey, error) {
	query := `
		SELECT id, agent_key, name, description, status, created_at, expires_at, last_used_at, agent_id
		FROM agent_keys
		WHERE id = ?
	`

	var agentKey domain.AgentKey
	var expiresAt sql.NullTime
	var lastUsedAt sql.NullTime
	var agentID sql.NullString

	err := r.db.DB().QueryRowContext(ctx, query, id.String()).Scan(
		&agentKey.ID,
		&agentKey.AgentKey,
		&agentKey.Name,
		&agentKey.Description,
		&agentKey.Status,
		&agentKey.CreatedAt,
		&expiresAt,
		&lastUsedAt,
		&agentID,
	)

	if err != nil {
		return nil, err
	}

	if expiresAt.Valid {
		agentKey.ExpiresAt = &expiresAt.Time
	}

	if lastUsedAt.Valid {
		agentKey.LastUsedAt = &lastUsedAt.Time
	}

	if agentID.Valid {
		id, err := uuid.Parse(agentID.String)
		if err == nil {
			agentKey.AgentID = &id
		}
	}

	return &agentKey, nil
}

func (r *agentKeyRepository) GetAll(ctx context.Context) ([]*domain.AgentKey, error) {
	query := `
		SELECT id, agent_key, name, description, status, created_at, expires_at, last_used_at, agent_id
		FROM agent_keys
		ORDER BY created_at DESC
	`

	rows, err := r.db.DB().QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agentKeys []*domain.AgentKey

	for rows.Next() {
		var agentKey domain.AgentKey
		var expiresAt sql.NullTime
		var lastUsedAt sql.NullTime
		var agentID sql.NullString

		err := rows.Scan(
			&agentKey.ID,
			&agentKey.AgentKey,
			&agentKey.Name,
			&agentKey.Description,
			&agentKey.Status,
			&agentKey.CreatedAt,
			&expiresAt,
			&lastUsedAt,
			&agentID,
		)

		if err != nil {
			return nil, err
		}

		if expiresAt.Valid {
			agentKey.ExpiresAt = &expiresAt.Time
		}

		if lastUsedAt.Valid {
			agentKey.LastUsedAt = &lastUsedAt.Time
		}

		if agentID.Valid {
			id, err := uuid.Parse(agentID.String)
			if err == nil {
				agentKey.AgentID = &id
			}
		}

		agentKeys = append(agentKeys, &agentKey)
	}

	return agentKeys, rows.Err()
}

func (r *agentKeyRepository) Update(ctx context.Context, agentKey *domain.AgentKey) error {
	query := `
		UPDATE agent_keys
		SET name = ?, description = ?, status = ?, expires_at = ?, last_used_at = ?, agent_id = ?
		WHERE id = ?
	`

	var expiresAt interface{}
	if agentKey.ExpiresAt != nil {
		expiresAt = *agentKey.ExpiresAt
	}

	var lastUsedAt interface{}
	if agentKey.LastUsedAt != nil {
		lastUsedAt = *agentKey.LastUsedAt
	}

	var agentID interface{}
	if agentKey.AgentID != nil {
		agentID = agentKey.AgentID.String()
	}

	_, err := r.db.DB().ExecContext(ctx, query,
		agentKey.Name,
		agentKey.Description,
		agentKey.Status,
		expiresAt,
		lastUsedAt,
		agentID,
		agentKey.ID.String(),
	)

	return err
}

func (r *agentKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM agent_keys WHERE id = ?`
	_, err := r.db.DB().ExecContext(ctx, query, id.String())
	return err
}

func (r *agentKeyRepository) UpdateStatus(ctx context.Context, key string, status string) error {
	query := `UPDATE agent_keys SET status = ? WHERE agent_key = ?`
	_, err := r.db.DB().ExecContext(ctx, query, status, key)
	return err
}

func (r *agentKeyRepository) UpdateLastUsed(ctx context.Context, key string) error {
	query := `UPDATE agent_keys SET last_used_at = ? WHERE agent_key = ?`
	_, err := r.db.DB().ExecContext(ctx, query, time.Now(), key)
	return err
}

func (r *agentKeyRepository) LinkToAgent(ctx context.Context, key string, agentID uuid.UUID) error {
	query := `UPDATE agent_keys SET agent_id = ?, last_used_at = ? WHERE agent_key = ?`
	_, err := r.db.DB().ExecContext(ctx, query, agentID.String(), time.Now(), key)
	return err
}
