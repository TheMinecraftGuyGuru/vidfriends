package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/vidfriends/backend/internal/models"
	"github.com/vidfriends/backend/internal/repositories"
	"github.com/vidfriends/backend/internal/videos"
)

// VideoHandler provides endpoints for sharing and fetching videos.
type VideoHandler struct {
	Videos   VideoStore
	Metadata VideoMetadataProvider
	NowFunc  func() time.Time
}

// Create handles POST /api/v1/videos.
func (h VideoHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if h.Videos == nil || h.Metadata == nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "video services unavailable"})
		return
	}

	var req createVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.OwnerID = strings.TrimSpace(req.OwnerID)
	req.URL = strings.TrimSpace(req.URL)
	if req.OwnerID == "" || req.URL == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "ownerId and url are required"})
		return
	}

	if _, err := url.ParseRequestURI(req.URL); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid url"})
		return
	}

	metadata, err := h.Metadata.Lookup(r.Context(), req.URL)
	if err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, videos.ErrProviderUnavailable) {
			status = http.StatusInternalServerError
		}
		respondJSON(w, status, map[string]string{"error": "failed to fetch video metadata"})
		return
	}

	now := h.now()
	share := models.VideoShare{
		ID:          uuid.NewString(),
		OwnerID:     req.OwnerID,
		URL:         req.URL,
		Title:       metadata.Title,
		Description: metadata.Description,
		Thumbnail:   metadata.Thumbnail,
		CreatedAt:   now,
	}

	if err := h.Videos.Create(r.Context(), share); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, repositories.ErrConflict) {
			status = http.StatusConflict
		}
		respondJSON(w, status, map[string]string{"error": "failed to store video share"})
		return
	}

	respondJSON(w, http.StatusCreated, createVideoResponse{Share: share})
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

func (h VideoHandler) now() time.Time {
	if h.NowFunc != nil {
		return h.NowFunc()
	}
	return time.Now().UTC()
}

type createVideoRequest struct {
	OwnerID string `json:"ownerId"`
	URL     string `json:"url"`
}

type createVideoResponse struct {
	Share models.VideoShare `json:"share"`
}
