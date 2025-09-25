package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Server wraps the http.Server with sensible defaults.
type Server struct {
	inner *http.Server
}

// New constructs a server listening on the provided port.
func New(port int, handler http.Handler) *Server {
	return &Server{
		inner: &http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      10 * time.Second,
		},
	}
}

// Start begins serving HTTP traffic.
func (s *Server) Start() error {
	return s.inner.ListenAndServe()
}

// Shutdown gracefully terminates the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.inner.Shutdown(ctx)
}
