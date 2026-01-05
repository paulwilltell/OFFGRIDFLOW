package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// HealthHandler provides comprehensive health check endpoints for Kubernetes
// Implements both liveness and readiness probes
type HealthHandler struct {
	db          *sql.DB
	redis       *redis.Client
	version     string
	startTime   time.Time
	degradedMu  sync.RWMutex
	degradedMsg string
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(db *sql.DB, redis *redis.Client, version string) *HealthHandler {
	return &HealthHandler{
		db:        db,
		redis:     redis,
		version:   version,
		startTime: time.Now(),
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                 `json:"status"` // "healthy", "degraded", "unhealthy"
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]CheckResult `json:"checks"`
	System    SystemInfo             `json:"system,omitempty"`
	Message   string                 `json:"message,omitempty"`
}

// CheckResult represents individual health check result
type CheckResult struct {
	Status       string      `json:"status"` // "pass", "warn", "fail"
	ResponseTime string      `json:"response_time,omitempty"`
	Message      string      `json:"message,omitempty"`
	LastChecked  time.Time   `json:"last_checked"`
	Details      interface{} `json:"details,omitempty"`
}

// SystemInfo provides system-level metrics
type SystemInfo struct {
	GoVersion     string `json:"go_version"`
	NumGoroutines int    `json:"num_goroutines"`
	MemoryAllocMB uint64 `json:"memory_alloc_mb"`
	MemorySysMB   uint64 `json:"memory_sys_mb"`
	NumCPU        int    `json:"num_cpu"`
}

// MarkDegraded marks the service as degraded with a reason
func (h *HealthHandler) MarkDegraded(message string) {
	h.degradedMu.Lock()
	defer h.degradedMu.Unlock()
	h.degradedMsg = message
}

// ClearDegraded clears the degraded status
func (h *HealthHandler) ClearDegraded() {
	h.degradedMu.Lock()
	defer h.degradedMu.Unlock()
	h.degradedMsg = ""
}

// LivenessProbe checks if the application is alive
// Returns 200 if the app is running, 503 if it should be restarted
// This is a lightweight check - just confirms the process is responsive
func (h *HealthHandler) LivenessProbe(w http.ResponseWriter, r *http.Request) {
	// Liveness check is simple - if we can respond, we're alive
	response := HealthResponse{
		Status:    "healthy",
		Version:   h.version,
		Uptime:    time.Since(h.startTime).String(),
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ReadinessProbe checks if the application is ready to serve traffic
// Returns 200 if ready, 503 if not ready (dependencies unavailable)
// Kubernetes will not route traffic to pods that fail readiness checks
func (h *HealthHandler) ReadinessProbe(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	checks := make(map[string]CheckResult)
	overallStatus := "healthy"

	// Check database connection
	dbResult := h.checkDatabase(ctx)
	checks["database"] = dbResult
	if dbResult.Status == "fail" {
		overallStatus = "unhealthy"
	}

	// Check Redis connection
	redisResult := h.checkRedis(ctx)
	checks["redis"] = redisResult
	if redisResult.Status == "fail" {
		overallStatus = "degraded" // Redis failure is degraded, not unhealthy
	}

	// Check for manually set degraded status
	h.degradedMu.RLock()
	degradedMsg := h.degradedMsg
	h.degradedMu.RUnlock()

	if degradedMsg != "" {
		overallStatus = "degraded"
	}

	response := HealthResponse{
		Status:    overallStatus,
		Version:   h.version,
		Uptime:    time.Since(h.startTime).String(),
		Timestamp: time.Now(),
		Checks:    checks,
		Message:   degradedMsg,
	}

	// Return 503 if unhealthy (Kubernetes will not route traffic)
	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// HealthCheck provides detailed health information (for monitoring/debugging)
// This is not used by Kubernetes but useful for ops teams
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	checks := make(map[string]CheckResult)
	overallStatus := "healthy"

	// Run all health checks
	dbResult := h.checkDatabase(ctx)
	checks["database"] = dbResult
	if dbResult.Status == "fail" {
		overallStatus = "unhealthy"
	} else if dbResult.Status == "warn" {
		overallStatus = "degraded"
	}

	redisResult := h.checkRedis(ctx)
	checks["redis"] = redisResult
	if redisResult.Status == "fail" && overallStatus != "unhealthy" {
		overallStatus = "degraded"
	}

	// Check system resources
	systemInfo := h.getSystemInfo()

	// Check for manually set degraded status
	h.degradedMu.RLock()
	degradedMsg := h.degradedMsg
	h.degradedMu.RUnlock()

	if degradedMsg != "" && overallStatus == "healthy" {
		overallStatus = "degraded"
	}

	response := HealthResponse{
		Status:    overallStatus,
		Version:   h.version,
		Uptime:    time.Since(h.startTime).String(),
		Timestamp: time.Now(),
		Checks:    checks,
		System:    systemInfo,
		Message:   degradedMsg,
	}

	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if overallStatus == "degraded" {
		statusCode = http.StatusOK // Still return 200 for degraded
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// checkDatabase verifies database connectivity and performance
func (h *HealthHandler) checkDatabase(ctx context.Context) CheckResult {
	if h.db == nil {
		return CheckResult{
			Status:      "fail",
			Message:     "database not configured",
			LastChecked: time.Now(),
		}
	}

	start := time.Now()

	// Ping database
	if err := h.db.PingContext(ctx); err != nil {
		return CheckResult{
			Status:       "fail",
			Message:      "database ping failed: " + err.Error(),
			ResponseTime: time.Since(start).String(),
			LastChecked:  time.Now(),
		}
	}

	responseTime := time.Since(start)

	// Get connection pool stats
	stats := h.db.Stats()
	details := map[string]interface{}{
		"open_connections":    stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}

	// Warn if response time is high
	status := "pass"
	message := "database is healthy"
	if responseTime > 100*time.Millisecond {
		status = "warn"
		message = "database response time is high"
	}

	// Warn if connection pool is saturated
	if stats.InUse > (stats.OpenConnections * 80 / 100) {
		status = "warn"
		message = "database connection pool nearing saturation"
	}

	return CheckResult{
		Status:       status,
		Message:      message,
		ResponseTime: responseTime.String(),
		LastChecked:  time.Now(),
		Details:      details,
	}
}

// checkRedis verifies Redis connectivity
func (h *HealthHandler) checkRedis(ctx context.Context) CheckResult {
	if h.redis == nil {
		return CheckResult{
			Status:      "warn",
			Message:     "redis not configured (using in-memory fallback)",
			LastChecked: time.Now(),
		}
	}

	start := time.Now()

	// Ping Redis
	if err := h.redis.Ping(ctx).Err(); err != nil {
		return CheckResult{
			Status:       "fail",
			Message:      "redis ping failed: " + err.Error(),
			ResponseTime: time.Since(start).String(),
			LastChecked:  time.Now(),
		}
	}

	responseTime := time.Since(start)

	// Get Redis info
	info, err := h.redis.Info(ctx, "stats").Result()
	if err != nil {
		return CheckResult{
			Status:       "warn",
			Message:      "redis info failed: " + err.Error(),
			ResponseTime: responseTime.String(),
			LastChecked:  time.Now(),
		}
	}

	details := map[string]interface{}{
		"info": info,
	}

	status := "pass"
	message := "redis is healthy"
	if responseTime > 50*time.Millisecond {
		status = "warn"
		message = "redis response time is high"
	}

	return CheckResult{
		Status:       status,
		Message:      message,
		ResponseTime: responseTime.String(),
		LastChecked:  time.Now(),
		Details:      details,
	}
}

// getSystemInfo collects system-level metrics
func (h *HealthHandler) getSystemInfo() SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemInfo{
		GoVersion:     runtime.Version(),
		NumGoroutines: runtime.NumGoroutine(),
		MemoryAllocMB: m.Alloc / 1024 / 1024,
		MemorySysMB:   m.Sys / 1024 / 1024,
		NumCPU:        runtime.NumCPU(),
	}
}

// RegisterRoutes registers all health check routes
func (h *HealthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health/live", h.LivenessProbe)
	mux.HandleFunc("/health/ready", h.ReadinessProbe)
	mux.HandleFunc("/health", h.HealthCheck)

	// Legacy endpoints for backwards compatibility
	mux.HandleFunc("/healthz", h.LivenessProbe)
	mux.HandleFunc("/readyz", h.ReadinessProbe)
}
