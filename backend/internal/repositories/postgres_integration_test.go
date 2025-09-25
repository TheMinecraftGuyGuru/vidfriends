package repositories

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach-go/v2/testserver"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vidfriends/backend/internal/auth"
	"github.com/vidfriends/backend/internal/models"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	server, err := testserver.NewTestServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "start cockroach test server: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, server.PGURL().String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect to cockroach test server: %v\n", err)
		server.Stop()
		os.Exit(1)
	}

	if err := applyMigrations(ctx, pool); err != nil {
		fmt.Fprintf(os.Stderr, "apply migrations: %v\n", err)
		pool.Close()
		server.Stop()
		os.Exit(1)
	}

	testPool = pool

	code := m.Run()

	pool.Close()
	server.Stop()

	os.Exit(code)
}

func TestPostgresUserRepository_CreateFindAndUpdate(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)

	repo := NewPostgresUserRepository(testPool)

	user := models.User{
		ID:        uuid.NewString(),
		Email:     "alice@example.com",
		Password:  "secret-hash",
		CreatedAt: time.Now().UTC().Truncate(time.Millisecond),
		UpdatedAt: time.Now().UTC().Truncate(time.Millisecond),
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	dup := models.User{
		ID:        uuid.NewString(),
		Email:     user.Email,
		Password:  "another-hash",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := repo.Create(ctx, dup); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict when creating duplicate email, got %v", err)
	}

	fetched, err := repo.FindByEmail(ctx, user.Email)
	if err != nil {
		t.Fatalf("find by email: %v", err)
	}

	if fetched.ID != user.ID || fetched.Email != user.Email || fetched.Password != user.Password {
		t.Fatalf("unexpected user fetched: %+v", fetched)
	}

	updated := user
	updated.Email = "updated@example.com"
	updated.Password = "rotated-hash"
	updated.UpdatedAt = time.Now().UTC().Add(time.Minute)

	if err := repo.Update(ctx, updated); err != nil {
		t.Fatalf("update user: %v", err)
	}

	fetched, err = repo.FindByEmail(ctx, updated.Email)
	if err != nil {
		t.Fatalf("find by updated email: %v", err)
	}

	if fetched.Email != updated.Email || fetched.Password != updated.Password {
		t.Fatalf("expected updated fields to persist, got %+v", fetched)
	}

	missing := models.User{
		ID:        uuid.NewString(),
		Email:     "missing@example.com",
		Password:  "hash",
		UpdatedAt: time.Now().UTC(),
	}

	if err := repo.Update(ctx, missing); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound updating missing user, got %v", err)
	}
}

func TestPostgresFriendRepository_CreateListAndUpdate(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)

	userRepo := NewPostgresUserRepository(testPool)
	viewer := createTestUser(t, userRepo, "viewer@example.com")
	friend := createTestUser(t, userRepo, "friend@example.com")
	stranger := createTestUser(t, userRepo, "stranger@example.com")

	repo := NewPostgresFriendRepository(testPool)

	request := models.FriendRequest{
		ID:        uuid.NewString(),
		Requester: viewer.ID,
		Receiver:  friend.ID,
		Status:    "pending",
		CreatedAt: time.Now().UTC().Add(-time.Hour),
	}

	if err := repo.CreateRequest(ctx, request); err != nil {
		t.Fatalf("create friend request: %v", err)
	}

	duplicate := request
	duplicate.ID = uuid.NewString()
	if err := repo.CreateRequest(ctx, duplicate); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict on duplicate friend request, got %v", err)
	}

	other := models.FriendRequest{
		ID:        uuid.NewString(),
		Requester: viewer.ID,
		Receiver:  stranger.ID,
		Status:    "pending",
		CreatedAt: time.Now().UTC(),
	}

	if err := repo.CreateRequest(ctx, other); err != nil {
		t.Fatalf("create second friend request: %v", err)
	}

	requests, err := repo.ListForUser(ctx, viewer.ID)
	if err != nil {
		t.Fatalf("list friend requests: %v", err)
	}

	if len(requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(requests))
	}

	if err := repo.UpdateStatus(ctx, request.ID, "accepted"); err != nil {
		t.Fatalf("update friend request status: %v", err)
	}

	requests, err = repo.ListForUser(ctx, friend.ID)
	if err != nil {
		t.Fatalf("list friend requests after update: %v", err)
	}

	if len(requests) != 1 {
		t.Fatalf("expected 1 request for friend, got %d", len(requests))
	}

	if requests[0].Status != "accepted" {
		t.Fatalf("expected status to be accepted, got %s", requests[0].Status)
	}

	if requests[0].RespondedAt == nil {
		t.Fatalf("expected responded_at to be set after acceptance")
	}

	if err := repo.UpdateStatus(ctx, uuid.NewString(), "accepted"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for unknown request, got %v", err)
	}
}

func TestPostgresSessionStore_SaveFindAndDelete(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)

	userRepo := NewPostgresUserRepository(testPool)
	user := createTestUser(t, userRepo, "owner@example.com")

	store := NewPostgresSessionStore(testPool)
	expires := time.Now().UTC().Add(24 * time.Hour)
	session := auth.Session{
		RefreshToken: uuid.NewString(),
		UserID:       user.ID,
		ExpiresAt:    expires,
	}

	if err := store.Save(ctx, session); err != nil {
		t.Fatalf("save session: %v", err)
	}

	loaded, err := store.Find(ctx, session.RefreshToken)
	if err != nil {
		t.Fatalf("find session: %v", err)
	}

	if loaded.UserID != session.UserID || !timesClose(loaded.ExpiresAt, expires.UTC(), time.Millisecond) {
		t.Fatalf("unexpected session loaded: %+v", loaded)
	}

	updated := session
	updated.ExpiresAt = expires.Add(48 * time.Hour)
	if err := store.Save(ctx, updated); err != nil {
		t.Fatalf("update session: %v", err)
	}

	loaded, err = store.Find(ctx, session.RefreshToken)
	if err != nil {
		t.Fatalf("find session after update: %v", err)
	}

	if !timesClose(loaded.ExpiresAt, updated.ExpiresAt.UTC(), time.Millisecond) {
		t.Fatalf("expected updated expiry, got %v", loaded.ExpiresAt)
	}

	if err := store.Delete(ctx, session.RefreshToken); err != nil {
		t.Fatalf("delete session: %v", err)
	}

	if _, err := store.Find(ctx, session.RefreshToken); !errors.Is(err, auth.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound after delete, got %v", err)
	}

	if err := store.Delete(ctx, session.RefreshToken); !errors.Is(err, auth.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound deleting twice, got %v", err)
	}
}

func TestPostgresVideoRepository_ListFeed(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)

	userRepo := NewPostgresUserRepository(testPool)
	friendRepo := NewPostgresFriendRepository(testPool)
	videoRepo := NewPostgresVideoRepository(testPool)

	viewer := createTestUser(t, userRepo, "viewer@example.com")
	acceptedFriend := createTestUser(t, userRepo, "accepted@example.com")
	pendingFriend := createTestUser(t, userRepo, "pending@example.com")
	stranger := createTestUser(t, userRepo, "stranger@example.com")

	acceptedReq := models.FriendRequest{
		ID:        uuid.NewString(),
		Requester: viewer.ID,
		Receiver:  acceptedFriend.ID,
		Status:    "accepted",
		CreatedAt: time.Now().UTC().Add(-2 * time.Hour),
	}
	if err := friendRepo.CreateRequest(ctx, acceptedReq); err != nil {
		t.Fatalf("create accepted request: %v", err)
	}
	if err := friendRepo.UpdateStatus(ctx, acceptedReq.ID, "accepted"); err != nil {
		t.Fatalf("confirm accepted request status: %v", err)
	}

	pendingReq := models.FriendRequest{
		ID:        uuid.NewString(),
		Requester: viewer.ID,
		Receiver:  pendingFriend.ID,
		Status:    "pending",
		CreatedAt: time.Now().UTC().Add(-time.Hour),
	}
	if err := friendRepo.CreateRequest(ctx, pendingReq); err != nil {
		t.Fatalf("create pending request: %v", err)
	}

	baseTime := time.Now().UTC().Add(-30 * time.Minute)
	ownShare := models.VideoShare{
		ID:        uuid.NewString(),
		OwnerID:   viewer.ID,
		URL:       "https://example.com/own",
		Title:     "Viewer Share",
		CreatedAt: baseTime.Add(2 * time.Minute),
	}
	acceptedShare := models.VideoShare{
		ID:        uuid.NewString(),
		OwnerID:   acceptedFriend.ID,
		URL:       "https://example.com/accepted",
		Title:     "Accepted Share",
		CreatedAt: baseTime.Add(5 * time.Minute),
	}
	pendingShare := models.VideoShare{
		ID:        uuid.NewString(),
		OwnerID:   pendingFriend.ID,
		URL:       "https://example.com/pending",
		Title:     "Pending Share",
		CreatedAt: baseTime.Add(10 * time.Minute),
	}
	strangerShare := models.VideoShare{
		ID:        uuid.NewString(),
		OwnerID:   stranger.ID,
		URL:       "https://example.com/stranger",
		Title:     "Stranger Share",
		CreatedAt: baseTime.Add(15 * time.Minute),
	}

	for _, share := range []models.VideoShare{ownShare, acceptedShare, pendingShare, strangerShare} {
		if err := videoRepo.Create(ctx, share); err != nil {
			t.Fatalf("create share %s: %v", share.ID, err)
		}
	}

	feed, err := videoRepo.ListFeed(ctx, viewer.ID)
	if err != nil {
		t.Fatalf("list feed: %v", err)
	}

	if len(feed) != 2 {
		t.Fatalf("expected 2 feed entries (viewer + accepted friend), got %d", len(feed))
	}

	if feed[0].ID != acceptedShare.ID || feed[1].ID != ownShare.ID {
		t.Fatalf("unexpected feed order: %+v", feed)
	}

	for _, share := range feed {
		if share.OwnerID == pendingFriend.ID || share.OwnerID == stranger.ID {
			t.Fatalf("unexpected share from owner %s in feed", share.OwnerID)
		}
	}
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrationsDir := filepath.Join("..", "..", "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		contents, err := os.ReadFile(filepath.Join(migrationsDir, entry.Name()))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		if _, err := pool.Exec(ctx, string(contents)); err != nil {
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}
	}

	return nil
}

func resetDatabase(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	conn, err := testPool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire connection: %v", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "TRUNCATE TABLE friend_requests, video_shares, sessions, users CASCADE"); err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}

func createTestUser(t *testing.T, repo *PostgresUserRepository, email string) models.User {
	t.Helper()
	user := models.User{
		ID:        uuid.NewString(),
		Email:     email,
		Password:  "password-hash",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := repo.Create(context.Background(), user); err != nil {
		t.Fatalf("create test user: %v", err)
	}
	return user
}

func timesClose(a, b time.Time, delta time.Duration) bool {
	diff := a.Sub(b)
	if diff < 0 {
		diff = -diff
	}
	return diff <= delta
}
