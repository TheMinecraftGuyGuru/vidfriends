package videos

import (
	"context"
	"sync"
	"time"
)

type cacheEntry struct {
	metadata Metadata
	expires  time.Time
}

// CachingProvider wraps another Provider with a TTL-based in-memory cache.
type CachingProvider struct {
	base Provider
	ttl  time.Duration

	mu    sync.RWMutex
	items map[string]cacheEntry
}

// NewCachingProvider returns a Provider that caches lookups for the provided TTL.
func NewCachingProvider(base Provider, ttl time.Duration) *CachingProvider {
	if ttl <= 0 {
		ttl = time.Minute
	}
	return &CachingProvider{
		base:  base,
		ttl:   ttl,
		items: make(map[string]cacheEntry),
	}
}

// Lookup returns cached metadata when available, otherwise it delegates to the
// underlying provider and stores the result.
func (c *CachingProvider) Lookup(ctx context.Context, url string) (Metadata, error) {
	if c == nil || c.base == nil {
		return Metadata{}, ErrProviderUnavailable
	}

	now := time.Now()

	c.mu.RLock()
	entry, ok := c.items[url]
	c.mu.RUnlock()
	if ok && now.Before(entry.expires) {
		return entry.metadata, nil
	}

	metadata, err := c.base.Lookup(ctx, url)
	if err != nil {
		return Metadata{}, err
	}

	c.mu.Lock()
	c.items[url] = cacheEntry{metadata: metadata, expires: now.Add(c.ttl)}
	c.mu.Unlock()

	return metadata, nil
}
