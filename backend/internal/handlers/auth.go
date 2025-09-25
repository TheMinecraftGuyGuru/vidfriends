package handlers

import (
	"encoding/json"
	"net/http"
)

// AuthHandler implements user authentication endpoints.
type AuthHandler struct {
	Users UserStore
}

// Login handles POST /api/v1/auth/login requests.
func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	respondJSON(w, http.StatusNotImplemented, map[string]string{
		"message": "login not yet implemented",
	})
}

// SignUp handles POST /api/v1/auth/signup requests.
func (h AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	respondJSON(w, http.StatusNotImplemented, map[string]string{
		"message": "signup not yet implemented",
	})
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
