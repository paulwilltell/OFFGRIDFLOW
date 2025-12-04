// Package events provides a domain event system for OffGridFlow.
// It enables loose coupling between components through publish/subscribe
// messaging patterns.
//
// The package supports multiple backends through the Bus interface:
//   - NoopBus: In-memory stub for testing and development
//   - (Future) NATS, Kafka, Redis Streams implementations
//
// Usage:
//
//	bus := events.NewInMemoryBus()
//
//	// Subscribe to events
//	bus.Subscribe(ctx, "emissions.calculated", func(e events.Event) {
//	    log.Printf("Emissions calculated: %v", e.Payload)
//	})
//
//	// Publish events
//	bus.Publish(ctx, events.Event{
//	    Type:    "emissions.calculated",
//	    Payload: emissionsData,
//	})
package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Event Type Constants
// =============================================================================

// Standard event types used throughout OffGridFlow.
const (
	// User events
	EventUserCreated       = "user.created"
	EventUserUpdated       = "user.updated"
	EventUserDeleted       = "user.deleted"
	EventUserAuthenticated = "user.authenticated"

	// Emissions events
	EventEmissionsCalculated = "emissions.calculated"
	EventEmissionsUpdated    = "emissions.updated"
	EventFactorUpdated       = "emissions.factor.updated"

	// Ingestion events
	EventDataIngested   = "ingestion.completed"
	EventIngestionError = "ingestion.error"

	// Compliance events
	EventReportGenerated = "compliance.report.generated"
	EventReportSubmitted = "compliance.report.submitted"

	// Billing events
	EventSubscriptionCreated = "billing.subscription.created"
	EventSubscriptionUpdated = "billing.subscription.updated"
	EventPaymentReceived     = "billing.payment.received"
	EventPaymentFailed       = "billing.payment.failed"

	// System events
	EventModeChanged      = "system.mode.changed"
	EventHealthCheckError = "system.health.error"
)

// =============================================================================
// Sentinel Errors
// =============================================================================

var (
	// ErrBusClosed is returned when publishing to a closed bus.
	ErrBusClosed = errors.New("events: bus is closed")

	// ErrNilHandler is returned when subscribing with a nil handler.
	ErrNilHandler = errors.New("events: nil handler")

	// ErrEmptyTopic is returned when subscribing to an empty topic.
	ErrEmptyTopic = errors.New("events: empty topic")

	// ErrEmptyEventType is returned when publishing an event with no type.
	ErrEmptyEventType = errors.New("events: empty event type")
)

// =============================================================================
// Event Types
// =============================================================================

// Event represents a domain event in the system.
// Events are immutable once created and carry all information
// needed to understand what happened.
type Event struct {
	// ID is a unique identifier for this event instance.
	ID string `json:"id"`

	// Type identifies the kind of event (e.g., "user.created").
	// Use dot notation for namespacing.
	Type string `json:"type"`

	// Payload contains the event-specific data.
	// Should be JSON-serializable.
	Payload any `json:"payload"`

	// Metadata contains optional contextual information.
	Metadata Metadata `json:"metadata,omitempty"`

	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`

	// Source identifies which component generated the event.
	Source string `json:"source,omitempty"`

	// CorrelationID links related events together.
	CorrelationID string `json:"correlation_id,omitempty"`

	// CausationID is the ID of the event that caused this event.
	CausationID string `json:"causation_id,omitempty"`

	// Version supports event schema evolution.
	Version int `json:"version,omitempty"`
}

// Metadata contains optional event context.
type Metadata struct {
	// UserID is the user who triggered the event, if applicable.
	UserID string `json:"user_id,omitempty"`

	// TenantID is the tenant context, for multi-tenant scenarios.
	TenantID string `json:"tenant_id,omitempty"`

	// RequestID correlates to the originating HTTP request.
	RequestID string `json:"request_id,omitempty"`

	// TraceID for distributed tracing integration.
	TraceID string `json:"trace_id,omitempty"`

	// Custom allows arbitrary key-value metadata.
	Custom map[string]any `json:"custom,omitempty"`
}

// NewEvent creates a new Event with a generated ID and current timestamp.
func NewEvent(eventType string, payload any) Event {
	return Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
		Version:   1,
	}
}

// NewEventWithMetadata creates a new Event with metadata.
func NewEventWithMetadata(eventType string, payload any, meta Metadata) Event {
	e := NewEvent(eventType, payload)
	e.Metadata = meta
	return e
}

// WithCorrelation sets the correlation ID and returns the event.
func (e Event) WithCorrelation(correlationID string) Event {
	e.CorrelationID = correlationID
	return e
}

// WithCausation sets the causation ID (the event that caused this one).
func (e Event) WithCausation(causationID string) Event {
	e.CausationID = causationID
	return e
}

// WithSource sets the source component.
func (e Event) WithSource(source string) Event {
	e.Source = source
	return e
}

// Validate checks that the event has required fields.
func (e Event) Validate() error {
	if e.Type == "" {
		return ErrEmptyEventType
	}
	return nil
}

// JSON serializes the event to JSON bytes.
func (e Event) JSON() ([]byte, error) {
	return json.Marshal(e)
}

// =============================================================================
// Bus Interface
// =============================================================================

// Handler is a function that processes events.
type Handler func(Event)

// Bus defines the interface for event publishing and subscription.
// Implementations must be safe for concurrent use.
type Bus interface {
	// Publish sends an event to all subscribers of its type.
	// Returns an error if the bus is closed or the event is invalid.
	Publish(ctx context.Context, event Event) error

	// Subscribe registers a handler for events matching the topic.
	// The topic can be an exact event type or a pattern (implementation-specific).
	// Returns an error if the topic is empty or handler is nil.
	Subscribe(ctx context.Context, topic string, handler Handler) error

	// Close shuts down the bus and releases resources.
	Close() error
}

// =============================================================================
// In-Memory Bus Implementation
// =============================================================================

// InMemoryBus is a simple in-memory event bus for development and testing.
// It synchronously dispatches events to handlers.
type InMemoryBus struct {
	mu          sync.RWMutex
	subscribers map[string][]Handler
	closed      bool

	// Options
	async      bool
	bufferSize int
	eventChan  chan Event
	wg         sync.WaitGroup
}

// InMemoryBusOption configures the InMemoryBus.
type InMemoryBusOption func(*InMemoryBus)

// WithAsyncDispatch enables asynchronous event dispatch with the given buffer size.
func WithAsyncDispatch(bufferSize int) InMemoryBusOption {
	return func(b *InMemoryBus) {
		b.async = true
		b.bufferSize = bufferSize
	}
}

// NewInMemoryBus creates a new in-memory event bus.
func NewInMemoryBus(opts ...InMemoryBusOption) *InMemoryBus {
	b := &InMemoryBus{
		subscribers: make(map[string][]Handler),
	}

	for _, opt := range opts {
		opt(b)
	}

	if b.async {
		if b.bufferSize <= 0 {
			b.bufferSize = 100
		}
		b.eventChan = make(chan Event, b.bufferSize)
		b.wg.Add(1)
		go b.dispatchLoop()
	}

	return b
}

// Publish sends an event to all matching subscribers.
func (b *InMemoryBus) Publish(ctx context.Context, event Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := event.Validate(); err != nil {
		return err
	}

	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return ErrBusClosed
	}
	b.mu.RUnlock()

	// Ensure event has ID and timestamp
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	if b.async {
		select {
		case b.eventChan <- event:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	b.dispatch(event)
	return nil
}

// Subscribe registers a handler for the given topic.
func (b *InMemoryBus) Subscribe(ctx context.Context, topic string, handler Handler) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if topic == "" {
		return ErrEmptyTopic
	}
	if handler == nil {
		return ErrNilHandler
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return ErrBusClosed
	}

	b.subscribers[topic] = append(b.subscribers[topic], handler)
	return nil
}

// Close shuts down the bus.
func (b *InMemoryBus) Close() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true
	b.mu.Unlock()

	if b.async && b.eventChan != nil {
		close(b.eventChan)
		b.wg.Wait()
	}

	return nil
}

// dispatch sends an event to all matching handlers.
func (b *InMemoryBus) dispatch(event Event) {
	b.mu.RLock()
	handlers := make([]Handler, len(b.subscribers[event.Type]))
	copy(handlers, b.subscribers[event.Type])

	// Also dispatch to wildcard subscribers
	wildcardHandlers := b.subscribers["*"]
	b.mu.RUnlock()

	for _, h := range handlers {
		b.safeCall(h, event)
	}

	for _, h := range wildcardHandlers {
		b.safeCall(h, event)
	}
}

// dispatchLoop processes events from the channel in async mode.
func (b *InMemoryBus) dispatchLoop() {
	defer b.wg.Done()
	for event := range b.eventChan {
		b.dispatch(event)
	}
}

// safeCall invokes a handler with panic recovery.
func (b *InMemoryBus) safeCall(h Handler, event Event) {
	defer func() {
		if r := recover(); r != nil {
			// In production, log this panic
			// For now, we silently recover to prevent one bad handler
			// from breaking the entire event system
		}
	}()
	h(event)
}

// SubscriberCount returns the number of subscribers for a topic.
func (b *InMemoryBus) SubscriberCount(topic string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribers[topic])
}

// Topics returns all subscribed topics.
func (b *InMemoryBus) Topics() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	topics := make([]string, 0, len(b.subscribers))
	for topic := range b.subscribers {
		topics = append(topics, topic)
	}
	return topics
}

// =============================================================================
// Noop Bus Implementation
// =============================================================================

// NoopBus is a no-operation event bus that discards all events.
// Use for testing or when event handling is disabled.
type NoopBus struct {
	publishCount int64
	mu           sync.Mutex
}

// NewNoopBus creates a new no-op event bus.
func NewNoopBus() *NoopBus {
	return &NoopBus{}
}

// Publish discards the event but tracks the count.
func (b *NoopBus) Publish(ctx context.Context, event Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	b.mu.Lock()
	b.publishCount++
	b.mu.Unlock()
	return nil
}

// Subscribe is a no-op that succeeds without storing the handler.
func (b *NoopBus) Subscribe(ctx context.Context, topic string, handler Handler) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if topic == "" {
		return ErrEmptyTopic
	}
	if handler == nil {
		return ErrNilHandler
	}
	return nil
}

// Close is a no-op.
func (b *NoopBus) Close() error {
	return nil
}

// PublishCount returns the number of events published.
func (b *NoopBus) PublishCount() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.publishCount
}

// =============================================================================
// Recording Bus (for Testing)
// =============================================================================

// RecordingBus records all published events for testing inspection.
type RecordingBus struct {
	mu     sync.Mutex
	events []Event
	inner  Bus
}

// NewRecordingBus creates a bus that records events, optionally forwarding to an inner bus.
func NewRecordingBus(inner Bus) *RecordingBus {
	return &RecordingBus{
		events: make([]Event, 0),
		inner:  inner,
	}
}

// Publish records the event and optionally forwards to the inner bus.
func (b *RecordingBus) Publish(ctx context.Context, event Event) error {
	b.mu.Lock()
	b.events = append(b.events, event)
	b.mu.Unlock()

	if b.inner != nil {
		return b.inner.Publish(ctx, event)
	}
	return nil
}

// Subscribe forwards to the inner bus if present.
func (b *RecordingBus) Subscribe(ctx context.Context, topic string, handler Handler) error {
	if b.inner != nil {
		return b.inner.Subscribe(ctx, topic, handler)
	}
	return nil
}

// Close forwards to the inner bus if present.
func (b *RecordingBus) Close() error {
	if b.inner != nil {
		return b.inner.Close()
	}
	return nil
}

// Events returns all recorded events.
func (b *RecordingBus) Events() []Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	result := make([]Event, len(b.events))
	copy(result, b.events)
	return result
}

// EventsOfType returns all recorded events of a specific type.
func (b *RecordingBus) EventsOfType(eventType string) []Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	var result []Event
	for _, e := range b.events {
		if e.Type == eventType {
			result = append(result, e)
		}
	}
	return result
}

// Clear removes all recorded events.
func (b *RecordingBus) Clear() {
	b.mu.Lock()
	b.events = b.events[:0]
	b.mu.Unlock()
}

// HasEvent returns true if an event of the given type was recorded.
func (b *RecordingBus) HasEvent(eventType string) bool {
	return len(b.EventsOfType(eventType)) > 0
}

// =============================================================================
// Compile-time Interface Checks
// =============================================================================

var (
	_ Bus = (*InMemoryBus)(nil)
	_ Bus = (*NoopBus)(nil)
	_ Bus = (*RecordingBus)(nil)
)

// =============================================================================
// Helper Functions
// =============================================================================

// PublishError creates and publishes an error event.
func PublishError(ctx context.Context, bus Bus, source string, err error, meta Metadata) error {
	if bus == nil {
		return nil
	}

	event := NewEventWithMetadata("error", map[string]any{
		"error":  err.Error(),
		"source": source,
	}, meta)
	event.Source = source

	return bus.Publish(ctx, event)
}

// MustPublish publishes an event and panics on error.
// Use only when event publication failure should be fatal.
func MustPublish(ctx context.Context, bus Bus, event Event) {
	if err := bus.Publish(ctx, event); err != nil {
		panic(fmt.Sprintf("events: failed to publish %s: %v", event.Type, err))
	}
}
