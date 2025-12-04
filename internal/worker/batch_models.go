package worker

import (
	cryptorand "crypto/rand"
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// JobStatus represents the status of a batch job
type JobStatus int

const (
	JobStatusPending JobStatus = iota
	JobStatusQueued
	JobStatusProcessing
	JobStatusComplete
	JobStatusFailed
	JobStatusCancelled
)

func (s JobStatus) String() string {
	switch s {
	case JobStatusPending:
		return "pending"
	case JobStatusQueued:
		return "queued"
	case JobStatusProcessing:
		return "processing"
	case JobStatusComplete:
		return "complete"
	case JobStatusFailed:
		return "failed"
	case JobStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

func (s JobStatus) Value() (driver.Value, error) {
	return s.String(), nil
}

func (s *JobStatus) Scan(value interface{}) error {
	if value == nil {
		*s = JobStatusPending
		return nil
	}

	v, ok := value.(string)
	if !ok {
		return fmt.Errorf("cannot scan %T into JobStatus", value)
	}

	switch strings.ToLower(v) {
	case "pending":
		*s = JobStatusPending
	case "queued":
		*s = JobStatusQueued
	case "processing":
		*s = JobStatusProcessing
	case "complete":
		*s = JobStatusComplete
	case "failed":
		*s = JobStatusFailed
	case "cancelled":
		*s = JobStatusCancelled
	default:
		return fmt.Errorf("unknown status: %s", v)
	}
	return nil
}

// BatchJob represents a batch processing job
type BatchJob struct {
	ID             string     `json:"id"`
	OrgID          string     `json:"org_id"`
	WorkspaceID    string     `json:"workspace_id"`
	Status         JobStatus  `json:"status"`
	ActivityCount  int        `json:"activity_count"`
	SuccessCount   int        `json:"success_count"`
	ErrorCount     int        `json:"error_count"`
	TotalEmissions float64    `json:"total_emissions"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	RetryCount     int        `json:"retry_count"`
	MaxRetries     int        `json:"max_retries"`
	Priority       int        `json:"priority"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	LockedBy       string     `json:"locked_by,omitempty"`
	LockedUntil    time.Time  `json:"locked_until,omitempty"`
}

// ProgressPercent calculates the progress percentage
func (b *BatchJob) ProgressPercent() float64 {
	if b.ActivityCount == 0 {
		return 0
	}
	return float64(b.SuccessCount+b.ErrorCount) / float64(b.ActivityCount)
}

// RemainingActivities returns the number of activities not yet processed
func (b *BatchJob) RemainingActivities() int {
	return b.ActivityCount - b.SuccessCount - b.ErrorCount
}

// IsLocked returns whether the batch is currently locked
func (b *BatchJob) IsLocked() bool {
	if b.LockedBy == "" {
		return false
	}
	return b.LockedUntil.After(time.Now())
}

// CanRetry returns whether the batch can be retried
func (b *BatchJob) CanRetry() bool {
	return b.Status == JobStatusFailed && b.RetryCount < b.MaxRetries
}

// BatchProgress represents batch processing progress
type BatchProgress struct {
	SuccessCount   int
	ErrorCount     int
	TotalEmissions float64
}

// BatchProgressLog represents a log entry for batch progress
type BatchProgressLog struct {
	ID             int64     `json:"id"`
	BatchID        string    `json:"batch_id"`
	EventType      string    `json:"event_type"`
	ProcessedCount int       `json:"processed_count"`
	ErrorCount     int       `json:"error_count"`
	TotalEmissions float64   `json:"total_emissions"`
	Timestamp      time.Time `json:"timestamp"`
}

// ActivityRef represents a reference to an activity in a batch
type ActivityRef struct {
	BatchID      string    `json:"batch_id"`
	ActivityID   string    `json:"activity_id"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// BatchFilter represents filtering options for batch queries
type BatchFilter struct {
	Status string
	Limit  int
	Offset int
}

// BatchDetailResponse is the response structure for batch details
type BatchDetailResponse struct {
	ID                      string     `json:"id"`
	OrgID                   string     `json:"org_id"`
	WorkspaceID             string     `json:"workspace_id"`
	Status                  string     `json:"status"`
	ActivityCount           int        `json:"activity_count"`
	SuccessCount            int        `json:"success_count"`
	ErrorCount              int        `json:"error_count"`
	TotalEmissions          float64    `json:"total_emissions"`
	ProgressPercent         float64    `json:"progress_percent"`
	RemainingCount          int        `json:"remaining_count"`
	StartedAt               *time.Time `json:"started_at,omitempty"`
	CompletedAt             *time.Time `json:"completed_at,omitempty"`
	ErrorMessage            string     `json:"error_message,omitempty"`
	RetryCount              int        `json:"retry_count"`
	MaxRetries              int        `json:"max_retries"`
	Priority                int        `json:"priority"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
	Duration                *int64     `json:"duration_seconds,omitempty"`
	AvgEmissionsPerActivity float64    `json:"avg_emissions_per_activity,omitempty"`
}

// SchedulerStats represents scheduler statistics
type SchedulerStats struct {
	BatchesProcessed   int64
	BatchesFailed      int64
	BatchesRetried     int64
	TotalActivities    int
	SuccessfulActivity int
	FailedActivity     int
	TotalEmissions     float64
	AverageProcessTime time.Duration
	LastPollingTime    time.Time
	NextPollingTime    time.Time
	WorkersActive      int
	PendingBatches     int
}

// HealthStatus represents the health status of the system
type HealthStatus struct {
	Status           string    `json:"status"`
	SchedulerRunning bool      `json:"scheduler_running"`
	BatchesProcessed int64     `json:"batches_processed"`
	WorkersActive    int       `json:"workers_active"`
	PendingBatches   int       `json:"pending_batches"`
	TotalEmissions   float64   `json:"total_emissions"`
	Timestamp        time.Time `json:"timestamp"`
}

// SubmitBatchRequest is the request structure for submitting a batch
type SubmitBatchRequest struct {
	ActivityIDs []string `json:"activity_ids"`
	MaxRetries  int      `json:"max_retries"`
	Priority    int      `json:"priority"`
}

// SubmitBatchResponse is the response structure for batch submission
type SubmitBatchResponse struct {
	BatchID       string    `json:"batch_id"`
	Status        string    `json:"status"`
	ActivityCount int       `json:"activity_count"`
	CreatedAt     time.Time `json:"created_at"`
}

// ListBatchesResponse is the response structure for listing batches
type ListBatchesResponse struct {
	Batches []BatchDetailResponse `json:"batches"`
	Total   int                   `json:"total"`
	Limit   int                   `json:"limit"`
	Offset  int                   `json:"offset"`
}

// ProgressResponse represents progress information
type ProgressResponse struct {
	BatchID             string         `json:"batch_id"`
	ProcessedCount      int            `json:"processed_count"`
	TotalCount          int            `json:"total_count"`
	SuccessCount        int            `json:"success_count"`
	ErrorCount          int            `json:"error_count"`
	PercentComplete     float64        `json:"percent_complete"`
	EstimatedRemaining  *time.Duration `json:"estimated_remaining,omitempty"`
	TotalEmissions      float64        `json:"total_emissions"`
	AvgEmissionsPerItem float64        `json:"avg_emissions_per_item"`
	Status              string         `json:"status"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error     string    `json:"error"`
	Code      string    `json:"code"`
	Details   string    `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// MarshalJSON implements custom JSON marshaling for BatchJob
func (b *BatchJob) MarshalJSON() ([]byte, error) {
	type Alias BatchJob
	return json.Marshal(&struct {
		Status string `json:"status"`
		*Alias
	}{
		Status: b.Status.String(),
		Alias:  (*Alias)(b),
	})
}

// Valid common error codes
const (
	ErrorCodeValidation = "VALIDATION_ERROR"
	ErrorCodeNotFound   = "NOT_FOUND"
	ErrorCodeConflict   = "CONFLICT"
	ErrorCodeInternal   = "INTERNAL_ERROR"
)

// ErrBatchNotFound is returned when a batch is not found
var ErrBatchNotFound = sql.ErrNoRows

// GenerateActivityID generates a unique activity ID
func GenerateActivityID() string {
	return fmt.Sprintf("act_%d", time.Now().UnixNano())
}

// GenerateBatchID generates a unique batch ID
func GenerateBatchID() string {
	b := make([]byte, 8)
	cryptorand.Read(b)
	return fmt.Sprintf("batch_%s", hex.EncodeToString(b))
}
