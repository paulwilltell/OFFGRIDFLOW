//go:build events_examples
// +build events_examples

package events_test

import (
	"context"
	"log"
	"time"

	"github.com/example/offgridflow/internal/events"
)

// Example demonstrates basic event bus usage
func Example_basicUsage() {
	ctx := context.Background()
	bus := events.NewInMemoryBus()
	defer bus.Close()

	// Subscribe to user events
	bus.Subscribe(ctx, events.EventUserCreated, func(e events.Event) {
		log.Printf("User created: %v", e.Payload)
	})

	// Publish an event
	event := events.NewEvent(events.EventUserCreated, map[string]string{
		"id":    "user-123",
		"email": "user@example.com",
	})

	bus.Publish(ctx, event)
}

// Example demonstrates async event processing
func Example_asyncProcessing() {
	ctx := context.Background()
	// Create async bus with 100 event buffer
	bus := events.NewInMemoryBus(events.WithAsyncDispatch(100))
	defer bus.Close()

	bus.Subscribe(ctx, "*", func(e events.Event) {
		log.Printf("Async event: %s", e.Type)
	})

	// Publish many events quickly
	for i := 0; i < 1000; i++ {
		bus.Publish(ctx, events.NewEvent("test.event", i))
	}

	time.Sleep(100 * time.Millisecond) // Wait for processing
}

// Example demonstrates event correlation
func Example_eventCorrelation() {
	ctx := context.Background()
	bus := events.NewInMemoryBus()
	defer bus.Close()

	correlationID := "request-123"

	// Publish correlated events
	event1 := events.NewEvent("step.1", "data").
		WithCorrelation(correlationID).
		WithSource("service-a")

	event2 := events.NewEvent("step.2", "data").
		WithCorrelation(correlationID).
		WithCausation(event1.ID).
		WithSource("service-b")

	bus.Publish(ctx, event1)
	bus.Publish(ctx, event2)
}

// Example demonstrates testing with RecordingBus
func Example_testing() {
	ctx := context.Background()
	bus := events.NewRecordingBus(nil)
	defer bus.Close()

	// Run code that publishes events
	bus.Publish(ctx, events.NewEvent(events.EventUserCreated, "data"))
	bus.Publish(ctx, events.NewEvent(events.EventUserUpdated, "data"))

	// Assert events were published
	if !bus.HasEvent(events.EventUserCreated) {
		log.Fatal("expected user.created event")
	}

	events := bus.EventsOfType(events.EventUserCreated)
	log.Printf("Found %d user.created events", len(events))

	bus.Clear() // Reset for next test
}

// Example demonstrates NATS distributed messaging (commented out - requires NATS server)
func Example_natsDistributed() {
	// ctx := context.Background()
	//
	// config := events.DefaultNATSConfig()
	// config.URL = "nats://localhost:4222"
	//
	// bus, err := events.NewNATSBus(config)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer bus.Close()
	//
	// // Subscribe on one service
	// bus.Subscribe(ctx, "order.created", func(e events.Event) {
	// 	log.Printf("Order created: %v", e.Payload)
	// })
	//
	// // Publish from another service
	// event := events.NewEvent("order.created", map[string]any{
	// 	"order_id": "order-123",
	// 	"total":    99.99,
	// })
	// bus.Publish(ctx, event)
}

// Example demonstrates event store for audit trail
func Example_eventStore() {
	// ctx := context.Background()
	// db, _ := sql.Open("postgres", "...")
	// bus := events.NewInMemoryBus()

	// store, err := events.NewPostgresEventStore(db, bus)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // Append events
	// event := events.NewEvent("order.created", orderData)
	// store.Append(ctx, event)

	// // Load historical events
	// criteria := events.EventCriteria{
	// 	EventTypes: []string{"order.created"},
	// 	Since:      time.Now().AddDate(0, -1, 0),
	// 	Limit:      100,
	// }

	// events, err := store.Load(ctx, criteria)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Printf("Found %d events", len(events))
}

// Example demonstrates wildcard subscription
func Example_wildcardSubscription() {
	ctx := context.Background()
	bus := events.NewInMemoryBus()
	defer bus.Close()

	// Subscribe to all events
	bus.Subscribe(ctx, "*", func(e events.Event) {
		log.Printf("[ALL] %s: %v", e.Type, e.Payload)
	})

	// Subscribe to specific events
	bus.Subscribe(ctx, "user.created", func(e events.Event) {
		log.Printf("[USER] New user: %v", e.Payload)
	})

	// Publish events
	bus.Publish(ctx, events.NewEvent("user.created", "user-data"))
	bus.Publish(ctx, events.NewEvent("order.created", "order-data"))
	bus.Publish(ctx, events.NewEvent("payment.received", "payment-data"))
}

// Example demonstrates metadata usage
func Example_metadata() {
	ctx := context.Background()
	bus := events.NewInMemoryBus()
	defer bus.Close()

	metadata := events.Metadata{
		UserID:    "user-123",
		TenantID:  "tenant-456",
		RequestID: "req-789",
		TraceID:   "trace-abc",
		Custom: map[string]any{
			"ip_address": "192.168.1.1",
			"user_agent": "Mozilla/5.0...",
		},
	}

	event := events.NewEventWithMetadata(
		events.EventUserAuthenticated,
		map[string]string{"username": "john"},
		metadata,
	)

	bus.Publish(ctx, event)
}

// Example demonstrates error publishing
func Example_errorPublishing() {
	ctx := context.Background()
	bus := events.NewInMemoryBus()
	defer bus.Close()

	bus.Subscribe(ctx, "error", func(e events.Event) {
		payload := e.Payload.(map[string]any)
		log.Printf("Error from %s: %s", payload["source"], payload["error"])
	})

	// Publish error event
	err := events.PublishError(
		ctx,
		bus,
		"payment-service",
		events.ErrBusClosed,
		events.Metadata{UserID: "user-123"},
	)

	if err != nil {
		log.Fatal(err)
	}
}

// Example demonstrates saga pattern
func Example_sagaPattern() {
	ctx := context.Background()
	bus := events.NewInMemoryBus()
	defer bus.Close()

	correlationID := "saga-123"

	// Step 1: Reserve inventory
	bus.Subscribe(ctx, "inventory.reserved", func(e events.Event) {
		if e.CorrelationID != correlationID {
			return
		}

		// Step 2: Process payment
		payment := events.NewEvent("payment.process", "payment-data").
			WithCorrelation(correlationID).
			WithCausation(e.ID)
		bus.Publish(ctx, payment)
	})

	// Step 2 handler
	bus.Subscribe(ctx, "payment.processed", func(e events.Event) {
		if e.CorrelationID != correlationID {
			return
		}

		// Step 3: Confirm order
		confirm := events.NewEvent("order.confirm", "order-data").
			WithCorrelation(correlationID).
			WithCausation(e.ID)
		bus.Publish(ctx, confirm)
	})

	// Start saga
	startEvent := events.NewEvent("inventory.reserve", "inventory-data").
		WithCorrelation(correlationID)
	bus.Publish(ctx, startEvent)
}
