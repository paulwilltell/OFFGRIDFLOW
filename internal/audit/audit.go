package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

// EventType represents the type of audit event
type EventType string

const (
	// Authentication events
	EventLogin           EventType = "auth.login"
	EventLoginFailed     EventType = "auth.login.failed"
	EventLogout          EventType = "auth.logout"
	EventPasswordChanged EventType = "auth.password.changed"
	EventPasswordReset   EventType = "auth.password.reset"
	EventMFAEnabled      EventType = "auth.mfa.enabled"
	EventMFADisabled     EventType = "auth.mfa.disabled"

	// API Key events
	EventAPIKeyCreated EventType = "auth.apikey.created"
	EventAPIKeyRevoked EventType = "auth.apikey.revoked"
	EventAPIKeyUsed    EventType = "auth.apikey.used"

	// User management events
	EventUserCreated     EventType = "user.created"
	EventUserUpdated     EventType = "user.updated"
	EventUserDeactivated EventType = "user.deactivated"
	EventUserRoleChanged EventType = "user.role.changed"

	// Tenant events
	EventTenantCreated EventType = "tenant.created"
	EventTenantUpdated EventType = "tenant.updated"
	EventTenantDeleted EventType = "tenant.deleted"

	// Permission events
	EventPermissionGranted EventType = "permission.granted"
	EventPermissionDenied  EventType = "permission.denied"

	// Data access events
	EventDataExported EventType = "data.exported"
	EventDataDeleted  EventType = "data.deleted"
	EventDataImported EventType = "data.imported"

	// Settings events
	EventSettingsChanged EventType = "settings.changed"
)

// Event represents an audit log event
type Event struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	EventType EventType              `json:"event_type"`
	ActorID   string                 `json:"actor_id"`
	ActorType string                 `json:"actor_type"` // user, apikey, system
	TenantID  string                 `json:"tenant_id"`
	Resource  string                 `json:"resource,omitempty"`
	Action    string                 `json:"action,omitempty"`
	Status    string                 `json:"status"` // success, failure
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// Logger provides audit logging functionality
type Logger struct {
	store  EventStore
	logger *slog.Logger
}

// EventStore defines persistence for audit logs
type EventStore interface {
	SaveEvent(ctx context.Context, event *Event) error
	QueryEvents(ctx context.Context, query EventQuery) ([]*Event, error)
	GetEvent(ctx context.Context, id string) (*Event, error)
}

// EventQuery represents an audit log query
type EventQuery struct {
	TenantID   string
	ActorID    string
	EventTypes []EventType
	StartTime  time.Time
	EndTime    time.Time
	Limit      int
	Offset     int
}

// NewLogger creates a new audit logger
func NewLogger(store EventStore, logger *slog.Logger) *Logger {
	if logger == nil {
		logger = slog.Default()
	}

	return &Logger{
		store:  store,
		logger: logger,
	}
}

// Log logs an audit event
func (l *Logger) Log(ctx context.Context, event Event) error {
	// Ensure required fields
	if event.ID == "" {
		event.ID = generateEventID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Validate event
	if err := validateEvent(&event); err != nil {
		return fmt.Errorf("invalid audit event: %w", err)
	}

	// Save to store
	if err := l.store.SaveEvent(ctx, &event); err != nil {
		l.logger.ErrorContext(ctx, "Failed to save audit event",
			"event_id", event.ID,
			"event_type", event.EventType,
			"error", err)
		return err
	}

	// Also log to structured logger
	l.logger.InfoContext(ctx, "Audit event",
		"event_id", event.ID,
		"event_type", event.EventType,
		"actor_id", event.ActorID,
		"tenant_id", event.TenantID,
		"status", event.Status)

	return nil
}

// LogLogin logs a successful login event
func (l *Logger) LogLogin(ctx context.Context, tenantID, userID, ip, userAgent string) error {
	return l.Log(ctx, Event{
		EventType: EventLogin,
		ActorID:   userID,
		ActorType: "user",
		TenantID:  tenantID,
		Status:    "success",
		IPAddress: ip,
		UserAgent: userAgent,
	})
}

// LogLoginFailed logs a failed login attempt
func (l *Logger) LogLoginFailed(ctx context.Context, email, ip, userAgent, reason string) error {
	return l.Log(ctx, Event{
		EventType: EventLoginFailed,
		ActorID:   email,
		ActorType: "user",
		Status:    "failure",
		IPAddress: ip,
		UserAgent: userAgent,
		Error:     reason,
		Details: map[string]interface{}{
			"email": email,
		},
	})
}

// LogAPIKeyCreated logs API key creation
func (l *Logger) LogAPIKeyCreated(ctx context.Context, tenantID, userID, keyID, keyLabel string) error {
	return l.Log(ctx, Event{
		EventType: EventAPIKeyCreated,
		ActorID:   userID,
		ActorType: "user",
		TenantID:  tenantID,
		Resource:  fmt.Sprintf("apikey:%s", keyID),
		Status:    "success",
		Details: map[string]interface{}{
			"key_id":    keyID,
			"key_label": keyLabel,
		},
	})
}

// LogAPIKeyRevoked logs API key revocation
func (l *Logger) LogAPIKeyRevoked(ctx context.Context, tenantID, userID, keyID string) error {
	return l.Log(ctx, Event{
		EventType: EventAPIKeyRevoked,
		ActorID:   userID,
		ActorType: "user",
		TenantID:  tenantID,
		Resource:  fmt.Sprintf("apikey:%s", keyID),
		Status:    "success",
		Details: map[string]interface{}{
			"key_id": keyID,
		},
	})
}

// LogPermissionDenied logs a permission denial
func (l *Logger) LogPermissionDenied(ctx context.Context, tenantID, userID, action, resource string) error {
	return l.Log(ctx, Event{
		EventType: EventPermissionDenied,
		ActorID:   userID,
		ActorType: "user",
		TenantID:  tenantID,
		Action:    action,
		Resource:  resource,
		Status:    "failure",
		Details: map[string]interface{}{
			"action":   action,
			"resource": resource,
		},
	})
}

// LogUserRoleChanged logs a user role change
func (l *Logger) LogUserRoleChanged(ctx context.Context, tenantID, actorID, targetUserID, oldRole, newRole string) error {
	return l.Log(ctx, Event{
		EventType: EventUserRoleChanged,
		ActorID:   actorID,
		ActorType: "user",
		TenantID:  tenantID,
		Resource:  fmt.Sprintf("user:%s", targetUserID),
		Status:    "success",
		Details: map[string]interface{}{
			"target_user_id": targetUserID,
			"old_role":       oldRole,
			"new_role":       newRole,
		},
	})
}

// LogDataExported logs a data export event
func (l *Logger) LogDataExported(ctx context.Context, tenantID, userID, format, resource string) error {
	return l.Log(ctx, Event{
		EventType: EventDataExported,
		ActorID:   userID,
		ActorType: "user",
		TenantID:  tenantID,
		Resource:  resource,
		Status:    "success",
		Details: map[string]interface{}{
			"format":   format,
			"resource": resource,
		},
	})
}

// QueryEvents retrieves audit events matching the query
func (l *Logger) QueryEvents(ctx context.Context, query EventQuery) ([]*Event, error) {
	return l.store.QueryEvents(ctx, query)
}

func validateEvent(event *Event) error {
	if event.EventType == "" {
		return fmt.Errorf("event_type is required")
	}
	if event.Status == "" {
		return fmt.Errorf("status is required")
	}
	if event.ActorID == "" {
		return fmt.Errorf("actor_id is required")
	}
	return nil
}

func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

// ToJSON serializes an event to JSON
func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON deserializes an event from JSON
func (e *Event) FromJSON(data []byte) error {
	return json.Unmarshal(data, e)
}
