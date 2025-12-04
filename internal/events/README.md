# Event Bus System

A comprehensive event-driven architecture for OffGridFlow, enabling loose coupling between components through publish/subscribe messaging patterns.

## Features

- ✅ **Multiple Backends**: In-memory, NATS, Redis, with easy extensibility
- ✅ **Event Store**: PostgreSQL-backed persistent event storage for audit trails and event sourcing
- ✅ **Async Processing**: Optional asynchronous event dispatch
- ✅ **Wildcard Subscriptions**: Subscribe to all events with `*`
- ✅ **Panic Recovery**: Handlers are protected with automatic panic recovery
- ✅ **Correlation & Causation**: Built-in support for event tracing
- ✅ **Testing Utilities**: Recording bus and noop bus for tests
- ✅ **Production Ready**: NATS and Redis adapters for distributed systems

## Quick Start

### In-Memory Bus (Development/Testing)

```go
import "github.com/example/offgridflow/internal/events"

// Create synchronous bus
bus := events.NewInMemoryBus()
defer bus.Close()

// Or async bus with buffering
bus := events.NewInMemoryBus(events.WithAsyncDispatch(100))
defer bus.Close()
```

### Subscribe to Events

```go
ctx := context.Background()

// Subscribe to specific event type
bus.Subscribe(ctx, events.EventEmissionsCalculated, func(e events.Event) {
    log.Printf("Emissions calculated: %v", e.Payload)
})

// Subscribe to all events
bus.Subscribe(ctx, "*", func(e events.Event) {
    log.Printf("Event %s: %v", e.Type, e.Payload)
})
```

### Publish Events

```go
// Create and publish event
event := events.NewEvent(events.EventEmissionsCalculated, map[string]any{
    "total_kg_co2e": 1234.56,
    "scope":         "scope1",
})

if err := bus.Publish(ctx, event); err != nil {
    log.Fatal(err)
}

// With metadata
event := events.NewEventWithMetadata(
    events.EventUserCreated,
    userData,
    events.Metadata{
        UserID:    "user-123",
        TenantID:  "tenant-456",
        RequestID: "req-789",
    },
)
bus.Publish(ctx, event)

// With correlation and causation
event = event.
    WithCorrelation("corr-123").
    WithCausation("cause-456").
    WithSource("user-service")
```

## Production Backends

### NATS (Recommended for Distributed Systems)

```go
import "github.com/example/offgridflow/internal/events"

config := events.DefaultNATSConfig()
config.URL = "nats://nats.example.com:4222"
config.EnableJetStream = true

bus, err := events.NewNATSBus(config)
if err != nil {
    log.Fatal(err)
}
defer bus.Close()

// Use like any other bus
bus.Subscribe(ctx, "user.created", handler)
bus.Publish(ctx, event)
```

**Features:**
- Persistent messaging with JetStream
- At-least-once delivery guarantees
- Automatic reconnection
- Message acknowledgment
- Durable subscriptions

### Redis (Simpler Distributed Option)

```go
config := events.DefaultRedisConfig()
config.Addrs = []string{"redis.example.com:6379"}
config.UseStreams = true // or false for simple pub/sub

bus, err := events.NewRedisBus(ctx, config)
if err != nil {
    log.Fatal(err)
}
defer bus.Close()
```

**Redis Streams Mode:**
- Persistent messages
- Consumer groups
- Message acknowledgment
- Ordered delivery

**Redis Pub/Sub Mode:**
- Lightweight, fire-and-forget
- Lower latency
- No persistence

## Event Store (Audit Trail & Event Sourcing)

```go
import "database/sql"

db, _ := sql.Open("postgres", dsn)

// Create event store with optional real-time bus
store, err := events.NewPostgresEventStore(db, bus)
if err != nil {
    log.Fatal(err)
}

// Append events
event := events.NewEvent("order.created", orderData)
store.Append(ctx, event)

// Load events by criteria
criteria := events.EventCriteria{
    EventTypes:    []string{"order.created", "order.updated"},
    Since:         time.Now().AddDate(0, -1, 0),
    CorrelationID: "order-123",
    Limit:         100,
}

events, err := store.Load(ctx, criteria)

// Load event stream for aggregate
events, err := store.LoadStream(ctx, "order-123", 0)

// Subscribe to new events
store.Subscribe(ctx, criteria, func(e events.Event) {
    // Process new event
})
```

## Standard Event Types

The package defines standard event types as constants:

```go
// User events
events.EventUserCreated
events.EventUserUpdated
events.EventUserDeleted
events.EventUserAuthenticated

// Emissions events
events.EventEmissionsCalculated
events.EventEmissionsUpdated
events.EventFactorUpdated

// Ingestion events
events.EventDataIngested
events.EventIngestionError

// Compliance events
events.EventReportGenerated
events.EventReportSubmitted

// Billing events
events.EventSubscriptionCreated
events.EventSubscriptionUpdated
events.EventPaymentReceived
events.EventPaymentFailed

// System events
events.EventModeChanged
events.EventHealthCheckError
```

## Testing Utilities

### Recording Bus (Inspect Events in Tests)

```go
func TestMyService(t *testing.T) {
    bus := events.NewRecordingBus(nil)
    defer bus.Close()

    // Run code that publishes events
    service.DoSomething(ctx)

    // Assert events were published
    if !bus.HasEvent(events.EventUserCreated) {
        t.Error("user.created event not published")
    }

    events := bus.EventsOfType(events.EventUserCreated)
    if len(events) != 1 {
        t.Errorf("expected 1 event, got %d", len(events))
    }

    // Inspect event payload
    payload := events[0].Payload.(map[string]any)
    if payload["email"] != "test@example.com" {
        t.Error("unexpected email in payload")
    }
}
```

### Noop Bus (Disable Events in Tests)

```go
bus := events.NewNoopBus()
defer bus.Close()

// Events are discarded but counted
service.Run(ctx)

count := bus.PublishCount()
log.Printf("Service published %d events", count)
```

## Error Handling

### Publishing Errors

```go
err := events.PublishError(ctx, bus, "my-service", err, metadata)
```

### Must Publish (Panics on Error)

```go
// Only use when event publication is critical
events.MustPublish(ctx, bus, event)
```

## Event Structure

```go
type Event struct {
    ID            string            // Unique event ID
    Type          string            // Event type (e.g., "user.created")
    Payload       any               // Event-specific data
    Metadata      Metadata          // Optional context
    Timestamp     time.Time         // When event occurred
    Source        string            // Component that generated event
    CorrelationID string            // Links related events
    CausationID   string            // ID of causing event
    Version       int               // Schema version
}

type Metadata struct {
    UserID    string            // User who triggered event
    TenantID  string            // Tenant context
    RequestID string            // HTTP request ID
    TraceID   string            // Distributed tracing ID
    Custom    map[string]any    // Custom metadata
}
```

## Best Practices

### 1. Use Descriptive Event Names

```go
// Good
events.EventEmissionsCalculated
events.EventUserAuthenticated

// Avoid
events.EventData
events.EventUpdate
```

### 2. Include Rich Metadata

```go
event := events.NewEventWithMetadata(
    events.EventReportGenerated,
    reportData,
    events.Metadata{
        UserID:    userID,
        TenantID:  tenantID,
        RequestID: requestID,
        Custom: map[string]any{
            "report_type": "SEC",
            "year":        2024,
        },
    },
)
```

### 3. Use Correlation for Request Tracing

```go
correlationID := r.Header.Get("X-Correlation-ID")
if correlationID == "" {
    correlationID = uuid.New().String()
}

event := events.NewEvent("user.login", loginData).
    WithCorrelation(correlationID).
    WithSource("auth-service")
```

### 4. Handle Errors Gracefully

```go
if err := bus.Publish(ctx, event); err != nil {
    // Log but don't fail the operation
    log.Error("failed to publish event", "error", err)
}
```

### 5. Use Event Store for Audit Requirements

```go
// Append to both bus and store
bus.Publish(ctx, event)
store.Append(ctx, event)

// Or use store with bus integration
store, _ := events.NewPostgresEventStore(db, bus)
store.Append(ctx, event) // Automatically publishes to bus
```

## Architecture Patterns

### Event Sourcing

```go
// Rebuild aggregate state from events
type Order struct {
    ID     string
    Status string
    Items  []Item
}

func (o *Order) ApplyEvent(e events.Event) {
    switch e.Type {
    case "order.created":
        o.ID = e.Payload.(map[string]any)["id"].(string)
        o.Status = "pending"
    case "order.item_added":
        // Add item
    case "order.completed":
        o.Status = "completed"
    }
}

events, _ := store.LoadStream(ctx, orderID, 0)
order := &Order{}
for _, e := range events {
    order.ApplyEvent(e)
}
```

### CQRS (Command Query Responsibility Segregation)

```go
// Command side - publishes events
func (s *OrderService) CreateOrder(ctx context.Context, cmd CreateOrderCommand) error {
    // Validate and create order
    order := createOrderFromCommand(cmd)

    // Publish event
    event := events.NewEvent("order.created", order)
    return s.bus.Publish(ctx, event)
}

// Query side - subscribes to events
func (s *OrderQueryService) Start(ctx context.Context) {
    s.bus.Subscribe(ctx, "order.*", func(e events.Event) {
        // Update read model
        s.updateReadModel(e)
    })
}
```

### Saga Pattern (Distributed Transactions)

```go
// Orchestrate multi-step workflow
func (s *CheckoutSaga) Start(ctx context.Context, orderID string) {
    correlationID := uuid.New().String()

    // Step 1: Reserve inventory
    s.bus.Publish(ctx, events.NewEvent("inventory.reserve", data).
        WithCorrelation(correlationID))

    // Subscribe to response
    s.bus.Subscribe(ctx, "inventory.reserved", func(e events.Event) {
        if e.CorrelationID != correlationID {
            return
        }

        // Step 2: Process payment
        s.bus.Publish(ctx, events.NewEvent("payment.process", data).
            WithCorrelation(correlationID).
            WithCausation(e.ID))
    })
}
```

## Performance Considerations

### Async Mode for High Throughput

```go
// Buffer 1000 events for async processing
bus := events.NewInMemoryBus(events.WithAsyncDispatch(1000))
```

### Batch Publishing

```go
events := []events.Event{
    events.NewEvent("type1", data1),
    events.NewEvent("type2", data2),
    events.NewEvent("type3", data3),
}

store.AppendBatch(ctx, events)
```

### Connection Pooling

```go
// NATS automatically handles connection pooling
config := events.DefaultNATSConfig()
config.MaxReconnects = 10
config.ReconnectWait = 2 * time.Second
```

## Monitoring

```go
// Check NATS status
if bus.IsConnected() {
    log.Info("NATS connected", "status", bus.Status())
}

// Monitor errors
go func() {
    for err := range bus.Errors() {
        log.Error("event bus error", "error", err)
    }
}()

// Redis health check
if err := bus.Ping(ctx); err != nil {
    log.Error("redis ping failed", "error", err)
}
```

## Migration Guide

### From In-Memory to NATS

```go
// Before
bus := events.NewInMemoryBus()

// After
config := events.DefaultNATSConfig()
config.URL = os.Getenv("NATS_URL")
bus, err := events.NewNATSBus(config)
if err != nil {
    log.Fatal(err)
}

// No changes needed to Publish/Subscribe code!
```

## Contributing

When adding new event types:

1. Define constant in `events.go`
2. Document in this README
3. Add tests in `events_test.go`
4. Update any relevant handlers

## License

Part of OffGridFlow - See main LICENSE file.
