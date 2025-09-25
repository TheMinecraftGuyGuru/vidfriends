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

	"golang.org/x/crypto/bcrypt"

	"github.com/vidfriends/backend/internal/auth"
	"github.com/vidfriends/backend/internal/models"
	"github.com/vidfriends/backend/internal/repositories"
)

type inMemoryUserStore struct {
	users map[string]models.User
}

func newInMemoryUserStore() *inMemoryUserStore {
	return &inMemoryUserStore{users: make(map[string]models.User)}
}

func (s *inMemoryUserStore) Create(_ context.Context, user models.User) error {
	if _, exists := s.users[user.Email]; exists {
		return repositories.ErrConflict
	}
	s.users[user.Email] = user
	return nil
}

func (s *inMemoryUserStore) FindByEmail(_ context.Context, email string) (models.User, error) {
	user, ok := s.users[email]
	if !ok {
		return models.User{}, repositories.ErrNotFound
	}
	return user, nil
}

type failingUserStore struct {
	createErr error
	findErr   error
}

func (s failingUserStore) Create(context.Context, models.User) error {
	return s.createErr
}

func (s failingUserStore) FindByEmail(context.Context, string) (models.User, error) {
	return models.User{}, s.findErr
}

type stubSessionManager struct {
	issueTokens   models.SessionTokens
	issueErr      error
	refreshTokens models.SessionTokens
	refreshErr    error
	issuedFor     string
	refreshedWith string
}

func (s *stubSessionManager) Issue(_ context.Context, userID string) (models.SessionTokens, error) {
	s.issuedFor = userID
	if s.issueErr != nil {
		return models.SessionTokens{}, s.issueErr
	}
	return s.issueTokens, nil
}

func (s *stubSessionManager) Refresh(_ context.Context, refreshToken string) (models.SessionTokens, error) {
	s.refreshedWith = refreshToken
	if s.refreshErr != nil {
		return models.SessionTokens{}, s.refreshErr
	}
	return s.refreshTokens, nil
}

func TestAuthHandlerSignUp(t *testing.T) {
	store := newInMemoryUserStore()
	manager := auth.NewManager(time.Minute, time.Hour)
	handler := AuthHandler{Users: store, Sessions: manager}

	body, err := json.Marshal(signUpRequest{Email: "test@example.com", Password: "supersafe"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.SignUp(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d got %d", http.StatusCreated, rec.Code)
	}

	var resp authResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Tokens.AccessToken == "" || resp.Tokens.RefreshToken == "" {
		t.Fatalf("expected tokens to be issued, got %+v", resp.Tokens)
	}

	stored, err := store.FindByEmail(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("expected user to be stored: %v", err)
	}

	if bcrypt.CompareHashAndPassword([]byte(stored.Password), []byte("supersafe")) != nil {
		t.Fatal("stored password is not hashed")
	}
}

func TestAuthHandlerSignUpValidationErrors(t *testing.T) {
	t.Parallel()

	manager := auth.NewManager(time.Minute, time.Hour)
	handler := AuthHandler{Users: newInMemoryUserStore(), Sessions: manager}

	cases := []struct {
		name       string
		body       []byte
		wantStatus int
	}{
		{"nonJSON", []byte("{"), http.StatusBadRequest},
		{"missingFields", []byte(`{"email":"","password":""}`), http.StatusBadRequest},
		{"invalidEmail", []byte(`{"email":"bad","password":"password"}`), http.StatusBadRequest},
		{"shortPassword", []byte(`{"email":"user@example.com","password":"short"}`), http.StatusBadRequest},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(tc.body))
			rec := httptest.NewRecorder()

			handler.SignUp(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d got %d", tc.wantStatus, rec.Code)
			}
		})
	}
}

func TestAuthHandlerSignUpExistingAccount(t *testing.T) {
	store := newInMemoryUserStore()
	store.users["taken@example.com"] = models.User{Email: "taken@example.com"}

	handler := AuthHandler{Users: store, Sessions: auth.NewManager(time.Minute, time.Hour)}

	body, _ := json.Marshal(signUpRequest{Email: "taken@example.com", Password: "password123"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.SignUp(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected conflict got %d", rec.Code)
	}
}

func TestAuthHandlerSignUpStorageFailures(t *testing.T) {
	handler := AuthHandler{
		Users:    failingUserStore{findErr: errors.New("db offline")},
		Sessions: &stubSessionManager{issueTokens: models.SessionTokens{AccessToken: "a", RefreshToken: "b"}},
	}

	body, _ := json.Marshal(signUpRequest{Email: "user@example.com", Password: "password123"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.SignUp(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal error got %d", rec.Code)
	}

	handler = AuthHandler{
		Users:    failingUserStore{createErr: errors.New("insert failed"), findErr: repositories.ErrNotFound},
		Sessions: &stubSessionManager{issueTokens: models.SessionTokens{AccessToken: "a", RefreshToken: "b"}},
	}
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	handler.SignUp(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal error got %d", rec.Code)
	}
}

func TestAuthHandlerLogin(t *testing.T) {
	store := newInMemoryUserStore()
	manager := auth.NewManager(time.Minute, time.Hour)
	handler := AuthHandler{Users: store, Sessions: manager}

	hashed, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	store.users["login@example.com"] = models.User{ID: "user-1", Email: "login@example.com", Password: string(hashed)}

	body, err := json.Marshal(loginRequest{Email: "login@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d got %d", http.StatusOK, rec.Code)
	}

	var resp authResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Tokens.AccessToken == "" || resp.Tokens.RefreshToken == "" {
		t.Fatalf("expected tokens to be issued, got %+v", resp.Tokens)
	}
}

func TestAuthHandlerLoginFailures(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	store := newInMemoryUserStore()
	store.users["user@example.com"] = models.User{ID: "user-1", Email: "user@example.com", Password: string(hashed)}

	handler := AuthHandler{Users: store, Sessions: auth.NewManager(time.Minute, time.Hour)}

	cases := []struct {
		name       string
		body       any
		wantStatus int
	}{
		{"badMethod", nil, http.StatusMethodNotAllowed},
		{"invalidJSON", "{", http.StatusBadRequest},
		{"missingFields", loginRequest{}, http.StatusBadRequest},
		{"notFound", loginRequest{Email: "missing@example.com", Password: "password123"}, http.StatusUnauthorized},
		{"wrongPassword", loginRequest{Email: "user@example.com", Password: "wrong"}, http.StatusUnauthorized},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body []byte
			var err error
			method := http.MethodPost
			if tc.name == "badMethod" {
				method = http.MethodGet
			}

			switch v := tc.body.(type) {
			case string:
				body = []byte(v)
			case nil:
				body = nil
			default:
				body, err = json.Marshal(v)
				if err != nil {
					t.Fatalf("marshal: %v", err)
				}
			}

			req := httptest.NewRequest(method, "/api/v1/auth/login", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.Login(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d got %d", tc.wantStatus, rec.Code)
			}
		})
	}

	handler = AuthHandler{Users: store, Sessions: &stubSessionManager{issueErr: errors.New("boom")}}

	body, _ := json.Marshal(loginRequest{Email: "user@example.com", Password: "password123"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal error got %d", rec.Code)
	}
}

func TestAuthHandlerRefresh(t *testing.T) {
	manager := auth.NewManager(time.Minute, time.Hour)
	tokens, err := manager.Issue(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("issue tokens: %v", err)
	}

	handler := AuthHandler{Sessions: manager}

	body, err := json.Marshal(refreshRequest{RefreshToken: tokens.RefreshToken})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.Refresh(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d got %d", http.StatusOK, rec.Code)
	}

	var resp authResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Tokens.RefreshToken == tokens.RefreshToken {
		t.Fatal("expected a new refresh token to be issued")
	}
}

func TestAuthHandlerRefreshFailures(t *testing.T) {
	manager := auth.NewManager(time.Minute, time.Hour)
	tokens, _ := manager.Issue(context.Background(), "user-123")

	cases := []struct {
		name       string
		handler    AuthHandler
		method     string
		body       []byte
		wantStatus int
	}{
		{"wrongMethod", AuthHandler{Sessions: manager}, http.MethodGet, nil, http.StatusMethodNotAllowed},
		{"missingSession", AuthHandler{}, http.MethodPost, nil, http.StatusInternalServerError},
		{"badJSON", AuthHandler{Sessions: manager}, http.MethodPost, []byte("{"), http.StatusBadRequest},
		{"missingToken", AuthHandler{Sessions: manager}, http.MethodPost, []byte(`{"refreshToken":""}`), http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/api/v1/auth/refresh", bytes.NewReader(tc.body))
			rec := httptest.NewRecorder()

			tc.handler.Refresh(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d got %d", tc.wantStatus, rec.Code)
			}
		})
	}

	handler := AuthHandler{Sessions: &stubSessionManager{refreshErr: errors.New("db down")}}
	body, _ := json.Marshal(refreshRequest{RefreshToken: tokens.RefreshToken})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.Refresh(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal error got %d", rec.Code)
	}

	handler = AuthHandler{Sessions: &stubSessionManager{refreshErr: auth.ErrSessionNotFound}}
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	handler.Refresh(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized got %d", rec.Code)
	}
}
