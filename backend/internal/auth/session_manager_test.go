package auth

import (
	"context"
	"testing"
	"time"
)

func TestManagerIssueAndRefresh(t *testing.T) {
	manager := NewManager(time.Minute, time.Hour)

	tokens, err := manager.Issue(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("expected non-empty tokens: %+v", tokens)
	}

	refreshed, err := manager.Refresh(context.Background(), tokens.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if refreshed.RefreshToken == tokens.RefreshToken {
		t.Fatal("expected new refresh token")
	}
	manager.mu.Lock()
	_, exists := manager.sessions[tokens.RefreshToken]
	manager.mu.Unlock()
	if exists {
		t.Fatal("old token should have been removed")
	}
}

func TestManagerIssueValidation(t *testing.T) {
	manager := NewManager(time.Minute, time.Hour)
	if _, err := manager.Issue(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty user id")
	}
}

func TestManagerRefreshFailures(t *testing.T) {
	manager := NewManager(time.Minute, time.Millisecond)

	if _, err := manager.Refresh(context.Background(), ""); err != ErrSessionNotFound {
		t.Fatalf("expected session not found got %v", err)
	}

	tokens, err := manager.Issue(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	time.Sleep(2 * time.Millisecond)

	if _, err := manager.Refresh(context.Background(), tokens.RefreshToken); err != ErrRefreshTokenExpired {
		t.Fatalf("expected refresh expired got %v", err)
	}

	tokens, err = manager.Issue(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	manager.Revoke(context.Background(), tokens.RefreshToken)
	if _, err := manager.Refresh(context.Background(), tokens.RefreshToken); err != ErrSessionNotFound {
		t.Fatalf("expected session not found after revoke got %v", err)
	}
}
