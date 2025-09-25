package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vidfriends/backend/internal/models"
	"github.com/vidfriends/backend/internal/repositories"
)

type inMemoryFriendStore struct {
	requests map[string]models.FriendRequest
}

func newInMemoryFriendStore() *inMemoryFriendStore {
	return &inMemoryFriendStore{requests: make(map[string]models.FriendRequest)}
}

func (s *inMemoryFriendStore) CreateRequest(_ context.Context, request models.FriendRequest) error {
	for _, existing := range s.requests {
		if existing.Requester == request.Requester && existing.Receiver == request.Receiver {
			return repositories.ErrConflict
		}
	}
	s.requests[request.ID] = request
	return nil
}

func (s *inMemoryFriendStore) ListForUser(_ context.Context, userID string) ([]models.FriendRequest, error) {
	var out []models.FriendRequest
	for _, request := range s.requests {
		if request.Requester == userID || request.Receiver == userID {
			out = append(out, request)
		}
	}
	return out, nil
}

func (s *inMemoryFriendStore) UpdateStatus(_ context.Context, requestID, status string) error {
	request, ok := s.requests[requestID]
	if !ok {
		return repositories.ErrNotFound
	}
	request.Status = status
	respondedAt := time.Now().UTC()
	request.RespondedAt = &respondedAt
	s.requests[requestID] = request
	return nil
}

type stubFriendStore struct {
	createErr error
	listErr   error
	updateErr error
}

func (s *stubFriendStore) CreateRequest(context.Context, models.FriendRequest) error {
	return s.createErr
}

func (s *stubFriendStore) ListForUser(context.Context, string) ([]models.FriendRequest, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return []models.FriendRequest{{ID: "req-1"}}, nil
}

func (s *stubFriendStore) UpdateStatus(context.Context, string, string) error {
	return s.updateErr
}

func TestFriendHandlerInvite(t *testing.T) {
	store := newInMemoryFriendStore()
	now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	handler := FriendHandler{Friends: store, NowFunc: func() time.Time { return now }}

	body, err := json.Marshal(inviteFriendRequest{RequesterID: "user-1", ReceiverID: "user-2"})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/friends/invite", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.Invite(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d got %d", http.StatusCreated, rec.Code)
	}

	var resp friendRequestResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Request.Status != friendStatusPending {
		t.Fatalf("expected status %q got %q", friendStatusPending, resp.Request.Status)
	}

	if resp.Request.CreatedAt != now {
		t.Fatalf("expected createdAt to use NowFunc")
	}

	if _, ok := store.requests[resp.Request.ID]; !ok {
		t.Fatalf("expected request to be stored")
	}
}

func TestFriendHandlerInviteFailures(t *testing.T) {
	body := []byte(`{"requesterId":"user-1","receiverId":"user-2"}`)

	cases := []struct {
		name       string
		handler    FriendHandler
		method     string
		body       []byte
		wantStatus int
	}{
		{"wrongMethod", FriendHandler{Friends: newInMemoryFriendStore()}, http.MethodGet, body, http.StatusMethodNotAllowed},
		{"missingStore", FriendHandler{}, http.MethodPost, body, http.StatusInternalServerError},
		{"badJSON", FriendHandler{Friends: newInMemoryFriendStore()}, http.MethodPost, []byte("{"), http.StatusBadRequest},
		{"missingFields", FriendHandler{Friends: newInMemoryFriendStore()}, http.MethodPost, []byte(`{"requesterId":"","receiverId":""}`), http.StatusBadRequest},
		{"selfInvite", FriendHandler{Friends: newInMemoryFriendStore()}, http.MethodPost, []byte(`{"requesterId":"same","receiverId":"same"}`), http.StatusBadRequest},
		{"conflict", FriendHandler{Friends: &stubFriendStore{createErr: repositories.ErrConflict}}, http.MethodPost, body, http.StatusConflict},
		{"notFound", FriendHandler{Friends: &stubFriendStore{createErr: repositories.ErrNotFound}}, http.MethodPost, body, http.StatusNotFound},
		{"internal", FriendHandler{Friends: &stubFriendStore{createErr: errors.New("boom")}}, http.MethodPost, body, http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/api/v1/friends/invite", bytes.NewReader(tc.body))
			rec := httptest.NewRecorder()

			tc.handler.Invite(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d got %d", tc.wantStatus, rec.Code)
			}
		})
	}
}

func TestFriendHandlerList(t *testing.T) {
	store := newInMemoryFriendStore()
	store.requests["req-1"] = models.FriendRequest{ID: "req-1", Requester: "user-1", Receiver: "user-2", Status: friendStatusPending}
	handler := FriendHandler{Friends: store}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/friends?user=user-1", nil)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d got %d", http.StatusOK, rec.Code)
	}

	var resp listFriendsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(resp.Requests) != 1 || resp.Requests[0].ID != "req-1" {
		t.Fatalf("unexpected response payload: %+v", resp)
	}
}

func TestFriendHandlerListFailures(t *testing.T) {
	handler := FriendHandler{Friends: &stubFriendStore{listErr: errors.New("db down")}}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/friends", nil)
	rec := httptest.NewRecorder()
	handler.List(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected method not allowed got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/friends", nil)
	rec = httptest.NewRecorder()
	handler.List(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/friends?user=user-1", nil)
	rec = httptest.NewRecorder()
	handler = FriendHandler{}
	handler.List(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal error got %d", rec.Code)
	}

	handler = FriendHandler{Friends: &stubFriendStore{listErr: errors.New("db down")}}
	rec = httptest.NewRecorder()
	handler.List(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal error got %d", rec.Code)
	}
}

func TestFriendHandlerRespond(t *testing.T) {
	store := newInMemoryFriendStore()
	store.requests["req-1"] = models.FriendRequest{ID: "req-1", Requester: "user-1", Receiver: "user-2", Status: friendStatusPending}
	handler := FriendHandler{Friends: store}

	body, err := json.Marshal(respondFriendRequest{RequestID: "req-1", Action: "accept"})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/friends/respond", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.Respond(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d got %d", http.StatusOK, rec.Code)
	}

	updated := store.requests["req-1"]
	if updated.Status != friendStatusAccepted {
		t.Fatalf("expected status %q got %q", friendStatusAccepted, updated.Status)
	}

	if updated.RespondedAt == nil {
		t.Fatalf("expected respondedAt to be set")
	}
}

func TestFriendHandlerRespondFailures(t *testing.T) {
	handler := FriendHandler{Friends: &stubFriendStore{updateErr: repositories.ErrNotFound}}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/friends/respond", nil)
	rec := httptest.NewRecorder()
	handler.Respond(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected method not allowed got %d", rec.Code)
	}

	handler = FriendHandler{}
	body := []byte(`{"requestId":"req-1","action":"accept"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/friends/respond", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	handler.Respond(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal error got %d", rec.Code)
	}

	handler = FriendHandler{Friends: &stubFriendStore{}}
	req = httptest.NewRequest(http.MethodPost, "/api/v1/friends/respond", bytes.NewReader([]byte("{")))
	rec = httptest.NewRecorder()
	handler.Respond(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/friends/respond", bytes.NewReader([]byte(`{"requestId":"","action":""}`)))
	rec = httptest.NewRecorder()
	handler.Respond(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/friends/respond", bytes.NewReader([]byte(`{"requestId":"req-1","action":"maybe"}`)))
	rec = httptest.NewRecorder()
	handler.Respond(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request got %d", rec.Code)
	}

	handler = FriendHandler{Friends: &stubFriendStore{updateErr: repositories.ErrNotFound}}
	req = httptest.NewRequest(http.MethodPost, "/api/v1/friends/respond", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	handler.Respond(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected not found got %d", rec.Code)
	}

	handler = FriendHandler{Friends: &stubFriendStore{updateErr: errors.New("db down")}}
	req = httptest.NewRequest(http.MethodPost, "/api/v1/friends/respond", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	handler.Respond(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal error got %d", rec.Code)
	}
}
