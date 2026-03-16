package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Payment struct {
	ID            string
	UserID        string
	MoneroAddress string
	MoneroAmount  float64
	Plan          string
	Country       string
	Status        string
	CreatedAt     time.Time
	ConfirmedAt   *time.Time
	NumberID      *string
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, p *Payment) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO payments (user_id, monero_address, monero_amount, plan, country)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`,
		p.UserID, p.MoneroAddress, p.MoneroAmount, p.Plan, p.Country,
	).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return fmt.Errorf("create payment: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id, userID string) (*Payment, error) {
	p := &Payment{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, monero_address, monero_amount, plan, country, status,
		        created_at, confirmed_at, number_id
		 FROM payments WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&p.ID, &p.UserID, &p.MoneroAddress, &p.MoneroAmount, &p.Plan, &p.Country,
		&p.Status, &p.CreatedAt, &p.ConfirmedAt, &p.NumberID)
	if err != nil {
		return nil, fmt.Errorf("get payment: %w", err)
	}
	return p, nil
}

func (r *Repository) ListPending(ctx context.Context) ([]*Payment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, monero_address, monero_amount, plan, country, created_at
		 FROM payments WHERE status = 'pending'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Payment
	for rows.Next() {
		p := &Payment{}
		if err := rows.Scan(&p.ID, &p.UserID, &p.MoneroAddress, &p.MoneroAmount,
			&p.Plan, &p.Country, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repository) Confirm(ctx context.Context, id, numberID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payments SET status='confirmed', confirmed_at=NOW(), number_id=$2
		 WHERE id=$1`,
		id, numberID)
	return err
}

func (r *Repository) ExpireOld(ctx context.Context) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payments SET status='expired'
		 WHERE status='pending' AND created_at < NOW() - INTERVAL '2 hours'`)
	return err
}
