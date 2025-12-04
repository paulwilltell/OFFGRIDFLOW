package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/ingestion"
	"github.com/example/offgridflow/internal/ingestion/sources/utility_bills"
)

// =============================================================================
// Configuration
// =============================================================================

// UtilityBillsHandlerConfig configures the utility bills upload endpoints.
type UtilityBillsHandlerConfig struct {
	// Adapter handles utility bill ingestion
	Adapter *utility_bills.Adapter

	// Store for retrieving uploaded bills
	Store ingestion.ActivityStore

	// Logger for operational logging
	Logger *slog.Logger

	// MaxUploadSize in bytes (default 50MB)
	MaxUploadSize int64

	// MaxBatchFiles is the maximum number of files in a batch upload
	MaxBatchFiles int
}

// DefaultUtilityBillsConfig returns a configuration with sensible defaults.
func DefaultUtilityBillsConfig(adapter *utility_bills.Adapter, store ingestion.ActivityStore) UtilityBillsHandlerConfig {
	return UtilityBillsHandlerConfig{
		Adapter:       adapter,
		Store:         store,
		Logger:        slog.Default(),
		MaxUploadSize: 50 * 1024 * 1024, // 50MB
		MaxBatchFiles: 20,
	}
}

// =============================================================================
// HTTP Handlers
// =============================================================================

// NewUtilityBillUploadHandler creates an HTTP handler for uploading utility bills.
//
// Endpoint: POST /api/ingestion/utility-bills/upload
// Content-Type: multipart/form-data
// Form fields:
//   - file: The utility bill file (required)
//   - org_id: Organization ID (optional if in JWT)
//   - strict: Enable strict validation (optional, default: false)
//
// Response:
//   - 200: Upload successful
//   - 400: Invalid request
//   - 413: File too large
//   - 500: Server error
func NewUtilityBillUploadHandler(cfg UtilityBillsHandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.Adapter == nil {
			responders.Error(w, http.StatusNotImplemented, "adapter_not_configured", "utility bills adapter not configured")
			return
		}

		// Parse multipart form
		if err := r.ParseMultipartForm(cfg.MaxUploadSize); err != nil {
			cfg.Logger.Error("failed to parse multipart form", "error", err)
			responders.Error(w, http.StatusBadRequest, "invalid_multipart", "failed to parse multipart form")
			return
		}

		// Get file from form
		file, header, err := r.FormFile("file")
		if err != nil {
			cfg.Logger.Error("no file in upload", "error", err)
			responders.Error(w, http.StatusBadRequest, "missing_file", "no file provided in upload")
			return
		}
		defer file.Close()

		// Check file size
		if header.Size > cfg.MaxUploadSize {
			responders.Error(w, http.StatusRequestEntityTooLarge, "file_too_large",
				fmt.Sprintf("file size %d exceeds maximum of %d bytes", header.Size, cfg.MaxUploadSize))
			return
		}

		// Get organization ID from form or JWT
		orgID := r.FormValue("org_id")
		if orgID == "" {
			// Try to get from JWT context (if auth middleware is enabled)
			if org, ok := r.Context().Value("org_id").(string); ok {
				orgID = org
			}
		}

		// Get strict mode flag
		strictMode := r.FormValue("strict") == "true"

		cfg.Logger.Info("processing utility bill upload",
			"filename", header.Filename,
			"size", header.Size,
			"org_id", orgID,
			"strict", strictMode)

		// Ingest the file
		activities, importErrors, err := cfg.Adapter.IngestFile(r.Context(), header.Filename, file)
		if err != nil {
			cfg.Logger.Error("ingestion failed", "error", err, "filename", header.Filename)
			responders.Error(w, http.StatusInternalServerError, "ingestion_failed", err.Error())
			return
		}

		// Build response
		response := UploadResponse{
			Success:         true,
			Filename:        header.Filename,
			ActivitiesCount: len(activities),
			ErrorsCount:     len(importErrors),
			Activities:      activities,
			Errors:          importErrors,
		}

		// If strict mode and there are errors, return as failure
	if strictMode && len(importErrors) > 0 {
		response.Success = false
		// Return validation payload directly to simplify client handling.
		responders.JSON(w, http.StatusBadRequest, response)
		return
	}

		cfg.Logger.Info("utility bill upload completed",
			"filename", header.Filename,
			"activities", len(activities),
			"errors", len(importErrors))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}

// NewUtilityBillBatchUploadHandler creates an HTTP handler for batch uploading utility bills.
//
// Endpoint: POST /api/ingestion/utility-bills/batch-upload
// Content-Type: multipart/form-data
// Form fields:
//   - files[]: Multiple utility bill files (required)
//   - org_id: Organization ID (optional if in JWT)
//
// Response:
//   - 200: Batch upload completed
//   - 400: Invalid request
//   - 413: Too many files or files too large
//   - 500: Server error
func NewUtilityBillBatchUploadHandler(cfg UtilityBillsHandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.Adapter == nil {
			responders.Error(w, http.StatusNotImplemented, "adapter_not_configured", "utility bills adapter not configured")
			return
		}

		// Parse multipart form
		if err := r.ParseMultipartForm(cfg.MaxUploadSize); err != nil {
			cfg.Logger.Error("failed to parse multipart form", "error", err)
			responders.Error(w, http.StatusBadRequest, "invalid_multipart", "failed to parse multipart form")
			return
		}

		// Get files from form
		formFiles := r.MultipartForm.File["files"]
		if len(formFiles) == 0 {
			// Try alternative field name
			formFiles = r.MultipartForm.File["file"]
		}

		if len(formFiles) == 0 {
			responders.Error(w, http.StatusBadRequest, "missing_files", "no files provided in upload")
			return
		}

		if len(formFiles) > cfg.MaxBatchFiles {
			responders.Error(w, http.StatusBadRequest, "too_many_files",
				fmt.Sprintf("batch upload limited to %d files, received %d", cfg.MaxBatchFiles, len(formFiles)))
			return
		}

		cfg.Logger.Info("processing batch utility bill upload",
			"file_count", len(formFiles))

		// Open all files
		files := make(map[string]io.Reader)
		var openFiles []io.Closer

		for _, fileHeader := range formFiles {
			file, err := fileHeader.Open()
			if err != nil {
				cfg.Logger.Error("failed to open file", "filename", fileHeader.Filename, "error", err)
				// Close previously opened files
				for _, f := range openFiles {
					f.Close()
				}
				responders.Error(w, http.StatusBadRequest, "file_open_error",
					fmt.Sprintf("failed to open file %s", fileHeader.Filename))
				return
			}
			openFiles = append(openFiles, file)
			files[fileHeader.Filename] = file
		}

		// Ensure all files are closed
		defer func() {
			for _, f := range openFiles {
				f.Close()
			}
		}()

		// Ingest files
		result, err := cfg.Adapter.IngestFiles(r.Context(), files)
		if err != nil {
			cfg.Logger.Error("batch ingestion failed", "error", err)
			responders.Error(w, http.StatusInternalServerError, "batch_ingestion_failed", err.Error())
			return
		}

		// Build response
		response := BatchUploadResponse{
			Success:         !result.HasErrors() || result.SuccessFiles > 0,
			TotalFiles:      result.TotalFiles,
			SuccessFiles:    result.SuccessFiles,
			FailedFiles:     result.FailedFiles,
			TotalActivities: result.TotalActivities,
			TotalErrors:     result.TotalErrors,
			Duration:        result.Duration().String(),
			FileResults:     make([]FileUploadResult, len(result.FileResults)),
		}

		for i, fr := range result.FileResults {
			response.FileResults[i] = FileUploadResult{
				Filename:   fr.Filename,
				Success:    fr.Error == nil,
				Activities: len(fr.Activities),
				Errors:     len(fr.Errors),
				Error:      errorString(fr.Error),
			}
		}

		cfg.Logger.Info("batch utility bill upload completed",
			"total_files", result.TotalFiles,
			"success_files", result.SuccessFiles,
			"failed_files", result.FailedFiles,
			"activities", result.TotalActivities,
			"duration", result.Duration())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}

// NewUtilityBillListHandler creates an HTTP handler for listing uploaded utility bills.
//
// Endpoint: GET /api/ingestion/utility-bills
// Query parameters:
//   - org_id: Filter by organization (required or from JWT)
//   - limit: Maximum number of records (default: 100)
//   - source: Filter by source (default: utility_bill)
//
// Response:
//   - 200: List of activities
//   - 400: Invalid request
//   - 500: Server error
func NewUtilityBillListHandler(cfg UtilityBillsHandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.Store == nil {
			responders.Error(w, http.StatusNotImplemented, "store_not_configured", "activity store not configured")
			return
		}

		// Get organization ID
		orgID := r.URL.Query().Get("org_id")
		if orgID == "" {
			if org, ok := r.Context().Value("org_id").(string); ok {
				orgID = org
			}
		}

		if orgID == "" {
			responders.Error(w, http.StatusBadRequest, "missing_org_id", "organization ID is required")
			return
		}

		// Get source filter
		source := r.URL.Query().Get("source")
		if source == "" {
			source = string(ingestion.SourceUtilityBill)
		}

		// Retrieve activities
		activities, err := cfg.Store.ListByOrgAndSource(r.Context(), orgID, source)
		if err != nil {
			cfg.Logger.Error("failed to list activities", "error", err, "org_id", orgID)
			responders.Error(w, http.StatusInternalServerError, "list_failed", "failed to retrieve activities")
			return
		}

		// Apply limit
		limit := 100
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		if len(activities) > limit {
			activities = activities[:limit]
		}

		response := ListResponse{
			Activities: activities,
			Count:      len(activities),
			Source:     source,
			OrgID:      orgID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}

// =============================================================================
// Response Types
// =============================================================================

// UploadResponse is the response for single file uploads.
type UploadResponse struct {
	Success         bool                      `json:"success"`
	Filename        string                    `json:"filename"`
	ActivitiesCount int                       `json:"activities_count"`
	ErrorsCount     int                       `json:"errors_count"`
	Activities      []ingestion.Activity      `json:"activities,omitempty"`
	Errors          []ingestion.ImportError   `json:"errors,omitempty"`
}

// BatchUploadResponse is the response for batch uploads.
type BatchUploadResponse struct {
	Success         bool                `json:"success"`
	TotalFiles      int                 `json:"total_files"`
	SuccessFiles    int                 `json:"success_files"`
	FailedFiles     int                 `json:"failed_files"`
	TotalActivities int                 `json:"total_activities"`
	TotalErrors     int                 `json:"total_errors"`
	Duration        string              `json:"duration"`
	FileResults     []FileUploadResult  `json:"file_results"`
}

// FileUploadResult represents the result of processing a single file in a batch.
type FileUploadResult struct {
	Filename   string `json:"filename"`
	Success    bool   `json:"success"`
	Activities int    `json:"activities"`
	Errors     int    `json:"errors"`
	Error      string `json:"error,omitempty"`
}

// ListResponse is the response for listing utility bill activities.
type ListResponse struct {
	Activities []ingestion.Activity `json:"activities"`
	Count      int                  `json:"count"`
	Source     string               `json:"source"`
	OrgID      string               `json:"org_id"`
}

// =============================================================================
// Helper Functions
// =============================================================================

// errorString safely converts an error to a string.
func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
