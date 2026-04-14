package audit

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// CustomerHealth represents the health score for an organization.
type CustomerHealth struct {
	ID                    string     `json:"id"`
	OrganizationID        string     `json:"organization_id"`
	OverallScore          int        `json:"overall_score"`
	DataFreshnessScore    int        `json:"data_freshness_score"`
	FeatureAdoptionScore  int        `json:"feature_adoption_score"`
	ReportCompletionScore int        `json:"report_completion_score"`
	UserEngagementScore   int        `json:"user_engagement_score"`
	DataQualityScore      int        `json:"data_quality_score"`
	HealthStatus          string     `json:"health_status"` // healthy, at_risk, critical, churning, unknown
	RenewalRiskPercent    int        `json:"renewal_risk_percent"`
	NextRenewalDate       *string    `json:"next_renewal_date,omitempty"`
	ContractValueCents    int64      `json:"contract_value_cents"`
	ActiveUsersCount      int        `json:"active_users_count"`
	TotalActivitiesCount  int        `json:"total_activities_count"`
	ReportsGeneratedCount int        `json:"reports_generated_count"`
	LastLoginAt           *time.Time `json:"last_login_at,omitempty"`
	LastReportAt          *time.Time `json:"last_report_at,omitempty"`
	LastDataUploadAt      *time.Time `json:"last_data_upload_at,omitempty"`
	ExpansionReady        bool       `json:"expansion_ready"`
	ExpansionTriggers     string     `json:"expansion_triggers"`
	ScoredAt              time.Time  `json:"scored_at"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// CalculateAndStoreHealthScore computes a health score from real data.
func (s *Store) CalculateAndStoreHealthScore(ctx context.Context, orgID string) (*CustomerHealth, error) {
	h := &CustomerHealth{
		OrganizationID: orgID,
		ScoredAt:       time.Now().UTC(),
	}

	// 1. Data Freshness: How recently was data uploaded?
	var lastActivity sql.NullTime
	s.db.QueryRowContext(ctx,
		`SELECT MAX(created_at) FROM activities WHERE organization_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&lastActivity)

	if lastActivity.Valid {
		h.LastDataUploadAt = &lastActivity.Time
		daysSince := time.Since(lastActivity.Time).Hours() / 24
		switch {
		case daysSince <= 7:
			h.DataFreshnessScore = 100
		case daysSince <= 14:
			h.DataFreshnessScore = 85
		case daysSince <= 30:
			h.DataFreshnessScore = 65
		case daysSince <= 60:
			h.DataFreshnessScore = 40
		case daysSince <= 90:
			h.DataFreshnessScore = 20
		default:
			h.DataFreshnessScore = 5
		}
	}

	// 2. Feature Adoption: Which features are being used?
	featureCount := 0
	totalFeatures := 6

	var count int
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM activities WHERE organization_id = $1 AND deleted_at IS NULL`, orgID).Scan(&count)
	if count > 0 {
		featureCount++
		h.TotalActivitiesCount = count
	}

	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM compliance_reports WHERE organization_id = $1`, orgID).Scan(&count)
	if count > 0 {
		featureCount++
		h.ReportsGeneratedCount = count
	}

	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM cloud_connectors WHERE organization_id = $1 AND enabled = true`, orgID).Scan(&count)
	if count > 0 {
		featureCount++
	}

	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM audit_logs WHERE organization_id = $1::uuid`, orgID).Scan(&count)
	if count > 10 {
		featureCount++
	}

	// Check if approval workflow is used
	var approvalCount int
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM approval_workflow WHERE tenant_id = $1`, orgID).Scan(&approvalCount)
	if approvalCount > 0 {
		featureCount++
	}

	// Check if factor snapshots are used
	var snapCount int
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM factor_snapshots WHERE organization_id = $1`, orgID).Scan(&snapCount)
	if snapCount > 0 {
		featureCount++
	}

	h.FeatureAdoptionScore = (featureCount * 100) / totalFeatures

	// 3. Report Completion: Are reports being completed through the workflow?
	var totalReports, approvedReports int
	s.db.QueryRowContext(ctx, `SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'approved' OR status = 'submitted') FROM compliance_reports WHERE organization_id = $1`, orgID).Scan(&totalReports, &approvedReports)
	if totalReports > 0 {
		h.ReportCompletionScore = (approvedReports * 100) / totalReports
	}

	var lastReport sql.NullTime
	s.db.QueryRowContext(ctx, `SELECT MAX(created_at) FROM compliance_reports WHERE organization_id = $1`, orgID).Scan(&lastReport)
	if lastReport.Valid {
		h.LastReportAt = &lastReport.Time
	}

	// 4. User Engagement: Active users, login frequency
	s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users WHERE organization_id = $1::uuid AND deleted_at IS NULL AND last_login_at > NOW() - INTERVAL '30 days'`, orgID,
	).Scan(&h.ActiveUsersCount)

	var totalUsers int
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE organization_id = $1::uuid AND deleted_at IS NULL`, orgID).Scan(&totalUsers)
	if totalUsers > 0 {
		h.UserEngagementScore = (h.ActiveUsersCount * 100) / totalUsers
	}

	var lastLogin sql.NullTime
	s.db.QueryRowContext(ctx, `SELECT MAX(last_login_at) FROM users WHERE organization_id = $1::uuid AND deleted_at IS NULL`, orgID).Scan(&lastLogin)
	if lastLogin.Valid {
		h.LastLoginAt = &lastLogin.Time
	}

	// 5. Data Quality: Ratio of open anomalies to total activities
	var openAnomalies int
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM data_quality_anomalies WHERE organization_id = $1 AND status = 'open'`, orgID).Scan(&openAnomalies)
	if h.TotalActivitiesCount > 0 {
		anomalyRate := float64(openAnomalies) / float64(h.TotalActivitiesCount)
		switch {
		case anomalyRate == 0:
			h.DataQualityScore = 100
		case anomalyRate < 0.02:
			h.DataQualityScore = 90
		case anomalyRate < 0.05:
			h.DataQualityScore = 75
		case anomalyRate < 0.10:
			h.DataQualityScore = 50
		default:
			h.DataQualityScore = 25
		}
	} else {
		h.DataQualityScore = 50 // No data = neutral
	}

	// Calculate overall score (weighted average)
	h.OverallScore = (h.DataFreshnessScore*25 +
		h.FeatureAdoptionScore*20 +
		h.ReportCompletionScore*20 +
		h.UserEngagementScore*15 +
		h.DataQualityScore*20) / 100

	// Determine health status
	switch {
	case h.OverallScore >= 80:
		h.HealthStatus = "healthy"
		h.RenewalRiskPercent = 5
	case h.OverallScore >= 60:
		h.HealthStatus = "at_risk"
		h.RenewalRiskPercent = 25
	case h.OverallScore >= 40:
		h.HealthStatus = "critical"
		h.RenewalRiskPercent = 50
	default:
		h.HealthStatus = "churning"
		h.RenewalRiskPercent = 80
	}

	// Expansion signals
	expansionReady := h.OverallScore >= 80 && h.FeatureAdoptionScore >= 80 && h.TotalActivitiesCount > 50
	h.ExpansionReady = expansionReady
	if expansionReady {
		h.ExpansionTriggers = `["high_adoption","active_usage","report_completion"]`
	} else {
		h.ExpansionTriggers = "[]"
	}

	// Store the score
	query := `
INSERT INTO customer_health_scores (
  organization_id, overall_score, data_freshness_score, feature_adoption_score,
  report_completion_score, user_engagement_score, data_quality_score,
  health_status, renewal_risk_percent, active_users_count, total_activities_count,
  reports_generated_count, last_login_at, last_report_at, last_data_upload_at,
  expansion_ready, expansion_triggers, scored_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17::jsonb, $18)
RETURNING id, created_at, updated_at`

	err := s.db.QueryRowContext(ctx, query,
		orgID, h.OverallScore, h.DataFreshnessScore, h.FeatureAdoptionScore,
		h.ReportCompletionScore, h.UserEngagementScore, h.DataQualityScore,
		h.HealthStatus, h.RenewalRiskPercent, h.ActiveUsersCount, h.TotalActivitiesCount,
		h.ReportsGeneratedCount, h.LastLoginAt, h.LastReportAt, h.LastDataUploadAt,
		h.ExpansionReady, h.ExpansionTriggers, h.ScoredAt,
	).Scan(&h.ID, &h.CreatedAt, &h.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("store health score: %w", err)
	}

	return h, nil
}

// GetLatestHealthScore returns the most recent health score for an organization.
func (s *Store) GetLatestHealthScore(ctx context.Context, orgID string) (*CustomerHealth, error) {
	h := &CustomerHealth{}
	query := `
SELECT id, organization_id, overall_score, data_freshness_score, feature_adoption_score,
  report_completion_score, user_engagement_score, data_quality_score,
  health_status, renewal_risk_percent, next_renewal_date, COALESCE(contract_value_cents, 0),
  active_users_count, total_activities_count, reports_generated_count,
  last_login_at, last_report_at, last_data_upload_at,
  expansion_ready, COALESCE(expansion_triggers::text, '[]'),
  scored_at, created_at, updated_at
FROM customer_health_scores
WHERE organization_id = $1
ORDER BY scored_at DESC LIMIT 1`

	err := s.db.QueryRowContext(ctx, query, orgID).Scan(
		&h.ID, &h.OrganizationID, &h.OverallScore, &h.DataFreshnessScore, &h.FeatureAdoptionScore,
		&h.ReportCompletionScore, &h.UserEngagementScore, &h.DataQualityScore,
		&h.HealthStatus, &h.RenewalRiskPercent, &h.NextRenewalDate, &h.ContractValueCents,
		&h.ActiveUsersCount, &h.TotalActivitiesCount, &h.ReportsGeneratedCount,
		&h.LastLoginAt, &h.LastReportAt, &h.LastDataUploadAt,
		&h.ExpansionReady, &h.ExpansionTriggers,
		&h.ScoredAt, &h.CreatedAt, &h.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get latest health score: %w", err)
	}
	return h, nil
}

// GetHealthScoreHistory returns health score trend data.
func (s *Store) GetHealthScoreHistory(ctx context.Context, orgID string, limit int) ([]CustomerHealth, error) {
	query := `
SELECT id, organization_id, overall_score, data_freshness_score, feature_adoption_score,
  report_completion_score, user_engagement_score, data_quality_score,
  health_status, renewal_risk_percent,
  active_users_count, total_activities_count, reports_generated_count,
  scored_at, created_at
FROM customer_health_scores
WHERE organization_id = $1
ORDER BY scored_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("get health score history: %w", err)
	}
	defer rows.Close()

	var scores []CustomerHealth
	for rows.Next() {
		var h CustomerHealth
		if err := rows.Scan(
			&h.ID, &h.OrganizationID, &h.OverallScore, &h.DataFreshnessScore, &h.FeatureAdoptionScore,
			&h.ReportCompletionScore, &h.UserEngagementScore, &h.DataQualityScore,
			&h.HealthStatus, &h.RenewalRiskPercent,
			&h.ActiveUsersCount, &h.TotalActivitiesCount, &h.ReportsGeneratedCount,
			&h.ScoredAt, &h.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan health score: %w", err)
		}
		scores = append(scores, h)
	}
	return scores, rows.Err()
}
