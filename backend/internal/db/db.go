package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool abstracts the pgx connection pool to make testing easier.
type Pool interface {
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
	Close()
}

// Connect initialises a PostgreSQL connection pool using the provided database URL.
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}
	return pool, nil
}
