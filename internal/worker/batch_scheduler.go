package worker

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// SchedulerConfig holds configuration for the batch scheduler
type SchedulerConfig struct {
	PollingInterval    time.Duration
	WorkerPoolSize     int
	JitterRange        time.Duration
	MaxBatchesPerPoll  int
	LockTimeout        time.Duration
	MaxRetries         int
}

// DefaultSchedulerConfig returns default scheduler configuration
func DefaultSchedulerConfig() *SchedulerConfig {
	return &SchedulerConfig{
		PollingInterval:   30 * time.Second,
		WorkerPoolSize:    5,
		JitterRange:       5 * time.Second,
		MaxBatchesPerPoll: 10,
		LockTimeout:       5 * time.Minute,
		MaxRetries:        3,
	}
}

// BatchScheduler manages batch processing with a worker pool
type BatchScheduler struct {
	store     BatchStore
	logger    *slog.Logger
	config    *SchedulerConfig
	running   atomic.Bool
	mu        sync.RWMutex
	workers   chan struct{} // worker pool semaphore
	ctx       context.Context
	cancel    context.CancelFunc
	stats     *SchedulerStats
	statsMu   sync.RWMutex
	wg        sync.WaitGroup
}

// NewBatchScheduler creates a new batch scheduler
func NewBatchScheduler(store BatchStore, logger *slog.Logger, config *SchedulerConfig) *BatchScheduler {
	if config == nil {
		config = DefaultSchedulerConfig()
	}

	return &BatchScheduler{
		store:   store,
		logger:  logger,
		config:  config,
		workers: make(chan struct{}, config.WorkerPoolSize),
		stats: &SchedulerStats{
			LastPollingTime: time.Now(),
		},
	}
}

// Start starts the batch scheduler
func (s *BatchScheduler) Start(ctx context.Context) error {
	if s.running.Load() {
		return fmt.Errorf("scheduler already running")
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running.Store(true)

	// Initialize worker pool
	for i := 0; i < s.config.WorkerPoolSize; i++ {
		s.workers <- struct{}{}
	}

	s.logger.Info("batch scheduler started", "workers", s.config.WorkerPoolSize, "polling_interval", s.config.PollingInterval)

	s.wg.Add(1)
	go s.pollingLoop()

	return nil
}

// Stop stops the batch scheduler
func (s *BatchScheduler) Stop(ctx context.Context) error {
	if !s.running.Load() {
		return nil
	}

	s.logger.Info("stopping batch scheduler")
	s.cancel()
	s.running.Store(false)

	// Wait for pollers to complete
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// IsRunning returns whether the scheduler is running
func (s *BatchScheduler) IsRunning() bool {
	return s.running.Load()
}

// SubmitBatch submits a batch for processing
func (s *BatchScheduler) SubmitBatch(ctx context.Context, orgID, workspaceID string, activityIDs []string, maxRetries int) (string, error) {
	if !s.running.Load() {
		return "", fmt.Errorf("scheduler not running")
	}

	if len(activityIDs) == 0 {
		return "", fmt.Errorf("activity_ids cannot be empty")
	}

	batchID := GenerateBatchID()
	batch := BatchJob{
		ID:            batchID,
		OrgID:         orgID,
		WorkspaceID:   workspaceID,
		Status:        JobStatusPending,
		ActivityCount: len(activityIDs),
		MaxRetries:    maxRetries,
		Priority:      5,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	id, err := s.store.CreateBatch(ctx, batch)
	if err != nil {
		s.logger.Error("failed to create batch", "error", err)
		return "", err
	}

	// Add activity references
	for _, activityID := range activityIDs {
		err := s.store.AddActivityRef(ctx, id, activityID)
		if err != nil {
			s.logger.Error("failed to add activity ref", "error", err, "activity_id", activityID)
		}
	}

	s.logger.Debug("batch submitted", "batch_id", id, "activity_count", len(activityIDs))
	return id, nil
}

// GetStats returns current scheduler statistics
func (s *BatchScheduler) GetStats() SchedulerStats {
	s.statsMu.RLock()
	defer s.statsMu.RUnlock()
	return *s.stats
}

// HealthCheck returns the health status
func (s *BatchScheduler) HealthCheck() HealthStatus {
	s.statsMu.RLock()
	stats := *s.stats
	s.statsMu.RUnlock()

	return HealthStatus{
		Status:           "healthy",
		SchedulerRunning: s.running.Load(),
		BatchesProcessed: stats.BatchesProcessed,
		WorkersActive:    s.config.WorkerPoolSize - len(s.workers),
		PendingBatches:   stats.PendingBatches,
		TotalEmissions:   stats.TotalEmissions,
		Timestamp:        time.Now(),
	}
}

// pollingLoop is the main scheduler loop
func (s *BatchScheduler) pollingLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.PollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("polling loop stopped")
			return
		case <-ticker.C:
			s.pollAndProcess()
			s.updateNextPollingTime()
		}
	}
}

// pollAndProcess polls for pending batches and processes them
func (s *BatchScheduler) pollAndProcess() {
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	s.statsMu.Lock()
	s.stats.LastPollingTime = time.Now()
	s.statsMu.Unlock()

	// Get pending batches
	batches, err := s.store.GetPendingBatches(ctx, s.config.MaxBatchesPerPoll)
	if err != nil {
		s.logger.Error("failed to get pending batches", "error", err)
		return
	}

	if len(batches) == 0 {
		return
	}

	s.statsMu.Lock()
	s.stats.PendingBatches = len(batches)
	s.statsMu.Unlock()

	s.logger.Debug("processing batches", "count", len(batches))

	// Process each batch with worker pool
	for _, batch := range batches {
		select {
		case <-s.ctx.Done():
			return
		case s.workers <- struct{}{}: // Acquire worker
			s.wg.Add(1)
			go s.processBatch(batch)
		}
	}
}

// processBatch processes a single batch
func (s *BatchScheduler) processBatch(batch BatchJob) {
	defer func() {
		<-s.workers // Release worker
		s.wg.Done()
	}()

	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
	defer cancel()

	workerID := fmt.Sprintf("worker_%d", time.Now().UnixNano())

	// Try to acquire lock
	acquired, err := s.store.AcquireBatchLock(ctx, batch.ID, workerID, s.config.LockTimeout)
	if err != nil {
		s.logger.Error("failed to acquire lock", "error", err, "batch_id", batch.ID)
		return
	}

	if !acquired {
		s.logger.Debug("could not acquire lock", "batch_id", batch.ID)
		return
	}

	defer s.store.ReleaseBatchLock(ctx, batch.ID)

	// Update status to processing
	err = s.store.UpdateBatchStatus(ctx, batch.ID, JobStatusProcessing, nil)
	if err != nil {
		s.logger.Error("failed to update batch status", "error", err, "batch_id", batch.ID)
		return
	}

	s.logger.Debug("processing batch", "batch_id", batch.ID, "worker_id", workerID)

	// Simulate processing (in real implementation, call actual processors)
	time.Sleep(100 * time.Millisecond)

	// Mark activities
	for i := 0; i < batch.ActivityCount; i++ {
		activityID := GenerateActivityID()

		if i%10 == 0 {
			// Simulate some failures
			s.store.MarkActivityFailed(ctx, batch.ID, activityID, "simulated error")
		} else {
			s.store.MarkActivityComplete(ctx, batch.ID, activityID, nil)
		}
	}

	// Mark batch complete
	err = s.store.UpdateBatchStatus(ctx, batch.ID, JobStatusComplete, &BatchProgress{
		SuccessCount:   batch.ActivityCount - (batch.ActivityCount / 10),
		ErrorCount:     batch.ActivityCount / 10,
		TotalEmissions: 1000.0,
	})

	if err != nil {
		s.logger.Error("failed to mark batch complete", "error", err, "batch_id", batch.ID)
		return
	}

	s.statsMu.Lock()
	s.stats.BatchesProcessed++
	s.stats.TotalEmissions += 1000.0
	s.statsMu.Unlock()

	s.logger.Info("batch processed successfully", "batch_id", batch.ID)
}

// updateNextPollingTime updates the next polling time in stats
func (s *BatchScheduler) updateNextPollingTime() {
	jitter := time.Duration(rand.Int63n(int64(s.config.JitterRange)))
	nextTime := time.Now().Add(s.config.PollingInterval + jitter)

	s.statsMu.Lock()
	s.stats.NextPollingTime = nextTime
	s.stats.WorkersActive = s.config.WorkerPoolSize - len(s.workers)
	s.statsMu.Unlock()
}

// GetWorkerStats returns statistics for a specific worker
func (s *BatchScheduler) GetWorkerStats(workerID string) map[string]interface{} {
	return map[string]interface{}{
		"worker_id":     workerID,
		"active":        len(s.workers) < s.config.WorkerPoolSize,
		"timestamp":     time.Now(),
	}
}

// GetAllWorkerStats returns statistics for all workers
func (s *BatchScheduler) GetAllWorkerStats() []map[string]interface{} {
	stats := make([]map[string]interface{}, s.config.WorkerPoolSize)
	for i := 0; i < s.config.WorkerPoolSize; i++ {
		stats[i] = map[string]interface{}{
			"worker_id": fmt.Sprintf("worker_%d", i),
			"active":    i < (s.config.WorkerPoolSize - len(s.workers)),
		}
	}
	return stats
}
