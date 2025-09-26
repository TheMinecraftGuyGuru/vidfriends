package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/vidfriends/backend/internal/logging"
	"github.com/vidfriends/backend/internal/models"
	"github.com/vidfriends/backend/internal/repositories"
)

const (
	friendStatusPending  = "pending"
	friendStatusAccepted = "accepted"
	friendStatusBlocked  = "blocked"
)

// FriendHandler provides friend invite and listing endpoints.
type FriendHandler struct {
	Friends     FriendStore
	NowFunc     func() time.Time
	RateLimiter RateLimiter
}

// Invite handles POST /api/v1/friends/invite.
func (h FriendHandler) Invite(w http.ResponseWriter, r *http.Request) {
	ctx, span := logging.StartSpan(r.Context(), "FriendHandler.Invite")
	defer span.End()
	r = r.WithContext(ctx)

	logger := logging.FromContext(ctx)
	if r.Method != http.MethodPost {
		logger.Warn("method not allowed", "method", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if !allowRequest(h.RateLimiter, r, "friends:invite") {
		logger.Warn("rate limit exceeded", "scope", "friends:invite")
		respondJSON(ctx, w, http.StatusTooManyRequests, map[string]string{"error": "too many friend invites"})
		return
	}

	if h.Friends == nil {
		logger.Error("friend service unavailable")
		respondJSON(ctx, w, http.StatusInternalServerError, map[string]string{"error": "friend service unavailable"})
		return
	}

	var req inviteFriendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("invalid invite payload", "error", err)
		respondJSON(ctx, w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.RequesterID = strings.TrimSpace(req.RequesterID)
	req.ReceiverID = strings.TrimSpace(req.ReceiverID)

	if req.RequesterID == "" || req.ReceiverID == "" {
		logger.Warn("invite missing participants", "requesterId", req.RequesterID, "receiverId", req.ReceiverID)
		respondJSON(ctx, w, http.StatusBadRequest, map[string]string{"error": "requesterId and receiverId are required"})
		return
	}

	if req.RequesterID == req.ReceiverID {
		logger.Warn("invite attempted self", "userId", req.RequesterID)
		respondJSON(ctx, w, http.StatusBadRequest, map[string]string{"error": "cannot invite yourself"})
		return
	}

	if _, err := uuid.Parse(req.RequesterID); err != nil {
		logger.Warn("invite invalid requester id", "requesterId", req.RequesterID, "error", err)
		respondJSON(ctx, w, http.StatusBadRequest, map[string]string{"error": "invalid requesterId"})
		return
	}

	if _, err := uuid.Parse(req.ReceiverID); err != nil {
		logger.Warn("invite invalid receiver id", "receiverId", req.ReceiverID, "error", err)
		respondJSON(ctx, w, http.StatusBadRequest, map[string]string{"error": "invalid receiverId"})
		return
	}

	now := h.now()
	friendReq := models.FriendRequest{
		ID:        uuid.NewString(),
		Requester: req.RequesterID,
		Receiver:  req.ReceiverID,
		Status:    friendStatusPending,
		CreatedAt: now,
	}

	if err := h.Friends.CreateRequest(ctx, friendReq); err != nil {
		switch {
		case errors.Is(err, repositories.ErrConflict):
			logger.Warn("friend request already exists", "requesterId", req.RequesterID, "receiverId", req.ReceiverID)
			respondJSON(ctx, w, http.StatusConflict, map[string]string{"error": "friend request already exists"})
		case errors.Is(err, repositories.ErrNotFound):
			logger.Warn("friend invite target missing", "requesterId", req.RequesterID, "receiverId", req.ReceiverID)
			respondJSON(ctx, w, http.StatusNotFound, map[string]string{"error": "user not found"})
		default:
			logger.Error("failed to create friend request", "error", err, "requesterId", req.RequesterID, "receiverId", req.ReceiverID)
			respondJSON(ctx, w, http.StatusInternalServerError, map[string]string{"error": "failed to create friend request"})
		}
		return
	}

	respondJSON(ctx, w, http.StatusCreated, friendRequestResponse{Request: friendReq})
}

// List handles GET /api/v1/friends requests.
func (h FriendHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, span := logging.StartSpan(r.Context(), "FriendHandler.List")
	defer span.End()
	r = r.WithContext(ctx)

	logger := logging.FromContext(ctx)
	if r.Method != http.MethodGet {
		logger.Warn("method not allowed", "method", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if h.Friends == nil {
		logger.Error("friend service unavailable")
		respondJSON(ctx, w, http.StatusInternalServerError, map[string]string{"error": "friend service unavailable"})
		return
	}

	userID := strings.TrimSpace(r.URL.Query().Get("user"))
	if userID == "" {
		logger.Warn("list friends missing user id")
		respondJSON(ctx, w, http.StatusBadRequest, map[string]string{"error": "user query parameter is required"})
		return
	}

	requests, err := h.Friends.ListForUser(ctx, userID)
	if err != nil {
		logger.Error("failed to list friend requests", "error", err, "userId", userID)
		respondJSON(ctx, w, http.StatusInternalServerError, map[string]string{"error": "failed to list friend requests"})
		return
	}

	respondJSON(ctx, w, http.StatusOK, listFriendsResponse{Requests: requests})
}

// Respond handles POST /api/v1/friends/respond requests to accept or block invites.
func (h FriendHandler) Respond(w http.ResponseWriter, r *http.Request) {
	ctx, span := logging.StartSpan(r.Context(), "FriendHandler.Respond")
	defer span.End()
	r = r.WithContext(ctx)

	logger := logging.FromContext(ctx)
	if r.Method != http.MethodPost {
		logger.Warn("method not allowed", "method", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if h.Friends == nil {
		logger.Error("friend service unavailable")
		respondJSON(ctx, w, http.StatusInternalServerError, map[string]string{"error": "friend service unavailable"})
		return
	}

	var req respondFriendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("invalid respond payload", "error", err)
		respondJSON(ctx, w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.RequestID = strings.TrimSpace(req.RequestID)
	action := strings.ToLower(strings.TrimSpace(req.Action))
	if req.RequestID == "" || action == "" {
		logger.Warn("respond missing fields", "requestId", req.RequestID, "action", action)
		respondJSON(ctx, w, http.StatusBadRequest, map[string]string{"error": "requestId and action are required"})
		return
	}

	var status string
	switch action {
	case "accept":
		status = friendStatusAccepted
	case "block":
		status = friendStatusBlocked
	default:
		logger.Warn("invalid respond action", "action", action)
		respondJSON(ctx, w, http.StatusBadRequest, map[string]string{"error": "action must be accept or block"})
		return
	}

	if err := h.Friends.UpdateStatus(ctx, req.RequestID, status); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			logger.Warn("friend request not found", "requestId", req.RequestID)
			respondJSON(ctx, w, http.StatusNotFound, map[string]string{"error": "friend request not found"})
			return
		}
		logger.Error("failed to update friend request", "error", err, "requestId", req.RequestID, "status", status)
		respondJSON(ctx, w, http.StatusInternalServerError, map[string]string{"error": "failed to update friend request"})
		return
	}

	respondJSON(ctx, w, http.StatusOK, map[string]string{
		"requestId": req.RequestID,
		"status":    status,
	})
}

func (h FriendHandler) now() time.Time {
	if h.NowFunc != nil {
		return h.NowFunc()
	}
	return time.Now().UTC()
}

type inviteFriendRequest struct {
	RequesterID string `json:"requesterId"`
	ReceiverID  string `json:"receiverId"`
}

type respondFriendRequest struct {
	RequestID string `json:"requestId"`
	Action    string `json:"action"`
}

type friendRequestResponse struct {
	Request models.FriendRequest `json:"request"`
}

type listFriendsResponse struct {
	Requests []models.FriendRequest `json:"requests"`
}
