package middleware

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter controls how frequently a caller may perform an action.
type RateLimiter interface {
	Allow(key string) bool
}

// ipRateLimiter tracks request rates per key (typically an IP address) with expiration.
type ipRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    rate.Limit
	burst    int
	ttl      time.Duration
	now      func() time.Time
}

// NewIPRateLimiter constructs a per-key rate limiter that allows up to `requests` events per `window`
// with an additional burst capacity. Entries expire after the provided ttl when no longer used.
func NewIPRateLimiter(requests int, window time.Duration, burst int, ttl time.Duration) RateLimiter {
	if requests <= 0 {
		requests = 1
	}
	if window <= 0 {
		window = time.Second
	}
	if burst <= 0 {
		burst = 1
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}

	limit := rate.Every(window / time.Duration(requests))
	return &ipRateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		burst:    burst,
		ttl:      ttl,
		now:      time.Now,
	}
}

func (l *ipRateLimiter) Allow(key string) bool {
	if key == "" {
		key = "unknown"
	}

	now := l.now()

	l.mu.Lock()
	v := l.getVisitorLocked(key, now)
	l.gcLocked(now)
	l.mu.Unlock()

	return v.limiter.Allow()
}

func (l *ipRateLimiter) getVisitorLocked(key string, now time.Time) *visitor {
	if v, ok := l.visitors[key]; ok {
		v.lastSeen = now
		return v
	}

	limiter := rate.NewLimiter(l.limit, l.burst)
	v := &visitor{limiter: limiter, lastSeen: now}
	l.visitors[key] = v
	return v
}

func (l *ipRateLimiter) gcLocked(now time.Time) {
	for key, v := range l.visitors {
		if now.Sub(v.lastSeen) > l.ttl {
			delete(l.visitors, key)
		}
	}
}

// WithNowFunc allows tests to override the time source.
func (l *ipRateLimiter) WithNowFunc(now func() time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.now = now
}
