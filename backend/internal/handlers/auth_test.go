package handlers

import (
        "bytes"
        "context"
        "encoding/json"
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
