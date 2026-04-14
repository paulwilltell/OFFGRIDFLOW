package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// FactorSnapshot represents a locked set of emission factors for a reporting period.
// Once locked, the factors in the snapshot are immutable — calculations using
// these factors can be exactly reproduced at any future date.
type FactorSnapshot struct {
	ID                    string          `json:"id"`
	OrganizationID        string          `json:"organization_id"`
	ReportingPeriodStart  string          `json:"reporting_period_start"`
	ReportingPeriodEnd    string          `json:"reporting_period_end"`
	SnapshotName          string          `json:"snapshot_name"`
	Factors               json.RawMessage `json:"factors"`
	FactorCount           int             `json:"factor_count"`
	Status                string          `json:"status"` // draft, locked, superseded
	LockedAt              *time.Time      `json:"locked_at,omitempty"`
	LockedBy              *string         `json:"locked_by,omitempty"`
	SourceRegistryVersion string          `json:"source_registry_version,omitempty"`
	Notes                 string          `json:"notes,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

// SnapshotFactor is a single factor within a snapshot — a frozen copy.
type SnapshotFactor struct {
	FactorID           string  `json:"factor_id"`
	Scope              int     `json:"scope"`
	Region             string  `json:"region"`
	Source             string  `json:"source"`
	Category           string  `json:"category"`
	Unit               string  `json:"unit"`
	ValueKgCO2ePerUnit float64 `json:"value_kg_co2e_per_unit"`
	Method             string  `json:"method"`
	DataSource         string  `json:"data_source"`
	Year               int     `json:"year,omitempty"`
}

// CreateFactorSnapshot creates a new draft factor snapshot for a reporting period.
func (s *Store) CreateFactorSnapshot(ctx context.Context, orgID, periodStart, periodEnd, name string, factors []SnapshotFactor, registryVersion string) (*FactorSnapshot, error) {
	factorsJSON, err := json.Marshal(factors)
	if err != nil {
		return nil, fmt.Errorf("marshal factors: %w", err)
	}

	snap := &FactorSnapshot{
		OrganizationID:        orgID,
		ReportingPeriodStart:  periodStart,
		ReportingPeriodEnd:    periodEnd,
		SnapshotName:          name,
		Factors:               factorsJSON,
		FactorCount:           len(factors),
		Status:                "draft",
		SourceRegistryVersion: registryVersion,
	}

	query := `
INSERT INTO factor_snapshots (
  organization_id, reporting_period_start, reporting_period_end, snapshot_name,
  factors, factor_count, status, source_registry_version
) VALUES ($1, $2, $3, $4, $5, $6, 'draft', $7)
RETURNING id, created_at, updated_at`

	err = s.db.QueryRowContext(ctx, query,
		orgID, periodStart, periodEnd, name, factorsJSON, len(factors), registryVersion,
	).Scan(&snap.ID, &snap.CreatedAt, &snap.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create factor snapshot: %w", err)
	}

	return snap, nil
}

// LockFactorSnapshot transitions a snapshot from draft to locked.
// Once locked, the factors are immutable.
func (s *Store) LockFactorSnapshot(ctx context.Context, id, orgID, lockedBy string) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE factor_snapshots SET status = 'locked', locked_at = NOW(), locked_by = $3, updated_at = NOW()
		 WHERE id = $1 AND organization_id = $2 AND status = 'draft'`,
		id, orgID, lockedBy)
	if err != nil {
		return fmt.Errorf("lock factor snapshot: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("snapshot %s not found or already locked", id)
	}
	return nil
}

// GetFactorSnapshot retrieves a single snapshot by ID.
func (s *Store) GetFactorSnapshot(ctx context.Context, id, orgID string) (*FactorSnapshot, error) {
	snap := &FactorSnapshot{}
	query := `
SELECT id, organization_id, reporting_period_start, reporting_period_end, snapshot_name,
  factors, factor_count, status, locked_at, locked_by, COALESCE(source_registry_version,''),
  COALESCE(notes,''), created_at, updated_at
FROM factor_snapshots WHERE id = $1 AND organization_id = $2`

	err := s.db.QueryRowContext(ctx, query, id, orgID).Scan(
		&snap.ID, &snap.OrganizationID, &snap.ReportingPeriodStart, &snap.ReportingPeriodEnd,
		&snap.SnapshotName, &snap.Factors, &snap.FactorCount, &snap.Status,
		&snap.LockedAt, &snap.LockedBy, &snap.SourceRegistryVersion,
		&snap.Notes, &snap.CreatedAt, &snap.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("factor snapshot %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get factor snapshot: %w", err)
	}
	return snap, nil
}

// ListFactorSnapshots returns all snapshots for an organization.
func (s *Store) ListFactorSnapshots(ctx context.Context, orgID string) ([]FactorSnapshot, error) {
	query := `
SELECT id, organization_id, reporting_period_start, reporting_period_end, snapshot_name,
  '[]'::jsonb, factor_count, status, locked_at, locked_by, COALESCE(source_registry_version,''),
  COALESCE(notes,''), created_at, updated_at
FROM factor_snapshots WHERE organization_id = $1 ORDER BY reporting_period_start DESC`

	rows, err := s.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("list factor snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []FactorSnapshot
	for rows.Next() {
		var snap FactorSnapshot
		if err := rows.Scan(
			&snap.ID, &snap.OrganizationID, &snap.ReportingPeriodStart, &snap.ReportingPeriodEnd,
			&snap.SnapshotName, &snap.Factors, &snap.FactorCount, &snap.Status,
			&snap.LockedAt, &snap.LockedBy, &snap.SourceRegistryVersion,
			&snap.Notes, &snap.CreatedAt, &snap.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan factor snapshot: %w", err)
		}
		snapshots = append(snapshots, snap)
	}
	return snapshots, rows.Err()
}

// GetLockedSnapshotForPeriod returns the locked snapshot covering a specific date.
func (s *Store) GetLockedSnapshotForPeriod(ctx context.Context, orgID string, date time.Time) (*FactorSnapshot, error) {
	snap := &FactorSnapshot{}
	query := `
SELECT id, organization_id, reporting_period_start, reporting_period_end, snapshot_name,
  factors, factor_count, status, locked_at, locked_by, COALESCE(source_registry_version,''),
  COALESCE(notes,''), created_at, updated_at
FROM factor_snapshots
WHERE organization_id = $1 AND status = 'locked'
  AND reporting_period_start <= $2 AND reporting_period_end >= $2
ORDER BY locked_at DESC LIMIT 1`

	err := s.db.QueryRowContext(ctx, query, orgID, date.Format("2006-01-02")).Scan(
		&snap.ID, &snap.OrganizationID, &snap.ReportingPeriodStart, &snap.ReportingPeriodEnd,
		&snap.SnapshotName, &snap.Factors, &snap.FactorCount, &snap.Status,
		&snap.LockedAt, &snap.LockedBy, &snap.SourceRegistryVersion,
		&snap.Notes, &snap.CreatedAt, &snap.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get locked snapshot for period: %w", err)
	}
	return snap, nil
}
