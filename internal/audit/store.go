// Package audit provides evidence vault, calculation ledger, approval workflow,
// and change logging for GHG Protocol audit-readiness.
package audit

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Store provides persistence for all audit-related operations.
type Store struct {
	db *sql.DB
}

// NewStore creates a new audit store.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// ============================================================================
// Calculation Ledger
// ============================================================================

// LedgerEntry represents one immutable calculation record.
type LedgerEntry struct {
	ID                   string    `json:"id"`
	TenantID             string    `json:"tenant_id"`
	Scope                string    `json:"scope"`
	Category             string    `json:"category,omitempty"`
	ActivityID           string    `json:"activity_id,omitempty"`
	Quantity             float64   `json:"quantity"`
	Unit                 string    `json:"unit"`
	EmissionFactorID     string    `json:"emission_factor_id"`
	EmissionFactorValue  float64   `json:"emission_factor_value"`
	EmissionFactorSource string    `json:"emission_factor_source"`
	EmissionFactorRegion string    `json:"emission_factor_region,omitempty"`
	Method               string    `json:"method"`
	ResultKgCO2e         float64   `json:"result_kg_co2e"`
	ResultTonnesCO2e     float64   `json:"result_tonnes_co2e"`
	PeriodStart          string    `json:"period_start"`
	PeriodEnd            string    `json:"period_end"`
	CalculatedAt         time.Time `json:"calculated_at"`
	Formula              string    `json:"formula"`
	Notes                string    `json:"notes,omitempty"`
	IsLocked             bool      `json:"is_locked"`
}

// RecordCalculation inserts an immutable calculation record into the ledger.
func (s *Store) RecordCalculation(ctx context.Context, entry *LedgerEntry) error {
	query := `
INSERT INTO calculation_ledger (
  tenant_id, scope, category, activity_id, quantity, unit,
  emission_factor_id, emission_factor_value, emission_factor_source, emission_factor_region,
  method, result_kg_co2e, result_tonnes_co2e,
  reporting_period_start, reporting_period_end, calculated_at, formula, notes
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
RETURNING id`
	return s.db.QueryRowContext(ctx, query,
		entry.TenantID, entry.Scope, entry.Category, entry.ActivityID,
		entry.Quantity, entry.Unit,
		entry.EmissionFactorID, entry.EmissionFactorValue, entry.EmissionFactorSource, entry.EmissionFactorRegion,
		entry.Method, entry.ResultKgCO2e, entry.ResultTonnesCO2e,
		entry.PeriodStart, entry.PeriodEnd, entry.CalculatedAt, entry.Formula, entry.Notes,
	).Scan(&entry.ID)
}

// GetLedgerByTenant returns all calculation records for a tenant, optionally filtered.
func (s *Store) GetLedgerByTenant(ctx context.Context, tenantID, scope string, limit int) ([]LedgerEntry, error) {
	query := `
SELECT id, tenant_id, scope, COALESCE(category,''), COALESCE(activity_id,''),
  quantity, unit, emission_factor_id, emission_factor_value, emission_factor_source,
  COALESCE(emission_factor_region,''), method, result_kg_co2e, result_tonnes_co2e,
  reporting_period_start, reporting_period_end, calculated_at, formula, COALESCE(notes,''), is_locked
FROM calculation_ledger WHERE tenant_id = $1`
	args := []any{tenantID}
	if scope != "" {
		query += " AND scope = $2"
		args = append(args, scope)
	}
	query += " ORDER BY calculated_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LedgerEntry
	for rows.Next() {
		var e LedgerEntry
		if err := rows.Scan(
			&e.ID, &e.TenantID, &e.Scope, &e.Category, &e.ActivityID,
			&e.Quantity, &e.Unit, &e.EmissionFactorID, &e.EmissionFactorValue, &e.EmissionFactorSource,
			&e.EmissionFactorRegion, &e.Method, &e.ResultKgCO2e, &e.ResultTonnesCO2e,
			&e.PeriodStart, &e.PeriodEnd, &e.CalculatedAt, &e.Formula, &e.Notes, &e.IsLocked,
		); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// ============================================================================
// Approval Workflow
// ============================================================================

// ApprovalRecord tracks the review/approval state of a reportable entity.
type ApprovalRecord struct {
	ID              string     `json:"id"`
	TenantID        string     `json:"tenant_id"`
	EntityType      string     `json:"entity_type"` // "report", "inventory", "scope2_data"
	EntityID        string     `json:"entity_id"`
	Status          string     `json:"status"` // draft, submitted, reviewed, approved, rejected
	PreparedBy      *string    `json:"prepared_by,omitempty"`
	PreparedAt      *time.Time `json:"prepared_at,omitempty"`
	ReviewedBy      *string    `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	ReviewNotes     *string    `json:"review_notes,omitempty"`
	ApprovedBy      *string    `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
	ApprovalNotes   *string    `json:"approval_notes,omitempty"`
	RejectedBy      *string    `json:"rejected_by,omitempty"`
	RejectedAt      *time.Time `json:"rejected_at,omitempty"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CreateApproval creates a new approval workflow record in draft status.
func (s *Store) CreateApproval(ctx context.Context, tenantID, entityType, entityID, preparedBy string) (*ApprovalRecord, error) {
	now := time.Now()
	rec := &ApprovalRecord{
		TenantID:   tenantID,
		EntityType: entityType,
		EntityID:   entityID,
		Status:     "draft",
		PreparedBy: &preparedBy,
		PreparedAt: &now,
	}
	query := `
INSERT INTO approval_workflow (tenant_id, entity_type, entity_id, status, prepared_by, prepared_at, created_at, updated_at)
VALUES ($1, $2, $3, 'draft', $4, $5, $5, $5) RETURNING id, created_at, updated_at`
	err := s.db.QueryRowContext(ctx, query, tenantID, entityType, entityID, preparedBy, now).
		Scan(&rec.ID, &rec.CreatedAt, &rec.UpdatedAt)
	return rec, err
}

// SubmitForReview moves a draft to submitted status.
func (s *Store) SubmitForReview(ctx context.Context, id, tenantID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE approval_workflow SET status = 'submitted', updated_at = NOW() WHERE id = $1 AND tenant_id = $2 AND status = 'draft'`,
		id, tenantID)
	return err
}

// Review marks an approval as reviewed.
func (s *Store) Review(ctx context.Context, id, tenantID, reviewerID, notes string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE approval_workflow SET status = 'reviewed', reviewed_by = $3, reviewed_at = NOW(), review_notes = $4, updated_at = NOW()
		 WHERE id = $1 AND tenant_id = $2 AND status = 'submitted'`,
		id, tenantID, reviewerID, notes)
	return err
}

// Approve marks an approval as approved and locks it.
func (s *Store) Approve(ctx context.Context, id, tenantID, approverID, notes string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE approval_workflow SET status = 'approved', approved_by = $3, approved_at = NOW(), approval_notes = $4, updated_at = NOW()
		 WHERE id = $1 AND tenant_id = $2 AND status = 'reviewed'`,
		id, tenantID, approverID, notes)
	return err
}

// Reject marks an approval as rejected with a reason.
func (s *Store) Reject(ctx context.Context, id, tenantID, rejectorID, reason string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE approval_workflow SET status = 'rejected', rejected_by = $3, rejected_at = NOW(), rejection_reason = $4, updated_at = NOW()
		 WHERE id = $1 AND tenant_id = $2 AND (status = 'submitted' OR status = 'reviewed')`,
		id, tenantID, rejectorID, reason)
	return err
}

// GetApprovalsByTenant returns all approval records for a tenant.
func (s *Store) GetApprovalsByTenant(ctx context.Context, tenantID string) ([]ApprovalRecord, error) {
	query := `
SELECT id, tenant_id, entity_type, entity_id, status,
  prepared_by, prepared_at, reviewed_by, reviewed_at, review_notes,
  approved_by, approved_at, approval_notes, rejected_by, rejected_at, rejection_reason,
  created_at, updated_at
FROM approval_workflow WHERE tenant_id = $1 ORDER BY updated_at DESC`
	rows, err := s.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []ApprovalRecord
	for rows.Next() {
		var r ApprovalRecord
		if err := rows.Scan(
			&r.ID, &r.TenantID, &r.EntityType, &r.EntityID, &r.Status,
			&r.PreparedBy, &r.PreparedAt, &r.ReviewedBy, &r.ReviewedAt, &r.ReviewNotes,
			&r.ApprovedBy, &r.ApprovedAt, &r.ApprovalNotes, &r.RejectedBy, &r.RejectedAt, &r.RejectionReason,
			&r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// ============================================================================
// Change Log
// ============================================================================

// ChangeEntry records a single data modification.
type ChangeEntry struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	Action     string    `json:"action"` // create, update, delete, lock, approve
	FieldName  string    `json:"field_name,omitempty"`
	OldValue   string    `json:"old_value,omitempty"`
	NewValue   string    `json:"new_value,omitempty"`
	ChangedBy  string    `json:"changed_by"`
	ChangedAt  time.Time `json:"changed_at"`
}

// LogChange records a data modification in the immutable change log.
func (s *Store) LogChange(ctx context.Context, tenantID, entityType, entityID, action, field, oldVal, newVal, changedBy string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO change_log (tenant_id, entity_type, entity_id, action, field_name, old_value, new_value, changed_by, changed_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())`,
		tenantID, entityType, entityID, action, field, oldVal, newVal, changedBy)
	return err
}

// GetChangeLog returns change history for a tenant, optionally filtered by entity.
func (s *Store) GetChangeLog(ctx context.Context, tenantID, entityType, entityID string, limit int) ([]ChangeEntry, error) {
	query := `SELECT id, tenant_id, entity_type, entity_id, action, COALESCE(field_name,''), COALESCE(old_value,''), COALESCE(new_value,''), COALESCE(changed_by::text,''), changed_at
FROM change_log WHERE tenant_id = $1`
	args := []any{tenantID}
	n := 2
	if entityType != "" {
		query += fmt.Sprintf(" AND entity_type = $%d", n)
		args = append(args, entityType)
		n++
	}
	if entityID != "" {
		query += fmt.Sprintf(" AND entity_id = $%d", n)
		args = append(args, entityID)
		n++
	}
	query += " ORDER BY changed_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []ChangeEntry
	for rows.Next() {
		var e ChangeEntry
		if err := rows.Scan(&e.ID, &e.TenantID, &e.EntityType, &e.EntityID, &e.Action, &e.FieldName, &e.OldValue, &e.NewValue, &e.ChangedBy, &e.ChangedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
