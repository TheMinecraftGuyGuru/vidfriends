package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/vidfriends/backend/internal/models"
)

var (
	// ErrSessionNotFound indicates the provided refresh token does not map to an active session.
	ErrSessionNotFound = errors.New("session not found")
	// ErrRefreshTokenExpired indicates the refresh token has expired and cannot be used.
	ErrRefreshTokenExpired = errors.New("refresh token expired")
)

// SessionStore persists issued refresh tokens so they can survive process restarts.
type SessionStore interface {
	Save(ctx context.Context, session Session) error
	Find(ctx context.Context, refreshToken string) (Session, error)
	Delete(ctx context.Context, refreshToken string) error
}

// Session represents a refresh token issued to a user.
type Session struct {
	RefreshToken string
	UserID       string
	ExpiresAt    time.Time
}

// Manager manages the lifecycle of issued session tokens backed by a persistent store.
type Manager struct {
	accessTTL  time.Duration
	refreshTTL time.Duration

	store SessionStore
}

// NewManager constructs a Manager that issues access and refresh tokens with the provided TTLs.
func NewManager(accessTTL, refreshTTL time.Duration, store SessionStore) *Manager {
	if store == nil {
		panic("auth: session store must not be nil")
	}
	return &Manager{
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		store:      store,
	}
}

// Issue creates a new pair of access and refresh tokens for the provided user identifier.
func (m *Manager) Issue(ctx context.Context, userID string) (models.SessionTokens, error) {
	if userID == "" {
		return models.SessionTokens{}, errors.New("user id must be provided")
	}

	now := time.Now().UTC()
	accessToken, err := randomToken()
	if err != nil {
		return models.SessionTokens{}, err
	}

	refreshToken, err := randomToken()
	if err != nil {
		return models.SessionTokens{}, err
	}

	tokens := models.SessionTokens{
		AccessToken:      accessToken,
		AccessExpiresAt:  now.Add(m.accessTTL),
		RefreshToken:     refreshToken,
		RefreshExpiresAt: now.Add(m.refreshTTL),
	}

	if err := m.store.Save(ctx, Session{
		RefreshToken: refreshToken,
		UserID:       userID,
		ExpiresAt:    tokens.RefreshExpiresAt,
	}); err != nil {
		return models.SessionTokens{}, err
	}

	return tokens, nil
}

// Refresh exchanges a refresh token for a new session token pair.
func (m *Manager) Refresh(ctx context.Context, refreshToken string) (models.SessionTokens, error) {
	if refreshToken == "" {
		return models.SessionTokens{}, ErrSessionNotFound
	}

	session, err := m.store.Find(ctx, refreshToken)
	if err != nil {
		return models.SessionTokens{}, err
	}

	if time.Now().UTC().After(session.ExpiresAt) {
		_ = m.store.Delete(ctx, refreshToken)
		return models.SessionTokens{}, ErrRefreshTokenExpired
	}

	if err := m.store.Delete(ctx, refreshToken); err != nil {
		return models.SessionTokens{}, err
	}

	return m.Issue(ctx, session.UserID)
}

// Revoke removes the provided refresh token from the active session store.
func (m *Manager) Revoke(ctx context.Context, refreshToken string) {
	if refreshToken == "" {
		return
	}
	_ = m.store.Delete(ctx, refreshToken)
}

func randomToken() (string, error) {
	const size = 32
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
