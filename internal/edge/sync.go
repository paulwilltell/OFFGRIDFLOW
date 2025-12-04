// Package edge provides offline-first edge computing with sync capabilities.
//
// This package enables the system to operate without internet connectivity,
// storing data locally and syncing when connection is restored.
package edge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// =============================================================================
// Connectivity State
// =============================================================================

// ConnectivityState represents the current network state.
type ConnectivityState string

const (
	StateOnline     ConnectivityState = "online"
	StateOffline    ConnectivityState = "offline"
	StateConnecting ConnectivityState = "connecting"
	StateSyncing    ConnectivityState = "syncing"
)

// =============================================================================
// Sync Queue
// =============================================================================

// SyncOperation represents a queued operation.
type SyncOperation struct {
	ID          string          `json:"id"`
	Type        OperationType   `json:"type"`
	Resource    string          `json:"resource"`
	Data        json.RawMessage `json:"data"`
	TenantID    string          `json:"tenantId"`
	CreatedAt   time.Time       `json:"createdAt"`
	Retries     int             `json:"retries"`
	LastError   string          `json:"lastError,omitempty"`
	Priority    int             `json:"priority"`
	Idempotency string          `json:"idempotency,omitempty"`
}

// OperationType defines the type of operation.
type OperationType string

const (
	OpCreate OperationType = "create"
	OpUpdate OperationType = "update"
	OpDelete OperationType = "delete"
	OpSync   OperationType = "sync"
)

// SyncQueue manages pending operations.
type SyncQueue struct {
	operations []SyncOperation
	store      QueueStore
	logger     *slog.Logger
	mu         sync.RWMutex
}

// QueueStore persists the sync queue.
type QueueStore interface {
	// Enqueue adds an operation to the queue
	Enqueue(ctx context.Context, op SyncOperation) error

	// Dequeue removes and returns the next operation
	Dequeue(ctx context.Context) (*SyncOperation, error)

	// Peek returns the next operation without removing
	Peek(ctx context.Context) (*SyncOperation, error)

	// GetAll returns all queued operations
	GetAll(ctx context.Context) ([]SyncOperation, error)

	// Remove removes an operation by ID
	Remove(ctx context.Context, id string) error

	// UpdateRetry updates retry count and error
	UpdateRetry(ctx context.Context, id string, retries int, err string) error
}

// NewSyncQueue creates a new sync queue.
func NewSyncQueue(store QueueStore, logger *slog.Logger) *SyncQueue {
	return &SyncQueue{
		operations: make([]SyncOperation, 0),
		store:      store,
		logger:     logger.With("component", "sync-queue"),
	}
}

// Enqueue adds an operation to the queue.
func (sq *SyncQueue) Enqueue(op SyncOperation) error {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	if op.ID == "" {
		op.ID = fmt.Sprintf("op-%d", time.Now().UnixNano())
	}
	if op.CreatedAt.IsZero() {
		op.CreatedAt = time.Now()
	}

	sq.operations = append(sq.operations, op)

	// Sort by priority (lower = higher priority)
	sq.sortByPriority()

	if sq.store != nil {
		return sq.store.Enqueue(context.Background(), op)
	}

	sq.logger.Debug("operation enqueued",
		"operationId", op.ID,
		"type", op.Type,
		"resource", op.Resource)

	return nil
}

// sortByPriority sorts operations by priority.
func (sq *SyncQueue) sortByPriority() {
	for i := 0; i < len(sq.operations)-1; i++ {
		for j := i + 1; j < len(sq.operations); j++ {
			if sq.operations[j].Priority < sq.operations[i].Priority {
				sq.operations[i], sq.operations[j] = sq.operations[j], sq.operations[i]
			}
		}
	}
}

// Dequeue removes and returns the next operation.
func (sq *SyncQueue) Dequeue() (*SyncOperation, error) {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	if len(sq.operations) == 0 {
		return nil, nil
	}

	op := sq.operations[0]
	sq.operations = sq.operations[1:]

	if sq.store != nil {
		sq.store.Remove(context.Background(), op.ID)
	}

	return &op, nil
}

// Size returns the number of queued operations.
func (sq *SyncQueue) Size() int {
	sq.mu.RLock()
	defer sq.mu.RUnlock()
	return len(sq.operations)
}

// GetPending returns all pending operations.
func (sq *SyncQueue) GetPending() []SyncOperation {
	sq.mu.RLock()
	defer sq.mu.RUnlock()

	result := make([]SyncOperation, len(sq.operations))
	copy(result, sq.operations)
	return result
}

// =============================================================================
// Edge Sync Manager
// =============================================================================

// SyncManager handles offline/online sync.
type SyncManager struct {
	queue    *SyncQueue
	state    ConnectivityState
	handlers map[string]SyncHandler
	logger   *slog.Logger
	mu       sync.RWMutex

	// Configuration
	maxRetries int
	retryDelay time.Duration
	batchSize  int

	// Callbacks
	onStateChange  func(state ConnectivityState)
	onSyncComplete func(stats SyncStats)
}

// SyncHandler processes sync operations.
type SyncHandler interface {
	// Handle processes an operation, returns error if should retry
	Handle(ctx context.Context, op SyncOperation) error

	// CanHandle returns true if handler can process this resource
	CanHandle(resource string) bool
}

// SyncManagerConfig configures the sync manager.
type SyncManagerConfig struct {
	Queue      *SyncQueue
	MaxRetries int
	RetryDelay time.Duration
	BatchSize  int
	Logger     *slog.Logger
}

// NewSyncManager creates a new sync manager.
func NewSyncManager(cfg SyncManagerConfig) *SyncManager {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 5
	}
	if cfg.RetryDelay <= 0 {
		cfg.RetryDelay = 30 * time.Second
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 50
	}

	return &SyncManager{
		queue:      cfg.Queue,
		state:      StateOffline,
		handlers:   make(map[string]SyncHandler),
		logger:     cfg.Logger.With("component", "sync-manager"),
		maxRetries: cfg.MaxRetries,
		retryDelay: cfg.RetryDelay,
		batchSize:  cfg.BatchSize,
	}
}

// RegisterHandler registers a sync handler.
func (sm *SyncManager) RegisterHandler(name string, handler SyncHandler) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.handlers[name] = handler
}

// OnStateChange sets the state change callback.
func (sm *SyncManager) OnStateChange(fn func(ConnectivityState)) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.onStateChange = fn
}

// OnSyncComplete sets the sync complete callback.
func (sm *SyncManager) OnSyncComplete(fn func(SyncStats)) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.onSyncComplete = fn
}

// SetState updates connectivity state.
func (sm *SyncManager) SetState(state ConnectivityState) {
	sm.mu.Lock()
	oldState := sm.state
	sm.state = state
	callback := sm.onStateChange
	sm.mu.Unlock()

	if oldState != state {
		sm.logger.Info("connectivity state changed",
			"from", oldState,
			"to", state)

		if callback != nil {
			callback(state)
		}

		// Trigger sync when coming online
		if state == StateOnline && oldState == StateOffline {
			go sm.Sync(context.Background())
		}
	}
}

// GetState returns current connectivity state.
func (sm *SyncManager) GetState() ConnectivityState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state
}

// SyncStats tracks sync operation statistics.
type SyncStats struct {
	StartedAt   time.Time     `json:"startedAt"`
	CompletedAt time.Time     `json:"completedAt"`
	Duration    time.Duration `json:"duration"`
	Total       int           `json:"total"`
	Succeeded   int           `json:"succeeded"`
	Failed      int           `json:"failed"`
	Retried     int           `json:"retried"`
}

// Sync processes all queued operations.
func (sm *SyncManager) Sync(ctx context.Context) (*SyncStats, error) {
	sm.mu.Lock()
	if sm.state != StateOnline && sm.state != StateSyncing {
		sm.mu.Unlock()
		return nil, errors.New("not connected")
	}
	sm.state = StateSyncing
	sm.mu.Unlock()

	defer func() {
		sm.mu.Lock()
		sm.state = StateOnline
		sm.mu.Unlock()
	}()

	stats := &SyncStats{StartedAt: time.Now()}
	sm.logger.Info("starting sync",
		"pendingOperations", sm.queue.Size())

	processed := 0
	for processed < sm.batchSize {
		op, err := sm.queue.Dequeue()
		if err != nil {
			return stats, err
		}
		if op == nil {
			break
		}

		stats.Total++
		processed++

		// Find handler
		handler := sm.findHandler(op.Resource)
		if handler == nil {
			sm.logger.Warn("no handler for resource",
				"resource", op.Resource)
			stats.Failed++
			continue
		}

		// Process operation
		if err := handler.Handle(ctx, *op); err != nil {
			sm.logger.Error("operation failed",
				"operationId", op.ID,
				"error", err)

			if op.Retries < sm.maxRetries {
				// Re-queue for retry
				op.Retries++
				op.LastError = err.Error()
				sm.queue.Enqueue(*op)
				stats.Retried++
			} else {
				stats.Failed++
			}
		} else {
			stats.Succeeded++
			sm.logger.Debug("operation succeeded",
				"operationId", op.ID)
		}
	}

	stats.CompletedAt = time.Now()
	stats.Duration = stats.CompletedAt.Sub(stats.StartedAt)

	sm.logger.Info("sync completed",
		"succeeded", stats.Succeeded,
		"failed", stats.Failed,
		"duration", stats.Duration)

	// Notify callback
	sm.mu.RLock()
	callback := sm.onSyncComplete
	sm.mu.RUnlock()
	if callback != nil {
		callback(*stats)
	}

	return stats, nil
}

// findHandler finds a handler for a resource.
func (sm *SyncManager) findHandler(resource string) SyncHandler {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, h := range sm.handlers {
		if h.CanHandle(resource) {
			return h
		}
	}
	return nil
}

// =============================================================================
// Local Storage
// =============================================================================

// LocalCache provides local caching for offline access.
type LocalCache struct {
	data   map[string]CacheEntry
	logger *slog.Logger
	mu     sync.RWMutex
}

// CacheEntry represents a cached item.
type CacheEntry struct {
	Key       string          `json:"key"`
	Data      json.RawMessage `json:"data"`
	Version   int64           `json:"version"`
	CachedAt  time.Time       `json:"cachedAt"`
	ExpiresAt *time.Time      `json:"expiresAt,omitempty"`
	SyncedAt  *time.Time      `json:"syncedAt,omitempty"`
}

// NewLocalCache creates a new local cache.
func NewLocalCache(logger *slog.Logger) *LocalCache {
	return &LocalCache{
		data:   make(map[string]CacheEntry),
		logger: logger.With("component", "local-cache"),
	}
}

// Set stores an item in the cache.
func (lc *LocalCache) Set(key string, data interface{}, ttl time.Duration) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	entry := CacheEntry{
		Key:      key,
		Data:     jsonData,
		Version:  time.Now().UnixNano(),
		CachedAt: time.Now(),
	}

	if ttl > 0 {
		expiresAt := time.Now().Add(ttl)
		entry.ExpiresAt = &expiresAt
	}

	lc.data[key] = entry
	return nil
}

// Get retrieves an item from the cache.
func (lc *LocalCache) Get(key string, v interface{}) error {
	lc.mu.RLock()
	entry, ok := lc.data[key]
	lc.mu.RUnlock()

	if !ok {
		return errors.New("not found")
	}

	// Check expiration
	if entry.ExpiresAt != nil && time.Now().After(*entry.ExpiresAt) {
		lc.mu.Lock()
		delete(lc.data, key)
		lc.mu.Unlock()
		return errors.New("expired")
	}

	return json.Unmarshal(entry.Data, v)
}

// Delete removes an item from the cache.
func (lc *LocalCache) Delete(key string) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	delete(lc.data, key)
}

// GetModified returns all entries modified since a timestamp.
func (lc *LocalCache) GetModified(since time.Time) []CacheEntry {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	var result []CacheEntry
	for _, entry := range lc.data {
		if entry.CachedAt.After(since) {
			result = append(result, entry)
		}
	}
	return result
}

// =============================================================================
// Conflict Resolution
// =============================================================================

// ConflictResolver handles sync conflicts.
type ConflictResolver struct {
	strategy ConflictStrategy
	logger   *slog.Logger
}

// ConflictStrategy defines how to resolve conflicts.
type ConflictStrategy string

const (
	StrategyServerWins ConflictStrategy = "server_wins"
	StrategyClientWins ConflictStrategy = "client_wins"
	StrategyLastWrite  ConflictStrategy = "last_write"
	StrategyMerge      ConflictStrategy = "merge"
)

// Conflict represents a sync conflict.
type Conflict struct {
	Key        string          `json:"key"`
	LocalData  json.RawMessage `json:"localData"`
	ServerData json.RawMessage `json:"serverData"`
	LocalTime  time.Time       `json:"localTime"`
	ServerTime time.Time       `json:"serverTime"`
}

// Resolution contains the resolved data.
type Resolution struct {
	Key    string          `json:"key"`
	Data   json.RawMessage `json:"data"`
	Source string          `json:"source"` // "local", "server", "merged"
}

// NewConflictResolver creates a new conflict resolver.
func NewConflictResolver(strategy ConflictStrategy, logger *slog.Logger) *ConflictResolver {
	return &ConflictResolver{
		strategy: strategy,
		logger:   logger.With("component", "conflict-resolver"),
	}
}

// Resolve resolves a conflict based on strategy.
func (cr *ConflictResolver) Resolve(conflict Conflict) (*Resolution, error) {
	cr.logger.Info("resolving conflict",
		"key", conflict.Key,
		"strategy", cr.strategy)

	switch cr.strategy {
	case StrategyServerWins:
		return &Resolution{
			Key:    conflict.Key,
			Data:   conflict.ServerData,
			Source: "server",
		}, nil

	case StrategyClientWins:
		return &Resolution{
			Key:    conflict.Key,
			Data:   conflict.LocalData,
			Source: "local",
		}, nil

	case StrategyLastWrite:
		if conflict.LocalTime.After(conflict.ServerTime) {
			return &Resolution{
				Key:    conflict.Key,
				Data:   conflict.LocalData,
				Source: "local",
			}, nil
		}
		return &Resolution{
			Key:    conflict.Key,
			Data:   conflict.ServerData,
			Source: "server",
		}, nil

	case StrategyMerge:
		// For merge, attempt to combine both versions
		merged, err := cr.merge(conflict)
		if err != nil {
			// Fall back to last write
			if conflict.LocalTime.After(conflict.ServerTime) {
				return &Resolution{
					Key:    conflict.Key,
					Data:   conflict.LocalData,
					Source: "local",
				}, nil
			}
			return &Resolution{
				Key:    conflict.Key,
				Data:   conflict.ServerData,
				Source: "server",
			}, nil
		}
		return merged, nil

	default:
		return nil, fmt.Errorf("unknown strategy: %s", cr.strategy)
	}
}

// merge attempts to merge two JSON documents.
func (cr *ConflictResolver) merge(conflict Conflict) (*Resolution, error) {
	var local, server map[string]interface{}

	if err := json.Unmarshal(conflict.LocalData, &local); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(conflict.ServerData, &server); err != nil {
		return nil, err
	}

	// Simple merge: server base with local overrides
	merged := make(map[string]interface{})
	for k, v := range server {
		merged[k] = v
	}
	for k, v := range local {
		merged[k] = v
	}

	data, err := json.Marshal(merged)
	if err != nil {
		return nil, err
	}

	return &Resolution{
		Key:    conflict.Key,
		Data:   data,
		Source: "merged",
	}, nil
}

// =============================================================================
// Edge Client
// =============================================================================

// Client provides the main edge computing interface.
type Client struct {
	syncManager *SyncManager
	cache       *LocalCache
	resolver    *ConflictResolver
	logger      *slog.Logger
}

// ClientConfig configures the edge client.
type ClientConfig struct {
	SyncQueue        *SyncQueue
	ConflictStrategy ConflictStrategy
	Logger           *slog.Logger
}

// NewClient creates a new edge client.
func NewClient(cfg ClientConfig) *Client {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	return &Client{
		syncManager: NewSyncManager(SyncManagerConfig{
			Queue:  cfg.SyncQueue,
			Logger: cfg.Logger,
		}),
		cache:    NewLocalCache(cfg.Logger),
		resolver: NewConflictResolver(cfg.ConflictStrategy, cfg.Logger),
		logger:   cfg.Logger.With("component", "edge-client"),
	}
}

// Store stores data locally and queues for sync.
func (c *Client) Store(resource, key string, data interface{}) error {
	// Cache locally
	if err := c.cache.Set(key, data, 0); err != nil {
		return err
	}

	// Queue for sync
	jsonData, _ := json.Marshal(data)
	return c.syncManager.queue.Enqueue(SyncOperation{
		Type:        OpCreate,
		Resource:    resource,
		Data:        jsonData,
		CreatedAt:   time.Now(),
		Idempotency: key,
	})
}

// Load retrieves data from local cache.
func (c *Client) Load(key string, v interface{}) error {
	return c.cache.Get(key, v)
}

// Sync triggers a sync cycle.
func (c *Client) Sync(ctx context.Context) (*SyncStats, error) {
	return c.syncManager.Sync(ctx)
}

// GetState returns connectivity state.
func (c *Client) GetState() ConnectivityState {
	return c.syncManager.GetState()
}

// SetOnline marks the client as online.
func (c *Client) SetOnline() {
	c.syncManager.SetState(StateOnline)
}

// SetOffline marks the client as offline.
func (c *Client) SetOffline() {
	c.syncManager.SetState(StateOffline)
}
