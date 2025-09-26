package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/vidfriends/backend/internal/logging"
)

// HealthHandler responds with service health information.
type HealthHandler struct{}

// Handle implements GET /healthz.
func (HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx, span := logging.StartSpan(r.Context(), "HealthHandler.Handle")
	defer span.End()
	r = r.WithContext(ctx)

	logger := logging.FromContext(ctx)
	if r.Method != http.MethodGet {
		logger.Warn("method not allowed", "method", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	payload := map[string]string{
		"status": "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logger.Error("encode health response", "error", err)
	}
}
