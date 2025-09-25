package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

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
	Friends FriendStore
	NowFunc func() time.Time
}

// Invite handles POST /api/v1/friends/invite.
func (h FriendHandler) Invite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if h.Friends == nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "friend service unavailable"})
		return
	}

	var req inviteFriendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.RequesterID = strings.TrimSpace(req.RequesterID)
	req.ReceiverID = strings.TrimSpace(req.ReceiverID)

	if req.RequesterID == "" || req.ReceiverID == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "requesterId and receiverId are required"})
		return
	}

	if req.RequesterID == req.ReceiverID {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot invite yourself"})
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

	if err := h.Friends.CreateRequest(r.Context(), friendReq); err != nil {
		switch {
		case errors.Is(err, repositories.ErrConflict):
			respondJSON(w, http.StatusConflict, map[string]string{"error": "friend request already exists"})
		case errors.Is(err, repositories.ErrNotFound):
			respondJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		default:
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create friend request"})
		}
		return
	}

	respondJSON(w, http.StatusCreated, friendRequestResponse{Request: friendReq})
}

// List handles GET /api/v1/friends requests.
func (h FriendHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if h.Friends == nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "friend service unavailable"})
		return
	}

	userID := strings.TrimSpace(r.URL.Query().Get("user"))
	if userID == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "user query parameter is required"})
		return
	}

	requests, err := h.Friends.ListForUser(r.Context(), userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list friend requests"})
		return
	}

	respondJSON(w, http.StatusOK, listFriendsResponse{Requests: requests})
}

// Respond handles POST /api/v1/friends/respond requests to accept or block invites.
func (h FriendHandler) Respond(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if h.Friends == nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "friend service unavailable"})
		return
	}

	var req respondFriendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.RequestID = strings.TrimSpace(req.RequestID)
	action := strings.ToLower(strings.TrimSpace(req.Action))
	if req.RequestID == "" || action == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "requestId and action are required"})
		return
	}

	var status string
	switch action {
	case "accept":
		status = friendStatusAccepted
	case "block":
		status = friendStatusBlocked
	default:
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "action must be accept or block"})
		return
	}

	if err := h.Friends.UpdateStatus(r.Context(), req.RequestID, status); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			respondJSON(w, http.StatusNotFound, map[string]string{"error": "friend request not found"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update friend request"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
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
