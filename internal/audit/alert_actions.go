package audit

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// AlertAction represents a trackable action item triggered by an alert.
type AlertAction struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	SourceType     string     `json:"source_type"` // anomaly, threshold, deadline, compliance, connector, approval
	SourceID       string     `json:"source_id"`
	Title          string     `json:"title"`
	Description    string     `json:"description,omitempty"`
	Priority       string     `json:"priority"` // low, medium, high, critical
	Category       string     `json:"category"`
	AssignedTo     *string    `json:"assigned_to,omitempty"`
	EscalatedTo    *string    `json:"escalated_to,omitempty"`
	Status         string     `json:"status"` // open, in_progress, blocked, resolved, dismissed, escalated
	DueDate        *time.Time `json:"due_date,omitempty"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	EscalatedAt    *time.Time `json:"escalated_at,omitempty"`
	CreatedBy      *string    `json:"created_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// AlertComment is a comment on an alert action.
type AlertComment struct {
	ID        string    `json:"id"`
	AlertID   string    `json:"alert_id"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateAlertAction creates a new alert action item.
func (s *Store) CreateAlertAction(ctx context.Context, a *AlertAction) error {
	query := `
INSERT INTO alert_actions (
  organization_id, source_type, source_id, title, description,
  priority, category, assigned_to, status, due_date, created_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'open', $9, $10)
RETURNING id, created_at, updated_at`

	return s.db.QueryRowContext(ctx, query,
		a.OrganizationID, a.SourceType, a.SourceID, a.Title, a.Description,
		a.Priority, a.Category, a.AssignedTo, a.DueDate, a.CreatedBy,
	).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
}

// GetAlertActions returns alert actions for an organization.
func (s *Store) GetAlertActions(ctx context.Context, orgID, status, priority string, limit int) ([]AlertAction, error) {
	query := `
SELECT id, organization_id, source_type, source_id, title, COALESCE(description,''),
  priority, category, assigned_to, escalated_to, status,
  due_date, acknowledged_at, resolved_at, escalated_at, created_by,
  created_at, updated_at
FROM alert_actions WHERE organization_id = $1`
	args := []any{orgID}
	n := 2

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", n)
		args = append(args, status)
		n++
	}
	if priority != "" {
		query += fmt.Sprintf(" AND priority = $%d", n)
		args = append(args, priority)
		n++
	}

	query += " ORDER BY CASE priority WHEN 'critical' THEN 0 WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ELSE 3 END, created_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get alert actions: %w", err)
	}
	defer rows.Close()

	var alerts []AlertAction
	for rows.Next() {
		var a AlertAction
		if err := rows.Scan(
			&a.ID, &a.OrganizationID, &a.SourceType, &a.SourceID, &a.Title, &a.Description,
			&a.Priority, &a.Category, &a.AssignedTo, &a.EscalatedTo, &a.Status,
			&a.DueDate, &a.AcknowledgedAt, &a.ResolvedAt, &a.EscalatedAt, &a.CreatedBy,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan alert action: %w", err)
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// AssignAlertAction assigns an alert to a user.
func (s *Store) AssignAlertAction(ctx context.Context, id, orgID, assignedTo string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE alert_actions SET assigned_to = $3, status = 'in_progress', acknowledged_at = COALESCE(acknowledged_at, NOW()), updated_at = NOW()
		 WHERE id = $1 AND organization_id = $2 AND status IN ('open', 'in_progress')`,
		id, orgID, assignedTo)
	return err
}

// EscalateAlertAction escalates an alert to a higher authority.
func (s *Store) EscalateAlertAction(ctx context.Context, id, orgID, escalatedTo string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE alert_actions SET escalated_to = $3, status = 'escalated', escalated_at = NOW(), updated_at = NOW()
		 WHERE id = $1 AND organization_id = $2`,
		id, orgID, escalatedTo)
	return err
}

// ResolveAlertAction marks an alert as resolved.
func (s *Store) ResolveAlertAction(ctx context.Context, id, orgID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE alert_actions SET status = 'resolved', resolved_at = NOW(), updated_at = NOW()
		 WHERE id = $1 AND organization_id = $2 AND status IN ('open', 'in_progress', 'escalated')`,
		id, orgID)
	return err
}

// DismissAlertAction dismisses an alert.
func (s *Store) DismissAlertAction(ctx context.Context, id, orgID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE alert_actions SET status = 'dismissed', resolved_at = NOW(), updated_at = NOW()
		 WHERE id = $1 AND organization_id = $2 AND status IN ('open', 'in_progress')`,
		id, orgID)
	return err
}

// AddAlertComment adds a comment to an alert action.
func (s *Store) AddAlertComment(ctx context.Context, alertID, userID, content string) (*AlertComment, error) {
	c := &AlertComment{
		AlertID: alertID,
		UserID:  userID,
		Content: content,
	}
	query := `INSERT INTO alert_comments (alert_id, user_id, content) VALUES ($1, $2, $3) RETURNING id, created_at`
	err := s.db.QueryRowContext(ctx, query, alertID, userID, content).Scan(&c.ID, &c.CreatedAt)
	return c, err
}

// GetAlertComments returns comments for an alert action.
func (s *Store) GetAlertComments(ctx context.Context, alertID string) ([]AlertComment, error) {
	query := `
SELECT ac.id, ac.alert_id, ac.user_id, COALESCE(u.name, u.email), ac.content, ac.created_at
FROM alert_comments ac
LEFT JOIN users u ON ac.user_id = u.id::text
WHERE ac.alert_id = $1
ORDER BY ac.created_at ASC`

	rows, err := s.db.QueryContext(ctx, query, alertID)
	if err != nil {
		return nil, fmt.Errorf("get alert comments: %w", err)
	}
	defer rows.Close()

	var comments []AlertComment
	for rows.Next() {
		var c AlertComment
		if err := rows.Scan(&c.ID, &c.AlertID, &c.UserID, &c.UserName, &c.Content, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan alert comment: %w", err)
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

// GetAlertActionCounts returns counts by status for dashboard display.
func (s *Store) GetAlertActionCounts(ctx context.Context, orgID string) (map[string]int, error) {
	query := `
SELECT
  COUNT(*) FILTER (WHERE status = 'open') AS open_count,
  COUNT(*) FILTER (WHERE status = 'in_progress') AS in_progress_count,
  COUNT(*) FILTER (WHERE status = 'escalated') AS escalated_count,
  COUNT(*) FILTER (WHERE status = 'resolved') AS resolved_count,
  COUNT(*) FILTER (WHERE priority = 'critical' AND status IN ('open', 'in_progress', 'escalated')) AS critical_active,
  COUNT(*) AS total_count
FROM alert_actions WHERE organization_id = $1`

	var open, inProgress, escalated, resolved, criticalActive, total int
	err := s.db.QueryRowContext(ctx, query, orgID).Scan(&open, &inProgress, &escalated, &resolved, &criticalActive, &total)
	if err != nil {
		return nil, fmt.Errorf("get alert action counts: %w", err)
	}

	return map[string]int{
		"open":            open,
		"in_progress":     inProgress,
		"escalated":       escalated,
		"resolved":        resolved,
		"critical_active": criticalActive,
		"total":           total,
	}, nil
}

// CreateAlertFromAnomaly creates an alert action from a detected anomaly.
func (s *Store) CreateAlertFromAnomaly(ctx context.Context, anomaly *Anomaly) (*AlertAction, error) {
	alert := &AlertAction{
		OrganizationID: anomaly.OrganizationID,
		SourceType:     "anomaly",
		SourceID:       anomaly.ID,
		Title:          fmt.Sprintf("Data Quality: %s", anomaly.Description),
		Description:    fmt.Sprintf("Anomaly detected: %s\nSeverity: %s\nDetection rule: %s", anomaly.Description, anomaly.Severity, anomaly.DetectionRule),
		Priority:       anomaly.Severity,
		Category:       "data_quality",
	}

	if anomaly.Severity == "info" {
		alert.Priority = "low"
	}

	if err := s.CreateAlertAction(ctx, alert); err != nil {
		return nil, err
	}
	return alert, nil
}
