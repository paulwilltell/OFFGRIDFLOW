//go:build events_redis
// +build events_redis

package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisBus implements the Bus interface using Redis Pub/Sub or Streams.
// It provides lightweight distributed messaging suitable for smaller deployments.
type RedisBus struct {
	client     redis.UniversalClient
	pubsub     *redis.PubSub
	mu         sync.RWMutex
	subs       map[string]context.CancelFunc
	closed     bool
	config     RedisConfig
	wg         sync.WaitGroup
	useStreams bool
}

// RedisConfig configures the Redis event bus.
type RedisConfig struct {
	// Addrs are Redis server addresses
	Addrs []string

	// Password for Redis authentication
	Password string

	// DB is the Redis database number
	DB int

	// UseCluster enables Redis Cluster mode
	UseCluster bool

	// UseStreams uses Redis Streams instead of Pub/Sub for persistence
	UseStreams bool

	// StreamName is the stream key for events
	StreamName string

	// ConsumerGroup is the consumer group name
	ConsumerGroup string

	// ConsumerName is this consumer's name
	ConsumerName string

	// MaxLen is the maximum stream length (0 = unlimited)
	MaxLen int64

	// BlockTime is how long to block when reading streams
	BlockTime time.Duration

	// BatchSize is the number of messages to read per batch
	BatchSize int64
}

// DefaultRedisConfig returns a configuration with sensible defaults.
func DefaultRedisConfig() RedisConfig {
	return RedisConfig{
		Addrs:         []string{"localhost:6379"},
		DB:            0,
		UseStreams:    true,
		StreamName:    "events",
		ConsumerGroup: "offgridflow",
		ConsumerName:  "worker-1",
		MaxLen:        10000,
		BlockTime:     1 * time.Second,
		BatchSize:     10,
	}
}

// NewRedisBus creates a new Redis-based event bus.
func NewRedisBus(ctx context.Context, config RedisConfig) (*RedisBus, error) {
	var client redis.UniversalClient

	if config.UseCluster {
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    config.Addrs,
			Password: config.Password,
		})
	} else {
		if len(config.Addrs) == 0 {
			config.Addrs = []string{"localhost:6379"}
		}
		client = redis.NewClient(&redis.Options{
			Addr:     config.Addrs[0],
			Password: config.Password,
			DB:       config.DB,
		})
	}

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	bus := &RedisBus{
		client:     client,
		subs:       make(map[string]context.CancelFunc),
		config:     config,
		useStreams: config.UseStreams,
	}

	if config.UseStreams {
		// Create consumer group
		err := client.XGroupCreateMkStream(ctx, config.StreamName, config.ConsumerGroup, "0").Err()
		if err != nil && !errors.Is(err, redis.Nil) && err.Error() != "BUSYGROUP Consumer Group name already exists" {
			client.Close()
			return nil, fmt.Errorf("create consumer group: %w", err)
		}
	} else {
		// Initialize Pub/Sub
		bus.pubsub = client.Subscribe(ctx)
	}

	return bus, nil
}

// Publish sends an event to Redis.
func (b *RedisBus) Publish(ctx context.Context, event Event) error {
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
		event.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	if b.useStreams {
		// Publish to Redis Stream
		args := &redis.XAddArgs{
			Stream: b.config.StreamName,
			MaxLen: b.config.MaxLen,
			Approx: true,
			Values: map[string]interface{}{
				"type":    event.Type,
				"payload": string(data),
			},
		}

		if _, err := b.client.XAdd(ctx, args).Result(); err != nil {
			return fmt.Errorf("xadd: %w", err)
		}
	} else {
		// Publish to Redis Pub/Sub
		channel := b.getChannel(event.Type)
		if err := b.client.Publish(ctx, channel, data).Err(); err != nil {
			return fmt.Errorf("publish: %w", err)
		}
	}

	return nil
}

// Subscribe registers a handler for events matching the topic.
func (b *RedisBus) Subscribe(ctx context.Context, topic string, handler Handler) error {
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

	// Create cancellable context for this subscription
	subCtx, cancel := context.WithCancel(ctx)
	b.subs[topic] = cancel

	if b.useStreams {
		// Subscribe using Redis Streams
		b.wg.Add(1)
		go b.consumeStream(subCtx, topic, handler)
	} else {
		// Subscribe using Redis Pub/Sub
		channel := b.getChannel(topic)
		if err := b.pubsub.Subscribe(ctx, channel); err != nil {
			cancel()
			return fmt.Errorf("subscribe: %w", err)
		}

		b.wg.Add(1)
		go b.consumePubSub(subCtx, topic, handler)
	}

	return nil
}

// consumeStream processes events from Redis Streams.
func (b *RedisBus) consumeStream(ctx context.Context, topic string, handler Handler) {
	defer b.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Read from stream
		streams, err := b.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    b.config.ConsumerGroup,
			Consumer: b.config.ConsumerName,
			Streams:  []string{b.config.StreamName, ">"},
			Count:    b.config.BatchSize,
			Block:    b.config.BlockTime,
		}).Result()

		if err != nil {
			if errors.Is(err, redis.Nil) || errors.Is(err, context.Canceled) {
				continue
			}
			// Log error
			time.Sleep(time.Second)
			continue
		}

		for _, stream := range streams {
			for _, message := range stream.Messages {
				b.processStreamMessage(ctx, topic, message, handler)
			}
		}
	}
}

// processStreamMessage handles a single stream message.
func (b *RedisBus) processStreamMessage(ctx context.Context, topic string, msg redis.XMessage, handler Handler) {
	payloadStr, ok := msg.Values["payload"].(string)
	if !ok {
		return
	}

	var event Event
	if err := json.Unmarshal([]byte(payloadStr), &event); err != nil {
		return
	}

	// Check if event type matches topic
	if topic != "*" && event.Type != topic {
		// Acknowledge and skip
		b.client.XAck(ctx, b.config.StreamName, b.config.ConsumerGroup, msg.ID)
		return
	}

	// Call handler with panic recovery
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Log panic
			}
		}()
		handler(event)
	}()

	// Acknowledge message
	b.client.XAck(ctx, b.config.StreamName, b.config.ConsumerGroup, msg.ID)
}

// consumePubSub processes events from Redis Pub/Sub.
func (b *RedisBus) consumePubSub(ctx context.Context, topic string, handler Handler) {
	defer b.wg.Done()

	ch := b.pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}

			var event Event
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				continue
			}

			// Call handler with panic recovery
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Log panic
					}
				}()
				handler(event)
			}()
		}
	}
}

// Close shuts down the Redis bus.
func (b *RedisBus) Close() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true

	// Cancel all subscriptions
	for _, cancel := range b.subs {
		cancel()
	}
	b.subs = nil
	b.mu.Unlock()

	// Wait for consumers to finish
	b.wg.Wait()

	// Close pub/sub if using it
	if b.pubsub != nil {
		if err := b.pubsub.Close(); err != nil {
			return err
		}
	}

	// Close Redis client
	return b.client.Close()
}

// getChannel converts an event type to a Redis channel name.
func (b *RedisBus) getChannel(eventType string) string {
	if eventType == "*" {
		return "events:*"
	}
	return "events:" + eventType
}

// Ping checks if the Redis connection is alive.
func (b *RedisBus) Ping(ctx context.Context) error {
	return b.client.Ping(ctx).Err()
}

// Compile-time interface check
var _ Bus = (*RedisBus)(nil)
