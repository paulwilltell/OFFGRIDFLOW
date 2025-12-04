package events

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestEvent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		event   Event
		wantErr bool
	}{
		{
			name:    "valid event",
			event:   Event{Type: "test.event", Payload: "data"},
			wantErr: false,
		},
		{
			name:    "empty type",
			event:   Event{Type: "", Payload: "data"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Event.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEvent_WithMethods(t *testing.T) {
	e := NewEvent("test", "payload")

	e = e.WithCorrelation("corr-123").
		WithCausation("cause-456").
		WithSource("test-service")

	if e.CorrelationID != "corr-123" {
		t.Errorf("CorrelationID = %v, want corr-123", e.CorrelationID)
	}
	if e.CausationID != "cause-456" {
		t.Errorf("CausationID = %v, want cause-456", e.CausationID)
	}
	if e.Source != "test-service" {
		t.Errorf("Source = %v, want test-service", e.Source)
	}
}

func TestEvent_JSON(t *testing.T) {
	e := NewEvent("test", map[string]string{"key": "value"})
	data, err := e.JSON()
	if err != nil {
		t.Fatalf("JSON() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("JSON() returned empty data")
	}
}

func TestInMemoryBus_PublishSubscribe(t *testing.T) {
	ctx := context.Background()
	bus := NewInMemoryBus()
	defer bus.Close()

	var received Event
	var wg sync.WaitGroup
	wg.Add(1)

	err := bus.Subscribe(ctx, "test.event", func(e Event) {
		received = e
		wg.Done()
	})
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	event := NewEvent("test.event", "test payload")
	err = bus.Publish(ctx, event)
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	wg.Wait()

	if received.Type != "test.event" {
		t.Errorf("received.Type = %v, want test.event", received.Type)
	}
	if received.Payload != "test payload" {
		t.Errorf("received.Payload = %v, want test payload", received.Payload)
	}
}

func TestInMemoryBus_MultipleSubscribers(t *testing.T) {
	ctx := context.Background()
	bus := NewInMemoryBus()
	defer bus.Close()

	var count int32
	handler := func(e Event) {
		atomic.AddInt32(&count, 1)
	}

	// Subscribe multiple handlers
	bus.Subscribe(ctx, "test", handler)
	bus.Subscribe(ctx, "test", handler)
	bus.Subscribe(ctx, "test", handler)

	event := NewEvent("test", "data")
	bus.Publish(ctx, event)

	// Give handlers time to execute
	time.Sleep(10 * time.Millisecond)

	if atomic.LoadInt32(&count) != 3 {
		t.Errorf("handler count = %v, want 3", count)
	}
}

func TestInMemoryBus_WildcardSubscription(t *testing.T) {
	ctx := context.Background()
	bus := NewInMemoryBus()
	defer bus.Close()

	var received []Event
	var mu sync.Mutex

	bus.Subscribe(ctx, "*", func(e Event) {
		mu.Lock()
		received = append(received, e)
		mu.Unlock()
	})

	bus.Publish(ctx, NewEvent("test.one", "1"))
	bus.Publish(ctx, NewEvent("test.two", "2"))
	bus.Publish(ctx, NewEvent("test.three", "3"))

	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	count := len(received)
	mu.Unlock()

	if count != 3 {
		t.Errorf("received %v events, want 3", count)
	}
}

func TestInMemoryBus_AsyncMode(t *testing.T) {
	ctx := context.Background()
	bus := NewInMemoryBus(WithAsyncDispatch(10))
	defer bus.Close()

	var received Event
	var wg sync.WaitGroup
	wg.Add(1)

	bus.Subscribe(ctx, "test", func(e Event) {
		received = e
		wg.Done()
	})

	event := NewEvent("test", "async data")
	err := bus.Publish(ctx, event)
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	wg.Wait()

	if received.Payload != "async data" {
		t.Errorf("received.Payload = %v, want async data", received.Payload)
	}
}

func TestInMemoryBus_ClosedBus(t *testing.T) {
	ctx := context.Background()
	bus := NewInMemoryBus()
	bus.Close()

	err := bus.Publish(ctx, NewEvent("test", "data"))
	if err != ErrBusClosed {
		t.Errorf("Publish() error = %v, want ErrBusClosed", err)
	}

	err = bus.Subscribe(ctx, "test", func(e Event) {})
	if err != ErrBusClosed {
		t.Errorf("Subscribe() error = %v, want ErrBusClosed", err)
	}
}

func TestInMemoryBus_ValidationErrors(t *testing.T) {
	ctx := context.Background()
	bus := NewInMemoryBus()
	defer bus.Close()

	// Empty topic
	err := bus.Subscribe(ctx, "", func(e Event) {})
	if err != ErrEmptyTopic {
		t.Errorf("Subscribe() error = %v, want ErrEmptyTopic", err)
	}

	// Nil handler
	err = bus.Subscribe(ctx, "test", nil)
	if err != ErrNilHandler {
		t.Errorf("Subscribe() error = %v, want ErrNilHandler", err)
	}

	// Empty event type
	err = bus.Publish(ctx, Event{})
	if err != ErrEmptyEventType {
		t.Errorf("Publish() error = %v, want ErrEmptyEventType", err)
	}
}

func TestInMemoryBus_PanicRecovery(t *testing.T) {
	ctx := context.Background()
	bus := NewInMemoryBus()
	defer bus.Close()

	var executed bool

	bus.Subscribe(ctx, "test", func(e Event) {
		panic("handler panic")
	})

	bus.Subscribe(ctx, "test", func(e Event) {
		executed = true
	})

	// Should not panic
	bus.Publish(ctx, NewEvent("test", "data"))
	time.Sleep(10 * time.Millisecond)

	if !executed {
		t.Error("second handler was not executed after panic in first")
	}
}

func TestInMemoryBus_SubscriberCount(t *testing.T) {
	ctx := context.Background()
	bus := NewInMemoryBus()
	defer bus.Close()

	if count := bus.SubscriberCount("test"); count != 0 {
		t.Errorf("SubscriberCount() = %v, want 0", count)
	}

	bus.Subscribe(ctx, "test", func(e Event) {})
	bus.Subscribe(ctx, "test", func(e Event) {})

	if count := bus.SubscriberCount("test"); count != 2 {
		t.Errorf("SubscriberCount() = %v, want 2", count)
	}
}

func TestInMemoryBus_Topics(t *testing.T) {
	ctx := context.Background()
	bus := NewInMemoryBus()
	defer bus.Close()

	bus.Subscribe(ctx, "topic1", func(e Event) {})
	bus.Subscribe(ctx, "topic2", func(e Event) {})
	bus.Subscribe(ctx, "topic3", func(e Event) {})

	topics := bus.Topics()
	if len(topics) != 3 {
		t.Errorf("Topics() returned %v topics, want 3", len(topics))
	}
}

func TestNoopBus(t *testing.T) {
	ctx := context.Background()
	bus := NewNoopBus()
	defer bus.Close()

	// Should succeed without doing anything
	err := bus.Publish(ctx, NewEvent("test", "data"))
	if err != nil {
		t.Errorf("Publish() error = %v", err)
	}

	err = bus.Subscribe(ctx, "test", func(e Event) {})
	if err != nil {
		t.Errorf("Subscribe() error = %v", err)
	}

	// Validate errors
	err = bus.Subscribe(ctx, "", func(e Event) {})
	if err != ErrEmptyTopic {
		t.Errorf("Subscribe() error = %v, want ErrEmptyTopic", err)
	}

	err = bus.Subscribe(ctx, "test", nil)
	if err != ErrNilHandler {
		t.Errorf("Subscribe() error = %v, want ErrNilHandler", err)
	}

	if count := bus.PublishCount(); count != 1 {
		t.Errorf("PublishCount() = %v, want 1", count)
	}
}

func TestRecordingBus(t *testing.T) {
	ctx := context.Background()
	bus := NewRecordingBus(nil)
	defer bus.Close()

	event1 := NewEvent("test.one", "data1")
	event2 := NewEvent("test.two", "data2")
	event3 := NewEvent("test.one", "data3")

	bus.Publish(ctx, event1)
	bus.Publish(ctx, event2)
	bus.Publish(ctx, event3)

	events := bus.Events()
	if len(events) != 3 {
		t.Errorf("Events() returned %v events, want 3", len(events))
	}

	oneEvents := bus.EventsOfType("test.one")
	if len(oneEvents) != 2 {
		t.Errorf("EventsOfType() returned %v events, want 2", len(oneEvents))
	}

	if !bus.HasEvent("test.one") {
		t.Error("HasEvent() returned false for test.one")
	}

	if bus.HasEvent("nonexistent") {
		t.Error("HasEvent() returned true for nonexistent event")
	}

	bus.Clear()
	if len(bus.Events()) != 0 {
		t.Error("Clear() did not remove events")
	}
}

func TestRecordingBus_WithInner(t *testing.T) {
	ctx := context.Background()
	inner := NewInMemoryBus()
	recording := NewRecordingBus(inner)
	defer recording.Close()

	var received Event
	var wg sync.WaitGroup
	wg.Add(1)

	inner.Subscribe(ctx, "test", func(e Event) {
		received = e
		wg.Done()
	})

	event := NewEvent("test", "forwarded")
	recording.Publish(ctx, event)

	wg.Wait()

	if received.Payload != "forwarded" {
		t.Errorf("inner bus did not receive event")
	}

	if !recording.HasEvent("test") {
		t.Error("recording bus did not record event")
	}
}

func TestPublishError(t *testing.T) {
	ctx := context.Background()
	bus := NewRecordingBus(nil)
	defer bus.Close()

	meta := Metadata{
		UserID: "user-123",
	}

	err := PublishError(ctx, bus, "test-service", ErrBusClosed, meta)
	if err != nil {
		t.Fatalf("PublishError() error = %v", err)
	}

	events := bus.EventsOfType("error")
	if len(events) != 1 {
		t.Fatalf("EventsOfType() returned %v events, want 1", len(events))
	}

	event := events[0]
	if event.Source != "test-service" {
		t.Errorf("event.Source = %v, want test-service", event.Source)
	}
	if event.Metadata.UserID != "user-123" {
		t.Errorf("event.Metadata.UserID = %v, want user-123", event.Metadata.UserID)
	}
}

func TestMustPublish(t *testing.T) {
	ctx := context.Background()
	bus := NewInMemoryBus()
	defer bus.Close()

	// Should not panic
	MustPublish(ctx, bus, NewEvent("test", "data"))

	// Should panic
	bus.Close()
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustPublish() did not panic on closed bus")
		}
	}()
	MustPublish(ctx, bus, NewEvent("test", "data"))
}

func TestContextCancellation(t *testing.T) {
	bus := NewInMemoryBus()
	defer bus.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := bus.Publish(ctx, NewEvent("test", "data"))
	if err == nil {
		t.Error("Publish() did not return error for cancelled context")
	}

	err = bus.Subscribe(ctx, "test", func(e Event) {})
	if err == nil {
		t.Error("Subscribe() did not return error for cancelled context")
	}
}

func BenchmarkInMemoryBus_Publish(b *testing.B) {
	ctx := context.Background()
	bus := NewInMemoryBus()
	defer bus.Close()

	event := NewEvent("benchmark", "data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish(ctx, event)
	}
}

func BenchmarkInMemoryBus_PublishWithSubscribers(b *testing.B) {
	ctx := context.Background()
	bus := NewInMemoryBus()
	defer bus.Close()

	// Add 10 subscribers
	for i := 0; i < 10; i++ {
		bus.Subscribe(ctx, "benchmark", func(e Event) {})
	}

	event := NewEvent("benchmark", "data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish(ctx, event)
	}
}

func BenchmarkInMemoryBus_AsyncPublish(b *testing.B) {
	ctx := context.Background()
	bus := NewInMemoryBus(WithAsyncDispatch(1000))
	defer bus.Close()

	event := NewEvent("benchmark", "data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish(ctx, event)
	}
}
