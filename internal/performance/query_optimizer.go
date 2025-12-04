package performance

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// QueryOptimizer optimizes database queries
type QueryOptimizer struct {
	db              *sql.DB
	logger          *slog.Logger
	queryStats      map[string]*QueryStats
	queryStatsMutex sync.RWMutex
	indexHints      map[string][]string
	indexMutex      sync.RWMutex
}

// QueryStats tracks query performance
type QueryStats struct {
	Query          string
	ExecutionCount int64
	TotalDuration  time.Duration
	AvgDuration    time.Duration
	MinDuration    time.Duration
	MaxDuration    time.Duration
	ErrorCount     int64
	LastExecuted   time.Time
	QueryHash      string
}

// NewQueryOptimizer creates a new query optimizer
func NewQueryOptimizer(db *sql.DB, logger *slog.Logger) *QueryOptimizer {
	return &QueryOptimizer{
		db:         db,
		logger:     logger,
		queryStats: make(map[string]*QueryStats),
		indexHints: make(map[string][]string),
	}
}

// RegisterIndexHint registers an index hint for a query
func (qo *QueryOptimizer) RegisterIndexHint(queryPattern string, indexes ...string) {
	qo.indexMutex.Lock()
	defer qo.indexMutex.Unlock()
	qo.indexHints[queryPattern] = indexes
}

// OptimizedQuery represents an optimized query execution
type OptimizedQuery struct {
	optimizer      *QueryOptimizer
	query          string
	args           []interface{}
	indexHints     []string
	startTime      time.Time
	executionPlan  *ExecutionPlan
	cacheStrategy  CacheStrategy
	batchSize      int
}

// CacheStrategy defines caching behavior
type CacheStrategy int

const (
	CacheStrategyNone CacheStrategy = iota
	CacheStrategyAlways
	CacheStrategyTTL
	CacheStrategyConditional
)

// ExecutionPlan represents a query execution plan
type ExecutionPlan struct {
	Query              string
	EstimatedCost      float64
	EstimatedRows      int64
	Indexes            []string
	JoinOrder          []string
	FilterPushdown      bool
	ParallelExecution  bool
	RecommendedBatches int
}

// NewOptimizedQuery creates a new optimized query
func (qo *QueryOptimizer) NewOptimizedQuery(query string, args ...interface{}) *OptimizedQuery {
	return &OptimizedQuery{
		optimizer:     qo,
		query:         query,
		args:          args,
		startTime:     time.Now(),
		cacheStrategy: CacheStrategyNone,
		batchSize:     1000,
	}
}

// WithCacheStrategy sets the cache strategy
func (oq *OptimizedQuery) WithCacheStrategy(strategy CacheStrategy) *OptimizedQuery {
	oq.cacheStrategy = strategy
	return oq
}

// WithBatchSize sets the batch size for bulk operations
func (oq *OptimizedQuery) WithBatchSize(size int) *OptimizedQuery {
	oq.batchSize = size
	return oq
}

// GenerateQueryHash generates a hash for the query
func (oq *OptimizedQuery) GenerateQueryHash() string {
	hash := md5.Sum([]byte(oq.query))
	return fmt.Sprintf("%x", hash)
}

// AnalyzeExecutionPlan analyzes the query execution plan
func (oq *OptimizedQuery) AnalyzeExecutionPlan() *ExecutionPlan {
	plan := &ExecutionPlan{
		Query:          oq.query,
		EstimatedCost:  0.0,
		EstimatedRows:  0,
		Indexes:        oq.indexHints,
		FilterPushdown:  true,
		ParallelExecution: false,
	}

	// Analyze query characteristics
	if len(oq.args) > 100 {
		plan.ParallelExecution = true
		plan.RecommendedBatches = (len(oq.args) / 1000) + 1
	}

	oq.executionPlan = plan
	return plan
}

// RecordQueryStats records query performance statistics
func (oq *OptimizedQuery) RecordQueryStats(duration time.Duration, err error) {
	oq.optimizer.queryStatsMutex.Lock()
	defer oq.optimizer.queryStatsMutex.Unlock()

	stats, exists := oq.optimizer.queryStats[oq.query]
	if !exists {
		stats = &QueryStats{
			Query:        oq.query,
			QueryHash:    oq.GenerateQueryHash(),
			MinDuration:  duration,
			MaxDuration:  duration,
		}
	}

	stats.ExecutionCount++
	stats.TotalDuration += duration
	stats.AvgDuration = stats.TotalDuration / time.Duration(stats.ExecutionCount)
	stats.LastExecuted = time.Now()

	if duration < stats.MinDuration {
		stats.MinDuration = duration
	}
	if duration > stats.MaxDuration {
		stats.MaxDuration = duration
	}

	if err != nil {
		stats.ErrorCount++
	}

	oq.optimizer.queryStats[oq.query] = stats
}

// GetQueryStats returns statistics for all tracked queries
func (qo *QueryOptimizer) GetQueryStats() map[string]*QueryStats {
	qo.queryStatsMutex.RLock()
	defer qo.queryStatsMutex.RUnlock()

	// Return a copy
	stats := make(map[string]*QueryStats)
	for k, v := range qo.queryStats {
		stats[k] = v
	}
	return stats
}

// GetSlowQueries returns queries that exceed the duration threshold
func (qo *QueryOptimizer) GetSlowQueries(threshold time.Duration) []*QueryStats {
	qo.queryStatsMutex.RLock()
	defer qo.queryStatsMutex.RUnlock()

	var slowQueries []*QueryStats
	for _, stats := range qo.queryStats {
		if stats.AvgDuration > threshold {
			slowQueries = append(slowQueries, stats)
		}
	}
	return slowQueries
}

// RecommendIndexes provides index recommendations
func (qo *QueryOptimizer) RecommendIndexes() map[string][]string {
	qo.queryStatsMutex.RLock()
	defer qo.queryStatsMutex.RUnlock()

	recommendations := make(map[string][]string)

	for query, stats := range qo.queryStats {
		if stats.AvgDuration > 100*time.Millisecond && stats.ExecutionCount > 10 {
			recommendations[query] = []string{
				"Consider adding indexes on frequently filtered columns",
				"Review JOIN conditions for optimization opportunities",
				"Consider denormalization for frequently accessed data",
			}
		}
	}

	return recommendations
}

// BatchedQuery represents a batched query for bulk operations
type BatchedQuery struct {
	query     string
	batchSize int
	totalRows int64
	metrics   *BatchMetrics
}

// BatchMetrics tracks batch operation metrics
type BatchMetrics struct {
	TotalRows       int64
	ProcessedRows   int64
	SkippedRows     int64
	ErrorRows       int64
	TotalDuration   time.Duration
	AvgRowDuration  time.Duration
	ThroughputRows  float64 // rows per second
	BatchCount      int64
	AvgBatchSize    int
}

// NewBatchedQuery creates a new batched query
func (qo *QueryOptimizer) NewBatchedQuery(query string, batchSize int) *BatchedQuery {
	return &BatchedQuery{
		query:     query,
		batchSize: batchSize,
		metrics: &BatchMetrics{
			AvgBatchSize: batchSize,
		},
	}
}

// ExecuteBatched executes a query in batches
func (bq *BatchedQuery) ExecuteBatched(
	executor func(batch []interface{}) error,
	totalRows int64,
	data []interface{},
) error {
	start := time.Now()
	bq.totalRows = totalRows
	bq.metrics.BatchCount = (totalRows / int64(bq.batchSize)) + 1

	for i := 0; i < len(data); i += bq.batchSize {
		end := i + bq.batchSize
		if end > len(data) {
			end = len(data)
		}

		batch := data[i:end]
		if err := executor(batch); err != nil {
			bq.metrics.ErrorRows += int64(len(batch))
		} else {
			bq.metrics.ProcessedRows += int64(len(batch))
		}
	}

	duration := time.Since(start)
	bq.metrics.TotalDuration = duration
	if bq.metrics.ProcessedRows > 0 {
		bq.metrics.AvgRowDuration = duration / time.Duration(bq.metrics.ProcessedRows)
		bq.metrics.ThroughputRows = float64(bq.metrics.ProcessedRows) / duration.Seconds()
	}

	return nil
}

// GetMetrics returns batch operation metrics
func (bq *BatchedQuery) GetMetrics() *BatchMetrics {
	return bq.metrics
}

// ConnectionPool manages database connections
type ConnectionPool struct {
	db              *sql.DB
	maxConnections  int
	maxIdleConns    int
	connMaxLifetime time.Duration
	logger          *slog.Logger
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(db *sql.DB, logger *slog.Logger) *ConnectionPool {
	return &ConnectionPool{
		db:              db,
		maxConnections:  25,
		maxIdleConns:    5,
		connMaxLifetime: 5 * time.Minute,
		logger:          logger,
	}
}

// Configure configures the connection pool
func (cp *ConnectionPool) Configure(maxConns, maxIdleConns int, maxLifetime time.Duration) {
	cp.maxConnections = maxConns
	cp.maxIdleConns = maxIdleConns
	cp.connMaxLifetime = maxLifetime

	cp.db.SetMaxOpenConns(maxConns)
	cp.db.SetMaxIdleConns(maxIdleConns)
	cp.db.SetConnMaxLifetime(maxLifetime)

	cp.logger.Info("connection pool configured",
		slog.Int("max_connections", maxConns),
		slog.Int("max_idle", maxIdleConns),
		slog.Duration("max_lifetime", maxLifetime),
	)
}

// GetStats returns connection pool statistics
func (cp *ConnectionPool) GetStats() sql.DBStats {
	return cp.db.Stats()
}
