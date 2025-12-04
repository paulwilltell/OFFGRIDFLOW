// Package utility_bills provides comprehensive ingestion capabilities for utility bills
// from various sources and formats.
//
// This adapter supports:
//   - File-based uploads (CSV, JSON, PDF, Excel)
//   - API integrations with utility providers
//   - Email attachment processing
//   - Automated bill scanning and extraction
//
// The adapter implements intelligent data extraction, validation, deduplication,
// and enrichment to ensure high-quality utility consumption data flows into
// the emissions calculation pipeline.
package utility_bills

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"strings"
	"sync"
	"time"

	"github.com/example/offgridflow/internal/ingestion"
	"github.com/example/offgridflow/internal/ingestion/parser"
)

// =============================================================================
// Adapter Configuration
// =============================================================================

// Config configures the utility bills ingestion adapter.
type Config struct {
	// DefaultOrgID is used when bills don't specify an organization
	DefaultOrgID string

	// DefaultLocation is used when location cannot be determined
	DefaultLocation string

	// Store persists ingested activities
	Store ingestion.ActivityStore

	// Logger for operational logging
	Logger *slog.Logger

	// EnableDeduplication prevents duplicate bills from being processed
	EnableDeduplication bool

	// DeduplicationWindow is how far back to check for duplicates
	DeduplicationWindow time.Duration

	// StrictValidation fails the entire import on any validation error
	StrictValidation bool

	// MaxConcurrentParsing limits parallel file processing
	MaxConcurrentParsing int

	// MaxFileSize in bytes (default 50MB)
	MaxFileSize int64

	// AutoEnrichLocation attempts to determine location from bill data
	AutoEnrichLocation bool

	// UtilityProviderMappings maps provider names to locations
	UtilityProviderMappings map[string]string
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig(orgID string) Config {
	return Config{
		DefaultOrgID:            orgID,
		DefaultLocation:         "US",
		EnableDeduplication:     true,
		DeduplicationWindow:     90 * 24 * time.Hour, // 90 days
		StrictValidation:        false,
		MaxConcurrentParsing:    4,
		MaxFileSize:             50 * 1024 * 1024, // 50MB
		AutoEnrichLocation:      true,
		UtilityProviderMappings: make(map[string]string),
	}
}

// =============================================================================
// Adapter Implementation
// =============================================================================

// Adapter ingests utility bills from various sources and formats.
type Adapter struct {
	config Config
	logger *slog.Logger
	parser *parser.UtilityBillParser

	// deduplicationCache tracks recently processed bills
	deduplicationCache map[string]time.Time
	mu                 sync.RWMutex
}

// NewAdapter creates a new utility bills adapter with the given configuration.
func NewAdapter(config Config) *Adapter {
	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	billParser := parser.NewUtilityBillParser(config.DefaultOrgID, config.DefaultLocation)
	billParser.StrictMode = config.StrictValidation
	if config.MaxFileSize > 0 {
		billParser.MaxFileSize = config.MaxFileSize
	}

	return &Adapter{
		config:             config,
		logger:             logger,
		parser:             billParser,
		deduplicationCache: make(map[string]time.Time),
	}
}

// Ingest implements SourceIngestionAdapter.
// This returns an empty slice as utility bills are typically uploaded manually
// rather than pulled from an external source. Use IngestFile or IngestFiles
// for actual bill processing.
func (a *Adapter) Ingest(ctx context.Context) ([]ingestion.Activity, error) {
	a.logger.Info("utility_bills adapter called via scheduled ingestion - no automatic source configured")
	return []ingestion.Activity{}, nil
}

// SetStore updates the activity store used by the adapter.
// This is useful when the adapter needs to be reconfigured after creation.
func (a *Adapter) SetStore(store ingestion.ActivityStore) {
	a.config.Store = store
}

// =============================================================================
// File-Based Ingestion Methods
// =============================================================================

// IngestFile processes a single utility bill file and returns the extracted activities.
//
// The file format is auto-detected from the filename and content.
// Supported formats: CSV, JSON, PDF (requires library), Excel (requires library)
//
// Returns:
//   - activities: Successfully parsed and validated activities
//   - importErrors: Non-fatal errors encountered during parsing
//   - error: Fatal error that prevented processing
func (a *Adapter) IngestFile(ctx context.Context, filename string, content io.Reader) (
	activities []ingestion.Activity,
	importErrors []ingestion.ImportError,
	err error,
) {
	a.logger.Info("ingesting utility bill file",
		"filename", filename,
		"org_id", a.config.DefaultOrgID)

	startTime := time.Now()

	// Parse the file
	result, err := a.parser.Parse(ctx, filename, content)
	if err != nil {
		a.logger.Error("failed to parse utility bill file",
			"filename", filename,
			"error", err)
		return nil, nil, fmt.Errorf("parse file: %w", err)
	}

	a.logger.Info("file parsed successfully",
		"filename", filename,
		"activities", len(result.Activities),
		"errors", len(result.Errors),
		"duration", time.Since(startTime))

	// Apply deduplication if enabled
	if a.config.EnableDeduplication {
		result.Activities = a.deduplicateActivities(ctx, result.Activities)
	}

	// Enrich activities if enabled
	if a.config.AutoEnrichLocation {
		result.Activities = a.enrichActivities(ctx, result.Activities)
	}

	// Persist activities if store is configured
	if a.config.Store != nil && len(result.Activities) > 0 {
		if err := a.config.Store.SaveBatch(ctx, result.Activities); err != nil {
			a.logger.Error("failed to persist activities",
				"count", len(result.Activities),
				"error", err)
			return result.Activities, result.Errors, fmt.Errorf("persist activities: %w", err)
		}

		a.logger.Info("activities persisted successfully",
			"count", len(result.Activities),
			"org_id", a.config.DefaultOrgID)
	}

	// Update deduplication cache
	if a.config.EnableDeduplication {
		a.updateDeduplicationCache(result.Activities)
	}

	return result.Activities, result.Errors, nil
}

// IngestMultipartFile processes a multipart file upload (from HTTP form).
func (a *Adapter) IngestMultipartFile(ctx context.Context, fileHeader *multipart.FileHeader) (
	activities []ingestion.Activity,
	importErrors []ingestion.ImportError,
	err error,
) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, nil, fmt.Errorf("open multipart file: %w", err)
	}
	defer file.Close()

	return a.IngestFile(ctx, fileHeader.Filename, file)
}

// IngestFiles processes multiple utility bill files concurrently.
//
// Files are processed in parallel up to MaxConcurrentParsing limit.
// This is useful for batch uploads or email attachment processing.
//
// Returns:
//   - BatchResult containing aggregated results from all files
func (a *Adapter) IngestFiles(ctx context.Context, files map[string]io.Reader) (*BatchResult, error) {
	if len(files) == 0 {
		return &BatchResult{}, nil
	}

	a.logger.Info("ingesting multiple utility bill files",
		"count", len(files),
		"org_id", a.config.DefaultOrgID)

	startTime := time.Now()

	// Create semaphore for concurrency control
	semaphore := make(chan struct{}, a.config.MaxConcurrentParsing)
	results := make(chan *FileResult, len(files))
	var wg sync.WaitGroup

	// Process files concurrently
	for filename, content := range files {
		wg.Add(1)
		go func(name string, r io.Reader) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Process file
			activities, errors, err := a.IngestFile(ctx, name, r)

			results <- &FileResult{
				Filename:   name,
				Activities: activities,
				Errors:     errors,
				Error:      err,
			}
		}(filename, content)
	}

	// Wait for all files to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Aggregate results
	batchResult := &BatchResult{
		FileResults:  make([]*FileResult, 0, len(files)),
		TotalFiles:   len(files),
		StartedAt:    startTime,
		Activities:   make([]ingestion.Activity, 0),
		ImportErrors: make([]ingestion.ImportError, 0),
	}

	for result := range results {
		batchResult.FileResults = append(batchResult.FileResults, result)
		batchResult.Activities = append(batchResult.Activities, result.Activities...)
		batchResult.ImportErrors = append(batchResult.ImportErrors, result.Errors...)

		if result.Error != nil {
			batchResult.FailedFiles++
		} else {
			batchResult.SuccessFiles++
		}
	}

	batchResult.CompletedAt = time.Now()
	batchResult.TotalActivities = len(batchResult.Activities)
	batchResult.TotalErrors = len(batchResult.ImportErrors)

	a.logger.Info("batch ingestion completed",
		"total_files", batchResult.TotalFiles,
		"success_files", batchResult.SuccessFiles,
		"failed_files", batchResult.FailedFiles,
		"total_activities", batchResult.TotalActivities,
		"duration", batchResult.CompletedAt.Sub(batchResult.StartedAt))

	return batchResult, nil
}

// =============================================================================
// Deduplication
// =============================================================================

// deduplicateActivities removes duplicate bills based on meter ID and period.
func (a *Adapter) deduplicateActivities(ctx context.Context, activities []ingestion.Activity) []ingestion.Activity {
	if !a.config.EnableDeduplication {
		return activities
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	cutoff := time.Now().Add(-a.config.DeduplicationWindow)
	deduplicated := make([]ingestion.Activity, 0, len(activities))

	for _, activity := range activities {
		key := deduplicationKey(activity)

		// Check cache
		if lastSeen, exists := a.deduplicationCache[key]; exists {
			if lastSeen.After(cutoff) {
				a.logger.Debug("skipping duplicate activity",
					"meter_id", activity.MeterID,
					"period_start", activity.PeriodStart,
					"last_seen", lastSeen)
				continue
			}
		}

		deduplicated = append(deduplicated, activity)
	}

	if len(deduplicated) < len(activities) {
		a.logger.Info("deduplication removed activities",
			"original", len(activities),
			"deduplicated", len(deduplicated),
			"removed", len(activities)-len(deduplicated))
	}

	return deduplicated
}

// updateDeduplicationCache adds activities to the deduplication cache.
func (a *Adapter) updateDeduplicationCache(activities []ingestion.Activity) {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-a.config.DeduplicationWindow)

	// Clean old entries
	for key, timestamp := range a.deduplicationCache {
		if timestamp.Before(cutoff) {
			delete(a.deduplicationCache, key)
		}
	}

	// Add new activities
	for _, activity := range activities {
		key := deduplicationKey(activity)
		a.deduplicationCache[key] = now
	}
}

// deduplicationKey generates a unique key for an activity.
func deduplicationKey(activity ingestion.Activity) string {
	return fmt.Sprintf("%s:%s:%s:%d",
		activity.OrgID,
		activity.MeterID,
		activity.PeriodStart.Format("2006-01-02"),
		int64(activity.Quantity*1000)) // Include quantity to handle corrections
}

// =============================================================================
// Enrichment
// =============================================================================

// enrichActivities enhances activities with additional context and standardization.
func (a *Adapter) enrichActivities(ctx context.Context, activities []ingestion.Activity) []ingestion.Activity {
	enriched := make([]ingestion.Activity, len(activities))

	for i, activity := range activities {
		// Apply location mapping based on provider
		if provider, ok := activity.Metadata["provider"]; ok {
			if location, exists := a.config.UtilityProviderMappings[provider]; exists && activity.Location == a.config.DefaultLocation {
				activity.Location = location
				activity.Metadata["location_source"] = "provider_mapping"
			}
		}

		// Standardize category names
		activity.Category = standardizeCategory(activity.Category)

		// Add enrichment timestamp
		if activity.Metadata == nil {
			activity.Metadata = make(map[string]string)
		}
		activity.Metadata["enriched_at"] = time.Now().UTC().Format(time.RFC3339)

		enriched[i] = activity
	}

	return enriched
}

// standardizeCategory normalizes category names to standard values.
func standardizeCategory(category string) string {
	category = strings.ToLower(category)
	switch category {
	case "electric", "power", "electricity", "elec":
		return "electricity"
	case "gas", "natural_gas", "ng":
		return "natural_gas"
	case "water", "h2o":
		return "water"
	case "steam", "district_heating":
		return "steam"
	default:
		return category
	}
}

// =============================================================================
// Result Types
// =============================================================================

// FileResult contains the results of processing a single file.
type FileResult struct {
	Filename   string
	Activities []ingestion.Activity
	Errors     []ingestion.ImportError
	Error      error
}

// BatchResult contains the aggregated results of processing multiple files.
type BatchResult struct {
	FileResults     []*FileResult
	TotalFiles      int
	SuccessFiles    int
	FailedFiles     int
	TotalActivities int
	TotalErrors     int
	StartedAt       time.Time
	CompletedAt     time.Time
	Activities      []ingestion.Activity
	ImportErrors    []ingestion.ImportError
}

// HasErrors returns true if any files failed or had import errors.
func (b *BatchResult) HasErrors() bool {
	return b.FailedFiles > 0 || b.TotalErrors > 0
}

// SuccessRate returns the percentage of successful file imports (0.0 to 1.0).
func (b *BatchResult) SuccessRate() float64 {
	if b.TotalFiles == 0 {
		return 1.0
	}
	return float64(b.SuccessFiles) / float64(b.TotalFiles)
}

// Duration returns how long the batch processing took.
func (b *BatchResult) Duration() time.Duration {
	return b.CompletedAt.Sub(b.StartedAt)
}

// Summary returns a human-readable summary of the batch results.
func (b *BatchResult) Summary() string {
	return fmt.Sprintf("Processed %d files (%d succeeded, %d failed) in %v: %d activities, %d errors",
		b.TotalFiles, b.SuccessFiles, b.FailedFiles, b.Duration(),
		b.TotalActivities, b.TotalErrors)
}
