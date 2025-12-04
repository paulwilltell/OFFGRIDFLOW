package observability

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// BatchMetrics tracks metrics specific to batch processing
type BatchMetrics struct {
	meter metric.Meter
	mu    sync.RWMutex

	// Batch submission metrics
	BatchesSubmitted metric.Int64Counter
	SubmitDuration   metric.Float64Histogram

	// Batch processing metrics
	BatchesProcessing  metric.Int64UpDownCounter
	BatchesCompleted   metric.Int64Counter
	BatchesFailed      metric.Int64Counter
	BatchesCancelled   metric.Int64Counter
	ProcessingDuration metric.Float64Histogram

	// Activity metrics
	ActivitiesTotal   metric.Int64Counter
	ActivitiesSuccess metric.Int64Counter
	ActivitiesFailed  metric.Int64Counter
	ActivityDuration  metric.Float64Histogram

	// Emissions metrics
	TotalEmissionsKgCO2  metric.Float64Counter
	AvgEmissionsPerBatch metric.Float64Histogram

	// Worker pool metrics
	WorkersActive  metric.Int64UpDownCounter
	WorkerIdleTime metric.Float64Histogram
	WorkerTaskTime metric.Float64Histogram

	// Lock/Contention metrics
	LockAcquisitions metric.Int64Counter
	LockWaitTime     metric.Float64Histogram
	LockTimeouts     metric.Int64Counter

	// Queue metrics
	QueueSize     metric.Int64UpDownCounter
	QueueWaitTime metric.Float64Histogram

	// Error metrics
	BatchErrors    metric.Int64Counter
	RetryAttempts  metric.Int64Counter
	RetrySuccesses metric.Int64Counter

	// State transition metrics
	StateTransitions metric.Int64Counter

	// Error details (structured)
	errorRegistry map[string]int
	errorMutex    sync.RWMutex

	// Local counters for testing/synchronous access
	localSubmitted int64
	localCompleted int64
	localFailed    int64
	localCancelled int64
}

// NewBatchMetrics creates and initializes batch-specific metrics
func NewBatchMetrics(ctx context.Context, serviceName string) (*BatchMetrics, error) {
	meter := otel.Meter(fmt.Sprintf("%s-batch-processor", serviceName))

	bm := &BatchMetrics{
		meter:         meter,
		errorRegistry: make(map[string]int),
	}

	var err error

	// Batch submission metrics
	if bm.BatchesSubmitted, err = meter.Int64Counter("batch.submitted.total",
		metric.WithDescription("Total number of batches submitted"),
		metric.WithUnit("1"),
	); err != nil {
		return nil, fmt.Errorf("failed to create BatchesSubmitted metric: %w", err)
	}

	if bm.SubmitDuration, err = meter.Float64Histogram("batch.submit.duration",
		metric.WithDescription("Duration of batch submission operations"),
		metric.WithUnit("ms"),
	); err != nil {
		return nil, fmt.Errorf("failed to create SubmitDuration metric: %w", err)
	}

	// Batch processing metrics
	if bm.BatchesProcessing, err = meter.Int64UpDownCounter("batch.processing",
		metric.WithDescription("Number of batches currently processing"),
		metric.WithUnit("1"),
	); err != nil {
		return nil, fmt.Errorf("failed to create BatchesProcessing metric: %w", err)
	}

	if bm.BatchesCompleted, err = meter.Int64Counter("batch.completed.total",
		metric.WithDescription("Total number of batches completed successfully"),
		metric.WithUnit("1"),
	); err != nil {
		return nil, fmt.Errorf("failed to create BatchesCompleted metric: %w", err)
	}

	if bm.BatchesFailed, err = meter.Int64Counter("batch.failed.total",
		metric.WithDescription("Total number of batches that failed"),
		metric.WithUnit("1"),
	); err != nil {
		return nil, fmt.Errorf("failed to create BatchesFailed metric: %w", err)
	}

	if bm.BatchesCancelled, err = meter.Int64Counter("batch.cancelled.total",
		metric.WithDescription("Total number of batches cancelled"),
		metric.WithUnit("1"),
	); err != nil {
		return nil, fmt.Errorf("failed to create BatchesCancelled metric: %w", err)
	}

	if bm.ProcessingDuration, err = meter.Float64Histogram("batch.processing.duration",
		metric.WithDescription("Time taken to process a batch"),
		metric.WithUnit("ms"),
	); err != nil {
		return nil, fmt.Errorf("failed to create ProcessingDuration metric: %w", err)
	}

	// Activity metrics
	if bm.ActivitiesTotal, err = meter.Int64Counter("activity.total",
		metric.WithDescription("Total activities processed"),
		metric.WithUnit("1"),
	); err != nil {
		return nil, fmt.Errorf("failed to create ActivitiesTotal metric: %w", err)
	}

	if bm.ActivitiesSuccess, err = meter.Int64Counter("activity.success.total",
		metric.WithDescription("Successfully processed activities"),
		metric.WithUnit("1"),
	); err != nil {
		return nil, fmt.Errorf("failed to create ActivitiesSuccess metric: %w", err)
	}

	if bm.ActivitiesFailed, err = meter.Int64Counter("activity.failed.total",
		metric.WithDescription("Failed activities"),
		metric.WithUnit("1"),
	); err != nil {
		return nil, fmt.Errorf("failed to create ActivitiesFailed metric: %w", err)
	}

	if bm.ActivityDuration, err = meter.Float64Histogram("activity.duration",
		metric.WithDescription("Duration of activity processing"),
		metric.WithUnit("ms"),
	); err != nil {
		return nil, fmt.Errorf("failed to create ActivityDuration metric: %w", err)
	}

	// Emissions metrics
	if bm.TotalEmissionsKgCO2, err = meter.Float64Counter("emissions.total",
		metric.WithDescription("Total emissions in kg CO2e"),
		metric.WithUnit("kg"),
	); err != nil {
		return nil, fmt.Errorf("failed to create TotalEmissionsKgCO2 metric: %w", err)
	}

	if bm.AvgEmissionsPerBatch, err = meter.Float64Histogram("emissions.avg.batch",
		metric.WithDescription("Average emissions per batch"),
		metric.WithUnit("kg"),
	); err != nil {
		return nil, fmt.Errorf("failed to create AvgEmissionsPerBatch metric: %w", err)
	}

	// Worker pool metrics
	if bm.WorkersActive, err = meter.Int64UpDownCounter("workers.active",
		metric.WithDescription("Number of active workers"),
		metric.WithUnit("1"),
	); err != nil {
		return nil, fmt.Errorf("failed to create WorkersActive metric: %w", err)
	}

	if bm.WorkerIdleTime, err = meter.Float64Histogram("worker.idle.duration",
		metric.WithDescription("Duration worker was idle"),
		metric.WithUnit("ms"),
	); err != nil {
		return nil, fmt.Errorf("failed to create WorkerIdleTime metric: %w", err)
	}

	if bm.WorkerTaskTime, err = meter.Float64Histogram("worker.task.duration",
		metric.WithDescription("Duration of worker task execution"),
		metric.WithUnit("ms"),
	); err != nil {
		return nil, fmt.Errorf("failed to create WorkerTaskTime metric: %w", err)
	}

	// Lock/Contention metrics
	if bm.LockAcquisitions, err = meter.Int64Counter("lock.acquisitions.total",
		metric.WithDescription("Total lock acquisitions"),
		metric.WithUnit("1"),
	); err != nil {
		return nil, fmt.Errorf("failed to create LockAcquisitions metric: %w", err)
	}

	if bm.LockWaitTime, err = meter.Float64Histogram("lock.wait.duration",
		metric.WithDescription("Time spent waiting for lock"),
		metric.WithUnit("ms"),
	); err != nil {
		return nil, fmt.Errorf("failed to create LockWaitTime metric: %w", err)
	}

	if bm.LockTimeouts, err = meter.Int64Counter("lock.timeouts.total",
		metric.WithDescription("Total lock timeout occurrences"),
		metric.WithUnit("1"),
	); err != nil {
		return nil, fmt.Errorf("failed to create LockTimeouts metric: %w", err)
	}

	// Queue metrics
	if bm.QueueSize, err = meter.Int64UpDownCounter("queue.size",
		metric.WithDescription("Current size of processing queue"),
		metric.WithUnit("1"),
	); err != nil {
		return nil, fmt.Errorf("failed to create QueueSize metric: %w", err)
	}

	if bm.QueueWaitTime, err = meter.Float64Histogram("queue.wait.duration",
		metric.WithDescription("Time spent waiting in queue"),
		metric.WithUnit("ms"),
	); err != nil {
		return nil, fmt.Errorf("failed to create QueueWaitTime metric: %w", err)
	}

	// Error metrics
	if bm.BatchErrors, err = meter.Int64Counter("batch.errors.total",
		metric.WithDescription("Total batch errors"),
		metric.WithUnit("1"),
	); err != nil {
		return nil, fmt.Errorf("failed to create BatchErrors metric: %w", err)
	}

	if bm.RetryAttempts, err = meter.Int64Counter("batch.retries.total",
		metric.WithDescription("Total retry attempts"),
		metric.WithUnit("1"),
	); err != nil {
		return nil, fmt.Errorf("failed to create RetryAttempts metric: %w", err)
	}

	if bm.RetrySuccesses, err = meter.Int64Counter("batch.retries.success.total",
		metric.WithDescription("Successful retries"),
		metric.WithUnit("1"),
	); err != nil {
		return nil, fmt.Errorf("failed to create RetrySuccesses metric: %w", err)
	}

	// State transition metrics
	if bm.StateTransitions, err = meter.Int64Counter("batch.state.transitions.total",
		metric.WithDescription("Total batch state transitions"),
		metric.WithUnit("1"),
	); err != nil {
		return nil, fmt.Errorf("failed to create StateTransitions metric: %w", err)
	}

	return bm, nil
}

// RecordBatchSubmission records a batch submission
func (bm *BatchMetrics) RecordBatchSubmission(ctx context.Context, batchID string, activityCount int) {
	bm.mu.Lock()
	bm.localSubmitted++
	bm.mu.Unlock()

	if bm.BatchesSubmitted == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("batch_id", batchID),
		attribute.Int("activity_count", activityCount),
	)

	bm.BatchesSubmitted.Add(ctx, 1, attrs)
}

// RecordSubmitDuration records the duration of a batch submission
func (bm *BatchMetrics) RecordSubmitDuration(ctx context.Context, duration time.Duration, success bool) {
	if bm.SubmitDuration == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.Bool("success", success),
	)

	bm.SubmitDuration.Record(ctx, float64(duration.Milliseconds()), attrs)
}

// RecordBatchCompletion records batch completion
func (bm *BatchMetrics) RecordBatchCompletion(ctx context.Context, batchID string, duration time.Duration, successCount, errorCount int, totalEmissions float64) {
	if bm.BatchesCompleted == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("batch_id", batchID),
		attribute.Int("success_count", successCount),
		attribute.Int("error_count", errorCount),
	)

	bm.BatchesCompleted.Add(ctx, 1, attrs)
	bm.ProcessingDuration.Record(ctx, float64(duration.Milliseconds()), attrs)

	// Record emissions
	if bm.TotalEmissionsKgCO2 != nil {
		bm.TotalEmissionsKgCO2.Add(ctx, totalEmissions, attrs)
	}

	// Record average emissions
	if successCount+errorCount > 0 && bm.AvgEmissionsPerBatch != nil {
		avg := totalEmissions / float64(successCount+errorCount)
		bm.AvgEmissionsPerBatch.Record(ctx, avg, attrs)
	}
}

// RecordBatchFailure records batch failure
func (bm *BatchMetrics) RecordBatchFailure(ctx context.Context, batchID, reason string, duration time.Duration) {
	if bm.BatchesFailed == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("batch_id", batchID),
		attribute.String("reason", reason),
	)

	bm.BatchesFailed.Add(ctx, 1, attrs)
	if bm.ProcessingDuration != nil {
		bm.ProcessingDuration.Record(ctx, float64(duration.Milliseconds()), attrs)
	}

	bm.RecordError(reason)
}

// RecordActivityProcessing records activity processing
func (bm *BatchMetrics) RecordActivityProcessing(ctx context.Context, batchID, activityID string, success bool, duration time.Duration) {
	if bm.ActivitiesTotal == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("batch_id", batchID),
		attribute.String("activity_id", activityID),
		attribute.Bool("success", success),
	)

	bm.ActivitiesTotal.Add(ctx, 1, attrs)

	if success {
		if bm.ActivitiesSuccess != nil {
			bm.ActivitiesSuccess.Add(ctx, 1, attrs)
		}
	} else {
		if bm.ActivitiesFailed != nil {
			bm.ActivitiesFailed.Add(ctx, 1, attrs)
		}
	}

	if bm.ActivityDuration != nil {
		bm.ActivityDuration.Record(ctx, float64(duration.Milliseconds()), attrs)
	}
}

// RecordWorkerStateChange records worker state changes
func (bm *BatchMetrics) RecordWorkerStateChange(ctx context.Context, workerID string, active bool) {
	if bm.WorkersActive == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("worker_id", workerID),
	)

	if active {
		bm.WorkersActive.Add(ctx, 1, attrs)
	} else {
		bm.WorkersActive.Add(ctx, -1, attrs)
	}
}

// RecordLockAcquisition records lock acquisition
func (bm *BatchMetrics) RecordLockAcquisition(ctx context.Context, batchID, workerID string, waitTime time.Duration, success bool) {
	if bm.LockAcquisitions == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("batch_id", batchID),
		attribute.String("worker_id", workerID),
		attribute.Bool("success", success),
	)

	bm.LockAcquisitions.Add(ctx, 1, attrs)

	if bm.LockWaitTime != nil {
		bm.LockWaitTime.Record(ctx, float64(waitTime.Milliseconds()), attrs)
	}

	if !success && bm.LockTimeouts != nil {
		bm.LockTimeouts.Add(ctx, 1, attrs)
	}
}

// RecordQueueEvent records queue events
func (bm *BatchMetrics) RecordQueueEvent(ctx context.Context, event string, waitTime time.Duration, queueSize int) {
	if bm.QueueSize == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("event", event),
	)

	if event == "enqueue" {
		bm.QueueSize.Add(ctx, 1, attrs)
	} else if event == "dequeue" {
		bm.QueueSize.Add(ctx, -1, attrs)
	}

	if bm.QueueWaitTime != nil {
		bm.QueueWaitTime.Record(ctx, float64(waitTime.Milliseconds()), attrs)
	}
}

// RecordBatchRetry records a batch retry
func (bm *BatchMetrics) RecordBatchRetry(ctx context.Context, batchID string, attempt int, success bool) {
	if bm.RetryAttempts == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("batch_id", batchID),
		attribute.Int("attempt", attempt),
		attribute.Bool("success", success),
	)

	bm.RetryAttempts.Add(ctx, 1, attrs)

	if success && bm.RetrySuccesses != nil {
		bm.RetrySuccesses.Add(ctx, 1, attrs)
	}
}

// RecordStateTransition records state transition
func (bm *BatchMetrics) RecordStateTransition(ctx context.Context, batchID, fromState, toState string) {
	if bm.StateTransitions == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("batch_id", batchID),
		attribute.String("from_state", fromState),
		attribute.String("to_state", toState),
	)

	bm.StateTransitions.Add(ctx, 1, attrs)
}

// RecordError records error occurrence
func (bm *BatchMetrics) RecordError(errorType string) {
	if bm.BatchErrors == nil {
		return
	}

	bm.errorMutex.Lock()
	defer bm.errorMutex.Unlock()
	bm.errorRegistry[errorType]++
}

// GetErrorStats returns error statistics
func (bm *BatchMetrics) GetErrorStats() map[string]int {
	bm.errorMutex.RLock()
	defer bm.errorMutex.RUnlock()

	// Return a copy
	stats := make(map[string]int)
	for k, v := range bm.errorRegistry {
		stats[k] = v
	}
	return stats
}

// GetLocalStats returns local counters for testing/synchronous access
func (bm *BatchMetrics) GetLocalStats() (submitted, completed, failed, cancelled int64) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.localSubmitted, bm.localCompleted, bm.localFailed, bm.localCancelled
}

// RecordBatchCancellation records batch cancellation
func (bm *BatchMetrics) RecordBatchCancellation(ctx context.Context, batchID string) {
	if bm.BatchesCancelled == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("batch_id", batchID),
	)

	bm.BatchesCancelled.Add(ctx, 1, attrs)
}

// BatchMetricsCollector provides synchronized access to batch metrics
type BatchMetricsCollector struct {
	metrics *BatchMetrics
	logger  *slog.Logger
}

// NewBatchMetricsCollector creates a new batch metrics collector
func NewBatchMetricsCollector(metrics *BatchMetrics, logger *slog.Logger) *BatchMetricsCollector {
	return &BatchMetricsCollector{
		metrics: metrics,
		logger:  logger,
	}
}

// CollectMetrics returns a snapshot of current metrics
func (bmc *BatchMetricsCollector) CollectMetrics() map[string]interface{} {
	return map[string]interface{}{
		"errors": bmc.metrics.GetErrorStats(),
	}
}
