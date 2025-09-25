package handlers

import "net/http"

// FriendHandler provides friend invite and listing endpoints.
type FriendHandler struct {
	Friends FriendStore
}

// Invite handles POST /api/v1/friends/invite.
func (h FriendHandler) Invite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	respondJSON(w, http.StatusNotImplemented, map[string]string{
		"message": "friend invite not yet implemented",
	})
}

// List handles GET /api/v1/friends requests.
func (h FriendHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	respondJSON(w, http.StatusNotImplemented, map[string]string{
		"message": "friend listing not yet implemented",
	})
}
