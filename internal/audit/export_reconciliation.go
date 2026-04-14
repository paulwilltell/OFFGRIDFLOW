package audit

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// ReportExport tracks an individual export for reconciliation.
type ReportExport struct {
	ID              string    `json:"id"`
	OrganizationID  string    `json:"organization_id"`
	ReportID        string    `json:"report_id"`
	ReportType      string    `json:"report_type"`
	ExportFormat    string    `json:"export_format"`
	ExportPurpose   string    `json:"export_purpose"`
	DataChecksum    string    `json:"data_checksum"`
	Scope1AtExport  float64   `json:"scope1_at_export"`
	Scope2AtExport  float64   `json:"scope2_at_export"`
	Scope3AtExport  float64   `json:"scope3_at_export"`
	TotalAtExport   float64   `json:"total_at_export"`
	FileURL         string    `json:"file_url,omitempty"`
	FileSizeBytes   int64     `json:"file_size_bytes"`
	ExportedBy      *string   `json:"exported_by,omitempty"`
	ExportedAt      time.Time `json:"exported_at"`
	CreatedAt       time.Time `json:"created_at"`
}

// ExportReconciliation represents the result of comparing an export to current data.
type ExportReconciliation struct {
	ExportID          string  `json:"export_id"`
	ExportedAt        string  `json:"exported_at"`
	ExportFormat      string  `json:"export_format"`
	ExportPurpose     string  `json:"export_purpose"`
	Scope1AtExport    float64 `json:"scope1_at_export"`
	Scope2AtExport    float64 `json:"scope2_at_export"`
	Scope3AtExport    float64 `json:"scope3_at_export"`
	TotalAtExport     float64 `json:"total_at_export"`
	Scope1Current     float64 `json:"scope1_current"`
	Scope2Current     float64 `json:"scope2_current"`
	Scope3Current     float64 `json:"scope3_current"`
	TotalCurrent      float64 `json:"total_current"`
	IsReconciled      bool    `json:"is_reconciled"`
	DriftPercent      float64 `json:"drift_percent"`
	ChecksumMatch     bool    `json:"checksum_match"`
}

// ComputeChecksum creates a deterministic hash for reconciliation.
func ComputeChecksum(scope1, scope2, scope3 float64, reportType string, year int) string {
	data := fmt.Sprintf("%.6f|%.6f|%.6f|%s|%d", scope1, scope2, scope3, reportType, year)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash[:16])
}

// RecordExport creates an export record for reconciliation tracking.
func (s *Store) RecordExport(ctx context.Context, e *ReportExport) error {
	query := `
INSERT INTO report_exports (
  organization_id, report_id, report_type, export_format, export_purpose,
  data_checksum, scope1_at_export, scope2_at_export, scope3_at_export, total_at_export,
  file_url, file_size_bytes, exported_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING id, exported_at, created_at`

	return s.db.QueryRowContext(ctx, query,
		e.OrganizationID, e.ReportID, e.ReportType, e.ExportFormat, e.ExportPurpose,
		e.DataChecksum, e.Scope1AtExport, e.Scope2AtExport, e.Scope3AtExport, e.TotalAtExport,
		e.FileURL, e.FileSizeBytes, e.ExportedBy,
	).Scan(&e.ID, &e.ExportedAt, &e.CreatedAt)
}

// GetExportHistory returns export records for a report.
func (s *Store) GetExportHistory(ctx context.Context, orgID, reportID string) ([]ReportExport, error) {
	query := `
SELECT id, organization_id, report_id, report_type, export_format, export_purpose,
  data_checksum, scope1_at_export, scope2_at_export, scope3_at_export, total_at_export,
  COALESCE(file_url,''), file_size_bytes, exported_by, exported_at, created_at
FROM report_exports WHERE organization_id = $1`
	args := []any{orgID}

	if reportID != "" {
		query += " AND report_id = $2"
		args = append(args, reportID)
	}
	query += " ORDER BY exported_at DESC LIMIT 50"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get export history: %w", err)
	}
	defer rows.Close()

	var exports []ReportExport
	for rows.Next() {
		var e ReportExport
		if err := rows.Scan(
			&e.ID, &e.OrganizationID, &e.ReportID, &e.ReportType, &e.ExportFormat, &e.ExportPurpose,
			&e.DataChecksum, &e.Scope1AtExport, &e.Scope2AtExport, &e.Scope3AtExport, &e.TotalAtExport,
			&e.FileURL, &e.FileSizeBytes, &e.ExportedBy, &e.ExportedAt, &e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan export: %w", err)
		}
		exports = append(exports, e)
	}
	return exports, rows.Err()
}

// ReconcileExport compares an export's data to current report values.
func (s *Store) ReconcileExport(ctx context.Context, exportID, orgID string) (*ExportReconciliation, error) {
	// Get the export record
	var e ReportExport
	query := `
SELECT id, report_id, report_type, export_format, export_purpose,
  data_checksum, scope1_at_export, scope2_at_export, scope3_at_export, total_at_export,
  exported_at
FROM report_exports WHERE id = $1 AND organization_id = $2`

	err := s.db.QueryRowContext(ctx, query, exportID, orgID).Scan(
		&e.ID, &e.ReportID, &e.ReportType, &e.ExportFormat, &e.ExportPurpose,
		&e.DataChecksum, &e.Scope1AtExport, &e.Scope2AtExport, &e.Scope3AtExport, &e.TotalAtExport,
		&e.ExportedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("export %s not found", exportID)
	}
	if err != nil {
		return nil, fmt.Errorf("get export for reconciliation: %w", err)
	}

	// Get current report values
	var scope1, scope2, scope3 float64
	reportQuery := `
SELECT COALESCE(scope1_emissions, 0), COALESCE(scope2_emissions, 0), COALESCE(scope3_emissions, 0)
FROM compliance_reports WHERE id = $1 AND organization_id = $2`

	err = s.db.QueryRowContext(ctx, reportQuery, e.ReportID, orgID).Scan(&scope1, &scope2, &scope3)
	if err != nil {
		return nil, fmt.Errorf("get current report data: %w", err)
	}

	totalCurrent := scope1 + scope2 + scope3
	currentChecksum := ComputeChecksum(scope1, scope2, scope3, e.ReportType, e.ExportedAt.Year())

	drift := 0.0
	if e.TotalAtExport > 0 {
		drift = ((totalCurrent - e.TotalAtExport) / e.TotalAtExport) * 100
	}

	recon := &ExportReconciliation{
		ExportID:       e.ID,
		ExportedAt:     e.ExportedAt.Format(time.RFC3339),
		ExportFormat:   e.ExportFormat,
		ExportPurpose:  e.ExportPurpose,
		Scope1AtExport: e.Scope1AtExport,
		Scope2AtExport: e.Scope2AtExport,
		Scope3AtExport: e.Scope3AtExport,
		TotalAtExport:  e.TotalAtExport,
		Scope1Current:  scope1,
		Scope2Current:  scope2,
		Scope3Current:  scope3,
		TotalCurrent:   totalCurrent,
		IsReconciled:   currentChecksum == e.DataChecksum,
		DriftPercent:   drift,
		ChecksumMatch:  currentChecksum == e.DataChecksum,
	}

	return recon, nil
}

// StakeholderExportConfig defines parameters for stakeholder-ready outputs.
type StakeholderExportConfig struct {
	Purpose        string `json:"purpose"` // board_package, auditor_package, regulator_submission, stakeholder_review
	IncludeScope1  bool   `json:"include_scope1"`
	IncludeScope2  bool   `json:"include_scope2"`
	IncludeScope3  bool   `json:"include_scope3"`
	IncludeFactors bool   `json:"include_factors"`
	IncludeMethodology bool `json:"include_methodology"`
	IncludeAuditTrail  bool `json:"include_audit_trail"`
}

// StakeholderExportData contains all data needed for a stakeholder-ready export.
type StakeholderExportData struct {
	ReportType       string          `json:"report_type"`
	ReportYear       int             `json:"report_year"`
	OrganizationName string          `json:"organization_name"`
	GeneratedAt      time.Time       `json:"generated_at"`
	Scope1Total      float64         `json:"scope1_total"`
	Scope2Total      float64         `json:"scope2_total"`
	Scope3Total      float64         `json:"scope3_total"`
	GrandTotal       float64         `json:"grand_total"`
	Status           string          `json:"status"`
	ApprovedBy       string          `json:"approved_by,omitempty"`
	ApprovedAt       string          `json:"approved_at,omitempty"`
	FactorSources    []string        `json:"factor_sources,omitempty"`
	Methodology      string          `json:"methodology"`
	AuditTrail       json.RawMessage `json:"audit_trail,omitempty"`
	Checksum         string          `json:"checksum"`
	ExportPurpose    string          `json:"export_purpose"`
}

// GenerateStakeholderExport builds a complete stakeholder-ready data package.
func (s *Store) GenerateStakeholderExport(ctx context.Context, orgID, reportID string, config StakeholderExportConfig) (*StakeholderExportData, error) {
	var data StakeholderExportData
	data.GeneratedAt = time.Now().UTC()
	data.ExportPurpose = config.Purpose
	data.Methodology = "GHG Protocol Corporate Standard (Scope 1, 2, 3). Location-based and market-based methods for Scope 2. Activity-based and spend-based for Scope 3."

	// Get report data
	query := `
SELECT r.report_type, r.report_year, COALESCE(r.scope1_emissions,0), COALESCE(r.scope2_emissions,0),
  COALESCE(r.scope3_emissions,0), COALESCE(r.total_emissions_co2e,0), r.status,
  COALESCE(u.name, u.email, ''), COALESCE(r.approved_at::text, ''),
  COALESCE(o.name, '')
FROM compliance_reports r
LEFT JOIN users u ON r.approved_by = u.id
LEFT JOIN organizations o ON r.organization_id = o.id
WHERE r.id = $1 AND r.organization_id = $2`

	err := s.db.QueryRowContext(ctx, query, reportID, orgID).Scan(
		&data.ReportType, &data.ReportYear, &data.Scope1Total, &data.Scope2Total,
		&data.Scope3Total, &data.GrandTotal, &data.Status,
		&data.ApprovedBy, &data.ApprovedAt, &data.OrganizationName,
	)
	if err != nil {
		return nil, fmt.Errorf("get report for stakeholder export: %w", err)
	}

	data.Checksum = ComputeChecksum(data.Scope1Total, data.Scope2Total, data.Scope3Total, data.ReportType, data.ReportYear)

	// Get factor sources used
	if config.IncludeFactors {
		factorRows, err := s.db.QueryContext(ctx,
			`SELECT DISTINCT emission_factor_source FROM calculation_ledger WHERE tenant_id = $1 AND emission_factor_source != ''`, orgID)
		if err == nil {
			defer factorRows.Close()
			for factorRows.Next() {
				var src string
				if factorRows.Scan(&src) == nil {
					data.FactorSources = append(data.FactorSources, src)
				}
			}
		}
	}

	// Get audit trail
	if config.IncludeAuditTrail {
		trailRows, err := s.db.QueryContext(ctx,
			`SELECT json_agg(json_build_object(
				'action', action, 'field', field_name, 'old', old_value, 'new', new_value,
				'by', changed_by, 'at', changed_at
			) ORDER BY changed_at DESC)
			FROM change_log WHERE tenant_id = $1 AND entity_type = 'report' AND entity_id = $2`,
			orgID, reportID)
		if err == nil {
			defer trailRows.Close()
			if trailRows.Next() {
				trailRows.Scan(&data.AuditTrail)
			}
		}
	}

	return &data, nil
}
