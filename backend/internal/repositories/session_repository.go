package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/vidfriends/backend/internal/auth"
	"github.com/vidfriends/backend/internal/db"
)

// PostgresSessionStore persists refresh tokens to PostgreSQL.
type PostgresSessionStore struct {
	pool db.Pool
}

// NewPostgresSessionStore constructs a session store backed by PostgreSQL.
func NewPostgresSessionStore(pool db.Pool) *PostgresSessionStore {
	return &PostgresSessionStore{pool: pool}
}

// Save stores or updates a session record.
func (s *PostgresSessionStore) Save(ctx context.Context, session auth.Session) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, `
        INSERT INTO sessions (refresh_token, user_id, expires_at)
        VALUES ($1, $2, $3)
        ON CONFLICT (refresh_token)
        DO UPDATE SET user_id = EXCLUDED.user_id, expires_at = EXCLUDED.expires_at
    `, session.RefreshToken, session.UserID, session.ExpiresAt.UTC())
	if err != nil {
		return fmt.Errorf("upsert session: %w", err)
	}

	return nil
}

// Find loads a session by its refresh token.
func (s *PostgresSessionStore) Find(ctx context.Context, refreshToken string) (auth.Session, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return auth.Session{}, fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	row := conn.QueryRow(ctx, `
        SELECT refresh_token, user_id, expires_at
        FROM sessions
        WHERE refresh_token = $1
    `, refreshToken)

	var session auth.Session
	var expiresAt time.Time
	if err := row.Scan(&session.RefreshToken, &session.UserID, &expiresAt); err != nil {
		if err == pgx.ErrNoRows {
			return auth.Session{}, auth.ErrSessionNotFound
		}
		return auth.Session{}, fmt.Errorf("select session: %w", err)
	}

	session.ExpiresAt = expiresAt.UTC()
	return session, nil
}

// Delete removes a session by its refresh token.
func (s *PostgresSessionStore) Delete(ctx context.Context, refreshToken string) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	tag, err := conn.Exec(ctx, `
        DELETE FROM sessions
        WHERE refresh_token = $1
    `, refreshToken)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return auth.ErrSessionNotFound
	}

	return nil
}
