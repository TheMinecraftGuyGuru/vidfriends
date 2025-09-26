package handlers

import (
	"net/http"
	"time"

	"github.com/vidfriends/backend/internal/middleware"
)

// RegisterRoutes wires HTTP handlers into the provided ServeMux.
func RegisterRoutes(mux *http.ServeMux, deps Dependencies) {
	health := HealthHandler{}

	authLimiter := middleware.NewIPRateLimiter(10, time.Minute, 5, 15*time.Minute)
	inviteLimiter := middleware.NewIPRateLimiter(5, time.Minute, 3, 15*time.Minute)

	auth := AuthHandler{Users: deps.Users, Sessions: deps.Sessions, RateLimiter: authLimiter}
	friends := FriendHandler{Friends: deps.Friends, RateLimiter: inviteLimiter}
	videos := VideoHandler{Videos: deps.Videos, Metadata: deps.VideoMetadata, Assets: deps.VideoAssets}

	mux.HandleFunc("/healthz", health.Handle)
	mux.HandleFunc("/api/v1/auth/login", auth.Login)
	mux.HandleFunc("/api/v1/auth/signup", auth.SignUp)
	mux.HandleFunc("/api/v1/auth/refresh", auth.Refresh)
	mux.HandleFunc("/api/v1/auth/password-reset", auth.RequestPasswordReset)
	mux.HandleFunc("/api/v1/friends", friends.List)
	mux.HandleFunc("/api/v1/friends/invite", friends.Invite)
	mux.HandleFunc("/api/v1/friends/respond", friends.Respond)
	mux.HandleFunc("/api/v1/videos", videos.Create)
	mux.HandleFunc("/api/v1/videos/feed", videos.Feed)
}

// Dependencies aggregates collaborators required by HTTP handlers.
type Dependencies struct {
	Users         UserStore
	Sessions      SessionManager
	Friends       FriendStore
	Videos        VideoStore
	VideoMetadata VideoMetadataProvider
	VideoAssets   VideoAssetIngestor
}
