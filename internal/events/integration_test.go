package events_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/offgridflow/internal/events"
)

// TestIntegration_EventFlowWithCorrelation tests a complete event flow with correlation
func TestIntegration_EventFlowWithCorrelation(t *testing.T) {
	ctx := context.Background()
	bus := events.NewInMemoryBus()
	defer bus.Close()

	correlationID := "test-correlation-123"
	var processedEvents []events.Event
	var mu sync.Mutex

	// Subscribe to all events to track the flow
	bus.Subscribe(ctx, "*", func(e events.Event) {
		mu.Lock()
		processedEvents = append(processedEvents, e)
		mu.Unlock()
	})

	// Simulate a multi-step workflow
	bus.Subscribe(ctx, "step.1.complete", func(e events.Event) {
		if e.CorrelationID != correlationID {
			return
		}
		// Trigger step 2
		step2 := events.NewEvent("step.2.start", "step2-data").
			WithCorrelation(correlationID).
			WithCausation(e.ID)
		bus.Publish(ctx, step2)
	})

	bus.Subscribe(ctx, "step.2.start", func(e events.Event) {
		if e.CorrelationID != correlationID {
			return
		}
		// Trigger step 3
		step3 := events.NewEvent("step.3.complete", "step3-data").
			WithCorrelation(correlationID).
			WithCausation(e.ID)
		bus.Publish(ctx, step3)
	})

	// Start the workflow
	startEvent := events.NewEvent("step.1.complete", "step1-data").
		WithCorrelation(correlationID)
	bus.Publish(ctx, startEvent)

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	count := len(processedEvents)
	mu.Unlock()

	// Should have 4 events: initial + step1.complete + step2.start + step3.complete
	// Plus the initial event gets published to wildcard
	if count < 3 {
		t.Errorf("expected at least 3 events, got %d", count)
	}

	// Verify correlation IDs
	mu.Lock()
	for _, e := range processedEvents {
		if e.Type != "step.1.complete" && e.Type != "step.2.start" && e.Type != "step.3.complete" {
			continue
		}
		if e.CorrelationID != correlationID {
			t.Errorf("event %s missing correlation ID", e.Type)
		}
	}
	mu.Unlock()
}

// TestIntegration_AsyncBusHighLoad tests async bus under load
func TestIntegration_AsyncBusHighLoad(t *testing.T) {
	ctx := context.Background()
	bus := events.NewInMemoryBus(events.WithAsyncDispatch(1000))
	defer bus.Close()

	var received int64
	bus.Subscribe(ctx, "load.test", func(e events.Event) {
		atomic.AddInt64(&received, 1)
	})

	// Publish many events
	const numEvents = 10000
	for i := 0; i < numEvents; i++ {
		bus.Publish(ctx, events.NewEvent("load.test", i))
	}

	// Wait for processing
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt64(&received) >= numEvents {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	got := atomic.LoadInt64(&received)
	if got < numEvents {
		t.Errorf("received %d events, want %d", got, numEvents)
	}
}

// TestIntegration_MultipleSubscribers tests event fanout
func TestIntegration_MultipleSubscribers(t *testing.T) {
	ctx := context.Background()
	bus := events.NewInMemoryBus()
	defer bus.Close()

	const numSubscribers = 10
	counters := make([]int64, numSubscribers)

	// Create multiple subscribers
	for i := 0; i < numSubscribers; i++ {
		idx := i
		bus.Subscribe(ctx, "fanout.test", func(e events.Event) {
			atomic.AddInt64(&counters[idx], 1)
		})
	}

	// Publish events
	const numEvents = 100
	for i := 0; i < numEvents; i++ {
		bus.Publish(ctx, events.NewEvent("fanout.test", i))
	}

	time.Sleep(50 * time.Millisecond)

	// Verify all subscribers received all events
	for i, count := range counters {
		if count != numEvents {
			t.Errorf("subscriber %d received %d events, want %d", i, count, numEvents)
		}
	}
}

// TestIntegration_EventMetadata tests rich metadata handling
func TestIntegration_EventMetadata(t *testing.T) {
	ctx := context.Background()
	bus := events.NewInMemoryBus()
	defer bus.Close()

	var receivedEvent events.Event
	var wg sync.WaitGroup
	wg.Add(1)

	bus.Subscribe(ctx, "metadata.test", func(e events.Event) {
		receivedEvent = e
		wg.Done()
	})

	metadata := events.Metadata{
		UserID:    "user-123",
		TenantID:  "tenant-456",
		RequestID: "req-789",
		TraceID:   "trace-abc",
		Custom: map[string]any{
			"custom_field": "custom_value",
			"nested": map[string]string{
				"key": "value",
			},
		},
	}

	event := events.NewEventWithMetadata("metadata.test", "payload", metadata)
	bus.Publish(ctx, event)

	wg.Wait()

	if receivedEvent.Metadata.UserID != "user-123" {
		t.Errorf("UserID = %v, want user-123", receivedEvent.Metadata.UserID)
	}
	if receivedEvent.Metadata.TenantID != "tenant-456" {
		t.Errorf("TenantID = %v, want tenant-456", receivedEvent.Metadata.TenantID)
	}
	if receivedEvent.Metadata.Custom["custom_field"] != "custom_value" {
		t.Error("custom field not preserved")
	}
}

// TestIntegration_RecordingBusWithInnerBus tests recording bus forwarding
func TestIntegration_RecordingBusWithInnerBus(t *testing.T) {
	ctx := context.Background()
	inner := events.NewInMemoryBus()
	recording := events.NewRecordingBus(inner)
	defer recording.Close()

	var innerReceived events.Event
	var wg sync.WaitGroup
	wg.Add(1)

	inner.Subscribe(ctx, "test", func(e events.Event) {
		innerReceived = e
		wg.Done()
	})

	event := events.NewEvent("test", "data")
	recording.Publish(ctx, event)

	wg.Wait()

	// Check inner bus received it
	if innerReceived.Type != "test" {
		t.Error("inner bus did not receive event")
	}

	// Check recording bus recorded it
	if !recording.HasEvent("test") {
		t.Error("recording bus did not record event")
	}

	recorded := recording.Events()
	if len(recorded) != 1 {
		t.Errorf("recorded %d events, want 1", len(recorded))
	}
}

// TestIntegration_PanicRecovery tests that panicking handlers don't crash
func TestIntegration_PanicRecovery(t *testing.T) {
	ctx := context.Background()
	bus := events.NewInMemoryBus()
	defer bus.Close()

	var safeHandlerExecuted bool

	// Handler that panics
	bus.Subscribe(ctx, "panic.test", func(e events.Event) {
		panic("handler panic!")
	})

	// Handler that should still execute
	bus.Subscribe(ctx, "panic.test", func(e events.Event) {
		safeHandlerExecuted = true
	})

	// Should not panic
	bus.Publish(ctx, events.NewEvent("panic.test", "data"))
	time.Sleep(10 * time.Millisecond)

	if !safeHandlerExecuted {
		t.Error("safe handler was not executed after panic")
	}
}

// TestIntegration_ContextCancellation tests graceful shutdown
func TestIntegration_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	bus := events.NewInMemoryBus()
	defer bus.Close()

	var count int64
	bus.Subscribe(ctx, "cancel.test", func(e events.Event) {
		atomic.AddInt64(&count, 1)
	})

	// Publish some events
	for i := 0; i < 5; i++ {
		bus.Publish(ctx, events.NewEvent("cancel.test", i))
	}

	time.Sleep(10 * time.Millisecond)

	// Cancel context
	cancel()

	// Try to publish after cancellation
	err := bus.Publish(ctx, events.NewEvent("cancel.test", "after-cancel"))
	if err == nil {
		t.Error("expected error when publishing with cancelled context")
	}

	// Should have processed the first 5 events
	got := atomic.LoadInt64(&count)
	if got != 5 {
		t.Errorf("processed %d events, want 5", got)
	}
}

// TestIntegration_WildcardAndSpecific tests wildcard with specific subscriptions
func TestIntegration_WildcardAndSpecific(t *testing.T) {
	ctx := context.Background()
	bus := events.NewInMemoryBus()
	defer bus.Close()

	var wildcardCount, specificCount int64

	bus.Subscribe(ctx, "*", func(e events.Event) {
		atomic.AddInt64(&wildcardCount, 1)
	})

	bus.Subscribe(ctx, "specific.event", func(e events.Event) {
		atomic.AddInt64(&specificCount, 1)
	})

	// Publish specific event
	bus.Publish(ctx, events.NewEvent("specific.event", "data"))

	// Publish other event
	bus.Publish(ctx, events.NewEvent("other.event", "data"))

	time.Sleep(10 * time.Millisecond)

	if atomic.LoadInt64(&wildcardCount) != 2 {
		t.Errorf("wildcard received %d events, want 2", wildcardCount)
	}

	if atomic.LoadInt64(&specificCount) != 1 {
		t.Errorf("specific received %d events, want 1", specificCount)
	}
}

// TestIntegration_BusLifecycle tests proper bus lifecycle management
func TestIntegration_BusLifecycle(t *testing.T) {
	ctx := context.Background()

	// Test sync bus
	syncBus := events.NewInMemoryBus()
	syncBus.Subscribe(ctx, "test", func(e events.Event) {})
	syncBus.Publish(ctx, events.NewEvent("test", "data"))
	if err := syncBus.Close(); err != nil {
		t.Errorf("sync bus close error: %v", err)
	}

	// Test async bus
	asyncBus := events.NewInMemoryBus(events.WithAsyncDispatch(10))
	asyncBus.Subscribe(ctx, "test", func(e events.Event) {})
	asyncBus.Publish(ctx, events.NewEvent("test", "data"))
	if err := asyncBus.Close(); err != nil {
		t.Errorf("async bus close error: %v", err)
	}

	// Test double close
	if err := asyncBus.Close(); err != nil {
		t.Errorf("double close should not error: %v", err)
	}
}

// BenchmarkIntegration_EndToEndLatency measures end-to-end event latency
func BenchmarkIntegration_EndToEndLatency(b *testing.B) {
	ctx := context.Background()
	bus := events.NewInMemoryBus()
	defer bus.Close()

	received := make(chan struct{}, b.N)
	bus.Subscribe(ctx, "bench", func(e events.Event) {
		received <- struct{}{}
	})

	event := events.NewEvent("bench", "data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish(ctx, event)
		<-received
	}
}
