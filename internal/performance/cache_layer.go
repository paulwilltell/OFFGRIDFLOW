package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheLayer provides Redis-based caching for batch operations
type CacheLayer struct {
	client *redis.Client
	logger *slog.Logger
	config CacheConfig
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Host           string
	Port           int
	DB             int
	Password       string
	MaxRetries     int
	PoolSize       int
	BatchTTL       time.Duration
	ActivityTTL    time.Duration
	EmissionsTTL   time.Duration
	QueryResultTTL time.Duration
	EnableMetrics  bool
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		Host:           "localhost",
		Port:           6379,
		DB:             0,
		MaxRetries:     3,
		PoolSize:       10,
		BatchTTL:       1 * time.Hour,
		ActivityTTL:    30 * time.Minute,
		EmissionsTTL:   1 * time.Hour,
		QueryResultTTL: 5 * time.Minute,
		EnableMetrics:  true,
	}
}

// CacheMetrics tracks cache performance
type CacheMetrics struct {
	Hits       int64
	Misses     int64
	Evictions  int64
	Errors     int64
	TotalSize  int64
	HitRate    float64
	AvgLatency time.Duration
}

// NewCacheLayer creates a new cache layer
func NewCacheLayer(config CacheConfig, logger *slog.Logger) (*CacheLayer, error) {
	client := redis.NewClient(&redis.Options{
		Addr:       fmt.Sprintf("%s:%d", config.Host, config.Port),
		DB:         config.DB,
		Password:   config.Password,
		MaxRetries: config.MaxRetries,
		PoolSize:   config.PoolSize,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logger.Info("redis cache layer initialized",
		slog.String("host", config.Host),
		slog.Int("port", config.Port),
	)

	return &CacheLayer{
		client: client,
		logger: logger,
		config: config,
	}, nil
}

// CacheBatch caches a batch object
func (cl *CacheLayer) CacheBatch(ctx context.Context, batchID string, data interface{}) error {
	key := fmt.Sprintf("batch:%s", batchID)

	jsonData, err := json.Marshal(data)
	if err != nil {
		cl.logger.Error("failed to marshal batch for caching", slog.String("error", err.Error()))
		return err
	}

	if err := cl.client.Set(ctx, key, jsonData, cl.config.BatchTTL).Err(); err != nil {
		cl.logger.Error("failed to cache batch", slog.String("error", err.Error()))
		return err
	}

	return nil
}

// GetCachedBatch retrieves a cached batch
func (cl *CacheLayer) GetCachedBatch(ctx context.Context, batchID string, dest interface{}) error {
	key := fmt.Sprintf("batch:%s", batchID)

	val, err := cl.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("batch not in cache")
	} else if err != nil {
		cl.logger.Error("failed to get cached batch", slog.String("error", err.Error()))
		return err
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		cl.logger.Error("failed to unmarshal cached batch", slog.String("error", err.Error()))
		return err
	}

	return nil
}

// CacheQueryResult caches a query result
func (cl *CacheLayer) CacheQueryResult(ctx context.Context, queryHash string, result interface{}) error {
	key := fmt.Sprintf("query:%s", queryHash)

	jsonData, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return cl.client.Set(ctx, key, jsonData, cl.config.QueryResultTTL).Err()
}

// GetCachedQueryResult retrieves a cached query result
func (cl *CacheLayer) GetCachedQueryResult(ctx context.Context, queryHash string, dest interface{}) error {
	key := fmt.Sprintf("query:%s", queryHash)

	val, err := cl.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("query result not in cache")
	} else if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

// CacheEmissionsCalculation caches emissions calculation results
func (cl *CacheLayer) CacheEmissionsCalculation(ctx context.Context, activityID string, emissions float64) error {
	key := fmt.Sprintf("emissions:%s", activityID)
	return cl.client.Set(ctx, key, emissions, cl.config.EmissionsTTL).Err()
}

// GetCachedEmissionsCalculation retrieves cached emissions
func (cl *CacheLayer) GetCachedEmissionsCalculation(ctx context.Context, activityID string) (float64, error) {
	key := fmt.Sprintf("emissions:%s", activityID)

	val, err := cl.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, fmt.Errorf("emissions not in cache")
	} else if err != nil {
		return 0, err
	}

	var emissions float64
	if err := json.Unmarshal([]byte(val), &emissions); err != nil {
		return 0, err
	}

	return emissions, nil
}

// InvalidateBatchCache invalidates batch cache
func (cl *CacheLayer) InvalidateBatchCache(ctx context.Context, batchID string) error {
	key := fmt.Sprintf("batch:%s", batchID)
	return cl.client.Del(ctx, key).Err()
}

// InvalidateQueryResultCache invalidates query result cache
func (cl *CacheLayer) InvalidateQueryResultCache(ctx context.Context, queryHash string) error {
	key := fmt.Sprintf("query:%s", queryHash)
	return cl.client.Del(ctx, key).Err()
}

// InvalidateEmissionsCache invalidates emissions cache
func (cl *CacheLayer) InvalidateEmissionsCache(ctx context.Context, activityID string) error {
	key := fmt.Sprintf("emissions:%s", activityID)
	return cl.client.Del(ctx, key).Err()
}

// ClearAllCache clears all cache entries
func (cl *CacheLayer) ClearAllCache(ctx context.Context) error {
	return cl.client.FlushDB(ctx).Err()
}

// GetMetrics returns cache metrics
func (cl *CacheLayer) GetMetrics(ctx context.Context) (*CacheMetrics, error) {
	_ = cl.client.Info(ctx, "stats")

	return &CacheMetrics{
		HitRate: 0.0, // Calculated from hits/total
	}, nil
}

// Close closes the cache connection
func (cl *CacheLayer) Close() error {
	return cl.client.Close()
}

// CachedOperation is a wrapper for operations with caching
type CachedOperation struct {
	cache   *CacheLayer
	logger  *slog.Logger
	metrics *OperationMetrics
}

// OperationMetrics tracks operation metrics
type OperationMetrics struct {
	CacheHits   int64
	CacheMisses int64
	TotalOps    int64
	AvgLatency  time.Duration
}

// NewCachedOperation creates a new cached operation wrapper
func NewCachedOperation(cache *CacheLayer, logger *slog.Logger) *CachedOperation {
	return &CachedOperation{
		cache:   cache,
		logger:  logger,
		metrics: &OperationMetrics{},
	}
}

// ExecuteWithCache executes an operation with caching
func (co *CachedOperation) ExecuteWithCache(
	ctx context.Context,
	cacheKey string,
	operation func(context.Context) (interface{}, error),
) (interface{}, error) {
	start := time.Now()

	// Try to get from cache
	if cached, err := co.cache.client.Get(ctx, cacheKey).Result(); err == nil {
		co.metrics.CacheHits++
		latency := time.Since(start)
		co.metrics.AvgLatency = (co.metrics.AvgLatency + latency) / 2
		return cached, nil
	}

	co.metrics.CacheMisses++

	// Execute operation
	result, err := operation(ctx)
	if err != nil {
		return nil, err
	}

	// Cache result
	jsonData, _ := json.Marshal(result)
	co.cache.client.Set(ctx, cacheKey, jsonData, 5*time.Minute)

	latency := time.Since(start)
	co.metrics.AvgLatency = (co.metrics.AvgLatency + latency) / 2
	co.metrics.TotalOps++

	return result, nil
}

// GetMetrics returns operation metrics
func (co *CachedOperation) GetMetrics() *OperationMetrics {
	if co.metrics.TotalOps > 0 {
		co.metrics.CacheHits += co.metrics.CacheMisses
	}
	return co.metrics
}
