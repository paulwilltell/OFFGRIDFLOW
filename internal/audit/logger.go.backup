// Package audit provides comprehensive audit logging for compliance requirements.
// Every significant action (especially compliance report generation) is logged
// with full context for audit trail purposes.
package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
)

// Action represents types of auditable actions
type Action string

const (
	ActionExportCSRD       Action = "export_csrd"
	ActionExportSEC        Action = "export_sec"
	ActionExportCalifornia Action = "export_california"
	ActionExportCBAM       Action = "export_cbam"
	ActionExportIFRS       Action = "export_ifrs"
	ActionCreateReport     Action = "create_report"
	ActionApproveReport    Action = "approve_report"
	ActionDeleteReport     Action = "delete_report"
	ActionCalculateScope1  Action = "calculate_scope1"
	ActionCalculateScope2  Action = "calculate_scope2"
	ActionCalculateScope3  Action = "calculate_scope3"
	ActionCreateActivity   Action = "create_activity"
	ActionUpdateActivity   Action = "update_activity"
	ActionDeleteActivity   Action = "delete_activity"
	ActionLogin            Action = "login"
	ActionLogout           Action = "logout"
	ActionAPIKeyCreate     Action = "api_key_create"
	ActionAPIKeyRevoke     Action = "api_key_revoke"
)

// ResourceType identifies what kind of resource was acted upon
type ResourceType string

const (
	ResourceComplianceReport ResourceType = "compliance_report"
	ResourceActivity         ResourceType = "activity"
	ResourceEmission         ResourceType = "emission"
	ResourceUser             ResourceType = "user"
	ResourceAPIKey           ResourceType = "api_key"
	ResourceTenant           ResourceType = "tenant"
)

// Status represents the outcome of an audited action
type Status string

const (
	StatusSuccess Status = "success"
	StatusFailure Status = "failure"
	StatusPartial Status = "partial"
)

// Log represents a single audit log entry
type Log struct {
	ID           uuid.UUID       `json:"id"`
	TenantID     uuid.UUID       `json:"tenant_id"`
	UserID       *uuid.UUID      `json:"user_id,omitempty"`
	Action       Action          `json:"action"`
	ResourceType ResourceType    `json:"resource_type"`
	ResourceID   *uuid.UUID      `json:"resource_id,omitempty"`
	Timestamp    time.Time       `json:"timestamp"`
	IPAddress    *net.IP         `json:"ip_address,omitempty"`
	UserAgent    string          `json:"user_agent,omitempty"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
	Status       Status          `json:"status"`
	ErrorMessage string          `json:"error_message,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

// Logger handles writing audit logs to the database
type Logger struct {
	db *sql.DB
}

// NewLogger creates a new audit logger
func NewLogger(db *sql.DB) *Logger {
	return &Logger{db: db}
}

// LogEntry creates and stores a new audit log entry
type LogEntry struct {
	TenantID     uuid.UUID
	UserID       *uuid.UUID
	Action       Action
	ResourceType ResourceType
	ResourceID   *uuid.UUID
	IPAddress    *net.IP
	UserAgent    string
	Metadata     map[string]interface{}
	Status       Status
	ErrorMessage string
}

// Log writes an audit log entry to the database
func (l *Logger) Log(ctx context.Context, entry LogEntry) error {
	// Serialize metadata to JSON
	var metadataJSON []byte
	var err error
	if entry.Metadata != nil {
		metadataJSON, err = json.Marshal(entry.Metadata)
		if err != nil {
			return fmt.Errorf("marshal metadata: %w", err)
		}
	}

	// Convert IP to string for PostgreSQL INET type
	var ipStr *string
	if entry.IPAddress != nil {
		s := entry.IPAddress.String()
		ipStr = &s
	}

	query := `
		INSERT INTO audit_logs (
			tenant_id, user_id, action, resource_type, resource_id,
			timestamp, ip_address, user_agent, metadata, status, error_message
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`

	_, err = l.db.ExecContext(ctx, query,
		entry.TenantID,
		entry.UserID,
		entry.Action,
		entry.ResourceType,
		entry.ResourceID,
		time.Now(),
		ipStr,
		entry.UserAgent,
		metadataJSON,
		entry.Status,
		entry.ErrorMessage,
	)

	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}

	return nil
}

// LogSuccess is a convenience method for logging successful actions
func (l *Logger) LogSuccess(ctx context.Context, tenantID, userID uuid.UUID, action Action, resourceType ResourceType, resourceID *uuid.UUID, metadata map[string]interface{}) error {
	return l.Log(ctx, LogEntry{
		TenantID:     tenantID,
		UserID:       &userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Metadata:     metadata,
		Status:       StatusSuccess,
	})
}

// LogFailure is a convenience method for logging failed actions
func (l *Logger) LogFailure(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, action Action, resourceType ResourceType, errorMsg string, metadata map[string]interface{}) error {
	return l.Log(ctx, LogEntry{
		TenantID:     tenantID,
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		Metadata:     metadata,
		Status:       StatusFailure,
		ErrorMessage: errorMsg,
	})
}

// QueryParams for filtering audit logs
type QueryParams struct {
	TenantID     uuid.UUID
	UserID       *uuid.UUID
	Action       *Action
	ResourceType *ResourceType
	ResourceID   *uuid.UUID
	StartTime    *time.Time
	EndTime      *time.Time
	Limit        int
	Offset       int
}

// Query retrieves audit logs based on filters
func (l *Logger) Query(ctx context.Context, params QueryParams) ([]Log, error) {
	query := `
		SELECT 
			id, tenant_id, user_id, action, resource_type, resource_id,
			timestamp, ip_address, user_agent, metadata, status, error_message, created_at
		FROM audit_logs
		WHERE tenant_id = $1
	`
	args := []interface{}{params.TenantID}
	argCount := 1

	if params.UserID != nil {
		argCount++
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, *params.UserID)
	}

	if params.Action != nil {
		argCount++
		query += fmt.Sprintf(" AND action = $%d", argCount)
		args = append(args, *params.Action)
	}

	if params.ResourceType != nil {
		argCount++
		query += fmt.Sprintf(" AND resource_type = $%d", argCount)
		args = append(args, *params.ResourceType)
	}

	if params.ResourceID != nil {
		argCount++
		query += fmt.Sprintf(" AND resource_id = $%d", argCount)
		args = append(args, *params.ResourceID)
	}

	if params.StartTime != nil {
		argCount++
		query += fmt.Sprintf(" AND timestamp >= $%d", argCount)
		args = append(args, *params.StartTime)
	}

	if params.EndTime != nil {
		argCount++
		query += fmt.Sprintf(" AND timestamp <= $%d", argCount)
		args = append(args, *params.EndTime)
	}

	query += " ORDER BY timestamp DESC"

	if params.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, params.Limit)
	}

	if params.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, params.Offset)
	}

	rows, err := l.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []Log
	for rows.Next() {
		var log Log
		var ipStr *string
		var metadataBytes []byte

		err := rows.Scan(
			&log.ID,
			&log.TenantID,
			&log.UserID,
			&log.Action,
			&log.ResourceType,
			&log.ResourceID,
			&log.Timestamp,
			&ipStr,
			&log.UserAgent,
			&metadataBytes,
			&log.Status,
			&log.ErrorMessage,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}

		// Parse IP address
		if ipStr != nil && *ipStr != "" {
			ip := net.ParseIP(*ipStr)
			log.IPAddress = &ip
		}

		// Set raw JSON metadata
		if len(metadataBytes) > 0 {
			log.Metadata = metadataBytes
		}

		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit logs: %w", err)
	}

	return logs, nil
}

// GetByID retrieves a single audit log by ID (must match tenant for security)
func (l *Logger) GetByID(ctx context.Context, tenantID, logID uuid.UUID) (*Log, error) {
	query := `
		SELECT 
			id, tenant_id, user_id, action, resource_type, resource_id,
			timestamp, ip_address, user_agent, metadata, status, error_message, created_at
		FROM audit_logs
		WHERE id = $1 AND tenant_id = $2
	`

	var log Log
	var ipStr *string
	var metadataBytes []byte

	err := l.db.QueryRowContext(ctx, query, logID, tenantID).Scan(
		&log.ID,
		&log.TenantID,
		&log.UserID,
		&log.Action,
		&log.ResourceType,
		&log.ResourceID,
		&log.Timestamp,
		&ipStr,
		&log.UserAgent,
		&metadataBytes,
		&log.Status,
		&log.ErrorMessage,
		&log.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("audit log not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get audit log: %w", err)
	}

	// Parse IP address
	if ipStr != nil && *ipStr != "" {
		ip := net.ParseIP(*ipStr)
		log.IPAddress = &ip
	}

	// Set raw JSON metadata
	if len(metadataBytes) > 0 {
		log.Metadata = metadataBytes
	}

	return &log, nil
}

// Count returns the total number of audit logs matching the criteria
func (l *Logger) Count(ctx context.Context, params QueryParams) (int64, error) {
	query := `SELECT COUNT(*) FROM audit_logs WHERE tenant_id = $1`
	args := []interface{}{params.TenantID}
	argCount := 1

	if params.UserID != nil {
		argCount++
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, *params.UserID)
	}

	if params.Action != nil {
		argCount++
		query += fmt.Sprintf(" AND action = $%d", argCount)
		args = append(args, *params.Action)
	}

	if params.ResourceType != nil {
		argCount++
		query += fmt.Sprintf(" AND resource_type = $%d", argCount)
		args = append(args, *params.ResourceType)
	}

	var count int64
	err := l.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count audit logs: %w", err)
	}

	return count, nil
}
