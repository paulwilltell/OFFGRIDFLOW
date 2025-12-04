package observability

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

// TestBatchMetricsCreation tests batch metrics creation
func TestBatchMetricsCreation(t *testing.T) {
	ctx := context.Background()
	metrics, err := NewBatchMetrics(ctx, "test-service")
	
	if err != nil {
		t.Fatalf("Failed to create batch metrics: %v", err)
	}
	
	if metrics == nil {
		t.Error("Batch metrics should not be nil")
	}
	
	if metrics.BatchesSubmitted == nil {
		t.Error("BatchesSubmitted metric should be initialized")
	}
}

// TestBatchSubmissionRecording tests recording batch submission
func TestBatchSubmissionRecording(t *testing.T) {
	ctx := context.Background()
	metrics, _ := NewBatchMetrics(ctx, "test-service")
	
	// Record a batch submission
	metrics.RecordBatchSubmission(ctx, "batch_1", 10)
	metrics.RecordSubmitDuration(ctx, 100*time.Millisecond, true)
	
	// No error should occur
	t.Log("Batch submission recorded successfully")
}

// TestBatchCompletionRecording tests recording batch completion
func TestBatchCompletionRecording(t *testing.T) {
	ctx := context.Background()
	metrics, _ := NewBatchMetrics(ctx, "test-service")
	
	// Record batch completion
	metrics.RecordBatchCompletion(ctx, "batch_1", 2*time.Second, 8, 2, 500.5)
	
	t.Log("Batch completion recorded successfully")
}

// TestActivityProcessingRecording tests recording activity processing
func TestActivityProcessingRecording(t *testing.T) {
	ctx := context.Background()
	metrics, _ := NewBatchMetrics(ctx, "test-service")
	
	// Record successful activity
	metrics.RecordActivityProcessing(ctx, "batch_1", "act_1", true, 100*time.Millisecond)
	
	// Record failed activity
	metrics.RecordActivityProcessing(ctx, "batch_1", "act_2", false, 50*time.Millisecond)
	
	t.Log("Activity processing recorded successfully")
}

// TestWorkerStateRecording tests recording worker state changes
func TestWorkerStateRecording(t *testing.T) {
	ctx := context.Background()
	metrics, _ := NewBatchMetrics(ctx, "test-service")
	
	// Worker becomes active
	metrics.RecordWorkerStateChange(ctx, "worker_1", true)
	
	// Worker becomes inactive
	metrics.RecordWorkerStateChange(ctx, "worker_1", false)
	
	t.Log("Worker state changes recorded successfully")
}

// TestLockAcquisitionRecording tests recording lock acquisitions
func TestLockAcquisitionRecording(t *testing.T) {
	ctx := context.Background()
	metrics, _ := NewBatchMetrics(ctx, "test-service")
	
	// Successful lock acquisition
	metrics.RecordLockAcquisition(ctx, "batch_1", "worker_1", 10*time.Millisecond, true)
	
	// Failed lock acquisition (timeout)
	metrics.RecordLockAcquisition(ctx, "batch_2", "worker_2", 100*time.Millisecond, false)
	
	t.Log("Lock acquisitions recorded successfully")
}

// TestErrorRecording tests error tracking
func TestErrorRecording(t *testing.T) {
	ctx := context.Background()
	metrics, _ := NewBatchMetrics(ctx, "test-service")
	
	// Record some errors
	metrics.RecordError("database_connection_error")
	metrics.RecordError("timeout_error")
	metrics.RecordError("database_connection_error")
	
	// Get error stats
	stats := metrics.GetErrorStats()
	
	if stats["database_connection_error"] != 2 {
		t.Errorf("Expected 2 database_connection_error, got %d", stats["database_connection_error"])
	}
	
	if stats["timeout_error"] != 1 {
		t.Errorf("Expected 1 timeout_error, got %d", stats["timeout_error"])
	}
}

// TestBatchTracerCreation tests batch tracer creation
func TestBatchTracerCreation(t *testing.T) {
	logger := slog.Default()
	tracer := NewBatchTracer("test-service", logger)
	
	if tracer == nil {
		t.Error("Batch tracer should not be nil")
	}
	
	if tracer.tracer == nil {
		t.Error("Tracer should be initialized")
	}
}

// TestBatchTracingSpans tests tracing span creation
func TestBatchTracingSpans(t *testing.T) {
	logger := slog.Default()
	tracer := NewBatchTracer("test-service", logger)
	ctx := context.Background()
	
	// Test span creation
	ctx, span := tracer.StartBatchSubmissionSpan(ctx, "batch_1", "org_1", "ws_1", 5)
	defer span.End()
	
	if span == nil {
		t.Error("Span should not be nil")
	}
	
	// Record completion
	tracer.RecordBatchSubmissionComplete(span, "batch_1", 100*time.Millisecond)
}

// TestPrometheusExporterCreation tests Prometheus exporter creation
func TestPrometheusExporterCreation(t *testing.T) {
	ctx := context.Background()
	metrics, _ := NewBatchMetrics(ctx, "test-service")
	logger := slog.Default()
	
	exporter := NewPrometheusExporter(metrics, logger)
	
	if exporter == nil {
		t.Error("Prometheus exporter should not be nil")
	}
}

// TestMetricsSnapshotGeneration tests metrics snapshot generation
func TestMetricsSnapshotGeneration(t *testing.T) {
	ctx := context.Background()
	metrics, _ := NewBatchMetrics(ctx, "test-service")
	logger := slog.Default()
	exporter := NewPrometheusExporter(metrics, logger)
	
	// Record some data
	metrics.RecordBatchSubmission(ctx, "batch_1", 10)
	metrics.RecordBatchCompletion(ctx, "batch_1", 1*time.Second, 8, 2, 500.0)
	
	// Generate snapshot
	snapshot := exporter.GenerateMetrics()
	
	if snapshot == nil {
		t.Error("Metrics snapshot should not be nil")
	}
	
	if snapshot.BatchesSubmitted != 1 {
		t.Errorf("Expected 1 submitted batch, got %d", snapshot.BatchesSubmitted)
	}
}

// TestPrometheusTextFormatExport tests Prometheus text format export
func TestPrometheusTextFormatExport(t *testing.T) {
	ctx := context.Background()
	metrics, _ := NewBatchMetrics(ctx, "test-service")
	logger := slog.Default()
	exporter := NewPrometheusExporter(metrics, logger)
	
	output := exporter.ExportPrometheusTextFormat()
	
	if output == "" {
		t.Error("Prometheus text format output should not be empty")
	}
	
	// Check for expected content
	if len(output) < 100 {
		t.Error("Prometheus text format output seems too short")
	}
}

// TestHealthCheckerCreation tests health checker creation
func TestHealthCheckerCreation(t *testing.T) {
	logger := slog.Default()
	checker := NewHealthChecker(logger)
	
	if checker == nil {
		t.Error("Health checker should not be nil")
	}
	
	if len(checker.checks) != 0 {
		t.Error("Health checker should have no checks initially")
	}
}

// TestHealthCheckRegistration tests registering health checks
func TestHealthCheckRegistration(t *testing.T) {
	logger := slog.Default()
	checker := NewHealthChecker(logger)
	
	// Register a check
	checker.RegisterCheck("database", func(ctx context.Context) CheckResult {
		return CheckResult{
			Name:   "database",
			Status: "healthy",
		}
	})
	
	// Check health
	ctx := context.Background()
	result := checker.CheckHealth(ctx)
	
	if result == nil {
		t.Error("Health check result should not be nil")
	}
	
	if _, ok := result.Checks["database"]; !ok {
		t.Error("Database check should be present")
	}
}

// TestHealthCheckResult tests health check result
func TestHealthCheckResult(t *testing.T) {
	logger := slog.Default()
	checker := NewHealthChecker(logger)
	
	// Register checks
	checker.RegisterCheck("check1", func(ctx context.Context) CheckResult {
		return CheckResult{Name: "check1", Status: "healthy"}
	})
	
	checker.RegisterCheck("check2", func(ctx context.Context) CheckResult {
		return CheckResult{Name: "check2", Status: "healthy"}
	})
	
	ctx := context.Background()
	result := checker.CheckHealth(ctx)
	
	if result.OverallStatus != "healthy" {
		t.Errorf("Expected healthy status, got %s", result.OverallStatus)
	}
	
	if len(result.Checks) != 2 {
		t.Errorf("Expected 2 checks, got %d", len(result.Checks))
	}
}

// TestDegradedHealthStatus tests degraded health status
func TestDegradedHealthStatus(t *testing.T) {
	logger := slog.Default()
	checker := NewHealthChecker(logger)
	
	// Register a degraded check
	checker.RegisterCheck("database", func(ctx context.Context) CheckResult {
		return CheckResult{
			Name:    "database",
			Status:  "degraded",
			Message: "slow response",
		}
	})
	
	ctx := context.Background()
	result := checker.CheckHealth(ctx)
	
	if result.OverallStatus != "degraded" {
		t.Errorf("Expected degraded status, got %s", result.OverallStatus)
	}
}

// TestMetricsCollectorCreation tests metrics collector creation
func TestMetricsCollectorCreation(t *testing.T) {
	ctx := context.Background()
	metrics, _ := NewBatchMetrics(ctx, "test-service")
	logger := slog.Default()
	
	collector := NewMetricsCollector(metrics, logger)
	
	if collector == nil {
		t.Error("Metrics collector should not be nil")
	}
}

// TestMetricsCollectorAggregation tests metrics aggregation
func TestMetricsCollectorAggregation(t *testing.T) {
	ctx := context.Background()
	metrics, _ := NewBatchMetrics(ctx, "test-service")
	logger := slog.Default()
	
	collector := NewMetricsCollector(metrics, logger)
	
	// Record some metrics
	metrics.RecordBatchSubmission(ctx, "batch_1", 10)
	
	// Collect metrics
	aggregated := collector.CollectAggregatedMetrics(ctx)
	
	if aggregated == nil {
		t.Error("Aggregated metrics should not be nil")
	}
	
	if aggregated["system"] == nil {
		t.Error("System metrics should be present")
	}
}

// TestBatchMetricsRecordingConcurrency tests concurrent metric recording
func TestBatchMetricsRecordingConcurrency(t *testing.T) {
	ctx := context.Background()
	metrics, _ := NewBatchMetrics(ctx, "test-service")
	
	// Record metrics concurrently
	done := make(chan bool)
	
	for i := 0; i < 10; i++ {
		go func(index int) {
			metrics.RecordBatchSubmission(ctx, "batch_"+string(rune(index)), 5)
			metrics.RecordActivityProcessing(ctx, "batch_"+string(rune(index)), "act_1", true, 100*time.Millisecond)
			done <- true
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	t.Log("Concurrent metric recording completed successfully")
}

// BenchmarkBatchSubmissionRecording benchmarks batch submission recording
func BenchmarkBatchSubmissionRecording(b *testing.B) {
	ctx := context.Background()
	metrics, _ := NewBatchMetrics(ctx, "test-service")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.RecordBatchSubmission(ctx, "batch_1", 10)
	}
}

// BenchmarkActivityProcessing benchmarks activity processing recording
func BenchmarkActivityProcessing(b *testing.B) {
	ctx := context.Background()
	metrics, _ := NewBatchMetrics(ctx, "test-service")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.RecordActivityProcessing(ctx, "batch_1", "act_1", true, 100*time.Millisecond)
	}
}

// BenchmarkHealthCheck benchmarks health check execution
func BenchmarkHealthCheck(b *testing.B) {
	logger := slog.Default()
	checker := NewHealthChecker(logger)
	
	checker.RegisterCheck("check1", func(ctx context.Context) CheckResult {
		return CheckResult{Name: "check1", Status: "healthy"}
	})
	
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.CheckHealth(ctx)
	}
}
