package http

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/example/offgridflow/internal/worker"
)

// BatchHandlers provides HTTP handlers for batch operations
type BatchHandlers struct {
	scheduler worker.BatchScheduler
	store     worker.BatchStore
	logger    *slog.Logger
}

// NewBatchHandlers creates new batch handlers
func NewBatchHandlers(scheduler worker.BatchScheduler, store worker.BatchStore, logger *slog.Logger) *BatchHandlers {
	return &BatchHandlers{
		scheduler: scheduler,
		store:     store,
		logger:    logger,
	}
}

// RegisterBatchRoutes registers batch-related routes
func (h *BatchHandlers) RegisterBatchRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/batches", h.SubmitBatch)
	mux.HandleFunc("GET /api/v1/batches", h.ListBatches)
	mux.HandleFunc("GET /api/v1/batches/{id}", h.GetBatch)
	mux.HandleFunc("POST /api/v1/batches/{id}/cancel", h.CancelBatch)
	mux.HandleFunc("POST /api/v1/batches/{id}/retry", h.RetryBatch)
	mux.HandleFunc("DELETE /api/v1/batches/{id}", h.DeleteBatch)
	mux.HandleFunc("GET /api/v1/batches/{id}/progress", h.GetProgress)
	mux.HandleFunc("GET /api/v1/health", h.HealthCheck)
}

// SubmitBatch handles batch submission
func (h *BatchHandlers) SubmitBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only POST allowed")
		return
	}

	// Extract headers
	orgID := r.Header.Get("X-Org-ID")
	workspaceID := r.Header.Get("X-Workspace-ID")

	if orgID == "" || workspaceID == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_HEADERS", "X-Org-ID and X-Workspace-ID required")
		return
	}

	// Parse request body
	var req worker.SubmitBatchRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	if len(req.ActivityIDs) == 0 {
		h.writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "activity_ids cannot be empty")
		return
	}

	// Submit batch
	batchID, err := h.scheduler.SubmitBatch(r.Context(), orgID, workspaceID, req.ActivityIDs, req.MaxRetries)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	// Write response
	resp := worker.SubmitBatchResponse{
		BatchID:       batchID,
		Status:        "pending",
		ActivityCount: len(req.ActivityIDs),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)

	h.logger.Debug("batch submitted", "batch_id", batchID, "org_id", orgID)
}

// ListBatches handles batch listing
func (h *BatchHandlers) ListBatches(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only GET allowed")
		return
	}

	orgID := r.Header.Get("X-Org-ID")
	if orgID == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_HEADER", "X-Org-ID required")
		return
	}

	// Parse query parameters
	status := r.URL.Query().Get("status")
	limit := 10
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 1000 {
			limit = v
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	// List batches
	batches, err := h.store.ListBatches(r.Context(), orgID, worker.BatchFilter{
		Status: status,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	// Convert to response format
	details := make([]worker.BatchDetailResponse, len(batches))
	for i, batch := range batches {
		details[i] = h.batchToDetail(batch)
	}

	resp := worker.ListBatchesResponse{
		Batches: details,
		Total:   len(details),
		Limit:   limit,
		Offset:  offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetBatch handles getting a single batch
func (h *BatchHandlers) GetBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only GET allowed")
		return
	}

	batchID := r.PathValue("id")
	batch, err := h.store.GetBatch(r.Context(), batchID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "batch not found")
		return
	}

	resp := h.batchToDetail(*batch)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CancelBatch handles batch cancellation
func (h *BatchHandlers) CancelBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only POST allowed")
		return
	}

	batchID := r.PathValue("id")
	batch, err := h.store.GetBatch(r.Context(), batchID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "batch not found")
		return
	}

	if batch.Status != worker.JobStatusPending && batch.Status != worker.JobStatusQueued {
		h.writeError(w, http.StatusConflict, "CONFLICT", "can only cancel pending or queued batches")
		return
	}

	err = h.store.UpdateBatchStatus(r.Context(), batchID, worker.JobStatusCancelled, nil)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	batch.Status = worker.JobStatusCancelled
	resp := h.batchToDetail(*batch)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	h.logger.Info("batch cancelled", "batch_id", batchID)
}

// RetryBatch handles batch retry
func (h *BatchHandlers) RetryBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only POST allowed")
		return
	}

	batchID := r.PathValue("id")
	batch, err := h.store.GetBatch(r.Context(), batchID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "batch not found")
		return
	}

	if batch.Status != worker.JobStatusFailed {
		h.writeError(w, http.StatusConflict, "CONFLICT", "can only retry failed batches")
		return
	}

	if batch.RetryCount >= batch.MaxRetries {
		h.writeError(w, http.StatusConflict, "CONFLICT", fmt.Sprintf("max retries (%d) reached", batch.MaxRetries))
		return
	}

	batch.RetryCount++
	err = h.store.UpdateBatchStatus(r.Context(), batchID, worker.JobStatusPending, nil)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	resp := h.batchToDetail(*batch)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	h.logger.Info("batch retried", "batch_id", batchID, "retry_count", batch.RetryCount)
}

// DeleteBatch handles batch deletion
func (h *BatchHandlers) DeleteBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only DELETE allowed")
		return
	}

	batchID := r.PathValue("id")
	err := h.store.DeleteBatch(r.Context(), batchID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "batch not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
	h.logger.Info("batch deleted", "batch_id", batchID)
}

// GetProgress handles progress retrieval
func (h *BatchHandlers) GetProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only GET allowed")
		return
	}

	batchID := r.PathValue("id")
	batch, err := h.store.GetBatch(r.Context(), batchID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "batch not found")
		return
	}

	processed := batch.SuccessCount + batch.ErrorCount
	avgEmissions := 0.0
	if processed > 0 {
		avgEmissions = batch.TotalEmissions / float64(processed)
	}

	var estimatedRemaining *time.Duration
	if batch.RemainingActivities() > 0 && batch.StartedAt != nil {
		elapsed := batch.UpdatedAt.Sub(*batch.StartedAt)
		if elapsed > 0 && processed > 0 {
			perItem := elapsed / time.Duration(processed)
			estimated := perItem * time.Duration(batch.RemainingActivities())
			estimatedRemaining = &estimated
		}
	}

	resp := worker.ProgressResponse{
		BatchID:             batchID,
		ProcessedCount:      processed,
		TotalCount:          batch.ActivityCount,
		SuccessCount:        batch.SuccessCount,
		ErrorCount:          batch.ErrorCount,
		PercentComplete:     batch.ProgressPercent(),
		EstimatedRemaining:  estimatedRemaining,
		TotalEmissions:      batch.TotalEmissions,
		AvgEmissionsPerItem: avgEmissions,
		Status:              batch.Status.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HealthCheck handles health check requests
func (h *BatchHandlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only GET allowed")
		return
	}

	// Always return 200 OK
	health := worker.HealthStatus{
		Status:           "healthy",
		SchedulerRunning: h.scheduler.IsRunning(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(health)
}

// Helper functions

func (h *BatchHandlers) batchToDetail(batch worker.BatchJob) worker.BatchDetailResponse {
	duration := int64(0)
	if batch.StartedAt != nil && batch.CompletedAt != nil {
		duration = int64(batch.CompletedAt.Sub(*batch.StartedAt).Seconds())
	}

	avgEmissions := 0.0
	if batch.ActivityCount > 0 {
		avgEmissions = batch.TotalEmissions / float64(batch.ActivityCount)
	}

	var durationPtr *int64
	if duration > 0 {
		durationPtr = &duration
	}

	return worker.BatchDetailResponse{
		ID:                      batch.ID,
		OrgID:                   batch.OrgID,
		WorkspaceID:             batch.WorkspaceID,
		Status:                  batch.Status.String(),
		ActivityCount:           batch.ActivityCount,
		SuccessCount:            batch.SuccessCount,
		ErrorCount:              batch.ErrorCount,
		TotalEmissions:          batch.TotalEmissions,
		ProgressPercent:         batch.ProgressPercent(),
		RemainingCount:          batch.RemainingActivities(),
		StartedAt:               batch.StartedAt,
		CompletedAt:             batch.CompletedAt,
		ErrorMessage:            batch.ErrorMessage,
		RetryCount:              batch.RetryCount,
		MaxRetries:              batch.MaxRetries,
		Priority:                batch.Priority,
		CreatedAt:               batch.CreatedAt,
		UpdatedAt:               batch.UpdatedAt,
		Duration:                durationPtr,
		AvgEmissionsPerActivity: avgEmissions,
	}
}

func (h *BatchHandlers) writeError(w http.ResponseWriter, status int, code, message string) {
	resp := worker.ErrorResponse{
		Error:   code,
		Code:    code,
		Details: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)

	h.logger.Warn("API error", "status", status, "code", code, "message", message)
}

// ReadBodyLimit reads request body with size limit
func (h *BatchHandlers) ReadBodyLimit(r io.Reader, maxSize int64) ([]byte, error) {
	limitedReader := &io.LimitedReader{
		R: r,
		N: maxSize,
	}
	return io.ReadAll(limitedReader)
}
