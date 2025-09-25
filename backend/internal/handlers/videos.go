package handlers

import "net/http"

// VideoHandler provides endpoints for sharing and fetching videos.
type VideoHandler struct {
	Videos VideoStore
}

// Create handles POST /api/v1/videos.
func (h VideoHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	respondJSON(w, http.StatusNotImplemented, map[string]string{
		"message": "video creation not yet implemented",
	})
}

// Feed handles GET /api/v1/videos/feed.
func (h VideoHandler) Feed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	respondJSON(w, http.StatusNotImplemented, map[string]string{
		"message": "video feed not yet implemented",
	})
}
