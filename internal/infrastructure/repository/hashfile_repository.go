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

type hashFileRepository struct {
	db          *database.SQLiteDB
	cache       cache.Cache
	getByIDStmt *sql.Stmt
	getByOrigNameStmt *sql.Stmt
	getAllStmt  *sql.Stmt
	deleteStmt  *sql.Stmt
}

func NewHashFileRepository(db *database.SQLiteDB) domain.HashFileRepository {
	repo := &hashFileRepository{
		db:    db,
		cache: cache.NewMemoryCache(60 * time.Second), // 60 second cache for hash files (they change less frequently)
	}

	// Prepare frequently used statements
	repo.prepareStatements()

	return repo
}

func (r *hashFileRepository) prepareStatements() {
	var err error

	r.getByIDStmt, err = r.db.DB().Prepare(`
		SELECT id, name, orig_name, path, size, type, created_at
		FROM hash_files WHERE id = ? LIMIT 1
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare getByID statement: %v", err))
	}

	r.getByOrigNameStmt, err = r.db.DB().Prepare(`
		SELECT id FROM hash_files WHERE orig_name = ? LIMIT 1
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare getByOrigName statement: %v", err))
	}

	r.getAllStmt, err = r.db.DB().Prepare(`
		SELECT id, name, orig_name, path, size, type, created_at
		FROM hash_files ORDER BY created_at DESC LIMIT 50
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare getAll statement: %v", err))
	}

	r.deleteStmt, err = r.db.DB().Prepare(`DELETE FROM hash_files WHERE id = ?`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare delete statement: %v", err))
	}
}

func (r *hashFileRepository) Create(ctx context.Context, hashFile *domain.HashFile) error {
	query := `
		INSERT INTO hash_files (id, name, orig_name, path, size, type, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	hashFile.CreatedAt = time.Now()

	_, err := r.db.DB().ExecContext(ctx, query,
		hashFile.ID.String(),
		hashFile.Name,
		hashFile.OrigName,
		hashFile.Path,
		hashFile.Size,
		hashFile.Type,
		hashFile.CreatedAt,
	)

	if err == nil {
		// Cache the new hash file
		r.cache.Set(ctx, "hashfile:"+hashFile.ID.String(), hashFile)
		// Invalidate list cache
		r.cache.Delete(ctx, "hashfiles:all")
	}

	return err
}

func (r *hashFileRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.HashFile, error) {
	cacheKey := "hashfile:" + id.String()

	// Try cache first
	var hashFile domain.HashFile
	if found, err := r.cache.Get(ctx, cacheKey, &hashFile); err == nil && found {
		return &hashFile, nil
	}

	// Fallback to database with prepared statement
	var idStr string

	err := r.getByIDStmt.QueryRowContext(ctx, id.String()).Scan(
		&idStr,
		&hashFile.Name,
		&hashFile.OrigName,
		&hashFile.Path,
		&hashFile.Size,
		&hashFile.Type,
		&hashFile.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("hash file not found")
		}
		return nil, err
	}

	hashFile.ID = uuid.MustParse(idStr)

	// Cache the result
	r.cache.Set(ctx, cacheKey, &hashFile)

	return &hashFile, nil
}

func (r *hashFileRepository) GetAll(ctx context.Context) ([]domain.HashFile, error) {
	cacheKey := "hashfiles:all"

	// Try cache first
	var hashFiles []domain.HashFile
	if found, err := r.cache.Get(ctx, cacheKey, &hashFiles); err == nil && found {
		return hashFiles, nil
	}

	// Fallback to database with prepared statement
	rows, err := r.getAllStmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hashFiles = make([]domain.HashFile, 0, 10) // Pre-allocate slice
	for rows.Next() {
		var hashFile domain.HashFile
		var idStr string

		err := rows.Scan(
			&idStr,
			&hashFile.Name,
			&hashFile.OrigName,
			&hashFile.Path,
			&hashFile.Size,
			&hashFile.Type,
			&hashFile.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		hashFile.ID = uuid.MustParse(idStr)
		hashFiles = append(hashFiles, hashFile)
	}

	// Cache the result
	r.cache.Set(ctx, cacheKey, hashFiles)

	return hashFiles, nil
}

func (r *hashFileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.deleteStmt.ExecContext(ctx, id.String())

	if err == nil {
		// Remove from cache
		r.cache.Delete(ctx, "hashfile:"+id.String())
		// Invalidate list cache
		r.cache.Delete(ctx, "hashfiles:all")
	}

	return err
}

func (r *hashFileRepository) GetByOrigName(ctx context.Context, origName string) (*domain.HashFile, error) {
	var idStr string
	err := r.getByOrigNameStmt.QueryRowContext(ctx, origName).Scan(&idStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("hash file not found")
		}
		return nil, err
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hash file ID format: %w", err)
	}

	return r.GetByID(ctx, id)
}
