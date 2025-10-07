package repository

import (
	"context"
	"database/sql"
	"time"

	"go-distributed-hashcat/internal/domain"

	"github.com/google/uuid"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, username, email, password, role, is_active, created_at, updated_at, last_login)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID.String(),
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
		user.LastLogin,
	)

	return err
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, username, email, password, role, is_active, created_at, updated_at, last_login
		FROM users
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id.String())

	user := &domain.User{}
	var lastLogin sql.NullTime

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLogin,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &domain.UserNotFoundError{Username: id.String()}
		}
		return nil, err
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return user, nil
}

// GetByUsername retrieves a user by username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password, role, is_active, created_at, updated_at, last_login
		FROM users
		WHERE username = ?
	`

	row := r.db.QueryRowContext(ctx, query, username)

	user := &domain.User{}
	var lastLogin sql.NullTime

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLogin,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &domain.UserNotFoundError{Username: username}
		}
		return nil, err
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password, role, is_active, created_at, updated_at, last_login
		FROM users
		WHERE email = ?
	`

	row := r.db.QueryRowContext(ctx, query, email)

	user := &domain.User{}
	var lastLogin sql.NullTime

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLogin,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &domain.UserNotFoundError{Username: email}
		}
		return nil, err
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return user, nil
}

// GetAll retrieves all users
func (r *userRepository) GetAll(ctx context.Context) ([]domain.User, error) {
	query := `
		SELECT id, username, email, password, role, is_active, created_at, updated_at, last_login
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User

	for rows.Next() {
		user := domain.User{}
		var lastLogin sql.NullTime

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
			&lastLogin,
		)

		if err != nil {
			return nil, err
		}

		if lastLogin.Valid {
			user.LastLogin = &lastLogin.Time
		}

		users = append(users, user)
	}

	return users, nil
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET username = ?, email = ?, password = ?, role = ?, is_active = ?, updated_at = ?, last_login = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		user.IsActive,
		user.UpdatedAt,
		user.LastLogin,
		user.ID.String(),
	)

	return err
}

// Delete deletes a user
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return &domain.UserNotFoundError{Username: id.String()}
	}

	return nil
}

// UpdateLastLogin updates the last login time for a user
func (r *userRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET last_login = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, now, now, id.String())

	return err
}
