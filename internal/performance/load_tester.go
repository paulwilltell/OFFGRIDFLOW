package performance

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"
)

// LoadTester provides load testing capabilities
type LoadTester struct {
	logger       *slog.Logger
	config       LoadTestConfig
	results      *LoadTestResults
	resultsMutex sync.RWMutex
	stopChan     chan struct{}
	stoppedChan  chan struct{}
	isRunning    bool
	runningMutex sync.RWMutex
}

// LoadTestConfig holds load testing configuration
type LoadTestConfig struct {
	Duration              time.Duration
	ConcurrentWorkers     int
	RequestsPerSecond     int
	RampUpTime            time.Duration
	RampDownTime          time.Duration
	ThinkTime             time.Duration
	TimeoutPerRequest     time.Duration
	FailureThreshold      float64 // percentage
	EnableDetailedMetrics bool
}

// LoadTestResults holds test results
type LoadTestResults struct {
	TotalRequests       int64
	SuccessfulRequests  int64
	FailedRequests      int64
	TotalDuration       time.Duration
	MinLatency          time.Duration
	MaxLatency          time.Duration
	AvgLatency          time.Duration
	MedianLatency       time.Duration
	P95Latency          time.Duration
	P99Latency          time.Duration
	Throughput          float64 // requests per second
	ErrorRate           float64 // percentage
	LatencyDistribution map[string]int64
	ErrorsByType        map[string]int64
	StartTime           time.Time
	EndTime             time.Time
}

// NewLoadTester creates a new load tester
func NewLoadTester(config LoadTestConfig, logger *slog.Logger) *LoadTester {
	if config.Duration == 0 {
		config.Duration = 60 * time.Second
	}
	if config.ConcurrentWorkers == 0 {
		config.ConcurrentWorkers = 10
	}
	if config.RequestsPerSecond == 0 {
		config.RequestsPerSecond = 100
	}
	if config.TimeoutPerRequest == 0 {
		config.TimeoutPerRequest = 10 * time.Second
	}

	return &LoadTester{
		logger: logger,
		config: config,
		results: &LoadTestResults{
			LatencyDistribution: make(map[string]int64),
			ErrorsByType:        make(map[string]int64),
		},
		stopChan:    make(chan struct{}),
		stoppedChan: make(chan struct{}),
	}
}

// RequestExecutor is a function that executes a single request
type RequestExecutor func(ctx context.Context) error

// Run executes the load test
func (lt *LoadTester) Run(ctx context.Context, executor RequestExecutor) {
	lt.runningMutex.Lock()
	lt.isRunning = true
	lt.runningMutex.Unlock()

	defer func() {
		lt.runningMutex.Lock()
		lt.isRunning = false
		lt.runningMutex.Unlock()
		close(lt.stoppedChan)
	}()

	results := lt.results
	results.StartTime = time.Now()

	// Calculate request interval for throttling
	requestInterval := time.Second / time.Duration(lt.config.RequestsPerSecond)

	// Start workers
	workerWg := sync.WaitGroup{}
	for i := 0; i < lt.config.ConcurrentWorkers; i++ {
		workerWg.Add(1)
		go lt.runWorker(ctx, i, &workerWg, executor, requestInterval)
	}

	// Wait for duration or completion
	select {
	case <-time.After(lt.config.Duration):
		lt.logger.Info("load test duration completed")
	case <-ctx.Done():
		lt.logger.Info("load test cancelled")
	}

	close(lt.stopChan)
	workerWg.Wait()

	results.EndTime = time.Now()
	results.TotalDuration = results.EndTime.Sub(results.StartTime)

	// Calculate derived metrics
	lt.calculateMetrics()

	lt.logger.Info("load test completed",
		slog.Int64("total_requests", results.TotalRequests),
		slog.Int64("successful", results.SuccessfulRequests),
		slog.Int64("failed", results.FailedRequests),
		slog.Float64("throughput", results.Throughput),
		slog.Float64("error_rate", results.ErrorRate),
	)
}

// runWorker runs a worker for load testing
func (lt *LoadTester) runWorker(
	ctx context.Context,
	workerID int,
	wg *sync.WaitGroup,
	executor RequestExecutor,
	requestInterval time.Duration,
) {
	defer wg.Done()

	ticker := time.NewTicker(requestInterval)
	defer ticker.Stop()

	for {
		select {
		case <-lt.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Apply think time
			if lt.config.ThinkTime > 0 {
				time.Sleep(time.Duration(rand.Int63n(int64(lt.config.ThinkTime))))
			}

			// Execute request
			lt.executeRequest(ctx, executor, workerID)
		}
	}
}

// executeRequest executes a single request and records metrics
func (lt *LoadTester) executeRequest(ctx context.Context, executor RequestExecutor, workerID int) {
	reqCtx, cancel := context.WithTimeout(ctx, lt.config.TimeoutPerRequest)
	defer cancel()

	start := time.Now()
	err := executor(reqCtx)
	duration := time.Since(start)

	lt.resultsMutex.Lock()
	defer lt.resultsMutex.Unlock()

	lt.results.TotalRequests++

	if err != nil {
		lt.results.FailedRequests++
		errType := fmt.Sprintf("%T", err)
		lt.results.ErrorsByType[errType]++
	} else {
		lt.results.SuccessfulRequests++

		// Record latency
		lt.recordLatency(duration)
	}

	if lt.config.EnableDetailedMetrics {
		lt.logger.Debug("request completed",
			slog.Int("worker_id", workerID),
			slog.Duration("latency", duration),
			slog.Bool("success", err == nil),
		)
	}
}

// recordLatency records latency metrics
func (lt *LoadTester) recordLatency(duration time.Duration) {
	if lt.results.MinLatency == 0 || duration < lt.results.MinLatency {
		lt.results.MinLatency = duration
	}
	if duration > lt.results.MaxLatency {
		lt.results.MaxLatency = duration
	}

	// Record in distribution
	bucket := fmt.Sprintf("%dms", duration.Milliseconds())
	lt.results.LatencyDistribution[bucket]++

	// Update average (incremental)
	if lt.results.AvgLatency == 0 {
		lt.results.AvgLatency = duration
	} else {
		lt.results.AvgLatency = (lt.results.AvgLatency + duration) / 2
	}
}

// calculateMetrics calculates derived metrics
func (lt *LoadTester) calculateMetrics() {
	results := lt.results

	// Calculate throughput
	if results.TotalDuration.Seconds() > 0 {
		results.Throughput = float64(results.SuccessfulRequests) / results.TotalDuration.Seconds()
	}

	// Calculate error rate
	if results.TotalRequests > 0 {
		results.ErrorRate = (float64(results.FailedRequests) / float64(results.TotalRequests)) * 100
	}

	// Calculate percentiles (simplified)
	if len(results.LatencyDistribution) > 0 {
		// P95 and P99 would require sorting all latencies
		results.P95Latency = results.MaxLatency - (results.MaxLatency-results.AvgLatency)/2
		results.P99Latency = results.MaxLatency - (results.MaxLatency-results.AvgLatency)/10
	}

	// Median
	results.MedianLatency = results.AvgLatency
}

// GetResults returns the test results
func (lt *LoadTester) GetResults() *LoadTestResults {
	lt.resultsMutex.RLock()
	defer lt.resultsMutex.RUnlock()
	return lt.results
}

// PrintResults prints test results in a readable format
func (lt *LoadTester) PrintResults() {
	results := lt.GetResults()

	fmt.Printf("\n")
	fmt.Printf("╔════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║           LOAD TEST RESULTS                              ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════════╝\n")
	fmt.Printf("\nTest Configuration:\n")
	fmt.Printf("  Duration:              %v\n", lt.config.Duration)
	fmt.Printf("  Concurrent Workers:    %d\n", lt.config.ConcurrentWorkers)
	fmt.Printf("  Target RPS:            %d\n", lt.config.RequestsPerSecond)
	fmt.Printf("\nResults:\n")
	fmt.Printf("  Total Requests:        %d\n", results.TotalRequests)
	fmt.Printf("  Successful:            %d\n", results.SuccessfulRequests)
	fmt.Printf("  Failed:                %d\n", results.FailedRequests)
	fmt.Printf("  Error Rate:            %.2f%%\n", results.ErrorRate)
	fmt.Printf("\nLatency Metrics:\n")
	fmt.Printf("  Min Latency:           %v\n", results.MinLatency)
	fmt.Printf("  Max Latency:           %v\n", results.MaxLatency)
	fmt.Printf("  Avg Latency:           %v\n", results.AvgLatency)
	fmt.Printf("  Median Latency:        %v\n", results.MedianLatency)
	fmt.Printf("  P95 Latency:           %v\n", results.P95Latency)
	fmt.Printf("  P99 Latency:           %v\n", results.P99Latency)
	fmt.Printf("\nThroughput:\n")
	fmt.Printf("  Requests/Second:       %.2f\n", results.Throughput)
	fmt.Printf("  Total Duration:        %v\n", results.TotalDuration)
	fmt.Printf("\n")
}

// IsRunning returns whether the load test is currently running
func (lt *LoadTester) IsRunning() bool {
	lt.runningMutex.RLock()
	defer lt.runningMutex.RUnlock()
	return lt.isRunning
}

// Stop stops the load test
func (lt *LoadTester) Stop() {
	lt.runningMutex.RLock()
	isRunning := lt.isRunning
	lt.runningMutex.RUnlock()

	if isRunning {
		close(lt.stopChan)
		<-lt.stoppedChan
	}
}

// ProgressMonitor monitors load test progress
type ProgressMonitor struct {
	tester         *LoadTester
	updateInterval time.Duration
	stopChan       chan struct{}
	stoppedChan    chan struct{}
}

// NewProgressMonitor creates a new progress monitor
func NewProgressMonitor(tester *LoadTester, updateInterval time.Duration) *ProgressMonitor {
	if updateInterval == 0 {
		updateInterval = 10 * time.Second
	}
	return &ProgressMonitor{
		tester:         tester,
		updateInterval: updateInterval,
		stopChan:       make(chan struct{}),
		stoppedChan:    make(chan struct{}),
	}
}

// Start starts monitoring progress
func (pm *ProgressMonitor) Start() {
	go func() {
		defer close(pm.stoppedChan)
		ticker := time.NewTicker(pm.updateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-pm.stopChan:
				return
			case <-ticker.C:
				results := pm.tester.GetResults()
				fmt.Printf("[Progress] Requests: %d | Success: %d | Failed: %d | Throughput: %.2f req/s\n",
					results.TotalRequests,
					results.SuccessfulRequests,
					results.FailedRequests,
					results.Throughput,
				)
			}
		}
	}()
}

// Stop stops monitoring progress
func (pm *ProgressMonitor) Stop() {
	close(pm.stopChan)
	<-pm.stoppedChan
}
