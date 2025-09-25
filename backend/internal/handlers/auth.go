package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/vidfriends/backend/internal/auth"
	"github.com/vidfriends/backend/internal/models"
	"github.com/vidfriends/backend/internal/repositories"
)

// AuthHandler implements user authentication endpoints.
type AuthHandler struct {
	Users    UserStore
	Sessions SessionManager
	NowFunc  func() time.Time
}

// Login handles POST /api/v1/auth/login requests.
func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if h.Users == nil || h.Sessions == nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "authentication services unavailable"})
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}

	user, err := h.Users.FindByEmail(r.Context(), req.Email)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	tokens, err := h.Sessions.Issue(r.Context(), user.ID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create session"})
		return
	}

	respondJSON(w, http.StatusOK, authResponse{Tokens: tokens})
}

// SignUp handles POST /api/v1/auth/signup requests.
func (h AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if h.Users == nil || h.Sessions == nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "authentication services unavailable"})
		return
	}

	var req signUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}

	if _, err := mail.ParseAddress(req.Email); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid email address"})
		return
	}

	if len(req.Password) < 8 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
		return
	}

	if _, err := h.Users.FindByEmail(r.Context(), req.Email); err == nil {
		respondJSON(w, http.StatusConflict, map[string]string{"error": "account already exists"})
		return
	} else if err != nil && !errors.Is(err, repositories.ErrNotFound) {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to verify existing accounts"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to secure password"})
		return
	}

	now := h.now()
	user := models.User{
		ID:        uuid.NewString(),
		Email:     req.Email,
		Password:  string(hashed),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.Users.Create(r.Context(), user); err != nil {
		if errors.Is(err, repositories.ErrConflict) {
			respondJSON(w, http.StatusConflict, map[string]string{"error": "account already exists"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create account"})
		return
	}

	tokens, err := h.Sessions.Issue(r.Context(), user.ID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create session"})
		return
	}

	respondJSON(w, http.StatusCreated, authResponse{Tokens: tokens})
}

// Refresh exchanges a refresh token for a new session.
func (h AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if h.Sessions == nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "session service unavailable"})
		return
	}

	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if req.RefreshToken == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "refresh token is required"})
		return
	}

	tokens, err := h.Sessions.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		status := http.StatusUnauthorized
		if errors.Is(err, auth.ErrRefreshTokenExpired) || errors.Is(err, auth.ErrSessionNotFound) {
			status = http.StatusUnauthorized
		} else {
			status = http.StatusInternalServerError
		}
		respondJSON(w, status, map[string]string{"error": "unable to refresh session"})
		return
	}

	respondJSON(w, http.StatusOK, authResponse{Tokens: tokens})
}

// RequestPasswordReset handles POST /api/v1/auth/password-reset requests.
func (h AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if h.Users == nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "authentication services unavailable"})
		return
	}

	var req passwordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "email is required"})
		return
	}

	if _, err := mail.ParseAddress(req.Email); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid email address"})
		return
	}

	if _, err := h.Users.FindByEmail(r.Context(), req.Email); err != nil {
		if !errors.Is(err, repositories.ErrNotFound) {
			respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to process password reset"})
			return
		}
	}

	respondJSON(w, http.StatusAccepted, map[string]string{
		"status": "If an account exists for that email, password reset instructions have been sent.",
	})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type passwordResetRequest struct {
	Email string `json:"email"`
}

type authResponse struct {
	Tokens models.SessionTokens `json:"tokens"`
}

func (h AuthHandler) now() time.Time {
	if h.NowFunc != nil {
		return h.NowFunc()
	}
	return time.Now().UTC()
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
