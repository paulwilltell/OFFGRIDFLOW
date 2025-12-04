package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/emissions"
)

// FactorsHandlerConfig configures factor endpoints.
type FactorsHandlerConfig struct {
	Registry emissions.FactorRegistry
}

// NewFactorsHandler exposes GET/POST for emission factors.
func NewFactorsHandler(cfg FactorsHandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.Registry == nil {
			responders.ServiceUnavailable(w, "factor registry not configured", 0)
			return
		}

		switch r.Method {
		case http.MethodGet:
			handleListFactors(w, r, cfg.Registry)
		case http.MethodPost:
			handleRegisterFactor(w, r, cfg.Registry)
		default:
			responders.MethodNotAllowed(w, http.MethodGet, http.MethodPost)
		}
	}
}

func handleListFactors(w http.ResponseWriter, r *http.Request, registry emissions.FactorRegistry) {
	query := emissions.FactorQuery{
		Region:   r.URL.Query().Get("region"),
		Source:   r.URL.Query().Get("source"),
		Category: r.URL.Query().Get("category"),
		Unit:     r.URL.Query().Get("unit"),
	}
	scopeVal := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))
	switch scopeVal {
	case "scope1", "1":
		query.Scope = emissions.Scope1
	case "scope2", "2":
		query.Scope = emissions.Scope2
	case "scope3", "3":
		query.Scope = emissions.Scope3
	}

	if ts := strings.TrimSpace(r.URL.Query().Get("valid_at")); ts != "" {
		if parsed, err := strconv.ParseInt(ts, 10, 64); err == nil {
			query.ValidAt = time.UnixMilli(parsed)
		} else if parsedTime, err := time.Parse(time.RFC3339, ts); err == nil {
			query.ValidAt = parsedTime
		}
	}

	factors, err := registry.ListFactors(r.Context(), query)
	if err != nil {
		responders.InternalError(w, "failed to list factors")
		return
	}
	responders.JSON(w, http.StatusOK, factors)
}

func handleRegisterFactor(w http.ResponseWriter, r *http.Request, registry emissions.FactorRegistry) {
	var factor emissions.EmissionFactor
	if err := json.NewDecoder(r.Body).Decode(&factor); err != nil {
		responders.BadRequest(w, "invalid_request", "invalid JSON payload")
		return
	}
	if factor.ID == "" {
		responders.BadRequest(w, "invalid_request", "id is required")
		return
	}
	if err := registry.RegisterFactor(r.Context(), factor); err != nil {
		responders.BadRequest(w, "register_failed", err.Error())
		return
	}
	responders.JSON(w, http.StatusCreated, factor)
}
