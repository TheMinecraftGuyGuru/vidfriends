package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/vidfriends/backend/internal/db"
	"github.com/vidfriends/backend/internal/models"
)

// PostgresUserRepository provides PostgreSQL-backed persistence for users.
type PostgresUserRepository struct {
	pool db.Pool
}

// NewPostgresUserRepository constructs a user repository backed by PostgreSQL.
func NewPostgresUserRepository(pool db.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

// Create persists a new user record.
func (r *PostgresUserRepository) Create(ctx context.Context, user models.User) error {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, `
        INSERT INTO users (id, email, password_hash, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
    `, user.ID, user.Email, user.Password, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrConflict
		}
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

// FindByEmail fetches a user by their email address.
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (models.User, error) {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return models.User{}, fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	row := conn.QueryRow(ctx, `
        SELECT id, email, password_hash, created_at, updated_at
        FROM users
        WHERE email = $1
    `, email)

	var user models.User
	if err := row.Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrNotFound
		}
		return models.User{}, fmt.Errorf("select user by email: %w", err)
	}

	return user, nil
}

// Update modifies an existing user record.
func (r *PostgresUserRepository) Update(ctx context.Context, user models.User) error {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	tag, err := conn.Exec(ctx, `
        UPDATE users
        SET email = $2, password_hash = $3, updated_at = $4
        WHERE id = $1
    `, user.ID, user.Email, user.Password, user.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrConflict
		}
		return fmt.Errorf("update user: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// PostgresFriendRepository provides PostgreSQL-backed persistence for friend requests.
type PostgresFriendRepository struct {
	pool db.Pool
}

// NewPostgresFriendRepository constructs a friend repository backed by PostgreSQL.
func NewPostgresFriendRepository(pool db.Pool) *PostgresFriendRepository {
	return &PostgresFriendRepository{pool: pool}
}

// CreateRequest persists a new friend request.
func (r *PostgresFriendRepository) CreateRequest(ctx context.Context, request models.FriendRequest) error {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, `
        INSERT INTO friend_requests (id, requester_id, receiver_id, status, created_at, responded_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, request.ID, request.Requester, request.Receiver, request.Status, request.CreatedAt, request.RespondedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return ErrConflict
			case "23503":
				return ErrNotFound
			}
		}
		return fmt.Errorf("insert friend request: %w", err)
	}

	return nil
}

// ListForUser returns friend requests where the user is the requester or receiver.
func (r *PostgresFriendRepository) ListForUser(ctx context.Context, userID string) ([]models.FriendRequest, error) {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, `
        SELECT id, requester_id, receiver_id, status, created_at, responded_at
        FROM friend_requests
        WHERE requester_id = $1 OR receiver_id = $1
        ORDER BY created_at DESC
    `, userID)
	if err != nil {
		return nil, fmt.Errorf("query friend requests: %w", err)
	}
	defer rows.Close()

	var requests []models.FriendRequest
	for rows.Next() {
		var (
			req         models.FriendRequest
			respondedAt sql.NullTime
		)

		if err := rows.Scan(&req.ID, &req.Requester, &req.Receiver, &req.Status, &req.CreatedAt, &respondedAt); err != nil {
			return nil, fmt.Errorf("scan friend request: %w", err)
		}

		if respondedAt.Valid {
			t := respondedAt.Time.UTC()
			req.RespondedAt = &t
		}

		requests = append(requests, req)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate friend requests: %w", err)
	}

	return requests, nil
}

// UpdateStatus updates the status (and responded_at) for a friend request.
func (r *PostgresFriendRepository) UpdateStatus(ctx context.Context, requestID, status string) error {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	respondedAt := sql.NullTime{}
	if status != "pending" {
		respondedAt = sql.NullTime{Valid: true, Time: time.Now().UTC()}
	}

	tag, err := conn.Exec(ctx, `
        UPDATE friend_requests
        SET status = $2, responded_at = $3
        WHERE id = $1
    `, requestID, status, respondedAt)
	if err != nil {
		return fmt.Errorf("update friend request: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// PostgresVideoRepository provides PostgreSQL-backed persistence for shared videos.
type PostgresVideoRepository struct {
	pool db.Pool
}

// NewPostgresVideoRepository constructs a video repository backed by PostgreSQL.
func NewPostgresVideoRepository(pool db.Pool) *PostgresVideoRepository {
	return &PostgresVideoRepository{pool: pool}
}

// Create stores a new shared video record.
func (r *PostgresVideoRepository) Create(ctx context.Context, share models.VideoShare) error {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, `
        INSERT INTO video_shares (id, owner_id, url, title, description, thumbnail, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `, share.ID, share.OwnerID, share.URL, share.Title, share.Description, share.Thumbnail, share.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrConflict
		}
		return fmt.Errorf("insert video share: %w", err)
	}

	return nil
}

// ListFeed returns a simple reverse chronological feed of shared videos.
func (r *PostgresVideoRepository) ListFeed(ctx context.Context, userID string) ([]models.VideoShare, error) {
	_ = userID

	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, `
        SELECT id, owner_id, url, title, description, thumbnail, created_at
        FROM video_shares
        ORDER BY created_at DESC
        LIMIT 100
    `)
	if err != nil {
		return nil, fmt.Errorf("query video feed: %w", err)
	}
	defer rows.Close()

	var shares []models.VideoShare
	for rows.Next() {
		var share models.VideoShare
		if err := rows.Scan(&share.ID, &share.OwnerID, &share.URL, &share.Title, &share.Description, &share.Thumbnail, &share.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan video share: %w", err)
		}
		shares = append(shares, share)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate video feed: %w", err)
	}

	return shares, nil
}

var _ UserRepository = (*PostgresUserRepository)(nil)
var _ FriendRepository = (*PostgresFriendRepository)(nil)
var _ VideoRepository = (*PostgresVideoRepository)(nil)
