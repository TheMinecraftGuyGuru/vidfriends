package videos

import (
	"context"
	"testing"
	"time"
)

type stubProvider struct {
	metadata Metadata
	err      error
	calls    int
}

func (s *stubProvider) Lookup(context.Context, string) (Metadata, error) {
	s.calls++
	if s.err != nil {
		return Metadata{}, s.err
	}
	return s.metadata, nil
}

func TestCachingProviderLookup(t *testing.T) {
	base := &stubProvider{metadata: Metadata{Title: "Test"}}
	cache := NewCachingProvider(base, time.Minute)

	ctx := context.Background()

	meta, err := cache.Lookup(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if meta.Title != "Test" {
		t.Fatalf("unexpected metadata: %+v", meta)
	}
	if base.calls != 1 {
		t.Fatalf("expected base called once got %d", base.calls)
	}

	meta, err = cache.Lookup(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if base.calls != 1 {
		t.Fatalf("expected cached result got %d calls", base.calls)
	}
}

func TestCachingProviderLookupErrors(t *testing.T) {
	cache := NewCachingProvider(nil, time.Minute)
	if _, err := cache.Lookup(context.Background(), "https://example.com"); err != ErrProviderUnavailable {
		t.Fatalf("expected provider unavailable got %v", err)
	}

	base := &stubProvider{err: ErrProviderUnavailable}
	cache = NewCachingProvider(base, time.Minute)
	if _, err := cache.Lookup(context.Background(), "https://example.com"); err != ErrProviderUnavailable {
		t.Fatalf("expected provider unavailable got %v", err)
	}
}

func TestCachingProviderExpiry(t *testing.T) {
	base := &stubProvider{metadata: Metadata{Title: "Test"}}
	cache := NewCachingProvider(base, time.Millisecond)

	if _, err := cache.Lookup(context.Background(), "https://example.com"); err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if base.calls != 1 {
		t.Fatalf("expected 1 call got %d", base.calls)
	}

	time.Sleep(2 * time.Millisecond)

	if _, err := cache.Lookup(context.Background(), "https://example.com"); err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if base.calls != 2 {
		t.Fatalf("expected cache miss after expiry got %d calls", base.calls)
	}
}

func TestCachingProviderDefaultTTL(t *testing.T) {
	base := &stubProvider{metadata: Metadata{Title: "Test"}}
	cache := NewCachingProvider(base, 0)

	if cache.ttl <= 0 {
		t.Fatalf("expected ttl to default positive got %v", cache.ttl)
	}
}
