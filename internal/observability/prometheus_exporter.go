package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// PrometheusExporter provides Prometheus-compatible metrics export
type PrometheusExporter struct {
	batchMetrics *BatchMetrics
	logger       *slog.Logger
	mu           sync.RWMutex
	lastSnapshot *MetricsSnapshot
}

// MetricsSnapshot represents a snapshot of metrics at a point in time
type MetricsSnapshot struct {
	Timestamp                time.Time      `json:"timestamp"`
	BatchesSubmitted         int64          `json:"batches_submitted"`
	BatchesProcessing        int64          `json:"batches_processing"`
	BatchesCompleted         int64          `json:"batches_completed"`
	BatchesFailed            int64          `json:"batches_failed"`
	BatchesCancelled         int64          `json:"batches_cancelled"`
	ActivitiesTotal          int64          `json:"activities_total"`
	ActivitiesSuccess        int64          `json:"activities_success"`
	ActivitiesFailed         int64          `json:"activities_failed"`
	TotalEmissionsKgCO2e     float64        `json:"total_emissions_kg_co2e"`
	WorkersActive            int64          `json:"workers_active"`
	QueueSize                int64          `json:"queue_size"`
	LockAcquisitions         int64          `json:"lock_acquisitions"`
	LockTimeouts             int64          `json:"lock_timeouts"`
	RetryAttempts            int64          `json:"retry_attempts"`
	RetrySuccesses           int64          `json:"retry_successes"`
	StateTransitions         int64          `json:"state_transitions"`
	ErrorsByType             map[string]int `json:"errors_by_type"`
	SuccessRate              float64        `json:"success_rate"`
	AverageProcessingTime    float64        `json:"average_processing_time_ms"`
	AverageEmissionsPerBatch float64        `json:"average_emissions_per_batch"`
}

// NewPrometheusExporter creates a new Prometheus exporter
func NewPrometheusExporter(batchMetrics *BatchMetrics, logger *slog.Logger) *PrometheusExporter {
	return &PrometheusExporter{
		batchMetrics: batchMetrics,
		logger:       logger,
		lastSnapshot: &MetricsSnapshot{
			Timestamp:    time.Now(),
			ErrorsByType: make(map[string]int),
		},
	}
}

// GenerateMetrics generates a snapshot of current metrics
func (pe *PrometheusExporter) GenerateMetrics() *MetricsSnapshot {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	// Get local stats for synchronous access
	submitted, completed, failed, cancelled := pe.batchMetrics.GetLocalStats()

	snapshot := &MetricsSnapshot{
		Timestamp:        time.Now(),
		BatchesSubmitted: submitted,
		BatchesCompleted: completed,
		BatchesFailed:    failed,
		BatchesCancelled: cancelled,
		ErrorsByType:     pe.batchMetrics.GetErrorStats(),
	}

	// Calculate derived metrics
	total := snapshot.ActivitiesSuccess + snapshot.ActivitiesFailed
	if total > 0 {
		snapshot.SuccessRate = float64(snapshot.ActivitiesSuccess) / float64(total) * 100
	}

	pe.lastSnapshot = snapshot
	return snapshot
}

// ExportMetricsHandler returns an HTTP handler for metrics export
func (pe *PrometheusExporter) ExportMetricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		metrics := pe.GenerateMetrics()

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

		if err := json.NewEncoder(w).Encode(metrics); err != nil {
			pe.logger.Error("failed to encode metrics", slog.String("error", err.Error()))
		}
	})
}

// ExportPrometheusTextFormat exports metrics in Prometheus text format
func (pe *PrometheusExporter) ExportPrometheusTextFormat() string {
	snapshot := pe.GenerateMetrics()

	output := "# HELP batch_processor_metrics OffGridFlow Batch Processor Metrics\n"
	output += "# TYPE batch_processor_metrics gauge\n\n"

	output += fmt.Sprintf("# Timestamp: %s\n", snapshot.Timestamp.Format(time.RFC3339))
	output += fmt.Sprintf("batch_submitted_total %d\n", snapshot.BatchesSubmitted)
	output += fmt.Sprintf("batch_processing %d\n", snapshot.BatchesProcessing)
	output += fmt.Sprintf("batch_completed_total %d\n", snapshot.BatchesCompleted)
	output += fmt.Sprintf("batch_failed_total %d\n", snapshot.BatchesFailed)
	output += fmt.Sprintf("batch_cancelled_total %d\n", snapshot.BatchesCancelled)
	output += fmt.Sprintf("activities_total %d\n", snapshot.ActivitiesTotal)
	output += fmt.Sprintf("activities_success_total %d\n", snapshot.ActivitiesSuccess)
	output += fmt.Sprintf("activities_failed_total %d\n", snapshot.ActivitiesFailed)
	output += fmt.Sprintf("emissions_total_kg_co2e %.2f\n", snapshot.TotalEmissionsKgCO2e)
	output += fmt.Sprintf("workers_active %d\n", snapshot.WorkersActive)
	output += fmt.Sprintf("queue_size %d\n", snapshot.QueueSize)
	output += fmt.Sprintf("lock_acquisitions_total %d\n", snapshot.LockAcquisitions)
	output += fmt.Sprintf("lock_timeouts_total %d\n", snapshot.LockTimeouts)
	output += fmt.Sprintf("retry_attempts_total %d\n", snapshot.RetryAttempts)
	output += fmt.Sprintf("retry_successes_total %d\n", snapshot.RetrySuccesses)
	output += fmt.Sprintf("state_transitions_total %d\n", snapshot.StateTransitions)
	output += fmt.Sprintf("success_rate %.2f\n", snapshot.SuccessRate)
	output += fmt.Sprintf("average_processing_time_ms %.2f\n", snapshot.AverageProcessingTime)
	output += fmt.Sprintf("average_emissions_per_batch %.4f\n", snapshot.AverageEmissionsPerBatch)

	output += "\n# Error counts by type\n"
	for errType, count := range snapshot.ErrorsByType {
		output += fmt.Sprintf("error_count_total{type=\"%s\"} %d\n", errType, count)
	}

	return output
}

// MetricsExportHandler returns handlers for different export formats
type MetricsExportHandler struct {
	exporter *PrometheusExporter
	logger   *slog.Logger
}

// NewMetricsExportHandler creates a new metrics export handler
func NewMetricsExportHandler(exporter *PrometheusExporter, logger *slog.Logger) *MetricsExportHandler {
	return &MetricsExportHandler{
		exporter: exporter,
		logger:   logger,
	}
}

// HandleJSON handles JSON format export
func (meh *MetricsExportHandler) HandleJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics := meh.exporter.GenerateMetrics()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		meh.logger.Error("failed to encode metrics", slog.String("error", err.Error()))
		http.Error(w, "failed to encode metrics", http.StatusInternalServerError)
	}
}

// HandlePrometheus handles Prometheus text format export
func (meh *MetricsExportHandler) HandlePrometheus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	output := meh.exporter.ExportPrometheusTextFormat()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	if _, err := w.Write([]byte(output)); err != nil {
		meh.logger.Error("failed to write prometheus metrics", slog.String("error", err.Error()))
	}
}

// RegisterMetricsRoutes registers metrics endpoints
func (meh *MetricsExportHandler) RegisterMetricsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /metrics", meh.HandlePrometheus)
	mux.HandleFunc("GET /metrics/json", meh.HandleJSON)
	mux.HandleFunc("GET /api/v1/metrics", meh.HandleJSON)
	mux.HandleFunc("GET /api/v1/metrics/prometheus", meh.HandlePrometheus)
}

// MetricsCollector provides aggregated metrics collection
type MetricsCollector struct {
	batchMetrics *BatchMetrics
	startTime    time.Time
	logger       *slog.Logger
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(batchMetrics *BatchMetrics, logger *slog.Logger) *MetricsCollector {
	return &MetricsCollector{
		batchMetrics: batchMetrics,
		startTime:    time.Now(),
		logger:       logger,
	}
}

// CollectSystemMetrics collects system-level metrics
func (mc *MetricsCollector) CollectSystemMetrics() map[string]interface{} {
	uptime := time.Since(mc.startTime)

	return map[string]interface{}{
		"uptime_seconds": int64(uptime.Seconds()),
		"errors":         mc.batchMetrics.GetErrorStats(),
	}
}

// CollectAggregatedMetrics collects aggregated metrics for reporting
func (mc *MetricsCollector) CollectAggregatedMetrics(ctx context.Context) map[string]interface{} {
	systemMetrics := mc.CollectSystemMetrics()

	return map[string]interface{}{
		"system":    systemMetrics,
		"timestamp": time.Now().Unix(),
	}
}
