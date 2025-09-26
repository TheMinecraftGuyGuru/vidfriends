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
		ObjectStore:      config.ObjectStoreConfig{Bucket: "test-bucket", Endpoint: "http://localhost:9000", Region: "us-east-1"},
	}

	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	deps, cleanup, err := buildDependencies(context.Background(), fakePool{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cleanup == nil {
		t.Fatal("expected cleanup function")
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = cleanup(ctx)
	}()

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
	if deps.VideoAssets == nil {
		t.Fatal("expected video asset ingestor to be configured")
	}
}
