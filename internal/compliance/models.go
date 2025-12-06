// Package compliance provides ESG/CSRD compliance report generation.
// Supports CSRD, SEC Climate Disclosure, California CCDAA, CBAM, and IFRS S2 standards.
package compliance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ReportType identifies different compliance report standards
type ReportType string

const (
	ReportTypeCSRD       ReportType = "CSRD"       // Corporate Sustainability Reporting Directive (EU)
	ReportTypeSEC        ReportType = "SEC"        // SEC Climate Disclosure (US)
	ReportTypeCalifornia ReportType = "CALIFORNIA" // California Climate Corporate Data Accountability Act
	ReportTypeCBAM       ReportType = "CBAM"       // Carbon Border Adjustment Mechanism (EU)
	ReportTypeIFRS       ReportType = "IFRS_S2"    // IFRS S2 Climate-related Disclosures
)

// ReportStatus tracks the lifecycle of a compliance report
type ReportStatus string

const (
	StatusDraft     ReportStatus = "draft"
	StatusGenerating ReportStatus = "generating"
	StatusReview    ReportStatus = "review"
	StatusApproved  ReportStatus = "approved"
	StatusPublished ReportStatus = "published"
	StatusArchived  ReportStatus = "archived"
)

// Report represents a compliance report with all required metadata
type Report struct {
	ID           uuid.UUID    `json:"id"`
	TenantID     uuid.UUID    `json:"tenant_id"`
	ReportType   ReportType   `json:"report_type"`
	ReportingYear int         `json:"reporting_year"`
	PeriodStart  time.Time    `json:"period_start"`
	PeriodEnd    time.Time    `json:"period_end"`

	// Required metadata for compliance
	ReportHash               string                 `json:"report_hash"`
	DataQualityScore         float64                `json:"data_quality_score"`
	CompletenessPercentage   float64                `json:"completeness_percentage"`
	MissingDataPoints        map[string]interface{} `json:"missing_data_points"`
	CalculationMethodology   string                 `json:"calculation_methodology"`

	// Emissions summary
	Scope1EmissionsTonnes float64 `json:"scope1_emissions_tonnes"`
	Scope2EmissionsTonnes float64 `json:"scope2_emissions_tonnes"`
	Scope3EmissionsTonnes float64 `json:"scope3_emissions_tonnes"`
	TotalEmissionsTonnes  float64 `json:"total_emissions_tonnes"`

	// File references
	PDFURL      string `json:"pdf_url,omitempty"`
	PDFFileSize int64  `json:"pdf_file_size,omitempty"`
	XBRLURL     string `json:"xbrl_url,omitempty"`
	XBRLFileSize int64  `json:"xbrl_file_size,omitempty"`

	// Audit trail
	GeneratedBy          *uuid.UUID `json:"generated_by,omitempty"`
	GenerationTimestamp  time.Time  `json:"generation_timestamp"`
	ApprovedBy           *uuid.UUID `json:"approved_by,omitempty"`
	ApprovedAt           *time.Time `json:"approved_at,omitempty"`

	// Status tracking
	Status  ReportStatus `json:"status"`
	Version int          `json:"version"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EmissionsData contains the calculated emissions for report generation
type EmissionsData struct {
	Scope1Tonnes float64                  `json:"scope1_tonnes"`
	Scope2Tonnes float64                  `json:"scope2_tonnes"`
	Scope3Tonnes float64                  `json:"scope3_tonnes"`
	TotalTonnes  float64                  `json:"total_tonnes"`
	Breakdown    map[string]interface{}   `json:"breakdown"` // Category-level details
	Activities   []ActivityEmission       `json:"activities"`
}

// ActivityEmission represents a single activity's emission contribution
type ActivityEmission struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Category     string    `json:"category"`
	Scope        string    `json:"scope"`
	Quantity     float64   `json:"quantity"`
	Unit         string    `json:"unit"`
	EmissionsTonnes float64 `json:"emissions_tonnes"`
	EmissionFactor  float64 `json:"emission_factor"`
	Location     string    `json:"location"`
	Period       string    `json:"period"`
}

// DataQualityMetrics tracks the quality and completeness of input data
type DataQualityMetrics struct {
	TotalActivities      int                    `json:"total_activities"`
	CompleteActivities   int                    `json:"complete_activities"`
	CompletenessPercentage float64              `json:"completeness_percentage"`
	MissingFields        map[string]int         `json:"missing_fields"`
	LowQualityFactors    int                    `json:"low_quality_factors"`
	OutdatedFactors      int                    `json:"outdated_factors"`
	DataQualityScore     float64                `json:"data_quality_score"`
	Warnings             []string               `json:"warnings"`
}

// CalculateDataQuality analyzes emissions data and calculates quality metrics
func CalculateDataQuality(data EmissionsData) DataQualityMetrics {
	metrics := DataQualityMetrics{
		TotalActivities:   len(data.Activities),
		MissingFields:     make(map[string]int),
		Warnings:          []string{},
	}

	completeCount := 0
	for _, activity := range data.Activities {
		isComplete := true

		if activity.Name == "" {
			metrics.MissingFields["name"]++
			isComplete = false
		}
		if activity.Category == "" {
			metrics.MissingFields["category"]++
			isComplete = false
		}
		if activity.Location == "" {
			metrics.MissingFields["location"]++
			isComplete = false
		}
		if activity.Period == "" {
			metrics.MissingFields["period"]++
			isComplete = false
		}
		if activity.Quantity == 0 {
			metrics.MissingFields["quantity"]++
			isComplete = false
		}
		if activity.Unit == "" {
			metrics.MissingFields["unit"]++
			isComplete = false
		}

		if isComplete {
			completeCount++
		}
	}

	metrics.CompleteActivities = completeCount
	if metrics.TotalActivities > 0 {
		metrics.CompletenessPercentage = (float64(completeCount) / float64(metrics.TotalActivities)) * 100.0
	}

	// Calculate overall data quality score (0-100)
	// Based on: completeness (60%), factor quality (20%), factor freshness (20%)
	completenessScore := metrics.CompletenessPercentage * 0.6

	factorQualityScore := 100.0
	if metrics.TotalActivities > 0 {
		factorQualityScore = (1.0 - (float64(metrics.LowQualityFactors) / float64(metrics.TotalActivities))) * 100.0 * 0.2
	}

	factorFreshnessScore := 100.0
	if metrics.TotalActivities > 0 {
		factorFreshnessScore = (1.0 - (float64(metrics.OutdatedFactors) / float64(metrics.TotalActivities))) * 100.0 * 0.2
	}

	metrics.DataQualityScore = completenessScore + factorQualityScore + factorFreshnessScore

	// Add warnings for low scores
	if metrics.CompletenessPercentage < 80 {
		metrics.Warnings = append(metrics.Warnings, "Data completeness below 80% - missing required fields")
	}
	if metrics.DataQualityScore < 70 {
		metrics.Warnings = append(metrics.Warnings, "Overall data quality score below 70 - review data before finalizing report")
	}

	return metrics
}

// CalculateReportHash generates SHA-256 hash of report content for integrity verification
func CalculateReportHash(report *Report, contentBytes []byte) string {
	// Create hash input from report metadata + content
	hashInput := struct {
		TenantID      uuid.UUID
		ReportType    ReportType
		Year          int
		Scope1        float64
		Scope2        float64
		Scope3        float64
		ContentHash   string
	}{
		TenantID:    report.TenantID,
		ReportType:  report.ReportType,
		Year:        report.ReportingYear,
		Scope1:      report.Scope1EmissionsTonnes,
		Scope2:      report.Scope2EmissionsTonnes,
		Scope3:      report.Scope3EmissionsTonnes,
		ContentHash: hashContent(contentBytes),
	}

	hashBytes, _ := json.Marshal(hashInput)
	hash := sha256.Sum256(hashBytes)
	return hex.EncodeToString(hash[:])
}

// hashContent generates SHA-256 hash of content bytes
func hashContent(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

// ReportRequest contains the input parameters for generating a compliance report
type ReportRequest struct {
	TenantID      uuid.UUID
	UserID        uuid.UUID
	ReportType    ReportType
	ReportingYear int
	PeriodStart   time.Time
	PeriodEnd     time.Time
}

// Validate checks if the report request is valid
func (r ReportRequest) Validate() error {
	if r.TenantID == uuid.Nil {
		return ErrInvalidTenant
	}
	if r.UserID == uuid.Nil {
		return ErrInvalidUser
	}
	if r.ReportingYear < 2000 || r.ReportingYear > time.Now().Year()+1 {
		return ErrInvalidYear
	}
	if r.PeriodStart.IsZero() || r.PeriodEnd.IsZero() {
		return ErrInvalidPeriod
	}
	if r.PeriodEnd.Before(r.PeriodStart) {
		return ErrInvalidPeriod
	}
	if !isValidReportType(r.ReportType) {
		return ErrInvalidReportType
	}
	return nil
}

func isValidReportType(rt ReportType) bool {
	switch rt {
	case ReportTypeCSRD, ReportTypeSEC, ReportTypeCalifornia, ReportTypeCBAM, ReportTypeIFRS:
		return true
	default:
		return false
	}
}
