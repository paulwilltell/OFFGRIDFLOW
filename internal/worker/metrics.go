package worker

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// MetricsRecorder records worker job metrics to OpenTelemetry.
type MetricsRecorder struct {
	meter             metric.Meter
	once              sync.Once
	jobStarts         metric.Int64Counter
	jobSuccesses      metric.Int64Counter
	jobFailures       metric.Int64Counter
	jobRetries        metric.Int64Counter
	jobDurationMillis metric.Float64Histogram
}

// NewMetricsRecorder constructs a metrics recorder using the global MeterProvider.
func NewMetricsRecorder() *MetricsRecorder {
	m := &MetricsRecorder{
		meter: otel.GetMeterProvider().Meter("offgridflow/worker"),
	}
	return m
}

func (m *MetricsRecorder) init() {
	m.once.Do(func() {
		var err error
		m.jobStarts, err = m.meter.Int64Counter("worker.jobs.started")
		if err != nil {
			return
		}
		m.jobSuccesses, _ = m.meter.Int64Counter("worker.jobs.succeeded")
		m.jobFailures, _ = m.meter.Int64Counter("worker.jobs.failed")
		m.jobRetries, _ = m.meter.Int64Counter("worker.jobs.retried")
		m.jobDurationMillis, _ = m.meter.Float64Histogram("worker.jobs.duration_ms")
	})
}

// RecordStart records a job start.
func (m *MetricsRecorder) RecordStart(ctx context.Context, jobName string) {
	if m == nil {
		return
	}
	m.init()
	if m.jobStarts != nil {
		m.jobStarts.Add(ctx, 1, metric.WithAttributes(attribute.String("job", jobName)))
	}
}

// RecordSuccess records a job success with duration.
func (m *MetricsRecorder) RecordSuccess(ctx context.Context, jobName string, duration time.Duration) {
	if m == nil {
		return
	}
	m.init()
	attrs := metric.WithAttributes(attribute.String("job", jobName))
	if m.jobSuccesses != nil {
		m.jobSuccesses.Add(ctx, 1, attrs)
	}
	if m.jobDurationMillis != nil {
		m.jobDurationMillis.Record(ctx, float64(duration.Milliseconds()), attrs)
	}
}

// RecordFailure records a job failure with duration.
func (m *MetricsRecorder) RecordFailure(ctx context.Context, jobName string, duration time.Duration) {
	if m == nil {
		return
	}
	m.init()
	attrs := metric.WithAttributes(attribute.String("job", jobName))
	if m.jobFailures != nil {
		m.jobFailures.Add(ctx, 1, attrs)
	}
	if m.jobDurationMillis != nil {
		m.jobDurationMillis.Record(ctx, float64(duration.Milliseconds()), attrs)
	}
}

// RecordRetry records a job retry attempt.
func (m *MetricsRecorder) RecordRetry(ctx context.Context, jobName string) {
	if m == nil {
		return
	}
	m.init()
	if m.jobRetries != nil {
		m.jobRetries.Add(ctx, 1, metric.WithAttributes(attribute.String("job", jobName)))
	}
}
