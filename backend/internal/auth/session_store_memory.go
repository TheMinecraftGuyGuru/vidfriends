package auth

import (
	"context"
	"sync"
)

// NewInMemorySessionStore returns a SessionStore backed by an in-memory map.
func NewInMemorySessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{sessions: make(map[string]Session)}
}

// InMemorySessionStore implements SessionStore for tests and local development.
type InMemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]Session
}

// Save persists the provided session record.
func (s *InMemorySessionStore) Save(_ context.Context, session Session) error {
	s.mu.Lock()
	s.sessions[session.RefreshToken] = session
	s.mu.Unlock()
	return nil
}

// Find retrieves a session by refresh token.
func (s *InMemorySessionStore) Find(_ context.Context, refreshToken string) (Session, error) {
	s.mu.RLock()
	session, ok := s.sessions[refreshToken]
	s.mu.RUnlock()
	if !ok {
		return Session{}, ErrSessionNotFound
	}
	return session, nil
}

// Delete removes the session associated with the refresh token.
func (s *InMemorySessionStore) Delete(_ context.Context, refreshToken string) error {
	s.mu.Lock()
	delete(s.sessions, refreshToken)
	s.mu.Unlock()
	return nil
}

// Has reports whether a refresh token exists. Useful for tests.
func (s *InMemorySessionStore) Has(refreshToken string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.sessions[refreshToken]
	return ok
}
