package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/ingestion"
)

// IngestionHandlerConfig configures ingestion status endpoints.
type IngestionHandlerConfig struct {
	LogStore ingestion.LogStore
}

// NewIngestionStatusHandler returns an HTTP handler that serves recent ingestion logs.
func NewIngestionStatusHandler(cfg IngestionHandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.LogStore == nil {
			responders.Error(w, http.StatusNotImplemented, "ingestion_logs_unavailable", "ingestion log store not configured")
			return
		}

		limit := 50
		if raw := r.URL.Query().Get("limit"); raw != "" {
			if v, err := strconv.Atoi(raw); err == nil && v > 0 {
				limit = v
			}
		}

		logs, err := cfg.LogStore.List(r.Context(), limit)
		if err != nil {
			responders.InternalError(w, "failed to load ingestion logs")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(logs)
	}
}
