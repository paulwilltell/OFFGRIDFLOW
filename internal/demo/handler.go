package demo

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/example/offgridflow/internal/api/http/responders"
)

// Handler provides demo mode endpoints and data.
type Handler struct {
	config   DemoConfig
	data     *DemoData
	dataOnce sync.Once
	mu       sync.RWMutex
}

// NewHandler creates a new demo handler.
func NewHandler(config DemoConfig) *Handler {
	return &Handler{
		config: config,
	}
}

// IsEnabled returns whether demo mode is active.
func (h *Handler) IsEnabled() bool {
	return h.config.Enabled
}

// GetData returns the demo data, generating it on first access.
func (h *Handler) GetData() *DemoData {
	h.dataOnce.Do(func() {
		h.data = GenerateDemoData(context.Background(), h.config)
	})
	return h.data
}

// RefreshData regenerates the demo data.
func (h *Handler) RefreshData() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.data = GenerateDemoData(context.Background(), h.config)
}

// ServeHTTP handles demo API requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.config.Enabled {
		responders.NotFound(w, "demo mode")
		return
	}

	switch r.URL.Path {
	case "/api/demo/data":
		h.handleGetData(w, r)
	case "/api/demo/summary":
		h.handleGetSummary(w, r)
	case "/api/demo/emissions":
		h.handleGetEmissions(w, r)
	case "/api/demo/compliance":
		h.handleGetCompliance(w, r)
	case "/api/demo/trends":
		h.handleGetTrends(w, r)
	case "/api/demo/benchmarks":
		h.handleGetBenchmarks(w, r)
	case "/api/demo/refresh":
		h.handleRefresh(w, r)
	default:
		responders.NotFound(w, "endpoint")
	}
}

func (h *Handler) handleGetData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}

	data := h.GetData()
	responders.SetCacheControl(w, 5*time.Minute, false)
	responders.JSON(w, http.StatusOK, data)
}

func (h *Handler) handleGetSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}

	data := h.GetData()
	responders.SetCacheControl(w, 5*time.Minute, false)
	responders.JSON(w, http.StatusOK, data.SummaryStats())
}

func (h *Handler) handleGetEmissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}

	data := h.GetData()

	// Filter by scope if specified
	scope := r.URL.Query().Get("scope")

	emissions := data.Emissions
	if scope != "" {
		filtered := make([]DemoEmission, 0)
		for _, e := range emissions {
			if (scope == "1" && e.Scope == 1) ||
				(scope == "2" && e.Scope == 2) ||
				(scope == "3" && e.Scope == 3) {
				filtered = append(filtered, e)
			}
		}
		emissions = filtered
	}

	responders.SetCacheControl(w, 5*time.Minute, false)
	responders.JSON(w, http.StatusOK, map[string]interface{}{
		"emissions": emissions,
		"count":     len(emissions),
	})
}

func (h *Handler) handleGetCompliance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}

	data := h.GetData()
	responders.SetCacheControl(w, 5*time.Minute, false)
	responders.JSON(w, http.StatusOK, data.Compliance)
}

func (h *Handler) handleGetTrends(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}

	data := h.GetData()
	responders.SetCacheControl(w, 5*time.Minute, false)
	responders.JSON(w, http.StatusOK, map[string]interface{}{
		"trends": data.TrendData,
		"count":  len(data.TrendData),
	})
}

func (h *Handler) handleGetBenchmarks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}

	data := h.GetData()
	responders.SetCacheControl(w, 5*time.Minute, false)
	responders.JSON(w, http.StatusOK, map[string]interface{}{
		"benchmarks":   data.Benchmarks,
		"organization": data.Organization,
	})
}

func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responders.MethodNotAllowed(w, http.MethodPost)
		return
	}

	h.RefreshData()
	responders.JSON(w, http.StatusOK, map[string]string{
		"status":  "refreshed",
		"message": "Demo data has been regenerated",
	})
}
