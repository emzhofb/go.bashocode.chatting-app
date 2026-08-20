package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type Handler struct {
	DB      *sql.DB
	Timeout time.Duration
}

func (h Handler) Live(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h Handler) Ready(w http.ResponseWriter, r *http.Request) {
	if h.DB == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "not_ready"})
		return
	}
	timeout := h.Timeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()
	if err := h.DB.PingContext(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "not_ready"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
