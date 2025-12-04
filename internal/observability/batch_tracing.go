package observability

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// BatchTracer provides tracing for batch operations
type BatchTracer struct {
	tracer trace.Tracer
	logger *slog.Logger
}

// NewBatchTracer creates a new batch tracer
func NewBatchTracer(serviceName string, logger *slog.Logger) *BatchTracer {
	tracer := otel.Tracer(fmt.Sprintf("%s-batch-processor", serviceName))
	return &BatchTracer{
		tracer: tracer,
		logger: logger,
	}
}

// StartBatchSubmissionSpan starts a span for batch submission
func (bt *BatchTracer) StartBatchSubmissionSpan(ctx context.Context, batchID, orgID, workspaceID string, activityCount int) (context.Context, trace.Span) {
	ctx, span := bt.tracer.Start(ctx, "batch.submission",
		trace.WithAttributes(
			attribute.String("batch.id", batchID),
			attribute.String("org.id", orgID),
			attribute.String("workspace.id", workspaceID),
			attribute.Int("activity.count", activityCount),
		),
	)

	bt.logger.Debug("batch submission span started", slog.String("batch_id", batchID))
	return ctx, span
}

// StartBatchProcessingSpan starts a span for batch processing
func (bt *BatchTracer) StartBatchProcessingSpan(ctx context.Context, batchID string, activityCount int) (context.Context, trace.Span) {
	ctx, span := bt.tracer.Start(ctx, "batch.processing",
		trace.WithAttributes(
			attribute.String("batch.id", batchID),
			attribute.Int("activity.count", activityCount),
		),
	)

	return ctx, span
}

// StartActivityProcessingSpan starts a span for activity processing
func (bt *BatchTracer) StartActivityProcessingSpan(ctx context.Context, batchID, activityID string) (context.Context, trace.Span) {
	ctx, span := bt.tracer.Start(ctx, "activity.processing",
		trace.WithAttributes(
			attribute.String("batch.id", batchID),
			attribute.String("activity.id", activityID),
		),
	)

	return ctx, span
}

// StartLockAcquisitionSpan starts a span for lock acquisition
func (bt *BatchTracer) StartLockAcquisitionSpan(ctx context.Context, batchID, workerID string) (context.Context, trace.Span) {
	ctx, span := bt.tracer.Start(ctx, "lock.acquisition",
		trace.WithAttributes(
			attribute.String("batch.id", batchID),
			attribute.String("worker.id", workerID),
		),
	)

	return ctx, span
}

// StartSchedulerPollSpan starts a span for scheduler polling
func (bt *BatchTracer) StartSchedulerPollSpan(ctx context.Context) (context.Context, trace.Span) {
	ctx, span := bt.tracer.Start(ctx, "scheduler.poll")
	return ctx, span
}

// RecordBatchSubmissionComplete records successful batch submission
func (bt *BatchTracer) RecordBatchSubmissionComplete(span trace.Span, batchID string, duration time.Duration) {
	span.AddEvent("batch.submission.complete", trace.WithAttributes(
		attribute.String("batch.id", batchID),
		attribute.Int64("duration.ms", duration.Milliseconds()),
	))

	span.SetStatus(codes.Ok, "batch submitted successfully")
}

// RecordBatchSubmissionError records batch submission error
func (bt *BatchTracer) RecordBatchSubmissionError(span trace.Span, err error) {
	span.RecordError(err, trace.WithAttributes(
		attribute.String("error.type", fmt.Sprintf("%T", err)),
	))

	span.SetStatus(codes.Error, err.Error())
}

// RecordBatchProcessingStart records the start of batch processing
func (bt *BatchTracer) RecordBatchProcessingStart(span trace.Span, batchID string, workerID string) {
	span.AddEvent("batch.processing.start", trace.WithAttributes(
		attribute.String("batch.id", batchID),
		attribute.String("worker.id", workerID),
	))
}

// RecordActivityComplete records activity completion
func (bt *BatchTracer) RecordActivityComplete(span trace.Span, activityID string, duration time.Duration) {
	span.AddEvent("activity.complete", trace.WithAttributes(
		attribute.String("activity.id", activityID),
		attribute.Int64("duration.ms", duration.Milliseconds()),
	))

	span.SetStatus(codes.Ok, "activity processed successfully")
}

// RecordActivityError records activity error
func (bt *BatchTracer) RecordActivityError(span trace.Span, activityID string, err error) {
	span.RecordError(err, trace.WithAttributes(
		attribute.String("activity.id", activityID),
		attribute.String("error.type", fmt.Sprintf("%T", err)),
	))

	span.SetStatus(codes.Error, err.Error())
}

// RecordLockAcquisitionSuccess records successful lock acquisition
func (bt *BatchTracer) RecordLockAcquisitionSuccess(span trace.Span, batchID, workerID string, waitTime time.Duration) {
	span.AddEvent("lock.acquired", trace.WithAttributes(
		attribute.String("batch.id", batchID),
		attribute.String("worker.id", workerID),
		attribute.Int64("wait.ms", waitTime.Milliseconds()),
	))

	span.SetStatus(codes.Ok, "lock acquired")
}

// RecordLockAcquisitionFailure records lock acquisition failure
func (bt *BatchTracer) RecordLockAcquisitionFailure(span trace.Span, batchID, workerID string, reason string) {
	span.AddEvent("lock.acquisition.failed", trace.WithAttributes(
		attribute.String("batch.id", batchID),
		attribute.String("worker.id", workerID),
		attribute.String("reason", reason),
	))

	span.SetStatus(codes.Error, reason)
}

// RecordBatchProcessingComplete records batch processing completion
func (bt *BatchTracer) RecordBatchProcessingComplete(span trace.Span, batchID string, successCount, errorCount int, duration time.Duration, emissions float64) {
	span.AddEvent("batch.processing.complete", trace.WithAttributes(
		attribute.String("batch.id", batchID),
		attribute.Int("success.count", successCount),
		attribute.Int("error.count", errorCount),
		attribute.Int64("duration.ms", duration.Milliseconds()),
		attribute.Float64("emissions.kg_co2e", emissions),
	))

	span.SetStatus(codes.Ok, "batch processed successfully")
}

// RecordBatchProcessingError records batch processing error
func (bt *BatchTracer) RecordBatchProcessingError(span trace.Span, batchID string, err error) {
	span.RecordError(err, trace.WithAttributes(
		attribute.String("batch.id", batchID),
		attribute.String("error.type", fmt.Sprintf("%T", err)),
	))

	span.SetStatus(codes.Error, err.Error())
}

// RecordSchedulerPollComplete records scheduler poll completion
func (bt *BatchTracer) RecordSchedulerPollComplete(span trace.Span, batchesFound int, duration time.Duration) {
	span.AddEvent("scheduler.poll.complete", trace.WithAttributes(
		attribute.Int("batches.found", batchesFound),
		attribute.Int64("duration.ms", duration.Milliseconds()),
	))

	span.SetStatus(codes.Ok, "poll completed")
}

// RecordBatchRetry records a batch retry
func (bt *BatchTracer) RecordBatchRetry(span trace.Span, batchID string, attempt int, reason string) {
	span.AddEvent("batch.retry", trace.WithAttributes(
		attribute.String("batch.id", batchID),
		attribute.Int("attempt", attempt),
		attribute.String("reason", reason),
	))
}

// RecordBatchStateTransition records batch state transition
func (bt *BatchTracer) RecordBatchStateTransition(span trace.Span, batchID, fromState, toState string) {
	span.AddEvent("batch.state.transition", trace.WithAttributes(
		attribute.String("batch.id", batchID),
		attribute.String("from.state", fromState),
		attribute.String("to.state", toState),
	))
}

// RecordWorkerBusyEvent records worker busy event
func (bt *BatchTracer) RecordWorkerBusyEvent(span trace.Span, workerID string) {
	span.AddEvent("worker.busy", trace.WithAttributes(
		attribute.String("worker.id", workerID),
	))
}

// RecordWorkerIdleEvent records worker idle event
func (bt *BatchTracer) RecordWorkerIdleEvent(span trace.Span, workerID string, idleDuration time.Duration) {
	span.AddEvent("worker.idle", trace.WithAttributes(
		attribute.String("worker.id", workerID),
		attribute.Int64("idle.ms", idleDuration.Milliseconds()),
	))
}

// RecordLockTimeout records lock timeout event
func (bt *BatchTracer) RecordLockTimeout(span trace.Span, batchID string) {
	span.AddEvent("lock.timeout", trace.WithAttributes(
		attribute.String("batch.id", batchID),
	))

	span.SetStatus(codes.Error, "lock acquisition timeout")
}

// BatchTraceContext provides context-aware tracing utilities
type BatchTraceContext struct {
	tracer *BatchTracer
	logger *slog.Logger
	spans  []trace.Span
}

// NewBatchTraceContext creates a new batch trace context
func NewBatchTraceContext(tracer *BatchTracer, logger *slog.Logger) *BatchTraceContext {
	return &BatchTraceContext{
		tracer: tracer,
		logger: logger,
		spans:  make([]trace.Span, 0),
	}
}

// AddSpan adds a span to the context
func (btc *BatchTraceContext) AddSpan(span trace.Span) {
	btc.spans = append(btc.spans, span)
}

// EndAllSpans ends all spans in the context
func (btc *BatchTraceContext) EndAllSpans() {
	for _, span := range btc.spans {
		span.End()
	}
}

// EndSpan ends the last span
func (btc *BatchTraceContext) EndSpan() {
	if len(btc.spans) > 0 {
		btc.spans[len(btc.spans)-1].End()
		btc.spans = btc.spans[:len(btc.spans)-1]
	}
}
