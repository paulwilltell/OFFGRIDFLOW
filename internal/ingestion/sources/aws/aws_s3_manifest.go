// Package aws provides S3 manifest parsing for AWS Cost and Usage Reports.
// CUR stores a manifest.json file that lists the actual report files.
package aws

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
)

// =============================================================================
// S3 Manifest Types (AWS CUR format)
// =============================================================================

// S3Manifest represents the AWS CUR manifest.json file structure.
type S3Manifest struct {
	AssemblyID    string         `json:"assemblyId"`
	InvoiceID     string         `json:"invoiceId"`
	BillingPeriod BillingPeriod  `json:"billingPeriod"`
	Files         []ManifestFile `json:"files"`
	Charset       string         `json:"charset"`
	ContentType   string         `json:"contentType"`
	ReportKeys    []string       `json:"reportKeys"`
	ReportName    string         `json:"reportName"`
	Bucket        string         `json:"bucket"`
	ColumnHeaders []string       `json:"columnHeaders"`
	IsTruncated   bool           `json:"isTruncated"`
}

// BillingPeriod represents the billing period in the manifest.
type BillingPeriod struct {
	Start string `json:"start"` // YYYY-MM-DDT00:00:00.000Z
	End   string `json:"end"`
}

// ManifestFile represents a single file entry in the manifest.
type ManifestFile struct {
	Key       string `json:"key"`       // S3 key (path)
	Size      int64  `json:"size"`      // File size in bytes
	ReportKey string `json:"reportKey"` // Report identifier
}

// =============================================================================
// S3 Manifest Parsing
// =============================================================================

// ParseS3Manifest parses an S3 manifest.json file.
// The manifest file is produced by AWS Cost and Usage Reports service.
func ParseS3Manifest(data []byte) (*S3Manifest, error) {
	var manifest S3Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("aws: failed to parse manifest.json: %w", err)
	}

	// Validate manifest has required fields
	if manifest.BillingPeriod.Start == "" || manifest.BillingPeriod.End == "" {
		return nil, fmt.Errorf("aws: manifest missing billing period")
	}

	if len(manifest.Files) == 0 {
		return nil, fmt.Errorf("aws: manifest contains no files")
	}

	return &manifest, nil
}

// ParseS3ManifestFromReader parses manifest from an io.Reader (e.g., S3 GetObject response).
func ParseS3ManifestFromReader(r io.Reader) (*S3Manifest, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("aws: failed to read manifest: %w", err)
	}

	return ParseS3Manifest(data)
}

// GetReportFiles returns only the report files (excluding parquet, etc).
// CUR can produce both CSV and Parquet formats; we prioritize CSV.
func (m *S3Manifest) GetReportFiles() []ManifestFile {
	var csvFiles []ManifestFile
	var otherFiles []ManifestFile

	for _, file := range m.Files {
		if isCURReportFile(file.Key) {
			if isCSVFile(file.Key) {
				csvFiles = append(csvFiles, file)
			} else {
				otherFiles = append(otherFiles, file)
			}
		}
	}

	// Return CSV files if available, otherwise other formats
	if len(csvFiles) > 0 {
		return csvFiles
	}
	return otherFiles
}

// GetFileByKey returns a specific file from the manifest by S3 key.
func (m *S3Manifest) GetFileByKey(key string) *ManifestFile {
	for i := range m.Files {
		if m.Files[i].Key == key {
			return &m.Files[i]
		}
	}
	return nil
}

// Summary returns a human-readable summary of the manifest.
func (m *S3Manifest) Summary() string {
	reportFiles := m.GetReportFiles()
	totalSize := int64(0)
	for _, f := range m.Files {
		totalSize += f.Size
	}

	return fmt.Sprintf(
		"AssemblyID=%s BillingPeriod=%s to %s Files=%d TotalSize=%.2fMB",
		m.AssemblyID,
		m.BillingPeriod.Start,
		m.BillingPeriod.End,
		len(reportFiles),
		float64(totalSize)/(1024*1024),
	)
}

// =============================================================================
// Helper Functions
// =============================================================================

// isCURReportFile checks if the S3 key is a CUR data file (not manifest, etc).
func isCURReportFile(key string) bool {
	// CUR files follow pattern: .../YYYY/MM/DDTHH/organizationid-accountid-curversion-region-service-guid.csv.gz
	// Manifest is: .../YYYY/MM/DDTHH/manifest.json

	// Skip manifest files and other metadata
	if contains(key, "manifest.json") ||
		contains(key, "report.json") ||
		contains(key, ".txt") ||
		contains(key, ".parquet.snappy") {
		return false
	}

	// Accept CSV and other data formats
	return contains(key, ".csv") || contains(key, ".parquet")
}

// isCSVFile checks if the file is CSV format (preferred over Parquet).
func isCSVFile(key string) bool {
	return contains(key, ".csv")
}

// contains is a helper for string containment checks.
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// =============================================================================
// Manifest Validation
// =============================================================================

// ValidateManifest checks if the manifest is valid for ingestion.
func ValidateManifest(m *S3Manifest, logger *slog.Logger) error {
	if m == nil {
		return fmt.Errorf("aws: manifest is nil")
	}

	if m.AssemblyID == "" {
		return fmt.Errorf("aws: manifest missing assemblyId")
	}

	if m.BillingPeriod.Start == "" || m.BillingPeriod.End == "" {
		return fmt.Errorf("aws: manifest missing billingPeriod")
	}

	if len(m.Files) == 0 {
		return fmt.Errorf("aws: manifest contains no files")
	}

	reportFiles := m.GetReportFiles()
	if len(reportFiles) == 0 {
		if logger != nil {
			logger.Warn("aws: no report files found in manifest", "totalFiles", len(m.Files))
		}
		return fmt.Errorf("aws: no report files found in manifest")
	}

	if logger != nil {
		logger.Info("aws: manifest validated", "assemblyId", m.AssemblyID, "reportFiles", len(reportFiles), "totalFiles", len(m.Files))
	}

	return nil
}

// =============================================================================
// Manifest Comparison
// =============================================================================

// IsNewManifest checks if this manifest is newer than a previous one.
// Used to determine if we need to re-ingest data.
func IsNewManifest(current, previous *S3Manifest) bool {
	if previous == nil {
		return true
	}

	if current == nil {
		return false
	}

	// Compare assembly IDs (each billing period has a unique assembly ID)
	return current.AssemblyID != previous.AssemblyID
}

// GetNewFiles returns files that are in current manifest but not in previous.
// Used for incremental ingestion (avoid re-downloading unchanged files).
func GetNewFiles(current, previous *S3Manifest) []ManifestFile {
	if previous == nil {
		return current.Files
	}

	// Build map of previous file keys
	prevKeys := make(map[string]bool)
	for _, f := range previous.Files {
		prevKeys[f.Key] = true
	}

	// Return files not in previous manifest
	var newFiles []ManifestFile
	for _, f := range current.Files {
		if !prevKeys[f.Key] {
			newFiles = append(newFiles, f)
		}
	}

	return newFiles
}
