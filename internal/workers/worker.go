package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// JobType represents the type of background job
type JobType string

const (
	JobTypeEmissionsCalculation JobType = "emissions.calculate"
	JobTypeConnectorSync        JobType = "connector.sync"
	JobTypeReportGeneration     JobType = "report.generate"
	JobTypeDataExport           JobType = "data.export"
	JobTypeAuditCleanup         JobType = "audit.cleanup"
	JobTypeBillingSync          JobType = "billing.sync"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusRetrying   JobStatus = "retrying"
)

// Job represents a background job
type Job struct {
	ID          string                 `json:"id"`
	Type        JobType                `json:"type"`
	Status      JobStatus              `json:"status"`
	TenantID    string                 `json:"tenant_id"`
	Payload     map[string]interface{} `json:"payload"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ScheduledAt time.Time              `json:"scheduled_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// JobHandler is a function that processes a job
type JobHandler func(ctx context.Context, job *Job) error

// Queue defines the interface for job queue implementations
type Queue interface {
	Enqueue(ctx context.Context, job *Job) error
	Dequeue(ctx context.Context, jobType JobType) (*Job, error)
	Complete(ctx context.Context, jobID string, result map[string]interface{}) error
	Fail(ctx context.Context, jobID string, err error) error
	Retry(ctx context.Context, jobID string, delay time.Duration) error
	GetJob(ctx context.Context, jobID string) (*Job, error)
	ListJobs(ctx context.Context, tenantID string, status JobStatus, limit int) ([]*Job, error)
}

// Worker processes jobs from a queue
type Worker struct {
	queue    Queue
	handlers map[JobType]JobHandler
	logger   *slog.Logger
	
	mu         sync.RWMutex
	workers    int
	stopChan   chan struct{}
	doneChan   chan struct{}
	isRunning  bool
}

// WorkerConfig holds worker configuration
type WorkerConfig struct {
	Workers     int           // Number of concurrent workers
	PollInterval time.Duration // How often to poll for jobs
}

// DefaultWorkerConfig returns sensible defaults
func DefaultWorkerConfig() WorkerConfig {
	return WorkerConfig{
		Workers:     5,
		PollInterval: 1 * time.Second,
	}
}

// NewWorker creates a new worker pool
func NewWorker(queue Queue, logger *slog.Logger) *Worker {
	if logger == nil {
		logger = slog.Default()
	}

	return &Worker{
		queue:    queue,
		handlers: make(map[JobType]JobHandler),
		logger:   logger,
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
	}
}

// RegisterHandler registers a handler for a job type
func (w *Worker) RegisterHandler(jobType JobType, handler JobHandler) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.handlers[jobType] = handler
}

// Start starts the worker pool
func (w *Worker) Start(ctx context.Context, config WorkerConfig) error {
	w.mu.Lock()
	if w.isRunning {
		w.mu.Unlock()
		return fmt.Errorf("worker already running")
	}
	w.isRunning = true
	w.workers = config.Workers
	w.mu.Unlock()

	w.logger.Info("Starting worker pool", "workers", config.Workers)

	var wg sync.WaitGroup
	for i := 0; i < config.Workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			w.run(ctx, workerID, config.PollInterval)
		}(i)
	}

	go func() {
		wg.Wait()
		close(w.doneChan)
	}()

	return nil
}

// Stop gracefully stops the worker pool
func (w *Worker) Stop(ctx context.Context) error {
	w.mu.Lock()
	if !w.isRunning {
		w.mu.Unlock()
		return nil
	}
	w.mu.Unlock()

	w.logger.Info("Stopping worker pool")
	close(w.stopChan)

	// Wait for all workers to finish with timeout
	select {
	case <-w.doneChan:
		w.logger.Info("Worker pool stopped")
	case <-ctx.Done():
		return fmt.Errorf("timeout waiting for workers to stop")
	}

	w.mu.Lock()
	w.isRunning = false
	w.mu.Unlock()

	return nil
}

func (w *Worker) run(ctx context.Context, workerID int, pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	w.logger.Info("Worker started", "worker_id", workerID)

	for {
		select {
		case <-w.stopChan:
			w.logger.Info("Worker stopped", "worker_id", workerID)
			return
		case <-ctx.Done():
			w.logger.Info("Worker context cancelled", "worker_id", workerID)
			return
		case <-ticker.C:
			w.processNext(ctx, workerID)
		}
	}
}

func (w *Worker) processNext(ctx context.Context, workerID int) {
	// Try to dequeue all job types
	w.mu.RLock()
	jobTypes := make([]JobType, 0, len(w.handlers))
	for jt := range w.handlers {
		jobTypes = append(jobTypes, jt)
	}
	w.mu.RUnlock()

	for _, jobType := range jobTypes {
		job, err := w.queue.Dequeue(ctx, jobType)
		if err != nil {
			continue // No jobs available
		}

		if job == nil {
			continue
		}

		w.processJob(ctx, workerID, job)
		return // Processed one job, return to polling
	}
}

func (w *Worker) processJob(ctx context.Context, workerID int, job *Job) {
	w.logger.Info("Processing job",
		"worker_id", workerID,
		"job_id", job.ID,
		"job_type", job.Type,
		"tenant_id", job.TenantID,
		"attempt", job.Attempts)

	// Get handler
	w.mu.RLock()
	handler, exists := w.handlers[job.Type]
	w.mu.RUnlock()

	if !exists {
		w.logger.Error("No handler for job type",
			"job_id", job.ID,
			"job_type", job.Type)
		w.queue.Fail(ctx, job.ID, fmt.Errorf("no handler for job type: %s", job.Type))
		return
	}

	// Execute handler with timeout
	jobCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	err := handler(jobCtx, job)
	if err != nil {
		w.logger.Error("Job failed",
			"worker_id", workerID,
			"job_id", job.ID,
			"error", err)

		// Retry logic
		if job.Attempts < job.MaxAttempts {
			delay := calculateRetryDelay(job.Attempts)
			w.logger.Info("Retrying job",
				"job_id", job.ID,
				"attempt", job.Attempts+1,
				"delay", delay)
			w.queue.Retry(ctx, job.ID, delay)
		} else {
			w.logger.Error("Job exceeded max attempts",
				"job_id", job.ID,
				"max_attempts", job.MaxAttempts)
			w.queue.Fail(ctx, job.ID, err)
		}
		return
	}

	// Mark as completed
	w.queue.Complete(ctx, job.ID, job.Result)
	w.logger.Info("Job completed",
		"worker_id", workerID,
		"job_id", job.ID,
		"duration", time.Since(job.UpdatedAt))
}

func calculateRetryDelay(attempts int) time.Duration {
	// Exponential backoff: 1min, 2min, 4min, 8min, 16min
	delay := time.Duration(1<<uint(attempts)) * time.Minute
	if delay > 30*time.Minute {
		delay = 30 * time.Minute
	}
	return delay
}

// NewJob creates a new job
func NewJob(jobType JobType, tenantID string, payload map[string]interface{}) *Job {
	return &Job{
		ID:          fmt.Sprintf("job_%d", time.Now().UnixNano()),
		Type:        jobType,
		Status:      JobStatusPending,
		TenantID:    tenantID,
		Payload:     payload,
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// ToJSON serializes a job to JSON
func (j *Job) ToJSON() ([]byte, error) {
	return json.Marshal(j)
}

// FromJSON deserializes a job from JSON
func FromJSON(data []byte) (*Job, error) {
	var job Job
	if err := json.Unmarshal(data, &job); err != nil {
		return nil, err
	}
	return &job, nil
}
