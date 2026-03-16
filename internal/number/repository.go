package number

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PhoneNumber struct {
	ID           string
	UserID       string
	TelnyxNumber string
	Country      string
	Plan         string
	ExpiresAt    time.Time
	Released     bool
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, n *PhoneNumber) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO phone_numbers (user_id, telnyx_number, country, plan, expires_at)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		n.UserID, n.TelnyxNumber, n.Country, n.Plan, n.ExpiresAt,
	).Scan(&n.ID)
	if err != nil {
		return fmt.Errorf("create phone number: %w", err)
	}
	return nil
}

func (r *Repository) ListByUser(ctx context.Context, userID string) ([]*PhoneNumber, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, telnyx_number, country, plan, expires_at
		 FROM phone_numbers WHERE user_id = $1 AND released = FALSE`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*PhoneNumber
	for rows.Next() {
		n := &PhoneNumber{}
		if err := rows.Scan(&n.ID, &n.UserID, &n.TelnyxNumber, &n.Country, &n.Plan, &n.ExpiresAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, id, userID string) (*PhoneNumber, error) {
	n := &PhoneNumber{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, telnyx_number, country, plan, expires_at, released
		 FROM phone_numbers WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&n.ID, &n.UserID, &n.TelnyxNumber, &n.Country, &n.Plan, &n.ExpiresAt, &n.Released)
	if err != nil {
		return nil, fmt.Errorf("get phone number: %w", err)
	}
	return n, nil
}

func (r *Repository) GetByTelnyxNumber(ctx context.Context, telnyxNumber string) (*PhoneNumber, error) {
	n := &PhoneNumber{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, telnyx_number, country, plan, expires_at
		 FROM phone_numbers WHERE telnyx_number = $1 AND released = FALSE`,
		telnyxNumber,
	).Scan(&n.ID, &n.UserID, &n.TelnyxNumber, &n.Country, &n.Plan, &n.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("get phone number by telnyx: %w", err)
	}
	return n, nil
}

func (r *Repository) MarkReleased(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE phone_numbers SET released = TRUE WHERE id = $1`, id)
	return err
}

func (r *Repository) ListExpired(ctx context.Context) ([]*PhoneNumber, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, telnyx_number FROM phone_numbers
		 WHERE expires_at <= NOW() AND released = FALSE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*PhoneNumber
	for rows.Next() {
		n := &PhoneNumber{}
		if err := rows.Scan(&n.ID, &n.UserID, &n.TelnyxNumber); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}
