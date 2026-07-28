package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"library/internal/models"
)

var (
	ErrUserNotFound              = errors.New("user not found")
	ErrDuplicateEmail            = errors.New("email address is already registered")
	ErrLibraryAlreadyHasEmployee = errors.New("this library already has an assigned employee")
)

type UserRepository interface {
	Create(ctx context.Context, exec GormExecutor, user *models.BookMember) error
	GetByEmail(ctx context.Context, exec GormExecutor, email string) (*models.BookMember, error)
	GetByID(ctx context.Context, exec GormExecutor, id int) (*models.BookMember, error)
}

type userRepository struct{}

func NewUserRepository() UserRepository {
	return &userRepository{}
}

// Create inserts a new user record into PostgreSQL and returns the generated ID & CreatedAt timestamp
func (r *userRepository) Create(ctx context.Context, exec GormExecutor, user *models.BookMember) error {
	query := `
		INSERT INTO members (email, password, role)
		VALUES ($1, $2, $3)
		RETURNING id, joined_at;
	`

	err := exec.QueryRowContext(ctx, query, user.Email, user.Password, user.Role).Scan(&user.ID, &user.JoinedAt)
	if err != nil {
		// Handle Postgres unique constraint violation (SQLSTATE 23505) for duplicate emails
		if isDuplicateKeyError(err) {
			return ErrDuplicateEmail
		}
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

// GetByEmail retrieves a user by their email address (used during login)
func (r *userRepository) GetByEmail(ctx context.Context, exec GormExecutor, email string) (*models.BookMember, error) {
	query := `
		SELECT id, email, password, role, joined_at
		FROM members
		WHERE email = $1;
	`

	var user models.BookMember
	err := exec.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.JoinedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user by email: %w", err)
	}

	return &user, nil
}

// GetByID retrieves a user by their primary key ID
func (r *userRepository) GetByID(ctx context.Context, exec GormExecutor, id int) (*models.BookMember, error) {
	query := `
		SELECT id, email, password, role, joined_at
		FROM members
		WHERE id = $1;
	`

	var user models.BookMember
	err := exec.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.JoinedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user by ID: %w", err)
	}

	return &user, nil
}

// Helper to catch duplicate email errors from PostgreSQL driver
func isDuplicateKeyError(err error) bool {
	return err != nil && (contains(err.Error(), "unique constraint") || contains(err.Error(), "23505"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
