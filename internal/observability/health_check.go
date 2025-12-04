package observability

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// HealthStatus represents the health status of the system
type HealthStatus int

const (
	HealthStatusHealthy HealthStatus = iota
	HealthStatusDegraded
	HealthStatusUnhealthy
)

func (hs HealthStatus) String() string {
	switch hs {
	case HealthStatusHealthy:
		return "healthy"
	case HealthStatusDegraded:
		return "degraded"
	case HealthStatusUnhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Status           string                 `json:"status"`
	Timestamp        time.Time              `json:"timestamp"`
	SystemUptime     int64                  `json:"system_uptime_seconds"`
	Checks           map[string]CheckResult `json:"checks"`
	OverallStatus    string                 `json:"overall_status"`
	Message          string                 `json:"message,omitempty"`
}

// CheckResult represents a single health check result
type CheckResult struct {
	Name    string      `json:"name"`
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

// HealthChecker manages health checks for the system
type HealthChecker struct {
	checks    map[string]func(context.Context) CheckResult
	checksMu  sync.RWMutex
	logger    *slog.Logger
	startTime time.Time
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(logger *slog.Logger) *HealthChecker {
	return &HealthChecker{
		checks:    make(map[string]func(context.Context) CheckResult),
		logger:    logger,
		startTime: time.Now(),
	}
}

// RegisterCheck registers a health check
func (hc *HealthChecker) RegisterCheck(name string, check func(context.Context) CheckResult) {
	hc.checksMu.Lock()
	defer hc.checksMu.Unlock()
	hc.checks[name] = check
}

// CheckHealth performs all registered health checks
func (hc *HealthChecker) CheckHealth(ctx context.Context) *HealthCheckResult {
	hc.checksMu.RLock()
	checks := make(map[string]func(context.Context) CheckResult)
	for name, check := range hc.checks {
		checks[name] = check
	}
	hc.checksMu.RUnlock()

	result := &HealthCheckResult{
		Timestamp:     time.Now(),
		SystemUptime:  int64(time.Since(hc.startTime).Seconds()),
		Checks:        make(map[string]CheckResult),
		OverallStatus: "healthy",
	}

	// Run all checks
	for name, check := range checks {
		result.Checks[name] = check(ctx)

		// Update overall status
		if result.Checks[name].Status != "healthy" {
			if result.OverallStatus == "healthy" {
				result.OverallStatus = result.Checks[name].Status
			}
		}
	}

	result.Status = result.OverallStatus

	return result
}

// HealthCheckHandler provides HTTP handlers for health checks
type HealthCheckHandler struct {
	checker *HealthChecker
	logger  *slog.Logger
}

// NewHealthCheckHandler creates a new health check handler
func NewHealthCheckHandler(checker *HealthChecker, logger *slog.Logger) *HealthCheckHandler {
	return &HealthCheckHandler{
		checker: checker,
		logger:  logger,
	}
}

// HandleHealth handles basic health check requests
func (hch *HealthCheckHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	result := hch.checker.CheckHealth(ctx)

	statusCode := http.StatusOK
	if result.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(result); err != nil {
		hch.logger.Error("failed to encode health check result", slog.String("error", err.Error()))
	}
}

// HandleLiveness handles liveness checks (is service running?)
func (hch *HealthCheckHandler) HandleLiveness(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// HandleReadiness handles readiness checks (is service ready to serve?)
func (hch *HealthCheckHandler) HandleReadiness(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	result := hch.checker.CheckHealth(ctx)

	statusCode := http.StatusOK
	if result.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(result); err != nil {
		hch.logger.Error("failed to encode readiness check result", slog.String("error", err.Error()))
	}
}

// RegisterHealthRoutes registers health check endpoints
func (hch *HealthCheckHandler) RegisterHealthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", hch.HandleHealth)
	mux.HandleFunc("GET /health/live", hch.HandleLiveness)
	mux.HandleFunc("GET /health/ready", hch.HandleReadiness)
	mux.HandleFunc("GET /api/v1/health", hch.HandleHealth)
	mux.HandleFunc("GET /api/v1/health/live", hch.HandleLiveness)
	mux.HandleFunc("GET /api/v1/health/ready", hch.HandleReadiness)
}

// SystemStatus provides detailed system status information
type SystemStatus struct {
	ServiceName     string                 `json:"service_name"`
	ServiceVersion  string                 `json:"service_version"`
	Environment     string                 `json:"environment"`
	Status          string                 `json:"status"`
	Uptime          int64                  `json:"uptime_seconds"`
	Timestamp       time.Time              `json:"timestamp"`
	Dependencies    map[string]interface{} `json:"dependencies"`
	Metrics         map[string]interface{} `json:"metrics,omitempty"`
}

// StatusHandler provides comprehensive status information
type StatusHandler struct {
	serviceInfo map[string]string
	checker     *HealthChecker
	metrics     *PrometheusExporter
	logger      *slog.Logger
}

// NewStatusHandler creates a new status handler
func NewStatusHandler(checker *HealthChecker, exporter *PrometheusExporter, logger *slog.Logger) *StatusHandler {
	return &StatusHandler{
		serviceInfo: make(map[string]string),
		checker:     checker,
		metrics:     exporter,
		logger:      logger,
	}
}

// SetServiceInfo sets service information
func (sh *StatusHandler) SetServiceInfo(key, value string) {
	sh.serviceInfo[key] = value
}

// HandleStatus handles status requests
func (sh *StatusHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	healthResult := sh.checker.CheckHealth(ctx)

	status := &SystemStatus{
		ServiceName:    sh.serviceInfo["name"],
		ServiceVersion: sh.serviceInfo["version"],
		Environment:    sh.serviceInfo["environment"],
		Status:         healthResult.Status,
		Uptime:         healthResult.SystemUptime,
		Timestamp:      time.Now(),
		Dependencies:   make(map[string]interface{}),
		Metrics:        make(map[string]interface{}),
	}

	// Add check details as dependencies
	for name, check := range healthResult.Checks {
		status.Dependencies[name] = map[string]interface{}{
			"status":  check.Status,
			"message": check.Message,
		}
	}

	// Add metrics
	if sh.metrics != nil {
		metricsSnapshot := sh.metrics.GenerateMetrics()
		status.Metrics = map[string]interface{}{
			"batches_submitted":     metricsSnapshot.BatchesSubmitted,
			"batches_processing":    metricsSnapshot.BatchesProcessing,
			"batches_completed":     metricsSnapshot.BatchesCompleted,
			"batches_failed":        metricsSnapshot.BatchesFailed,
			"workers_active":        metricsSnapshot.WorkersActive,
			"queue_size":            metricsSnapshot.QueueSize,
			"success_rate":          metricsSnapshot.SuccessRate,
			"total_emissions":       metricsSnapshot.TotalEmissionsKgCO2e,
		}
	}

	statusCode := http.StatusOK
	if status.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(status); err != nil {
		sh.logger.Error("failed to encode status", slog.String("error", err.Error()))
	}
}

// RegisterStatusRoutes registers status endpoints
func (sh *StatusHandler) RegisterStatusRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /status", sh.HandleStatus)
	mux.HandleFunc("GET /api/v1/status", sh.HandleStatus)
}
