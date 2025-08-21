package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go-distributed-hashcat/internal/domain"
	"go-distributed-hashcat/internal/infrastructure/cache"
	"go-distributed-hashcat/internal/infrastructure/database"

	"github.com/google/uuid"
)

type wordlistRepository struct {
	db          *database.SQLiteDB
	cache       cache.Cache
	getByIDStmt *sql.Stmt
	getByOrigNameStmt *sql.Stmt
	getAllStmt  *sql.Stmt
	deleteStmt  *sql.Stmt
}

func NewWordlistRepository(db *database.SQLiteDB) domain.WordlistRepository {
	repo := &wordlistRepository{
		db:    db,
		cache: cache.NewMemoryCache(60 * time.Second), // 60 second cache for wordlists
	}

	// Prepare frequently used statements
	repo.prepareStatements()

	return repo
}

func (r *wordlistRepository) prepareStatements() {
	var err error

	r.getByIDStmt, err = r.db.DB().Prepare(`
		SELECT id, name, orig_name, path, size, word_count, created_at
		FROM wordlists WHERE id = ? LIMIT 1
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare getByID statement: %v", err))
	}

	r.getByOrigNameStmt, err = r.db.DB().Prepare(`
		SELECT id FROM wordlists WHERE orig_name = ? LIMIT 1
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare getByOrigName statement: %v", err))
	}

	r.getAllStmt, err = r.db.DB().Prepare(`
		SELECT id, name, orig_name, path, size, word_count, created_at
		FROM wordlists ORDER BY created_at DESC LIMIT 50
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare getAll statement: %v", err))
	}

	r.deleteStmt, err = r.db.DB().Prepare(`DELETE FROM wordlists WHERE id = ?`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare delete statement: %v", err))
	}
}

func (r *wordlistRepository) Create(ctx context.Context, wordlist *domain.Wordlist) error {
	query := `
		INSERT INTO wordlists (id, name, orig_name, path, size, word_count, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	wordlist.CreatedAt = time.Now()

	_, err := r.db.DB().ExecContext(ctx, query,
		wordlist.ID.String(),
		wordlist.Name,
		wordlist.OrigName,
		wordlist.Path,
		wordlist.Size,
		wordlist.WordCount,
		wordlist.CreatedAt,
	)

	if err != nil {
		// Check if it's a duplicate orig_name constraint violation
		if strings.Contains(err.Error(), "UNIQUE constraint failed") && strings.Contains(err.Error(), "orig_name") {
			return fmt.Errorf("file already exists: %s", wordlist.OrigName)
		}
		return err
	}

	// Cache the new wordlist
	r.cache.Set(ctx, "wordlist:"+wordlist.ID.String(), wordlist)
	// Invalidate list cache
	r.cache.Delete(ctx, "wordlists:all")

	return nil
}

func (r *wordlistRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Wordlist, error) {
	cacheKey := "wordlist:" + id.String()

	// Try cache first
	var wordlist domain.Wordlist
	if found, err := r.cache.Get(ctx, cacheKey, &wordlist); err == nil && found {
		return &wordlist, nil
	}

	// Fallback to database with prepared statement
	var idStr string
	var wordCount sql.NullInt64

	err := r.getByIDStmt.QueryRowContext(ctx, id.String()).Scan(
		&idStr,
		&wordlist.Name,
		&wordlist.OrigName,
		&wordlist.Path,
		&wordlist.Size,
		&wordCount,
		&wordlist.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("wordlist not found")
		}
		return nil, err
	}

	wordlist.ID = uuid.MustParse(idStr)

	if wordCount.Valid {
		wordlist.WordCount = &wordCount.Int64
	}

	// Cache the result
	r.cache.Set(ctx, cacheKey, &wordlist)

	return &wordlist, nil
}

func (r *wordlistRepository) GetByOrigName(ctx context.Context, origName string) (*domain.Wordlist, error) {
	var idStr string
	
	err := r.getByOrigNameStmt.QueryRowContext(ctx, origName).Scan(&idStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("wordlist not found")
		}
		return nil, err
	}

	// Parse UUID
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid wordlist ID format: %w", err)
	}

	// Get full wordlist details using existing GetByID method
	return r.GetByID(ctx, id)
}

func (r *wordlistRepository) GetAll(ctx context.Context) ([]domain.Wordlist, error) {
	cacheKey := "wordlists:all"

	// Try cache first
	var wordlists []domain.Wordlist
	if found, err := r.cache.Get(ctx, cacheKey, &wordlists); err == nil && found {
		return wordlists, nil
	}

	// Fallback to database with prepared statement
	rows, err := r.getAllStmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	wordlists = make([]domain.Wordlist, 0, 10) // Pre-allocate slice
	for rows.Next() {
		var wordlist domain.Wordlist
		var idStr string
		var wordCount sql.NullInt64

		err := rows.Scan(
			&idStr,
			&wordlist.Name,
			&wordlist.OrigName,
			&wordlist.Path,
			&wordlist.Size,
			&wordCount,
			&wordlist.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		wordlist.ID = uuid.MustParse(idStr)

		if wordCount.Valid {
			wordlist.WordCount = &wordCount.Int64
		}

		wordlists = append(wordlists, wordlist)
	}

	// Cache the result
	r.cache.Set(ctx, cacheKey, wordlists)

	return wordlists, nil
}

func (r *wordlistRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.deleteStmt.ExecContext(ctx, id.String())

	if err == nil {
		// Remove from cache
		r.cache.Delete(ctx, "wordlist:"+id.String())
		// Invalidate list cache
		r.cache.Delete(ctx, "wordlists:all")
	}

	return err
}
