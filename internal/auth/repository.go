package auth

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateUser(ctx context.Context, recoveryHash string) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (recovery_hash) VALUES ($1) RETURNING id`,
		recoveryHash,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("create user: %w", err)
	}
	return id, nil
}

func (r *Repository) GetUserByHash(ctx context.Context, recoveryHash string) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`SELECT id FROM users WHERE recovery_hash = $1`,
		recoveryHash,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}
	return id, nil
}
