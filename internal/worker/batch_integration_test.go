package worker

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"
)

// Test helper: Create mock store for integration tests
type mockTestStore struct {
	mu      sync.RWMutex
	batches map[string]*BatchJob
}

func createMockTestStore() BatchStore {
	return &mockTestStore{
		batches: make(map[string]*BatchJob),
	}
}

func (m *mockTestStore) CreateBatch(ctx context.Context, job BatchJob) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.batches[job.ID] = &job
	return job.ID, nil
}

func (m *mockTestStore) GetBatch(ctx context.Context, batchID string) (*BatchJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	batch, ok := m.batches[batchID]
	if !ok {
		return nil, ErrBatchNotFound
	}
	return batch, nil
}

func (m *mockTestStore) ListBatches(ctx context.Context, orgID string, filter BatchFilter) ([]BatchJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []BatchJob
	for _, batch := range m.batches {
		if batch.OrgID == orgID {
			result = append(result, *batch)
		}
	}
	return result, nil
}

func (m *mockTestStore) UpdateBatchStatus(ctx context.Context, batchID string, status JobStatus, progress *BatchProgress) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	batch, ok := m.batches[batchID]
	if !ok {
		return ErrBatchNotFound
	}
	batch.Status = status
	batch.UpdatedAt = time.Now()
	if progress != nil {
		batch.SuccessCount = progress.SuccessCount
		batch.ErrorCount = progress.ErrorCount
		batch.TotalEmissions = progress.TotalEmissions
	}
	return nil
}

func (m *mockTestStore) AddActivityRef(ctx context.Context, batchID, activityID string) error {
	return nil
}

func (m *mockTestStore) MarkActivityComplete(ctx context.Context, batchID, activityID string, record interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	batch, ok := m.batches[batchID]
	if !ok {
		return ErrBatchNotFound
	}
	batch.SuccessCount++
	return nil
}

func (m *mockTestStore) MarkActivityFailed(ctx context.Context, batchID, activityID string, errMsg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	batch, ok := m.batches[batchID]
	if !ok {
		return ErrBatchNotFound
	}
	batch.ErrorCount++
	return nil
}

func (m *mockTestStore) GetPendingBatches(ctx context.Context, limit int) ([]BatchJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []BatchJob
	count := 0
	for _, batch := range m.batches {
		if (batch.Status == JobStatusPending || batch.Status == JobStatusQueued) && count < limit {
			result = append(result, *batch)
			count++
		}
	}
	return result, nil
}

func (m *mockTestStore) AcquireBatchLock(ctx context.Context, batchID, workerID string, timeout time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	batch, ok := m.batches[batchID]
	if !ok {
		return false, ErrBatchNotFound
	}
	if batch.LockedBy != "" && batch.LockedUntil.After(time.Now()) {
		return false, nil
	}
	batch.LockedBy = workerID
	batch.LockedUntil = time.Now().Add(timeout)
	return true, nil
}

func (m *mockTestStore) ReleaseBatchLock(ctx context.Context, batchID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	batch, ok := m.batches[batchID]
	if !ok {
		return ErrBatchNotFound
	}
	batch.LockedBy = ""
	batch.LockedUntil = time.Time{}
	return nil
}

func (m *mockTestStore) DeleteBatch(ctx context.Context, batchID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.batches, batchID)
	return nil
}

func (m *mockTestStore) GetProgressLog(ctx context.Context, batchID string) ([]BatchProgressLog, error) {
	return []BatchProgressLog{}, nil
}

// Integration Tests

func TestIntegration_SubmitAndRetrieveBatch(t *testing.T) {
	store := createMockTestStore()
	logger := slog.Default()
	scheduler := NewBatchScheduler(store, logger, DefaultSchedulerConfig())

	ctx := context.Background()
	scheduler.Start(ctx)
	defer scheduler.Stop(ctx)

	batchID, err := scheduler.SubmitBatch(ctx, "org_1", "ws_1", []string{"act_1", "act_2"}, 3)
	if err != nil {
		t.Fatalf("Failed to submit batch: %v", err)
	}

	if batchID == "" {
		t.Error("Batch ID should not be empty")
	}

	batch, err := store.GetBatch(ctx, batchID)
	if err != nil {
		t.Fatalf("Failed to get batch: %v", err)
	}

	if batch.ActivityCount != 2 {
		t.Errorf("Expected 2 activities, got %d", batch.ActivityCount)
	}

	if batch.OrgID != "org_1" {
		t.Errorf("Expected org_1, got %s", batch.OrgID)
	}
}

func TestIntegration_BatchProgress(t *testing.T) {
	store := createMockTestStore()
	logger := slog.Default()
	scheduler := NewBatchScheduler(store, logger, DefaultSchedulerConfig())

	ctx := context.Background()
	scheduler.Start(ctx)
	defer scheduler.Stop(ctx)

	batchID, err := scheduler.SubmitBatch(ctx, "org_1", "ws_1", []string{"act_1", "act_2", "act_3"}, 3)
	if err != nil {
		t.Fatalf("failed to submit batch: %v", err)
	}

	batch, err := store.GetBatch(ctx, batchID)
	if err != nil {
		t.Fatalf("failed to get batch: %v", err)
	}
	if batch.ProgressPercent() != 0 {
		t.Errorf("Initial progress should be 0, got %f", batch.ProgressPercent())
	}

	// Mark 2 activities complete
	store.MarkActivityComplete(ctx, batchID, "act_1", nil)
	store.MarkActivityComplete(ctx, batchID, "act_2", nil)

	batch, _ = store.GetBatch(ctx, batchID)
	// Expected progress is ~66% (2/3 activities complete)
	if batch.ProgressPercent() <= 0 {
		t.Errorf("Progress should increase after completing activities")
	}
}

func TestIntegration_ConcurrentBatches(t *testing.T) {
	store := createMockTestStore()
	logger := slog.Default()
	scheduler := NewBatchScheduler(store, logger, DefaultSchedulerConfig())

	ctx := context.Background()
	scheduler.Start(ctx)
	defer scheduler.Stop(ctx)

	var wg sync.WaitGroup
	batchIDs := make([]string, 5)
	errors := make([]error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			id, err := scheduler.SubmitBatch(ctx, "org_1", "ws_1", []string{"act_1", "act_2"}, 3)
			batchIDs[index] = id
			errors[index] = err
		}(i)
	}

	wg.Wait()

	for i, err := range errors {
		if err != nil {
			t.Errorf("Batch %d submission failed: %v", i, err)
		}
	}

	batches, _ := store.ListBatches(ctx, "org_1", BatchFilter{})
	if len(batches) != 5 {
		t.Errorf("Expected 5 batches, got %d", len(batches))
	}
}

func TestIntegration_SchedulerHealth(t *testing.T) {
	store := createMockTestStore()
	logger := slog.Default()
	scheduler := NewBatchScheduler(store, logger, DefaultSchedulerConfig())

	ctx := context.Background()
	scheduler.Start(ctx)
	defer scheduler.Stop(ctx)

	time.Sleep(100 * time.Millisecond)

	health := scheduler.HealthCheck()
	if health.Status != "healthy" {
		t.Errorf("Expected healthy status, got %s", health.Status)
	}

	if !health.SchedulerRunning {
		t.Error("Scheduler should be running")
	}

	stats := scheduler.GetStats()
	if stats.WorkersActive < 0 || stats.WorkersActive > 5 {
		t.Errorf("Invalid worker count: %d", stats.WorkersActive)
	}
}

func TestIntegration_BatchLocking(t *testing.T) {
	store := createMockTestStore()
	logger := slog.Default()
	scheduler := NewBatchScheduler(store, logger, DefaultSchedulerConfig())

	ctx := context.Background()
	scheduler.Start(ctx)
	defer scheduler.Stop(ctx)

	batchID, err := scheduler.SubmitBatch(ctx, "org_1", "ws_1", []string{"act_1"}, 3)
	if err != nil {
		t.Fatalf("failed to submit batch: %v", err)
	}

	// Worker 1 acquires lock
	acquired1, _ := store.AcquireBatchLock(ctx, batchID, "worker_1", 5*time.Second)
	if !acquired1 {
		t.Error("Worker 1 should acquire lock")
	}

	// Worker 2 tries to acquire same lock (should fail)
	acquired2, _ := store.AcquireBatchLock(ctx, batchID, "worker_2", 5*time.Second)
	if acquired2 {
		t.Error("Worker 2 should not acquire lock while held by worker 1")
	}

	// Worker 1 releases lock
	store.ReleaseBatchLock(ctx, batchID)

	// Worker 2 can now acquire lock
	acquired2, _ = store.AcquireBatchLock(ctx, batchID, "worker_2", 5*time.Second)
	if !acquired2 {
		t.Error("Worker 2 should acquire lock after release")
	}
}

func TestIntegration_BatchStatusTransitions(t *testing.T) {
	store := createMockTestStore()
	batchID, _ := store.CreateBatch(context.Background(), BatchJob{
		ID:        GenerateBatchID(),
		OrgID:     "org_1",
		Status:    JobStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	ctx := context.Background()

	// Transition: Pending -> Queued
	store.UpdateBatchStatus(ctx, batchID, JobStatusQueued, nil)
	batch, _ := store.GetBatch(ctx, batchID)
	if batch.Status != JobStatusQueued {
		t.Errorf("Expected queued status, got %v", batch.Status)
	}

	// Transition: Queued -> Processing
	store.UpdateBatchStatus(ctx, batchID, JobStatusProcessing, nil)
	batch, _ = store.GetBatch(ctx, batchID)
	if batch.Status != JobStatusProcessing {
		t.Errorf("Expected processing status, got %v", batch.Status)
	}

	// Transition: Processing -> Complete
	store.UpdateBatchStatus(ctx, batchID, JobStatusComplete, &BatchProgress{
		SuccessCount:   5,
		ErrorCount:     0,
		TotalEmissions: 100.0,
	})
	batch, _ = store.GetBatch(ctx, batchID)
	if batch.Status != JobStatusComplete {
		t.Errorf("Expected complete status, got %v", batch.Status)
	}
	if batch.SuccessCount != 5 {
		t.Errorf("Expected 5 successes, got %d", batch.SuccessCount)
	}
}

func TestIntegration_BatchDeletion(t *testing.T) {
	store := createMockTestStore()
	ctx := context.Background()

	batchID, _ := store.CreateBatch(ctx, BatchJob{
		ID:        GenerateBatchID(),
		OrgID:     "org_1",
		Status:    JobStatusComplete,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	// Batch exists
	batch, err := store.GetBatch(ctx, batchID)
	if err != nil || batch == nil {
		t.Error("Batch should exist before deletion")
	}

	// Delete batch
	store.DeleteBatch(ctx, batchID)

	// Batch no longer exists
	batch, err = store.GetBatch(ctx, batchID)
	if err != ErrBatchNotFound {
		t.Errorf("Batch should not be found after deletion")
	}
}

func BenchmarkSubmitBatch(b *testing.B) {
	store := createMockTestStore()
	logger := slog.Default()
	scheduler := NewBatchScheduler(store, logger, DefaultSchedulerConfig())

	ctx := context.Background()
	scheduler.Start(ctx)
	defer scheduler.Stop(ctx)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		scheduler.SubmitBatch(ctx, "org_1", "ws_1", []string{"act_1"}, 3)
	}
}

func BenchmarkAcquireLock(b *testing.B) {
	store := createMockTestStore()
	ctx := context.Background()

	batchID, _ := store.CreateBatch(ctx, BatchJob{
		ID:        GenerateBatchID(),
		OrgID:     "org_1",
		Status:    JobStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		store.AcquireBatchLock(ctx, batchID, "worker", 5*time.Second)
		store.ReleaseBatchLock(ctx, batchID)
	}
}
