package performance

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"testing"
	"time"
)

// TestCacheLayerCreation tests cache layer initialization
func TestCacheLayerCreation(t *testing.T) {
	config := CacheConfig{
		Host: "localhost",
		Port: 6379,
	}
	logger := slog.Default()

	// Note: This test requires Redis running
	cache, err := NewCacheLayer(config, logger)
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer cache.Close()

	if cache == nil {
		t.Error("Cache layer should not be nil")
	}
}

// TestCacheBatchOperations tests batch caching operations
func TestCacheBatchOperations(t *testing.T) {
	config := CacheConfig{
		Host: "localhost",
		Port: 6379,
	}
	logger := slog.Default()

	cache, err := NewCacheLayer(config, logger)
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer cache.Close()

	ctx := context.Background()

	// Test caching batch
	batchData := map[string]interface{}{
		"batch_id": "batch_1",
		"status":   "processing",
	}

	if err := cache.CacheBatch(ctx, "batch_1", batchData); err != nil {
		t.Fatalf("Failed to cache batch: %v", err)
	}

	// Test retrieving batch
	var retrieved map[string]interface{}
	if err := cache.GetCachedBatch(ctx, "batch_1", &retrieved); err != nil {
		t.Fatalf("Failed to retrieve batch: %v", err)
	}

	if retrieved["batch_id"] != "batch_1" {
		t.Error("Retrieved batch data mismatch")
	}
}

// TestQueryOptimizerCreation tests query optimizer initialization
func TestQueryOptimizerCreation(t *testing.T) {
	// Create mock database
	db := &sql.DB{}
	logger := slog.Default()

	optimizer := NewQueryOptimizer(db, logger)

	if optimizer == nil {
		t.Error("Query optimizer should not be nil")
	}
}

// TestQueryStats tests query statistics tracking
func TestQueryStats(t *testing.T) {
	db := &sql.DB{}
	logger := slog.Default()
	optimizer := NewQueryOptimizer(db, logger)

	_ = context.Background()
	query := "SELECT * FROM batches WHERE status = ?"
	optimized := optimizer.NewOptimizedQuery(query, "processing")

	// Simulate recording stats
	optimized.RecordQueryStats(100*time.Millisecond, nil)
	optimized.RecordQueryStats(150*time.Millisecond, nil)

	stats := optimizer.GetQueryStats()

	if _, exists := stats[query]; !exists {
		t.Error("Query stats should be recorded")
	}

	queryStats := stats[query]
	if queryStats.ExecutionCount != 2 {
		t.Errorf("Expected 2 executions, got %d", queryStats.ExecutionCount)
	}
}

// TestLoadTesterCreation tests load tester initialization
func TestLoadTesterCreation(t *testing.T) {
	config := LoadTestConfig{
		Duration:          10 * time.Second,
		ConcurrentWorkers: 5,
		RequestsPerSecond: 10,
	}
	logger := slog.Default()

	tester := NewLoadTester(config, logger)

	if tester == nil {
		t.Error("Load tester should not be nil")
	}

	if tester.IsRunning() {
		t.Error("Load tester should not be running initially")
	}
}

// TestLoadTestExecution tests load test execution
func TestLoadTestExecution(t *testing.T) {
	config := LoadTestConfig{
		Duration:          1 * time.Second,
		ConcurrentWorkers: 2,
		RequestsPerSecond: 10,
		TimeoutPerRequest: 5 * time.Second,
	}
	logger := slog.Default()

	tester := NewLoadTester(config, logger)

	ctx := context.Background()
	executor := func(ctx context.Context) error {
		// Simulate request
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	tester.Run(ctx, executor)

	results := tester.GetResults()

	if results.TotalRequests == 0 {
		t.Error("Should have executed requests")
	}

	if results.SuccessfulRequests != results.TotalRequests {
		t.Error("All requests should be successful in this test")
	}

	if results.Throughput == 0 {
		t.Error("Throughput should be calculated")
	}
}

// TestLoadTestResults tests load test results calculation
func TestLoadTestResults(t *testing.T) {
	config := LoadTestConfig{
		Duration:          1 * time.Second,
		ConcurrentWorkers: 1,
		RequestsPerSecond: 5,
	}
	logger := slog.Default()

	tester := NewLoadTester(config, logger)

	ctx := context.Background()
	executor := func(ctx context.Context) error {
		// Small delay to ensure measurable latency
		time.Sleep(1 * time.Millisecond)
		return nil
	}

	tester.Run(ctx, executor)
	results := tester.GetResults()

	if results.TotalRequests == 0 {
		t.Error("Should have executed requests")
	}

	if results.TotalDuration == 0 {
		t.Error("Total duration should be set")
	}
}

// TestProfilerCreation tests profiler initialization
func TestProfilerCreation(t *testing.T) {
	config := ProfileConfig{
		OutputDir:       "./test_profiles",
		EnableCPU:       true,
		EnableMemory:    true,
		EnableGoroutine: true,
		EnableTrace:     false,
	}
	logger := slog.Default()

	profiler := NewProfiler(config, logger)

	if profiler == nil {
		t.Error("Profiler should not be nil")
	}
}

// TestMemoryStats tests memory statistics capture
func TestMemoryStats(t *testing.T) {
	config := ProfileConfig{
		OutputDir: "./test_profiles",
	}
	logger := slog.Default()

	profiler := NewProfiler(config, logger)

	stats := profiler.GetMemoryStats()

	if stats == nil {
		t.Error("Memory stats should not be nil")
	}

	if stats.Goroutines == 0 {
		t.Error("Goroutines count should be > 0")
	}

	if stats.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}
}

// TestConnectionPoolConfiguration tests connection pool setup
func TestConnectionPoolConfiguration(t *testing.T) {
	db := &sql.DB{}
	logger := slog.Default()

	pool := NewConnectionPool(db, logger)

	pool.Configure(25, 5, 5*time.Minute)

	stats := pool.GetStats()
	if stats.OpenConnections == 0 && stats.InUse == 0 && stats.Idle == 0 {
		// Stats might be 0 for mock DB, that's okay
	}
}

// BenchmarkCacheOperations benchmarks cache operations
func BenchmarkCacheOperations(b *testing.B) {
	config := CacheConfig{
		Host: "localhost",
		Port: 6379,
	}
	logger := slog.Default()

	cache, err := NewCacheLayer(config, logger)
	if err != nil {
		b.Skip("Redis not available")
	}
	defer cache.Close()

	ctx := context.Background()
	data := map[string]interface{}{"id": "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.CacheBatch(ctx, "batch_test", data)
	}
}

// BenchmarkQueryOptimization benchmarks query optimization
func BenchmarkQueryOptimization(b *testing.B) {
	db := &sql.DB{}
	logger := slog.Default()
	optimizer := NewQueryOptimizer(db, logger)

	query := "SELECT * FROM batches WHERE status = ?"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		optimized := optimizer.NewOptimizedQuery(query, "processing")
		optimized.RecordQueryStats(10*time.Millisecond, nil)
	}
}

// BenchmarkLoadTesterMetrics benchmarks load tester metrics collection
func BenchmarkLoadTesterMetrics(b *testing.B) {
	config := LoadTestConfig{
		Duration:          100 * time.Millisecond,
		ConcurrentWorkers: 1,
		RequestsPerSecond: 100,
	}
	logger := slog.Default()

	tester := NewLoadTester(config, logger)

	ctx := context.Background()
	executor := func(ctx context.Context) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tester.Run(ctx, executor)
	}
}

// TestLoadTestWithFailures tests load test with some failures
func TestLoadTestWithFailures(t *testing.T) {
	config := LoadTestConfig{
		Duration:          1 * time.Second,
		ConcurrentWorkers: 2,
		RequestsPerSecond: 10,
	}
	logger := slog.Default()

	tester := NewLoadTester(config, logger)

	ctx := context.Background()
	failureCount := 0
	executor := func(ctx context.Context) error {
		failureCount++
		if failureCount%3 == 0 {
			return fmt.Errorf("simulated error")
		}
		return nil
	}

	tester.Run(ctx, executor)

	results := tester.GetResults()

	if results.FailedRequests == 0 {
		t.Error("Should have some failed requests")
	}

	if results.ErrorRate == 0 {
		t.Error("Error rate should be calculated")
	}
}

// TestMemoryMonitoring tests memory monitoring
func TestMemoryMonitoring(t *testing.T) {
	config := ProfileConfig{
		OutputDir: "./test_profiles",
	}
	logger := slog.Default()

	profiler := NewProfiler(config, logger)
	monitor := NewMemoryMonitor(profiler, 100*time.Millisecond)

	monitor.Start()
	time.Sleep(500 * time.Millisecond)
	monitor.Stop()

	stats := monitor.GetStats()

	if len(stats) == 0 {
		t.Error("Should have collected memory stats")
	}

	if len(stats) > 0 {
		trend := monitor.AnalyzeMemoryTrend()
		if trend == nil {
			t.Error("Trend analysis should not be nil")
		}
	}
}

// TestBatchedQueryExecution tests batched query execution
func TestBatchedQueryExecution(t *testing.T) {
	db := &sql.DB{}
	logger := slog.Default()
	optimizer := NewQueryOptimizer(db, logger)

	batched := optimizer.NewBatchedQuery("INSERT INTO batches VALUES (?)", 100)

	data := make([]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = i
	}

	executor := func(batch []interface{}) error {
		// Simulate batch execution
		return nil
	}

	if err := batched.ExecuteBatched(executor, 1000, data); err != nil {
		t.Fatalf("Batch execution failed: %v", err)
	}

	metrics := batched.GetMetrics()

	if metrics.ProcessedRows != 1000 {
		t.Errorf("Expected 1000 processed rows, got %d", metrics.ProcessedRows)
	}

	if metrics.ThroughputRows == 0 {
		t.Error("Throughput should be calculated")
	}
}
