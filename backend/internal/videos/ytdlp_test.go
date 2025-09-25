package videos

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestYTDLPProviderLookup(t *testing.T) {
	provider := NewYTDLPProvider("yt-dlp", time.Second)
	provider.Run = func(ctx context.Context, binary string, args ...string) ([]byte, error) {
		wantArgs := []string{"--dump-single-json", "--no-warnings", "--no-playlist", "--skip-download", "https://example.com"}
		if len(args) != len(wantArgs) {
			t.Fatalf("unexpected args length: got %d want %d", len(args), len(wantArgs))
		}
		for i, arg := range wantArgs {
			if args[i] != arg {
				t.Fatalf("unexpected arg at %d: got %q want %q", i, args[i], arg)
			}
		}
		return []byte(`{"title":"Example","description":"Desc","thumbnail":"thumb.jpg"}`), nil
	}

	meta, err := provider.Lookup(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if meta.Title != "Example" || meta.Description != "Desc" || meta.Thumbnail != "thumb.jpg" {
		t.Fatalf("unexpected metadata: %+v", meta)
	}
}

func TestYTDLPProviderLookupEmptyPayload(t *testing.T) {
	provider := NewYTDLPProvider("yt-dlp", time.Second)
	provider.Run = func(ctx context.Context, binary string, args ...string) ([]byte, error) {
		return []byte(`{"title":"","description":"","thumbnail":""}`), nil
	}

	if _, err := provider.Lookup(context.Background(), "https://example.com"); err == nil {
		t.Fatal("expected error for empty metadata")
	}
}

func TestCachingProvider(t *testing.T) {
	calls := 0
	base := ProviderFunc(func(ctx context.Context, url string) (Metadata, error) {
		calls++
		return Metadata{Title: "Test"}, nil
	})

	cache := NewCachingProvider(base, time.Hour)

	if _, err := cache.Lookup(context.Background(), "https://example.com"); err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if _, err := cache.Lookup(context.Background(), "https://example.com"); err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}

	if calls != 1 {
		t.Fatalf("expected base provider called once, got %d", calls)
	}
}

func TestCachingProviderNilBase(t *testing.T) {
	var cache *CachingProvider
	if _, err := cache.Lookup(context.Background(), "https://example.com"); !errors.Is(err, ErrProviderUnavailable) {
		t.Fatalf("expected ErrProviderUnavailable, got %v", err)
	}
}

func TestYTDLPProviderFetchDownloadsVideo(t *testing.T) {
	provider := NewYTDLPProvider("yt-dlp", time.Second)

	tmpDir := t.TempDir()
	videoPath := filepath.Join(tmpDir, "video.mp4")
	if err := os.WriteFile(videoPath, []byte("content"), 0o600); err != nil {
		t.Fatalf("failed to prepare video file: %v", err)
	}

	provider.Run = func(ctx context.Context, binary string, args ...string) ([]byte, error) {
		wantArgs := []string{"--dump-single-json", "--no-warnings", "--no-playlist", "https://example.com"}
		if len(args) != len(wantArgs) {
			return nil, fmt.Errorf("unexpected args length: got %d want %d", len(args), len(wantArgs))
		}
		for i, arg := range wantArgs {
			if args[i] != arg {
				return nil, fmt.Errorf("unexpected arg at %d: got %q want %q", i, args[i], arg)
			}
		}
		payload := fmt.Sprintf(`{"title":"Example","description":"Desc","thumbnail":"thumb.jpg","requested_downloads":[{"filepath":%q,"filesize":1234}]}`, videoPath)
		return []byte(payload), nil
	}

	storage := &stubStorage{saved: make(map[string][]byte)}

	meta, assets, err := provider.Fetch(context.Background(), "https://example.com", FetchOptions{DownloadVideo: true, Storage: storage})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if meta.Title != "Example" || meta.Description != "Desc" || meta.Thumbnail != "thumb.jpg" {
		t.Fatalf("unexpected metadata: %+v", meta)
	}
	if len(assets) != 1 {
		t.Fatalf("expected 1 asset, got %d", len(assets))
	}
	if assets[0].Type != AssetTypeVideo {
		t.Fatalf("unexpected asset type: %v", assets[0].Type)
	}
	if assets[0].Location != "stored://video.mp4" {
		t.Fatalf("unexpected asset location: %q", assets[0].Location)
	}
	if _, ok := storage.saved["video.mp4"]; !ok {
		t.Fatalf("expected storage to contain persisted asset")
	}
	if _, err := os.Stat(videoPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected temporary video to be removed, stat err = %v", err)
	}
}

func TestYTDLPProviderFetchRequiresStorage(t *testing.T) {
	provider := NewYTDLPProvider("yt-dlp", time.Second)
	provider.Run = func(ctx context.Context, binary string, args ...string) ([]byte, error) {
		return []byte(`{"title":"Example","description":"Desc","thumbnail":"thumb.jpg","requested_downloads":[]}`), nil
	}

	if _, _, err := provider.Fetch(context.Background(), "https://example.com", FetchOptions{DownloadVideo: true}); !errors.Is(err, ErrAssetStorageUnavailable) {
		t.Fatalf("expected ErrAssetStorageUnavailable, got %v", err)
	}
}

type stubStorage struct {
	saved map[string][]byte
}

func (s *stubStorage) Save(ctx context.Context, name string, r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	if s.saved == nil {
		s.saved = make(map[string][]byte)
	}
	s.saved[name] = data
	return "stored://" + name, nil
}

// ProviderFunc adapts a function to the Provider interface.
type ProviderFunc func(ctx context.Context, url string) (Metadata, error)

// Lookup implements Provider.
func (f ProviderFunc) Lookup(ctx context.Context, url string) (Metadata, error) {
	return f(ctx, url)
}
