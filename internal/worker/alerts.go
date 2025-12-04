package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/example/offgridflow/internal/events"
)

// Alert represents a worker alert event.
type Alert struct {
	Time     time.Time
	Job      string
	Severity string
	Message  string
	Error    string
	Metadata map[string]string
}

// AlertQueue buffers alerts and dispatches them to the event bus and logs.
type AlertQueue struct {
	ch     chan Alert
	bus    events.Bus
	logger *slog.Logger
}

// NewAlertQueue creates a buffered alert queue.
func NewAlertQueue(bus events.Bus, logger *slog.Logger, buffer int) *AlertQueue {
	if buffer <= 0 {
		buffer = 64
	}
	return &AlertQueue{
		ch:     make(chan Alert, buffer),
		bus:    bus,
		logger: logger,
	}
}

// Start begins draining alerts until ctx is cancelled.
func (q *AlertQueue) Start(ctx context.Context) {
	if q == nil {
		return
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case alert := <-q.ch:
				q.dispatch(ctx, alert)
			}
		}
	}()
}

// Publish enqueues an alert.
func (q *AlertQueue) Publish(alert Alert) {
	if q == nil {
		return
	}
	select {
	case q.ch <- alert:
	default:
		if q.logger != nil {
			q.logger.Warn("alert queue full, dropping alert", "job", alert.Job, "message", alert.Message)
		}
	}
}

func (q *AlertQueue) dispatch(ctx context.Context, alert Alert) {
	if q.logger != nil {
		q.logger.Error("worker alert",
			"job", alert.Job,
			"severity", alert.Severity,
			"message", alert.Message,
			"error", alert.Error,
			"metadata", alert.Metadata,
		)
	}

	if q.bus != nil {
		_ = q.bus.Publish(ctx, events.Event{
			Type:      "worker.alert",
			Timestamp: alert.Time,
			Payload: map[string]interface{}{
				"job":      alert.Job,
				"severity": alert.Severity,
				"message":  alert.Message,
				"error":    alert.Error,
				"metadata": alert.Metadata,
			},
		})
	}
}
