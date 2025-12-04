//go:build events_nats
// +build events_nats

package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// NATSBus implements the Bus interface using NATS for distributed messaging.
// It provides durable, scalable event distribution across multiple services.
type NATSBus struct {
	nc      *nats.Conn
	js      nats.JetStreamContext
	mu      sync.RWMutex
	subs    map[string]*nats.Subscription
	closed  bool
	config  NATSConfig
	errChan chan error
}

// NATSConfig configures the NATS event bus.
type NATSConfig struct {
	// URL is the NATS server URL (e.g., "nats://localhost:4222")
	URL string

	// StreamName is the JetStream stream name for events
	StreamName string

	// DurableName is the consumer durable name for persistence
	DurableName string

	// MaxReconnects is the maximum number of reconnection attempts
	MaxReconnects int

	// ReconnectWait is the time to wait between reconnection attempts
	ReconnectWait time.Duration

	// EnableJetStream enables JetStream for persistent messaging
	EnableJetStream bool

	// StreamRetention defines the stream retention policy
	StreamRetention nats.RetentionPolicy

	// StreamMaxAge is the maximum age of messages in the stream
	StreamMaxAge time.Duration

	// StreamMaxBytes is the maximum size of the stream in bytes
	StreamMaxBytes int64

	// MaxInFlight is the maximum number of unacknowledged messages
	MaxInFlight int

	// AckWait is the time to wait for message acknowledgment
	AckWait time.Duration
}

// DefaultNATSConfig returns a configuration with sensible defaults.
func DefaultNATSConfig() NATSConfig {
	return NATSConfig{
		URL:             nats.DefaultURL,
		StreamName:      "EVENTS",
		DurableName:     "offgridflow",
		MaxReconnects:   10,
		ReconnectWait:   2 * time.Second,
		EnableJetStream: true,
		StreamRetention: nats.InterestPolicy,
		StreamMaxAge:    24 * time.Hour,
		StreamMaxBytes:  10 * 1024 * 1024 * 1024, // 10GB
		MaxInFlight:     1000,
		AckWait:         30 * time.Second,
	}
}

// NewNATSBus creates a new NATS-based event bus.
func NewNATSBus(config NATSConfig) (*NATSBus, error) {
	if config.URL == "" {
		config.URL = nats.DefaultURL
	}

	opts := []nats.Option{
		nats.Name("OffGridFlow Event Bus"),
		nats.MaxReconnects(config.MaxReconnects),
		nats.ReconnectWait(config.ReconnectWait),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				// Log disconnection error
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			// Log reconnection
		}),
	}

	nc, err := nats.Connect(config.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}

	bus := &NATSBus{
		nc:      nc,
		subs:    make(map[string]*nats.Subscription),
		config:  config,
		errChan: make(chan error, 100),
	}

	if config.EnableJetStream {
		js, err := nc.JetStream()
		if err != nil {
			nc.Close()
			return nil, fmt.Errorf("jetstream context: %w", err)
		}
		bus.js = js

		// Create or update stream
		if err := bus.ensureStream(); err != nil {
			nc.Close()
			return nil, fmt.Errorf("ensure stream: %w", err)
		}
	}

	return bus, nil
}

// ensureStream creates or updates the JetStream stream.
func (b *NATSBus) ensureStream() error {
	streamConfig := &nats.StreamConfig{
		Name:      b.config.StreamName,
		Subjects:  []string{b.config.StreamName + ".*"},
		Retention: b.config.StreamRetention,
		MaxAge:    b.config.StreamMaxAge,
		MaxBytes:  b.config.StreamMaxBytes,
		Storage:   nats.FileStorage,
		Replicas:  1,
		Discard:   nats.DiscardOld,
	}

	_, err := b.js.StreamInfo(b.config.StreamName)
	if err != nil {
		if err == nats.ErrStreamNotFound {
			// Create new stream
			_, err = b.js.AddStream(streamConfig)
			return err
		}
		return err
	}

	// Update existing stream
	_, err = b.js.UpdateStream(streamConfig)
	return err
}

// Publish sends an event to NATS.
func (b *NATSBus) Publish(ctx context.Context, event Event) error {
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
		event.ID = nats.NewInbox()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	subject := b.getSubject(event.Type)

	if b.config.EnableJetStream && b.js != nil {
		// Publish with JetStream for persistence
		_, err = b.js.Publish(subject, data)
		if err != nil {
			return fmt.Errorf("jetstream publish: %w", err)
		}
	} else {
		// Simple publish without persistence
		if err := b.nc.Publish(subject, data); err != nil {
			return fmt.Errorf("publish: %w", err)
		}
	}

	return nil
}

// Subscribe registers a handler for events matching the topic.
func (b *NATSBus) Subscribe(ctx context.Context, topic string, handler Handler) error {
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

	subject := b.getSubject(topic)

	// Wrap handler to unmarshal events
	msgHandler := func(msg *nats.Msg) {
		var event Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			// Log error
			select {
			case b.errChan <- fmt.Errorf("unmarshal event: %w", err):
			default:
			}
			return
		}

		// Call user handler with panic recovery
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Log panic
				}
			}()
			handler(event)
		}()

		// Acknowledge message if using JetStream
		if msg.Reply != "" {
			msg.Ack()
		}
	}

	var sub *nats.Subscription
	var err error

	if b.config.EnableJetStream && b.js != nil {
		// Durable subscription with JetStream
		sub, err = b.js.Subscribe(subject, msgHandler, nats.Durable(b.config.DurableName+"-"+topic),
			nats.ManualAck(),
			nats.AckWait(b.config.AckWait),
			nats.MaxAckPending(b.config.MaxInFlight),
		)
	} else {
		// Simple subscription
		sub, err = b.nc.Subscribe(subject, msgHandler)
	}

	if err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}

	b.subs[topic] = sub
	return nil
}

// Close shuts down the NATS connection.
func (b *NATSBus) Close() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true

	// Unsubscribe all
	for _, sub := range b.subs {
		sub.Unsubscribe()
	}
	b.subs = nil
	b.mu.Unlock()

	// Drain and close connection
	if err := b.nc.Drain(); err != nil {
		b.nc.Close()
		return err
	}

	close(b.errChan)
	return nil
}

// getSubject converts an event type to a NATS subject.
func (b *NATSBus) getSubject(eventType string) string {
	if eventType == "*" {
		return b.config.StreamName + ".>"
	}
	return b.config.StreamName + "." + eventType
}

// Errors returns a channel for asynchronous error notifications.
func (b *NATSBus) Errors() <-chan error {
	return b.errChan
}

// IsConnected returns true if the NATS connection is active.
func (b *NATSBus) IsConnected() bool {
	return b.nc.IsConnected()
}

// Status returns the current NATS connection status.
func (b *NATSBus) Status() nats.Status {
	return b.nc.Status()
}

// Compile-time interface check
var _ Bus = (*NATSBus)(nil)
