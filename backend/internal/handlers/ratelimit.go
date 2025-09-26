package handlers

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

// RateLimiter is the minimal interface required to guard sensitive endpoints.
type RateLimiter interface {
	Allow(key string) bool
}

func allowRequest(limiter RateLimiter, r *http.Request, scope string) bool {
	if limiter == nil {
		return true
	}
	key := rateLimitKey(r, scope)
	return limiter.Allow(key)
}

func rateLimitKey(r *http.Request, scope string) string {
	ip := clientIP(r)
	if scope == "" {
		return ip
	}
	return fmt.Sprintf("%s:%s", scope, ip)
}

func clientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}
