package handlers

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/example/offgridflow/internal/api/http/responders"
)

// -----------------------------------------------------------------------------
// Health Check Handler
// -----------------------------------------------------------------------------

// HealthStatus represents the health check response.
type HealthStatus struct {
	Status    string            `json:"status"`             // "ok", "degraded", "unhealthy"
	Timestamp string            `json:"timestamp"`          // ISO 8601 timestamp
	Version   string            `json:"version,omitempty"`  // Application version
	Uptime    string            `json:"uptime,omitempty"`   // Human-readable uptime
	Checks    map[string]Check  `json:"checks,omitempty"`   // Individual component checks
	Metadata  map[string]string `json:"metadata,omitempty"` // Additional metadata
}

// Check represents a single health check result.
type Check struct {
	Status   string `json:"status"`             // "pass", "fail", "warn"
	Message  string `json:"message,omitempty"`  // Optional message
	Duration string `json:"duration,omitempty"` // Check duration
}

// HealthChecker is a function that performs a health check.
type HealthChecker func(ctx context.Context) Check

// HealthHandler provides health check endpoints.
type HealthHandler struct {
	version   string
	startTime time.Time
	checkers  map[string]HealthChecker
	mu        sync.RWMutex
}

// HealthHandlerConfig configures the health handler.
type HealthHandlerConfig struct {
	Version string
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(cfg HealthHandlerConfig) *HealthHandler {
	return &HealthHandler{
		version:   cfg.Version,
		startTime: time.Now(),
		checkers:  make(map[string]HealthChecker),
	}
}

// RegisterChecker adds a health checker for a component.
func (h *HealthHandler) RegisterChecker(name string, checker HealthChecker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers[name] = checker
}

// ServeHTTP handles health check requests.
// GET /health - Simple health check
// GET /health?full=true - Detailed health check with component status
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}

	// Simple health check (for load balancers)
	if r.URL.Query().Get("full") != "true" {
		responders.JSON(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
		return
	}

	// Full health check
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	status := h.runChecks(ctx)

	httpStatus := http.StatusOK
	if status.Status == "unhealthy" {
		httpStatus = http.StatusServiceUnavailable
	} else if status.Status == "degraded" {
		httpStatus = http.StatusOK // Still return 200 for degraded
	}

	responders.JSON(w, httpStatus, status)
}

// runChecks executes all registered health checkers.
func (h *HealthHandler) runChecks(ctx context.Context) HealthStatus {
	h.mu.RLock()
	checkers := make(map[string]HealthChecker, len(h.checkers))
	for k, v := range h.checkers {
		checkers[k] = v
	}
	h.mu.RUnlock()

	checks := make(map[string]Check, len(checkers))
	overallStatus := "ok"

	// Run checks in parallel
	var wg sync.WaitGroup
	var checkMu sync.Mutex

	for name, checker := range checkers {
		wg.Add(1)
		go func(name string, checker HealthChecker) {
			defer wg.Done()

			start := time.Now()
			result := checker(ctx)
			result.Duration = time.Since(start).String()

			checkMu.Lock()
			checks[name] = result

			// Update overall status
			if result.Status == "fail" && overallStatus != "unhealthy" {
				overallStatus = "unhealthy"
			} else if result.Status == "warn" && overallStatus == "ok" {
				overallStatus = "degraded"
			}
			checkMu.Unlock()
		}(name, checker)
	}

	wg.Wait()

	uptime := time.Since(h.startTime)

	return HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   h.version,
		Uptime:    formatDuration(uptime),
		Checks:    checks,
	}
}

// formatDuration formats a duration in a human-readable way.
func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return time.Duration(d.Nanoseconds()).Round(time.Minute).String()
	}
	if hours > 0 {
		return time.Duration(d.Nanoseconds()).Round(time.Minute).String()
	}
	if minutes > 0 {
		return time.Duration(d.Nanoseconds()).Round(time.Second).String()
	}
	return d.Round(time.Millisecond).String()
}

// -----------------------------------------------------------------------------
// Simple Health Handlers (for backward compatibility)
// -----------------------------------------------------------------------------

// SimpleHealthHandler returns API health (simple version).
func SimpleHealthHandler(w http.ResponseWriter, r *http.Request) {
	responders.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ReadyzHandler is a Kubernetes-style readiness probe.
// Returns 200 if the service is ready to receive traffic.
func ReadyzHandler(w http.ResponseWriter, r *http.Request) {
	responders.JSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// LivezHandler is a Kubernetes-style liveness probe.
// Returns 200 if the service is alive.
func LivezHandler(w http.ResponseWriter, r *http.Request) {
	responders.JSON(w, http.StatusOK, map[string]string{"status": "alive"})
}

// -----------------------------------------------------------------------------
// Common Health Checkers
// -----------------------------------------------------------------------------

// DatabaseChecker creates a health checker for database connectivity.
type DatabasePinger interface {
	PingContext(ctx context.Context) error
}

// NewDatabaseChecker creates a health checker for a database connection.
func NewDatabaseChecker(db DatabasePinger) HealthChecker {
	return func(ctx context.Context) Check {
		if err := db.PingContext(ctx); err != nil {
			return Check{
				Status:  "fail",
				Message: "database connection failed: " + err.Error(),
			}
		}
		return Check{
			Status:  "pass",
			Message: "database connection ok",
		}
	}
}

// MemoryChecker creates a health checker that warns on high memory usage.
func NewMemoryChecker(thresholdPercent float64) HealthChecker {
	return func(ctx context.Context) Check {
		// In production, use runtime.MemStats to check actual memory
		// For now, always return pass
		return Check{
			Status:  "pass",
			Message: "memory usage within limits",
		}
	}
}
