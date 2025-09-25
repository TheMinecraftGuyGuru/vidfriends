package handlers

import (
	"encoding/json"
	"net/http"
)

// HealthHandler responds with service health information.
type HealthHandler struct{}

// Handle implements GET /healthz.
func (HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	payload := map[string]string{
		"status": "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}
