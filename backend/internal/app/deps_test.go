package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vidfriends/backend/internal/config"
)

type fakePool struct{}

func (fakePool) Acquire(context.Context) (*pgxpool.Conn, error) {
	return nil, errors.New("not implemented")
}

func (fakePool) Close() {}

func TestBuildDependencies(t *testing.T) {
	cfg := config.Config{
		YTDLPPath:        "yt-dlp",
		YTDLPTimeout:     time.Second,
		MetadataCacheTTL: time.Minute,
	}

	deps := buildDependencies(fakePool{}, cfg)

	if deps.Users == nil {
		t.Fatal("expected user repository to be configured")
	}
	if deps.Sessions == nil {
		t.Fatal("expected session manager to be configured")
	}
	if deps.Friends == nil {
		t.Fatal("expected friend repository to be configured")
	}
	if deps.Videos == nil {
		t.Fatal("expected video repository to be configured")
	}
	if deps.VideoMetadata == nil {
		t.Fatal("expected video metadata provider to be configured")
	}
}
