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
	"github.com/vidfriends/backend/internal/videos"
)

type videoStoreStub struct {
	share     models.VideoShare
	feed      []models.VideoShare
	feedUser  string
	createErr error
	feedErr   error
}

func (s *videoStoreStub) Create(ctx context.Context, share models.VideoShare) error {
	_ = ctx
	s.share = share
	return s.createErr
}

func (s *videoStoreStub) ListFeed(ctx context.Context, userID string) ([]models.VideoShare, error) {
	_ = ctx
	s.feedUser = userID
	if s.feedErr != nil {
		return nil, s.feedErr
	}
	return s.feed, nil
}

type metadataProviderStub struct {
	metadata videos.Metadata
	err      error
}

func (m metadataProviderStub) Lookup(ctx context.Context, url string) (videos.Metadata, error) {
	return m.metadata, m.err
}

func TestVideoHandlerCreateSuccess(t *testing.T) {
	store := &videoStoreStub{}
	metadata := metadataProviderStub{metadata: videos.Metadata{Title: "Test", Description: "Desc", Thumbnail: "thumb.jpg"}}

	handler := VideoHandler{
		Videos:   store,
		Metadata: metadata,
		NowFunc: func() time.Time {
			return time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)
		},
	}

	body, _ := json.Marshal(map[string]string{
		"ownerId": "user-123",
		"url":     "https://example.com/watch?v=123",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/videos", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusCreated)
	}

	if store.share.ID == "" {
		t.Fatal("expected share ID to be set")
	}
	if store.share.Title != "Test" || store.share.Description != "Desc" || store.share.Thumbnail != "thumb.jpg" {
		t.Fatalf("unexpected share metadata: %+v", store.share)
	}
	if store.share.OwnerID != "user-123" || store.share.URL != "https://example.com/watch?v=123" {
		t.Fatalf("unexpected share data: %+v", store.share)
	}
	if !store.share.CreatedAt.Equal(time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected created at: %v", store.share.CreatedAt)
	}

	var resp createVideoResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Share.ID != store.share.ID {
		t.Fatalf("response share mismatch: got %s want %s", resp.Share.ID, store.share.ID)
	}
}

func TestVideoHandlerCreateMetadataError(t *testing.T) {
	handler := VideoHandler{
		Videos:   &videoStoreStub{},
		Metadata: metadataProviderStub{err: errors.New("boom")},
	}

	body := bytes.NewBufferString(`{"ownerId":"user","url":"https://example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/videos", body)
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusBadGateway)
	}
}

func TestVideoHandlerCreateStoreConflict(t *testing.T) {
	handler := VideoHandler{
		Videos:   &videoStoreStub{createErr: repositories.ErrConflict},
		Metadata: metadataProviderStub{metadata: videos.Metadata{Title: "Test"}},
	}

	body := bytes.NewBufferString(`{"ownerId":"user","url":"https://example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/videos", body)
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusConflict)
	}
}

func TestVideoHandlerCreateMissingDeps(t *testing.T) {
	handler := VideoHandler{}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/videos", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestVideoHandlerCreateValidationFailures(t *testing.T) {
	handler := VideoHandler{Videos: &videoStoreStub{}, Metadata: metadataProviderStub{metadata: videos.Metadata{}}}

	cases := []struct {
		name       string
		method     string
		body       string
		wantStatus int
	}{
		{"wrongMethod", http.MethodGet, `{"ownerId":"u","url":"https://example.com"}`, http.StatusMethodNotAllowed},
		{"badJSON", http.MethodPost, "{", http.StatusBadRequest},
		{"missingFields", http.MethodPost, `{"ownerId":"","url":""}`, http.StatusBadRequest},
		{"invalidURL", http.MethodPost, `{"ownerId":"user","url":"not-a-url"}`, http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/api/v1/videos", bytes.NewBufferString(tc.body))
			rec := httptest.NewRecorder()

			handler.Create(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d got %d", tc.wantStatus, rec.Code)
			}
		})
	}
}

func TestVideoHandlerCreateProviderUnavailable(t *testing.T) {
	handler := VideoHandler{
		Videos:   &videoStoreStub{},
		Metadata: metadataProviderStub{err: videos.ErrProviderUnavailable},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/videos", bytes.NewBufferString(`{"ownerId":"u","url":"https://example.com"}`))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d", rec.Code)
	}
}

func TestVideoHandlerCreateStoreFailure(t *testing.T) {
	handler := VideoHandler{
		Videos:   &videoStoreStub{createErr: errors.New("insert failed")},
		Metadata: metadataProviderStub{metadata: videos.Metadata{Title: "ok"}},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/videos", bytes.NewBufferString(`{"ownerId":"u","url":"https://example.com"}`))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d", rec.Code)
	}
}

func TestVideoHandlerFeedSuccess(t *testing.T) {
	now := time.Date(2024, time.January, 2, 15, 0, 0, 0, time.UTC)
	entries := []models.VideoShare{{
		ID:        "share-1",
		OwnerID:   "friend-1",
		URL:       "https://example.com/watch?v=abc",
		Title:     "Example",
		CreatedAt: now,
	}}
	store := &videoStoreStub{feed: entries}
	handler := VideoHandler{Videos: store}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/videos/feed?user=user-123", nil)
	rec := httptest.NewRecorder()

	handler.Feed(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
	if store.feedUser != "user-123" {
		t.Fatalf("expected feed query for user-123 got %s", store.feedUser)
	}

	var resp feedResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Entries) != len(entries) {
		t.Fatalf("expected %d entries got %d", len(entries), len(resp.Entries))
	}
	if resp.Entries[0].ID != entries[0].ID {
		t.Fatalf("unexpected feed response: %+v", resp.Entries[0])
	}
}

func TestVideoHandlerFeedValidation(t *testing.T) {
	handler := VideoHandler{Videos: &videoStoreStub{}}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/videos/feed", nil)
	rec := httptest.NewRecorder()
	handler.Feed(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/videos/feed", nil)
	rec = httptest.NewRecorder()
	handler.Feed(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", rec.Code)
	}
}

func TestVideoHandlerFeedServiceUnavailable(t *testing.T) {
	handler := VideoHandler{}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/videos/feed?user=user-123", nil)
	rec := httptest.NewRecorder()

	handler.Feed(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d", rec.Code)
	}
}

func TestVideoHandlerFeedStoreError(t *testing.T) {
	handler := VideoHandler{Videos: &videoStoreStub{feedErr: errors.New("query failed")}}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/videos/feed?user=user-123", nil)
	rec := httptest.NewRecorder()

	handler.Feed(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d", rec.Code)
	}
}
