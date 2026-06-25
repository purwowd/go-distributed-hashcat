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
	getAllStmt  *sql.Stmt
	deleteStmt  *sql.Stmt
}

func NewWordlistRepository(db *database.SQLiteDB) domain.WordlistRepository {
	repo := &wordlistRepository{
		db:    db,
		cache: cache.NewMemoryCache(60 * time.Second),
	}
	repo.prepareStatements()
	return repo
}

func (r *wordlistRepository) prepareStatements() {
	var err error

	r.getByIDStmt, err = r.db.DB().Prepare(`
		SELECT id, name, orig_name, path, size, word_count, COALESCE(source, 'uploaded'), created_at
		FROM wordlists WHERE id = ? LIMIT 1
	`)
	if err != nil {
		panic(fmt.Sprintf("Failed to prepare getByID statement: %v", err))
	}

	r.getAllStmt, err = r.db.DB().Prepare(`
		SELECT id, name, orig_name, path, size, word_count, COALESCE(source, 'uploaded'), created_at
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

func (r *wordlistRepository) scanWordlist(idStr string, wordlist *domain.Wordlist, wordCount sql.NullInt64, source sql.NullString, createdAt time.Time) {
	wordlist.ID = uuid.MustParse(idStr)
	if wordCount.Valid {
		wordlist.WordCount = &wordCount.Int64
	}
	if source.Valid && source.String != "" {
		wordlist.Source = source.String
	} else {
		wordlist.Source = domain.WordlistSourceUploaded
	}
	wordlist.CreatedAt = createdAt
}

func (r *wordlistRepository) Create(ctx context.Context, wordlist *domain.Wordlist) error {
	if wordlist.Source == "" {
		wordlist.Source = domain.WordlistSourceUploaded
	}
	query := `
		INSERT INTO wordlists (id, name, orig_name, path, size, word_count, source, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	wordlist.CreatedAt = time.Now()

	_, err := r.db.DB().ExecContext(ctx, query,
		wordlist.ID.String(),
		wordlist.Name,
		wordlist.OrigName,
		wordlist.Path,
		wordlist.Size,
		wordlist.WordCount,
		wordlist.Source,
		wordlist.CreatedAt,
	)
	if err == nil {
		r.cache.Set(ctx, "wordlist:"+wordlist.ID.String(), wordlist)
		r.cache.Delete(ctx, "wordlists:all")
		r.cache.Delete(ctx, "wordlist:orig:"+strings.ToLower(wordlist.OrigName))
	}
	return err
}

func (r *wordlistRepository) Update(ctx context.Context, wordlist *domain.Wordlist) error {
	if wordlist.Source == "" {
		wordlist.Source = domain.WordlistSourceUploaded
	}
	query := `
		UPDATE wordlists
		SET name = ?, orig_name = ?, path = ?, size = ?, word_count = ?, source = ?
		WHERE id = ?
	`
	_, err := r.db.DB().ExecContext(ctx, query,
		wordlist.Name,
		wordlist.OrigName,
		wordlist.Path,
		wordlist.Size,
		wordlist.WordCount,
		wordlist.Source,
		wordlist.ID.String(),
	)
	if err == nil {
		r.cache.Delete(ctx, "wordlist:"+wordlist.ID.String())
		r.cache.Delete(ctx, "wordlists:all")
		r.cache.Delete(ctx, "wordlist:orig:"+strings.ToLower(wordlist.OrigName))
	}
	return err
}

func (r *wordlistRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Wordlist, error) {
	cacheKey := "wordlist:" + id.String()
	var wordlist domain.Wordlist
	if found, err := r.cache.Get(ctx, cacheKey, &wordlist); err == nil && found {
		return &wordlist, nil
	}

	var idStr string
	var wordCount sql.NullInt64
	var source sql.NullString

	err := r.getByIDStmt.QueryRowContext(ctx, id.String()).Scan(
		&idStr,
		&wordlist.Name,
		&wordlist.OrigName,
		&wordlist.Path,
		&wordlist.Size,
		&wordCount,
		&source,
		&wordlist.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("wordlist not found")
		}
		return nil, err
	}

	r.scanWordlist(idStr, &wordlist, wordCount, source, wordlist.CreatedAt)
	r.cache.Set(ctx, cacheKey, &wordlist)
	return &wordlist, nil
}

func (r *wordlistRepository) GetByOrigName(ctx context.Context, origName string) (*domain.Wordlist, error) {
	key := strings.ToLower(strings.TrimSpace(origName))
	if key == "" {
		return nil, fmt.Errorf("wordlist not found")
	}

	cacheKey := "wordlist:orig:" + key
	var wordlist domain.Wordlist
	if found, err := r.cache.Get(ctx, cacheKey, &wordlist); err == nil && found {
		return &wordlist, nil
	}

	query := `
		SELECT id, name, orig_name, path, size, word_count, COALESCE(source, 'uploaded'), created_at
		FROM wordlists
		WHERE LOWER(orig_name) = ?
		ORDER BY CASE WHEN source = 'uploaded' THEN 0 ELSE 1 END, created_at DESC
		LIMIT 1
	`
	var idStr string
	var wordCount sql.NullInt64
	var source sql.NullString
	err := r.db.DB().QueryRowContext(ctx, query, key).Scan(
		&idStr,
		&wordlist.Name,
		&wordlist.OrigName,
		&wordlist.Path,
		&wordlist.Size,
		&wordCount,
		&source,
		&wordlist.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("wordlist not found")
		}
		return nil, err
	}

	r.scanWordlist(idStr, &wordlist, wordCount, source, wordlist.CreatedAt)
	r.cache.Set(ctx, cacheKey, &wordlist)
	r.cache.Set(ctx, "wordlist:"+wordlist.ID.String(), &wordlist)
	return &wordlist, nil
}

func (r *wordlistRepository) GetAll(ctx context.Context) ([]domain.Wordlist, error) {
	cacheKey := "wordlists:all"
	var wordlists []domain.Wordlist
	if found, err := r.cache.Get(ctx, cacheKey, &wordlists); err == nil && found {
		return wordlists, nil
	}

	rows, err := r.getAllStmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	wordlists = make([]domain.Wordlist, 0, 10)
	for rows.Next() {
		var wordlist domain.Wordlist
		var idStr string
		var wordCount sql.NullInt64
		var source sql.NullString

		err := rows.Scan(
			&idStr,
			&wordlist.Name,
			&wordlist.OrigName,
			&wordlist.Path,
			&wordlist.Size,
			&wordCount,
			&source,
			&wordlist.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		r.scanWordlist(idStr, &wordlist, wordCount, source, wordlist.CreatedAt)
		wordlists = append(wordlists, wordlist)
	}

	r.cache.Set(ctx, cacheKey, wordlists)
	return wordlists, nil
}

func (r *wordlistRepository) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := r.GetByID(ctx, id)
	if err == nil && existing != nil {
		r.cache.Delete(ctx, "wordlist:orig:"+strings.ToLower(existing.OrigName))
	}

	_, err = r.deleteStmt.ExecContext(ctx, id.String())
	if err == nil {
		r.cache.Delete(ctx, "wordlist:"+id.String())
		r.cache.Delete(ctx, "wordlists:all")
	}
	return err
}
