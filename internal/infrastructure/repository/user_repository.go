package repository

import (
	"database/sql"
	"time"

	"go-distributed-hashcat/internal/domain"

	"github.com/google/uuid"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) domain.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(user *domain.User) error {
	query := `
		INSERT INTO users (id, username, email, password, role, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		user.ID.String(),
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

func (r *userRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, username, email, password, role, is_active, created_at, updated_at, last_login
		FROM users
		WHERE id = ?
	`

	var user domain.User
	var lastLogin sql.NullTime

	err := r.db.QueryRow(query, id.String()).Scan(
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

	return &user, nil
}

func (r *userRepository) GetByUsername(username string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password, role, is_active, created_at, updated_at, last_login
		FROM users
		WHERE username = ?
	`

	var user domain.User
	var lastLogin sql.NullTime
	var idStr string

	err := r.db.QueryRow(query, username).Scan(
		&idStr,
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

	// Parse UUID
	user.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password, role, is_active, created_at, updated_at, last_login
		FROM users
		WHERE email = ?
	`

	var user domain.User
	var lastLogin sql.NullTime
	var idStr string

	err := r.db.QueryRow(query, email).Scan(
		&idStr,
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

	// Parse UUID
	user.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return &user, nil
}

func (r *userRepository) GetAll() ([]*domain.User, error) {
	query := `
		SELECT id, username, email, password, role, is_active, created_at, updated_at, last_login
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User

	for rows.Next() {
		var user domain.User
		var lastLogin sql.NullTime
		var idStr string

		err := rows.Scan(
			&idStr,
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

		// Parse UUID
		user.ID, err = uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}

		if lastLogin.Valid {
			user.LastLogin = &lastLogin.Time
		}

		users = append(users, &user)
	}

	return users, nil
}

func (r *userRepository) Update(user *domain.User) error {
	query := `
		UPDATE users
		SET username = ?, email = ?, password = ?, role = ?, is_active = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		user.IsActive,
		user.UpdatedAt,
		user.ID.String(),
	)

	return err
}

func (r *userRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := r.db.Exec(query, id.String())
	return err
}

func (r *userRepository) UpdateLastLogin(id uuid.UUID) error {
	query := `UPDATE users SET last_login = ? WHERE id = ?`
	_, err := r.db.Exec(query, time.Now(), id.String())
	return err
}
