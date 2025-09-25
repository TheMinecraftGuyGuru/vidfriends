package handlers

import "net/http"

// RegisterRoutes wires HTTP handlers into the provided ServeMux.
func RegisterRoutes(mux *http.ServeMux, deps Dependencies) {
	health := HealthHandler{}
	auth := AuthHandler{Users: deps.Users, Sessions: deps.Sessions}
	friends := FriendHandler{Friends: deps.Friends}
        videos := VideoHandler{Videos: deps.Videos, Metadata: deps.VideoMetadata}

	mux.HandleFunc("/healthz", health.Handle)
	mux.HandleFunc("/api/v1/auth/login", auth.Login)
	mux.HandleFunc("/api/v1/auth/signup", auth.SignUp)
	mux.HandleFunc("/api/v1/auth/refresh", auth.Refresh)
	mux.HandleFunc("/api/v1/friends", friends.List)
	mux.HandleFunc("/api/v1/friends/invite", friends.Invite)
	mux.HandleFunc("/api/v1/friends/respond", friends.Respond)
	mux.HandleFunc("/api/v1/videos", videos.Create)
	mux.HandleFunc("/api/v1/videos/feed", videos.Feed)
}

// Dependencies aggregates collaborators required by HTTP handlers.
type Dependencies struct {
        Users    UserStore
        Sessions SessionManager
        Friends  FriendStore
        Videos   VideoStore
        VideoMetadata VideoMetadataProvider
}
