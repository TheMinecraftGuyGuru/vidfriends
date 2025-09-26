package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/vidfriends/backend/internal/logging"
	"github.com/vidfriends/backend/internal/models"
	"github.com/vidfriends/backend/internal/repositories"
	"github.com/vidfriends/backend/internal/videos"
)

// VideoHandler provides endpoints for sharing and fetching videos.
type VideoHandler struct {
	Videos   VideoStore
	Metadata VideoMetadataProvider
	Assets   VideoAssetIngestor
	NowFunc  func() time.Time
}

// Create handles POST /api/v1/videos.
func (h VideoHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := logging.StartSpan(r.Context(), "VideoHandler.Create")
	defer span.End()
	r = r.WithContext(ctx)

	logger := logging.FromContext(ctx)
	if r.Method != http.MethodPost {
		logger.Warn("method not allowed", "method", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if h.Videos == nil || h.Metadata == nil {
		logger.Error("video services unavailable", "hasVideos", h.Videos != nil, "hasMetadata", h.Metadata != nil)
		respondJSON(ctx, w, http.StatusInternalServerError, map[string]string{"error": "video services unavailable"})
		return
	}

	var req createVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("invalid create video payload", "error", err)
		respondJSON(ctx, w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.OwnerID = strings.TrimSpace(req.OwnerID)
	req.URL = strings.TrimSpace(req.URL)
	if req.OwnerID == "" || req.URL == "" {
		logger.Warn("missing create video fields", "ownerId", req.OwnerID, "url", req.URL)
		respondJSON(ctx, w, http.StatusBadRequest, map[string]string{"error": "ownerId and url are required"})
		return
	}

	if _, err := url.ParseRequestURI(req.URL); err != nil {
		logger.Warn("invalid video url", "url", req.URL, "error", err)
		respondJSON(ctx, w, http.StatusBadRequest, map[string]string{"error": "invalid url"})
		return
	}

	metadata, err := h.Metadata.Lookup(ctx, req.URL)
	if err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, videos.ErrProviderUnavailable) {
			status = http.StatusInternalServerError
		}
		logger.Error("failed to lookup video metadata", "error", err, "url", req.URL)
		respondJSON(ctx, w, status, map[string]string{"error": "failed to fetch video metadata"})
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
		AssetStatus: models.AssetStatusPending,
	}

	if err := h.Videos.Create(ctx, share); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, repositories.ErrConflict) {
			status = http.StatusConflict
		}
		logger.Error("failed to persist video share", "error", err, "ownerId", share.OwnerID, "url", share.URL)
		respondJSON(ctx, w, status, map[string]string{"error": "failed to store video share"})
		return
	}

	if h.Assets != nil {
		if err := h.Assets.Enqueue(ctx, share); err != nil {
			logger.Error("failed to enqueue asset ingestion", "error", err, "shareId", share.ID)
		}
	}

	respondJSON(ctx, w, http.StatusCreated, createVideoResponse{Share: share})
}

// Feed handles GET /api/v1/videos/feed.
func (h VideoHandler) Feed(w http.ResponseWriter, r *http.Request) {
	ctx, span := logging.StartSpan(r.Context(), "VideoHandler.Feed")
	defer span.End()
	r = r.WithContext(ctx)

	logger := logging.FromContext(ctx)
	if r.Method != http.MethodGet {
		logger.Warn("method not allowed", "method", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if h.Videos == nil {
		logger.Error("video service unavailable")
		respondJSON(ctx, w, http.StatusInternalServerError, map[string]string{"error": "video service unavailable"})
		return
	}

	userID := strings.TrimSpace(r.URL.Query().Get("user"))
	if userID == "" {
		logger.Warn("feed missing user id")
		respondJSON(ctx, w, http.StatusBadRequest, map[string]string{"error": "user query parameter is required"})
		return
	}

	feed, err := h.Videos.ListFeed(ctx, userID)
	if err != nil {
		logger.Error("failed to load video feed", "error", err, "userId", userID)
		respondJSON(ctx, w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch video feed"})
		return
	}

	respondJSON(ctx, w, http.StatusOK, feedResponse{Entries: feed})
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

type feedResponse struct {
	Entries []models.VideoShare `json:"entries"`
}
