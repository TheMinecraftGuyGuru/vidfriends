package videos

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vidfriends/backend/internal/models"
)

type assetStorageStub struct {
	saved map[string][]byte
	err   error
}

func (s *assetStorageStub) Save(ctx context.Context, name string, r io.Reader) (string, error) {
	_ = ctx
	if s.err != nil {
		return "", s.err
	}
	if s.saved == nil {
		s.saved = make(map[string][]byte)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	s.saved[name] = data
	return fmt.Sprintf("https://cdn.example.com/%s", name), nil
}

type shareUpdaterStub struct {
	readyCalls  []string
	readyLoc    string
	readySize   int64
	failedCalls []string
	readyErr    error
	failedErr   error
}

func (s *shareUpdaterStub) MarkAssetReady(ctx context.Context, shareID, location string, size int64) error {
	_ = ctx
	s.readyCalls = append(s.readyCalls, shareID)
	s.readyLoc = location
	s.readySize = size
	return s.readyErr
}

func (s *shareUpdaterStub) MarkAssetFailed(ctx context.Context, shareID string) error {
	_ = ctx
	s.failedCalls = append(s.failedCalls, shareID)
	return s.failedErr
}

func TestAssetIngestorSuccess(t *testing.T) {
	dir := t.TempDir()
	provider := &YTDLPProvider{Binary: "yt-dlp", Timeout: time.Second}
	provider.Run = func(ctx context.Context, binary string, args ...string) ([]byte, error) {
		file := filepath.Join(dir, "video.mp4")
		if err := os.WriteFile(file, []byte("video-bytes"), 0o644); err != nil {
			return nil, err
		}
		payload := fmt.Sprintf(`{"title":"Test","description":"","thumbnail":"","requested_downloads":[{"filepath":"%s","filename":"video.mp4","filesize":%d}]}`, file, len("video-bytes"))
		return []byte(payload), nil
	}

	storage := &assetStorageStub{}
	updater := &shareUpdaterStub{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ingestor := NewAssetIngestor(provider, storage, updater, AssetIngestorConfig{QueueSize: 1, Workers: 1}, logger)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = ingestor.Shutdown(ctx)
	}()

	share := models.VideoShare{ID: "share-1", URL: "https://example.com/watch?v=123"}
	if err := ingestor.Enqueue(context.Background(), share); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	waitForCondition(t, func() bool { return len(updater.readyCalls) > 0 }, time.Second)

	if _, ok := storage.saved[filepath.Join(share.ID, "video.mp4")]; !ok {
		t.Fatalf("expected asset to be saved with share prefix")
	}
	if updater.readyLoc == "" {
		t.Fatalf("expected ready location to be populated")
	}
	if updater.readySize != int64(len("video-bytes")) {
		t.Fatalf("unexpected asset size: %d", updater.readySize)
	}
}

func TestAssetIngestorFailure(t *testing.T) {
	provider := &YTDLPProvider{Binary: "yt-dlp", Timeout: time.Second}
	provider.Run = func(ctx context.Context, binary string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("yt-dlp error")
	}

	storage := &assetStorageStub{}
	updater := &shareUpdaterStub{}
	ingestor := NewAssetIngestor(provider, storage, updater, AssetIngestorConfig{QueueSize: 1, Workers: 1}, nil)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = ingestor.Shutdown(ctx)
	}()

	share := models.VideoShare{ID: "share-2", URL: "https://example.com/fail"}
	if err := ingestor.Enqueue(context.Background(), share); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	waitForCondition(t, func() bool { return len(updater.failedCalls) > 0 }, time.Second)
	if len(updater.readyCalls) != 0 {
		t.Fatalf("expected no ready calls on failure")
	}
}

func waitForCondition(t *testing.T, predicate func() bool, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if predicate() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("condition not met within %v", timeout)
}
