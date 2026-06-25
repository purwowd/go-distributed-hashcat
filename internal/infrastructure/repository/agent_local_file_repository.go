package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go-distributed-hashcat/internal/domain"
	"go-distributed-hashcat/internal/infrastructure/database"

	"github.com/google/uuid"
)

type agentLocalFileRepository struct {
	db *database.SQLiteDB
}

func NewAgentLocalFileRepository(db *database.SQLiteDB) domain.AgentLocalFileRepository {
	return &agentLocalFileRepository{db: db}
}

func (r *agentLocalFileRepository) ReplaceInventory(ctx context.Context, agentID uuid.UUID, files []domain.AgentLocalFile) error {
	tx, err := r.db.DB().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM agent_local_files WHERE agent_id = ?`, agentID.String()); err != nil {
		return fmt.Errorf("delete old inventory: %w", err)
	}

	now := time.Now()
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO agent_local_files (id, agent_id, name, path, size, type, hash, mod_time, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, file := range files {
		fileID := file.ID
		if fileID == uuid.Nil {
			fileID = uuid.New()
		}
		modTime := file.ModTime
		if modTime.IsZero() {
			modTime = now
		}
		if _, err := stmt.ExecContext(
			ctx,
			fileID.String(),
			agentID.String(),
			file.Name,
			file.Path,
			file.Size,
			file.Type,
			nullIfEmpty(file.Hash),
			modTime,
			now,
		); err != nil {
			return fmt.Errorf("insert file %s: %w", file.Name, err)
		}
	}

	return tx.Commit()
}

func (r *agentLocalFileRepository) GetByAgentID(ctx context.Context, agentID uuid.UUID) ([]domain.AgentLocalFile, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, agent_id, name, path, size, type, COALESCE(hash, ''), mod_time, updated_at
		FROM agent_local_files
		WHERE agent_id = ?
		ORDER BY name ASC
	`, agentID.String())
	if err != nil {
		return nil, fmt.Errorf("query agent files: %w", err)
	}
	defer rows.Close()

	return scanAgentLocalFiles(rows)
}

func (r *agentLocalFileRepository) ListWithAgents(ctx context.Context, fileType, name string) ([]domain.AgentLocalFileEntry, error) {
	query := `
		SELECT
			f.id, f.agent_id, f.name, f.path, f.size, f.type, COALESCE(f.hash, ''), f.mod_time, f.updated_at,
			a.name, a.status
		FROM agent_local_files f
		INNER JOIN agents a ON a.id = f.agent_id
		WHERE 1=1
	`
	args := make([]interface{}, 0, 2)

	if fileType != "" {
		query += ` AND f.type = ?`
		args = append(args, fileType)
	}
	if name != "" {
		query += ` AND LOWER(f.name) = LOWER(?)`
		args = append(args, name)
	}
	query += ` ORDER BY f.name ASC, a.name ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query inventory: %w", err)
	}
	defer rows.Close()

	var out []domain.AgentLocalFileEntry
	for rows.Next() {
		var entry domain.AgentLocalFileEntry
		var idStr, agentIDStr string
		var modTime, updatedAt sql.NullTime
		if err := rows.Scan(
			&idStr, &agentIDStr, &entry.Name, &entry.Path, &entry.Size, &entry.Type,
			&entry.Hash, &modTime, &updatedAt,
			&entry.AgentName, &entry.AgentStatus,
		); err != nil {
			return nil, fmt.Errorf("scan inventory row: %w", err)
		}
		entry.ID, _ = uuid.Parse(idStr)
		entry.AgentID, _ = uuid.Parse(agentIDStr)
		if modTime.Valid {
			entry.ModTime = modTime.Time
		}
		if updatedAt.Valid {
			entry.UpdatedAt = updatedAt.Time
		}
		out = append(out, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func scanAgentLocalFiles(rows *sql.Rows) ([]domain.AgentLocalFile, error) {
	var out []domain.AgentLocalFile
	for rows.Next() {
		var file domain.AgentLocalFile
		var idStr, agentIDStr string
		var modTime, updatedAt sql.NullTime
		if err := rows.Scan(
			&idStr, &agentIDStr, &file.Name, &file.Path, &file.Size, &file.Type,
			&file.Hash, &modTime, &updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan agent file: %w", err)
		}
		file.ID, _ = uuid.Parse(idStr)
		file.AgentID, _ = uuid.Parse(agentIDStr)
		if modTime.Valid {
			file.ModTime = modTime.Time
		}
		if updatedAt.Valid {
			file.UpdatedAt = updatedAt.Time
		}
		out = append(out, file)
	}
	return out, rows.Err()
}

func (r *agentLocalFileRepository) AgentHasWordlist(ctx context.Context, agentID uuid.UUID, origName string) (bool, error) {
	name := strings.TrimSpace(origName)
	if name == "" {
		return false, nil
	}
	var count int
	err := r.db.DB().QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM agent_local_files
		WHERE agent_id = ? AND type = 'wordlist' AND LOWER(name) = LOWER(?)
	`, agentID.String(), name).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check agent wordlist: %w", err)
	}
	return count > 0, nil
}

func nullIfEmpty(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
