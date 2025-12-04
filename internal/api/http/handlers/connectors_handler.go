package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/auth"
	"github.com/example/offgridflow/internal/connectors"
	"github.com/example/offgridflow/internal/ingestion"
)

// Connector represents a data connector status.
type Connector struct {
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	LastRunAt *time.Time `json:"last_run_at,omitempty"`
	LastError string     `json:"last_error,omitempty"`
}

// ConnectorsHandlerConfig configures connector endpoints.
type ConnectorsHandlerConfig struct {
	IngestionSvc   *ingestion.Service
	ConnectorStore connectors.Store
	Orchestrator   *ingestion.Orchestrator
	Scheduler      *ingestion.Scheduler
}

// NewConnectorsHandler exposes list/test/run endpoints for connectors.
func NewConnectorsHandler(cfg ConnectorsHandlerConfig) http.Handler {
	mux := http.NewServeMux()
	resolveOrgID := func(r *http.Request) string {
		if tenant, ok := auth.TenantFromContext(r.Context()); ok && tenant != nil && tenant.ID != "" {
			return tenant.ID
		}
		return ""
	}

	mux.HandleFunc("/api/connectors/list", func(w http.ResponseWriter, r *http.Request) {
		orgID := resolveOrgID(r)
		if orgID == "" {
			responders.Unauthorized(w, "missing_org", "tenant context required")
			return
		}
		if cfg.ConnectorStore != nil {
			if list, err := cfg.ConnectorStore.List(r.Context(), orgID); err == nil {
				responders.JSON(w, http.StatusOK, ensureDefaults(mapStored(list)))
				return
			}
		}

		list := deriveConnectors(cfg.IngestionSvc)
		responders.JSON(w, http.StatusOK, list)
	})

	mux.HandleFunc("/api/connectors/test", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			responders.MethodNotAllowed(w, http.MethodPost)
			return
		}
		name := r.URL.Query().Get("name")
		orgID := resolveOrgID(r)
		if orgID == "" {
			responders.Unauthorized(w, "missing_org", "tenant context required")
			return
		}
		if name == "" {
			name = "unknown"
		}
		status := "ok"
		if cfg.ConnectorStore != nil {
			now := time.Now().UTC()
			_ = cfg.ConnectorStore.SetStatus(r.Context(), name, orgID, "connected", "", &now)
		}
		responders.JSON(w, http.StatusOK, map[string]string{"status": status})
	})

	mux.HandleFunc("/api/connectors/run", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			responders.MethodNotAllowed(w, http.MethodPost)
			return
		}
		if cfg.IngestionSvc == nil {
			responders.Error(w, http.StatusServiceUnavailable, "ingestion_unavailable", "ingestion service not configured")
			return
		}
		ctx := r.Context()
		orgID := resolveOrgID(r)
		if orgID == "" {
			responders.Unauthorized(w, "missing_org", "tenant context required")
			return
		}
		svc := cfg.IngestionSvc
		if orgID != "" && svc != nil {
			clone := *svc
			clone.OrgID = orgID
			svc = &clone
		}
		go func() {
			start := time.Now().UTC()
			if cfg.ConnectorStore != nil {
				_ = cfg.ConnectorStore.SetStatus(ctx, "all", orgID, "running", "", &start)
			}
			orchestrator := cfg.Orchestrator
			if orchestrator == nil {
				orchestrator = &ingestion.Orchestrator{
					Service: svc,
					Logger:  svc.Logger,
				}
			}
			if _, err := orchestrator.Run(ctx); err != nil {
				if cfg.ConnectorStore != nil {
					_ = cfg.ConnectorStore.SetStatus(ctx, "all", orgID, "error", err.Error(), &start)
					_ = cfg.ConnectorStore.LastError(ctx, "all", orgID, err)
				}
				return
			}
			if cfg.ConnectorStore != nil {
				_ = cfg.ConnectorStore.SetStatus(ctx, "all", orgID, "connected", "", &start)
			}
		}()
		responders.JSON(w, http.StatusAccepted, map[string]string{
			"status":    "started",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	mux.HandleFunc("/api/connectors/sync/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			responders.MethodNotAllowed(w, http.MethodPost)
			return
		}
		if cfg.IngestionSvc == nil {
			responders.Error(w, http.StatusServiceUnavailable, "ingestion_unavailable", "ingestion service not configured")
			return
		}

		// Extract provider from URL path
		provider := strings.TrimPrefix(r.URL.Path, "/api/connectors/sync/")
		provider = strings.ToLower(strings.TrimSpace(provider))

		if provider == "" || (provider != "aws" && provider != "azure" && provider != "gcp") {
			responders.Error(w, http.StatusBadRequest, "invalid_provider", "provider must be: aws, azure, or gcp")
			return
		}

		ctx := r.Context()
		orgID := resolveOrgID(r)

		// Create a job ID
		jobID := fmt.Sprintf("sync-%s-%s-%d", provider, orgID, time.Now().Unix())

		// Start sync in background
		go func() {
			start := time.Now().UTC()
			if cfg.ConnectorStore != nil {
				_ = cfg.ConnectorStore.SetStatus(ctx, provider, orgID, "running", "", &start)
			}

			// Find the specific adapter for this provider
			svc := cfg.IngestionSvc
			var targetAdapter ingestion.SourceIngestionAdapter
			for _, adapter := range svc.Adapters {
				adapterType := fmt.Sprintf("%T", adapter)
				if strings.Contains(strings.ToLower(adapterType), provider) {
					targetAdapter = adapter
					break
				}
			}

			if targetAdapter == nil {
				err := fmt.Errorf("provider %s not configured", provider)
				if cfg.ConnectorStore != nil {
					_ = cfg.ConnectorStore.SetStatus(ctx, provider, orgID, "error", err.Error(), &start)
					_ = cfg.ConnectorStore.LastError(ctx, provider, orgID, err)
				}
				return
			}

			// Run the specific adapter
			activities, err := targetAdapter.Ingest(ctx)
			if err != nil {
				if cfg.ConnectorStore != nil {
					_ = cfg.ConnectorStore.SetStatus(ctx, provider, orgID, "error", err.Error(), &start)
					_ = cfg.ConnectorStore.LastError(ctx, provider, orgID, err)
				}
				return
			}

			// Save activities
			if svc.Store != nil && len(activities) > 0 {
				if err := svc.Store.SaveBatch(ctx, activities); err != nil {
					if cfg.ConnectorStore != nil {
						_ = cfg.ConnectorStore.SetStatus(ctx, provider, orgID, "error", err.Error(), &start)
						_ = cfg.ConnectorStore.LastError(ctx, provider, orgID, err)
					}
					return
				}
			}

			if cfg.ConnectorStore != nil {
				_ = cfg.ConnectorStore.SetStatus(ctx, provider, orgID, "connected", "", &start)
			}
		}()

		responders.JSON(w, http.StatusAccepted, map[string]string{
			"job_id":    jobID,
			"provider":  provider,
			"status":    "started",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	mux.HandleFunc("/api/connectors/schedule", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			responders.MethodNotAllowed(w, http.MethodGet)
			return
		}
		if cfg.Scheduler == nil {
			responders.Error(w, http.StatusNotImplemented, "schedule_not_configured", "automated schedule is not configured")
			return
		}
		status := cfg.Scheduler.Status()
		responders.JSON(w, http.StatusOK, status)
	})

	return mux
}

func deriveConnectors(svc *ingestion.Service) []Connector {
	// Default list
	defaults := []string{"aws", "azure", "gcp"}
	status := make(map[string]string)
	for _, name := range defaults {
		status[name] = "disconnected"
	}

	if svc != nil {
		for _, a := range svc.Adapters {
			switch fmt.Sprintf("%T", a) {
			case "*aws.Adapter":
				status["aws"] = "connected"
			case "*azure.Adapter":
				status["azure"] = "connected"
			case "*gcp.Adapter":
				status["gcp"] = "connected"
			default:
				// leave default
			}
		}
	}

	var out []Connector
	for _, name := range defaults {
		out = append(out, Connector{Name: strings.ToUpper(name), Status: status[name]})
	}
	return out
}

// ensureDefaults adds default connectors if missing from store list.
func ensureDefaults(list []Connector) []Connector {
	defaults := []string{"AWS", "AZURE", "GCP"}
	seen := make(map[string]bool, len(list))
	for _, c := range list {
		seen[strings.ToUpper(c.Name)] = true
	}
	for _, name := range defaults {
		if !seen[name] {
			list = append(list, Connector{Name: name, Status: "disconnected"})
		}
	}
	return list
}

// mapStored converts connectors.Connector to handler Connector.
func mapStored(in []connectors.Connector) []Connector {
	out := make([]Connector, 0, len(in))
	for _, c := range in {
		out = append(out, Connector{
			Name:      c.Name,
			Status:    c.Status,
			LastRunAt: c.LastRunAt,
			LastError: c.LastError,
		})
	}
	return out
}
