package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"github.com/vidfriends/backend/internal/models"
)

var (
	// ErrSessionNotFound indicates the provided refresh token does not map to an active session.
	ErrSessionNotFound = errors.New("session not found")
	// ErrRefreshTokenExpired indicates the refresh token has expired and cannot be used.
	ErrRefreshTokenExpired = errors.New("refresh token expired")
)

// Manager manages the lifecycle of issued session tokens in-memory.
//
// It is intended for development environments where a shared cache such as Redis is not yet available.
// Callers should wrap it with appropriate persistence for production deployments.
type Manager struct {
	accessTTL  time.Duration
	refreshTTL time.Duration

	mu       sync.Mutex
	sessions map[string]sessionRecord
}

type sessionRecord struct {
	userID    string
	expiresAt time.Time
}

// NewManager constructs a Manager that issues access and refresh tokens with the provided TTLs.
func NewManager(accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		sessions:   make(map[string]sessionRecord),
	}
}

// Issue creates a new pair of access and refresh tokens for the provided user identifier.
func (m *Manager) Issue(_ context.Context, userID string) (models.SessionTokens, error) {
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

	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[refreshToken] = sessionRecord{userID: userID, expiresAt: tokens.RefreshExpiresAt}

	return tokens, nil
}

// Refresh exchanges a refresh token for a new session token pair.
func (m *Manager) Refresh(_ context.Context, refreshToken string) (models.SessionTokens, error) {
	if refreshToken == "" {
		return models.SessionTokens{}, ErrSessionNotFound
	}

	m.mu.Lock()
	record, ok := m.sessions[refreshToken]
	if !ok {
		m.mu.Unlock()
		return models.SessionTokens{}, ErrSessionNotFound
	}
	if time.Now().UTC().After(record.expiresAt) {
		delete(m.sessions, refreshToken)
		m.mu.Unlock()
		return models.SessionTokens{}, ErrRefreshTokenExpired
	}
	delete(m.sessions, refreshToken)
	m.mu.Unlock()

	return m.Issue(context.Background(), record.userID)
}

// Revoke removes the provided refresh token from the active session store.
func (m *Manager) Revoke(_ context.Context, refreshToken string) {
	if refreshToken == "" {
		return
	}
	m.mu.Lock()
	delete(m.sessions, refreshToken)
	m.mu.Unlock()
}

func randomToken() (string, error) {
	const size = 32
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
