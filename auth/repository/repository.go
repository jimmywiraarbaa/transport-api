// Package repository implements the auth usecase's Repository interface
// using sqlc-generated queries on top of pgx.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jimmywiraarbaa/transport-api/auth/domain"
	"github.com/jimmywiraarbaa/transport-api/internal/database/sqlcgen"
)

// ErrUserNotFound is returned when no user matches the lookup.
var ErrUserNotFound = errors.New("user not found")

// postgresRepository satisfies auth.usecase.Repository.
type postgresRepository struct {
	db *pgxpool.Pool
	q  *sqlcgen.Queries
}

// New builds a repository bound to the given connection pool.
func New(db *pgxpool.Pool) *postgresRepository {
	return &postgresRepository{
		db: db,
		q:  sqlcgen.New(db),
	}
}

// Create inserts a new user and returns it.
func (r *postgresRepository) Create(ctx context.Context, email, username, passwordHash, name string) (*domain.User, error) {
	row, err := r.q.CreateUser(ctx, sqlcgen.CreateUserParams{
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		Name:         name,
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return toDomain(row), nil
}

// GetByID returns the user with the given id.
func (r *postgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return toDomain(row), nil
}

// GetByUsername returns the user matching the username.
func (r *postgresRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	row, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return toDomain(row), nil
}

func toDomain(u sqlcgen.User) *domain.User {
	return &domain.User{
		ID:           u.ID,
		Email:        u.Email,
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
		Name:         u.Name,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}
