package audit

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

// AnomalyType classifies the kind of data quality issue detected.
type AnomalyType string

const (
	AnomalyQuantityOutlier       AnomalyType = "quantity_outlier"
	AnomalyDuplicateEntry        AnomalyType = "duplicate_entry"
	AnomalyMissingPeriod         AnomalyType = "missing_period"
	AnomalyUnitMismatch          AnomalyType = "unit_mismatch"
	AnomalyFactorDeviation       AnomalyType = "factor_deviation"
	AnomalySuddenChange          AnomalyType = "sudden_change"
	AnomalyNegativeValue         AnomalyType = "negative_value"
	AnomalyStaleData             AnomalyType = "stale_data"
	AnomalyScopeMisclassification AnomalyType = "scope_misclassification"
	AnomalyCompletenessGap       AnomalyType = "completeness_gap"
)

// Anomaly represents a detected data quality issue.
type Anomaly struct {
	ID               string      `json:"id"`
	OrganizationID   string      `json:"organization_id"`
	EntityType       string      `json:"entity_type"`
	EntityID         string      `json:"entity_id"`
	AnomalyType      AnomalyType `json:"anomaly_type"`
	Severity         string      `json:"severity"` // info, warning, critical
	Description      string      `json:"description"`
	ExpectedValue    *float64    `json:"expected_value,omitempty"`
	ActualValue      *float64    `json:"actual_value,omitempty"`
	DeviationPercent *float64    `json:"deviation_percent,omitempty"`
	DetectionRule    string      `json:"detection_rule,omitempty"`
	Status           string      `json:"status"` // open, acknowledged, resolved, dismissed
	ResolvedAt       *time.Time  `json:"resolved_at,omitempty"`
	ResolvedBy       *string     `json:"resolved_by,omitempty"`
	ResolutionNotes  string      `json:"resolution_notes,omitempty"`
	AssignedTo       *string     `json:"assigned_to,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

// RecordAnomaly inserts a detected data quality anomaly.
func (s *Store) RecordAnomaly(ctx context.Context, a *Anomaly) error {
	query := `
INSERT INTO data_quality_anomalies (
  organization_id, entity_type, entity_id, anomaly_type, severity,
  description, expected_value, actual_value, deviation_percent, detection_rule, status
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'open')
RETURNING id, created_at, updated_at`

	return s.db.QueryRowContext(ctx, query,
		a.OrganizationID, a.EntityType, a.EntityID, a.AnomalyType, a.Severity,
		a.Description, a.ExpectedValue, a.ActualValue, a.DeviationPercent, a.DetectionRule,
	).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
}

// GetAnomalies returns anomalies for an organization, filtered by status.
func (s *Store) GetAnomalies(ctx context.Context, orgID, status, severity string, limit int) ([]Anomaly, error) {
	query := `
SELECT id, organization_id, entity_type, entity_id, anomaly_type, severity,
  description, expected_value, actual_value, deviation_percent, COALESCE(detection_rule,''),
  status, resolved_at, resolved_by, COALESCE(resolution_notes,''), assigned_to,
  created_at, updated_at
FROM data_quality_anomalies WHERE organization_id = $1`
	args := []any{orgID}
	n := 2

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", n)
		args = append(args, status)
		n++
	}
	if severity != "" {
		query += fmt.Sprintf(" AND severity = $%d", n)
		args = append(args, severity)
		n++
	}

	query += " ORDER BY CASE severity WHEN 'critical' THEN 0 WHEN 'warning' THEN 1 ELSE 2 END, created_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		if isMissingAuditTableError(err) {
			return []Anomaly{}, nil
		}
		return nil, fmt.Errorf("get anomalies: %w", err)
	}
	defer rows.Close()

	var anomalies []Anomaly
	for rows.Next() {
		var a Anomaly
		if err := rows.Scan(
			&a.ID, &a.OrganizationID, &a.EntityType, &a.EntityID, &a.AnomalyType, &a.Severity,
			&a.Description, &a.ExpectedValue, &a.ActualValue, &a.DeviationPercent, &a.DetectionRule,
			&a.Status, &a.ResolvedAt, &a.ResolvedBy, &a.ResolutionNotes, &a.AssignedTo,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan anomaly: %w", err)
		}
		anomalies = append(anomalies, a)
	}
	return anomalies, rows.Err()
}

// ResolveAnomaly marks an anomaly as resolved.
func (s *Store) ResolveAnomaly(ctx context.Context, id, orgID, resolvedBy, notes string) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE data_quality_anomalies SET status = 'resolved', resolved_at = NOW(), resolved_by = $3,
		 resolution_notes = $4, updated_at = NOW()
		 WHERE id = $1 AND organization_id = $2 AND status IN ('open', 'acknowledged')`,
		id, orgID, resolvedBy, notes)
	if err != nil {
		return fmt.Errorf("resolve anomaly: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("anomaly %s not found or already resolved", id)
	}
	return nil
}

// DismissAnomaly marks an anomaly as dismissed (false positive).
func (s *Store) DismissAnomaly(ctx context.Context, id, orgID, dismissedBy, notes string) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE data_quality_anomalies SET status = 'dismissed', resolved_at = NOW(), resolved_by = $3,
		 resolution_notes = $4, updated_at = NOW()
		 WHERE id = $1 AND organization_id = $2 AND status IN ('open', 'acknowledged')`,
		id, orgID, dismissedBy, notes)
	if err != nil {
		return fmt.Errorf("dismiss anomaly: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("anomaly %s not found or already resolved", id)
	}
	return nil
}

// GetAnomalyCounts returns counts by status and severity for dashboard display.
func (s *Store) GetAnomalyCounts(ctx context.Context, orgID string) (map[string]int, error) {
	query := `
SELECT
  COUNT(*) FILTER (WHERE status = 'open') AS open_count,
  COUNT(*) FILTER (WHERE status = 'open' AND severity = 'critical') AS critical_count,
  COUNT(*) FILTER (WHERE status = 'open' AND severity = 'warning') AS warning_count,
  COUNT(*) FILTER (WHERE status = 'acknowledged') AS acknowledged_count,
  COUNT(*) FILTER (WHERE status = 'resolved') AS resolved_count,
  COUNT(*) AS total_count
FROM data_quality_anomalies WHERE organization_id = $1`

	var open, critical, warning, ack, resolved, total int
	err := s.db.QueryRowContext(ctx, query, orgID).Scan(&open, &critical, &warning, &ack, &resolved, &total)
	if err != nil {
		if isMissingAuditTableError(err) {
			return map[string]int{
				"open":         0,
				"critical":     0,
				"warning":      0,
				"acknowledged": 0,
				"resolved":     0,
				"total":        0,
			}, nil
		}
		return nil, fmt.Errorf("get anomaly counts: %w", err)
	}

	return map[string]int{
		"open":         open,
		"critical":     critical,
		"warning":      warning,
		"acknowledged": ack,
		"resolved":     resolved,
		"total":        total,
	}, nil
}

func isMissingAuditTableError(err error) bool {
	if err == nil {
		return false
	}

	message := err.Error()
	return strings.Contains(message, "42P01") || strings.Contains(strings.ToLower(message), "does not exist")
}

// RunAnomalyDetection scans activities for data quality issues.
// This is the core detection engine that creates anomaly records.
func (s *Store) RunAnomalyDetection(ctx context.Context, orgID string) (int, error) {
	detected := 0

	// 1. Detect quantity outliers (>3 standard deviations from mean per category)
	n, err := s.detectQuantityOutliers(ctx, orgID)
	if err != nil {
		return detected, fmt.Errorf("detect quantity outliers: %w", err)
	}
	detected += n

	// 2. Detect duplicate entries (same date, type, quantity within org)
	n, err = s.detectDuplicates(ctx, orgID)
	if err != nil {
		return detected, fmt.Errorf("detect duplicates: %w", err)
	}
	detected += n

	// 3. Detect sudden changes (>50% change from prior period)
	n, err = s.detectSuddenChanges(ctx, orgID)
	if err != nil {
		return detected, fmt.Errorf("detect sudden changes: %w", err)
	}
	detected += n

	// 4. Detect negative values
	n, err = s.detectNegativeValues(ctx, orgID)
	if err != nil {
		return detected, fmt.Errorf("detect negative values: %w", err)
	}
	detected += n

	return detected, nil
}

func (s *Store) detectQuantityOutliers(ctx context.Context, orgID string) (int, error) {
	// Find activities where quantity is >3 std deviations from mean for their category
	query := `
WITH stats AS (
  SELECT activity_type, AVG(quantity) as mean_qty, STDDEV(quantity) as std_qty, COUNT(*) as cnt
  FROM activities
  WHERE organization_id = $1 AND deleted_at IS NULL
  GROUP BY activity_type
  HAVING COUNT(*) >= 5 AND STDDEV(quantity) > 0
)
SELECT a.id, a.activity_type, a.quantity, s.mean_qty, s.std_qty,
  ABS(a.quantity - s.mean_qty) / s.std_qty as z_score
FROM activities a
JOIN stats s ON a.activity_type = s.activity_type
WHERE a.organization_id = $1 AND a.deleted_at IS NULL
  AND ABS(a.quantity - s.mean_qty) / s.std_qty > 3
  AND NOT EXISTS (
    SELECT 1 FROM data_quality_anomalies dqa
    WHERE dqa.entity_id = a.id::text AND dqa.anomaly_type = 'quantity_outlier' AND dqa.status IN ('open', 'acknowledged')
  )`

	rows, err := s.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var actID, actType string
		var quantity, meanQty, stdQty, zScore float64
		if err := rows.Scan(&actID, &actType, &quantity, &meanQty, &stdQty, &zScore); err != nil {
			return count, err
		}

		deviation := ((quantity - meanQty) / meanQty) * 100
		severity := "warning"
		if math.Abs(zScore) > 5 {
			severity = "critical"
		}

		a := &Anomaly{
			OrganizationID:   orgID,
			EntityType:       "activity",
			EntityID:         actID,
			AnomalyType:      AnomalyQuantityOutlier,
			Severity:         severity,
			Description:      fmt.Sprintf("%s quantity %.2f is %.1f standard deviations from mean (%.2f)", actType, quantity, zScore, meanQty),
			ExpectedValue:    &meanQty,
			ActualValue:      &quantity,
			DeviationPercent: &deviation,
			DetectionRule:    fmt.Sprintf("z-score > 3 (actual: %.2f)", zScore),
		}
		if err := s.RecordAnomaly(ctx, a); err != nil {
			return count, err
		}
		count++
	}
	return count, rows.Err()
}

func (s *Store) detectDuplicates(ctx context.Context, orgID string) (int, error) {
	query := `
WITH dupes AS (
  SELECT activity_type, activity_date, quantity, unit, COUNT(*) as cnt,
    array_agg(id ORDER BY created_at) as ids
  FROM activities
  WHERE organization_id = $1 AND deleted_at IS NULL
  GROUP BY activity_type, activity_date, quantity, unit
  HAVING COUNT(*) > 1
)
SELECT unnest(ids[2:]) as dupe_id, activity_type, activity_date, quantity, cnt
FROM dupes
WHERE NOT EXISTS (
  SELECT 1 FROM data_quality_anomalies dqa
  WHERE dqa.entity_id = ANY(dupes.ids[2:])::text AND dqa.anomaly_type = 'duplicate_entry' AND dqa.status IN ('open', 'acknowledged')
)`

	rows, err := s.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var dupeID, actType string
		var actDate time.Time
		var quantity float64
		var cnt int
		if err := rows.Scan(&dupeID, &actType, &actDate, &quantity, &cnt); err != nil {
			return count, err
		}

		a := &Anomaly{
			OrganizationID: orgID,
			EntityType:     "activity",
			EntityID:       dupeID,
			AnomalyType:    AnomalyDuplicateEntry,
			Severity:       "warning",
			Description:    fmt.Sprintf("Possible duplicate: %s on %s with quantity %.2f appears %d times", actType, actDate.Format("2006-01-02"), quantity, cnt),
			DetectionRule:  "same type+date+quantity+unit appears multiple times",
		}
		if err := s.RecordAnomaly(ctx, a); err != nil {
			return count, err
		}
		count++
	}
	return count, rows.Err()
}

func (s *Store) detectSuddenChanges(ctx context.Context, orgID string) (int, error) {
	query := `
WITH monthly AS (
  SELECT activity_type,
    DATE_TRUNC('month', activity_date) as month,
    SUM(quantity) as total_qty
  FROM activities
  WHERE organization_id = $1 AND deleted_at IS NULL
  GROUP BY activity_type, DATE_TRUNC('month', activity_date)
),
changes AS (
  SELECT m1.activity_type, m1.month, m1.total_qty as current_qty,
    m2.total_qty as prev_qty,
    CASE WHEN m2.total_qty > 0
      THEN ((m1.total_qty - m2.total_qty) / m2.total_qty) * 100
      ELSE NULL END as pct_change
  FROM monthly m1
  JOIN monthly m2 ON m1.activity_type = m2.activity_type
    AND m1.month = m2.month + INTERVAL '1 month'
)
SELECT activity_type, month, current_qty, prev_qty, pct_change
FROM changes
WHERE ABS(pct_change) > 50
  AND NOT EXISTS (
    SELECT 1 FROM data_quality_anomalies dqa
    WHERE dqa.organization_id = $1 AND dqa.anomaly_type = 'sudden_change'
      AND dqa.detection_rule LIKE '%' || activity_type || '%' || to_char(month, 'YYYY-MM') || '%'
      AND dqa.status IN ('open', 'acknowledged')
  )
ORDER BY month DESC LIMIT 20`

	rows, err := s.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var actType string
		var month time.Time
		var currentQty, prevQty, pctChange float64
		if err := rows.Scan(&actType, &month, &currentQty, &prevQty, &pctChange); err != nil {
			return count, err
		}

		severity := "warning"
		if math.Abs(pctChange) > 100 {
			severity = "critical"
		}

		direction := "increase"
		if pctChange < 0 {
			direction = "decrease"
		}

		a := &Anomaly{
			OrganizationID:   orgID,
			EntityType:       "activity",
			EntityID:         orgID,
			AnomalyType:      AnomalySuddenChange,
			Severity:         severity,
			Description:      fmt.Sprintf("%s shows %.0f%% %s in %s (%.2f vs prior %.2f)", actType, math.Abs(pctChange), direction, month.Format("Jan 2006"), currentQty, prevQty),
			ExpectedValue:    &prevQty,
			ActualValue:      &currentQty,
			DeviationPercent: &pctChange,
			DetectionRule:    fmt.Sprintf("month-over-month change >50%% for %s %s", actType, month.Format("2006-01")),
		}
		if err := s.RecordAnomaly(ctx, a); err != nil {
			return count, err
		}
		count++
	}
	return count, rows.Err()
}

func (s *Store) detectNegativeValues(ctx context.Context, orgID string) (int, error) {
	query := `
SELECT id, activity_type, quantity
FROM activities
WHERE organization_id = $1 AND deleted_at IS NULL AND quantity < 0
  AND NOT EXISTS (
    SELECT 1 FROM data_quality_anomalies dqa
    WHERE dqa.entity_id = id::text AND dqa.anomaly_type = 'negative_value' AND dqa.status IN ('open', 'acknowledged')
  )`

	rows, err := s.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var actID, actType string
		var quantity float64
		if err := rows.Scan(&actID, &actType, &quantity); err != nil {
			return count, err
		}

		a := &Anomaly{
			OrganizationID: orgID,
			EntityType:     "activity",
			EntityID:       actID,
			AnomalyType:    AnomalyNegativeValue,
			Severity:       "critical",
			Description:    fmt.Sprintf("%s has negative quantity %.2f — verify if this is a credit or data error", actType, quantity),
			ActualValue:    &quantity,
			DetectionRule:  "quantity < 0",
		}
		if err := s.RecordAnomaly(ctx, a); err != nil {
			return count, err
		}
		count++
	}
	return count, rows.Err()
}
