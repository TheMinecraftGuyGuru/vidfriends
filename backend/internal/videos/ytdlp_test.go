package videos

import (
	"context"
	"errors"
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

// ProviderFunc adapts a function to the Provider interface.
type ProviderFunc func(ctx context.Context, url string) (Metadata, error)

// Lookup implements Provider.
func (f ProviderFunc) Lookup(ctx context.Context, url string) (Metadata, error) {
	return f(ctx, url)
}
