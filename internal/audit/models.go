// Package audit provides comprehensive audit logging and change tracking.
//
// The audit system captures all significant actions and data changes within
// OffGridFlow, enabling compliance reporting, forensic analysis, and
// regulatory requirements for emissions data integrity.
//
// Key features:
//   - Immutable audit trail
//   - Change lineage tracking
//   - Actor attribution
//   - Queryable audit history
//   - Export for compliance reporting
package audit

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// =============================================================================
// Action Constants
// =============================================================================

// Action represents the type of audited action.
type Action string

const (
	// Entity lifecycle actions
	ActionCreate Action = "CREATE"
	ActionRead   Action = "READ"
	ActionUpdate Action = "UPDATE"
	ActionDelete Action = "DELETE"

	// Data import/export actions
	ActionImport Action = "IMPORT"
	ActionExport Action = "EXPORT"

	// Calculation actions
	ActionCalculate Action = "CALCULATE"
	ActionAllocate  Action = "ALLOCATE"

	// Access control actions
	ActionLogin      Action = "LOGIN"
	ActionLogout     Action = "LOGOUT"
	ActionGrantRole  Action = "GRANT_ROLE"
	ActionRevokeRole Action = "REVOKE_ROLE"

	// Configuration actions
	ActionConfigure Action = "CONFIGURE"

	// Approval workflow actions
	ActionSubmit  Action = "SUBMIT"
	ActionApprove Action = "APPROVE"
	ActionReject  Action = "REJECT"

	// Integration actions
	ActionSync    Action = "SYNC"
	ActionWebhook Action = "WEBHOOK"
	ActionAPICall Action = "API_CALL"
)

// String returns the string representation of the action.
func (a Action) String() string {
	return string(a)
}

// IsWrite returns true if the action modifies data.
func (a Action) IsWrite() bool {
	switch a {
	case ActionCreate, ActionUpdate, ActionDelete, ActionImport,
		ActionCalculate, ActionAllocate, ActionGrantRole, ActionRevokeRole,
		ActionConfigure, ActionApprove, ActionReject:
		return true
	default:
		return false
	}
}

// =============================================================================
// Entity Type Constants
// =============================================================================

// EntityType identifies the type of entity being audited.
type EntityType string

const (
	EntityActivity       EntityType = "activity"
	EntityEmission       EntityType = "emission"
	EntityEmissionFactor EntityType = "emission_factor"
	EntityAllocationRule EntityType = "allocation_rule"
	EntityUser           EntityType = "user"
	EntityOrganization   EntityType = "organization"
	EntityWorkspace      EntityType = "workspace"
	EntityReport         EntityType = "report"
	EntitySettings       EntityType = "settings"
	EntityIntegration    EntityType = "integration"
	EntityImportBatch    EntityType = "import_batch"
	EntitySubscription   EntityType = "subscription"
)

// String returns the string representation of the entity type.
func (e EntityType) String() string {
	return string(e)
}

// =============================================================================
// Outcome Constants
// =============================================================================

// Outcome indicates the result of an audited action.
type Outcome string

const (
	OutcomeSuccess Outcome = "SUCCESS"
	OutcomeFailure Outcome = "FAILURE"
	OutcomeDenied  Outcome = "DENIED"
	OutcomePending Outcome = "PENDING"
)

// =============================================================================
// Sentinel Errors
// =============================================================================

var (
	// ErrEmptyEntityType is returned when entity type is missing.
	ErrEmptyEntityType = errors.New("audit: entity type is required")

	// ErrEmptyEntityID is returned when entity ID is missing.
	ErrEmptyEntityID = errors.New("audit: entity ID is required")

	// ErrEmptyAction is returned when action is missing.
	ErrEmptyAction = errors.New("audit: action is required")

	// ErrEmptyActor is returned when actor is missing.
	ErrEmptyActor = errors.New("audit: actor is required")

	// ErrEntryNotFound is returned when an audit entry cannot be found.
	ErrEntryNotFound = errors.New("audit: entry not found")

	// ErrImmutable is returned when attempting to modify an audit entry.
	ErrImmutable = errors.New("audit: entries are immutable")
)

// =============================================================================
// AuditEntry Model
// =============================================================================

// AuditEntry captures a single audited action with full context.
//
// Audit entries are immutable once created. They provide a complete record
// of what happened, when, who performed the action, and what changed.
//
// For compliance, entries include:
//   - Precise timestamps (UTC)
//   - Actor identification
//   - Before/after state for changes
//   - Request context (IP, user agent)
//   - Correlation IDs for tracing
type AuditEntry struct {
	// ID is a unique identifier for this entry.
	ID string `json:"id"`

	// Entity identifies the type of resource affected.
	Entity EntityType `json:"entity"`

	// EntityID identifies the specific resource.
	EntityID string `json:"entity_id"`

	// Action describes what was done.
	Action Action `json:"action"`

	// Outcome indicates success, failure, or denial.
	Outcome Outcome `json:"outcome"`

	// Actor identifies who performed the action.
	Actor ActorInfo `json:"actor"`

	// Timestamp is when the action occurred (UTC).
	Timestamp time.Time `json:"timestamp"`

	// Changes captures before/after state for modifications.
	Changes *ChangeSet `json:"changes,omitempty"`

	// Context provides additional request information.
	Context RequestContext `json:"context,omitempty"`

	// Metadata contains additional key-value data.
	Metadata map[string]string `json:"metadata,omitempty"`

	// OrgID identifies the organization.
	OrgID string `json:"org_id,omitempty"`

	// WorkspaceID identifies the workspace.
	WorkspaceID string `json:"workspace_id,omitempty"`

	// CorrelationID links related audit entries.
	CorrelationID string `json:"correlation_id,omitempty"`

	// ParentID references a parent audit entry (for nested operations).
	ParentID string `json:"parent_id,omitempty"`

	// Duration is how long the action took (milliseconds).
	DurationMs int64 `json:"duration_ms,omitempty"`

	// ErrorMessage contains error details for failed actions.
	ErrorMessage string `json:"error_message,omitempty"`
}

// ActorInfo identifies who performed an action.
type ActorInfo struct {
	// ID is the actor's unique identifier (user ID, service ID, etc.).
	ID string `json:"id"`

	// Type indicates the actor type (user, service, system).
	Type string `json:"type"`

	// Name is the actor's display name.
	Name string `json:"name,omitempty"`

	// Email is the actor's email (for users).
	Email string `json:"email,omitempty"`

	// OnBehalfOf indicates impersonation (admin acting as user).
	OnBehalfOf string `json:"on_behalf_of,omitempty"`
}

// RequestContext captures the request environment.
type RequestContext struct {
	// IPAddress is the client's IP address.
	IPAddress string `json:"ip_address,omitempty"`

	// UserAgent is the client's user agent string.
	UserAgent string `json:"user_agent,omitempty"`

	// RequestID is the unique request identifier.
	RequestID string `json:"request_id,omitempty"`

	// SessionID is the user's session identifier.
	SessionID string `json:"session_id,omitempty"`

	// Method is the HTTP method (GET, POST, etc.).
	Method string `json:"method,omitempty"`

	// Path is the request path.
	Path string `json:"path,omitempty"`
}

// ChangeSet captures before/after state for modifications.
type ChangeSet struct {
	// Before is the state before the change (as JSON).
	Before json.RawMessage `json:"before,omitempty"`

	// After is the state after the change (as JSON).
	After json.RawMessage `json:"after,omitempty"`

	// Fields lists the specific fields that changed.
	Fields []FieldChange `json:"fields,omitempty"`
}

// FieldChange captures a single field modification.
type FieldChange struct {
	// Field is the field name or path.
	Field string `json:"field"`

	// OldValue is the previous value.
	OldValue interface{} `json:"old_value,omitempty"`

	// NewValue is the new value.
	NewValue interface{} `json:"new_value,omitempty"`
}

// Validate checks that the audit entry has required fields.
func (e AuditEntry) Validate() error {
	var errs []error

	if e.Entity == "" {
		errs = append(errs, ErrEmptyEntityType)
	}

	if strings.TrimSpace(e.EntityID) == "" {
		errs = append(errs, ErrEmptyEntityID)
	}

	if e.Action == "" {
		errs = append(errs, ErrEmptyAction)
	}

	if e.Actor.ID == "" {
		errs = append(errs, ErrEmptyActor)
	}

	if len(errs) > 0 {
		return fmt.Errorf("audit entry validation failed: %w", errors.Join(errs...))
	}

	return nil
}

// IsSuccess returns true if the action succeeded.
func (e AuditEntry) IsSuccess() bool {
	return e.Outcome == OutcomeSuccess
}

// IsWrite returns true if this was a write operation.
func (e AuditEntry) IsWrite() bool {
	return e.Action.IsWrite()
}

// String returns a concise human-readable representation.
func (e AuditEntry) String() string {
	return fmt.Sprintf(
		"AuditEntry{id=%s, entity=%s/%s, action=%s, actor=%s, outcome=%s}",
		e.ID, e.Entity, e.EntityID, e.Action, e.Actor.ID, e.Outcome,
	)
}

// JSON serializes the entry to JSON bytes.
func (e AuditEntry) JSON() ([]byte, error) {
	return json.Marshal(e)
}

// =============================================================================
// Entry Builder
// =============================================================================

// EntryBuilder provides a fluent interface for constructing audit entries.
type EntryBuilder struct {
	entry AuditEntry
}

// NewEntryBuilder creates a new audit entry builder.
func NewEntryBuilder() *EntryBuilder {
	return &EntryBuilder{
		entry: AuditEntry{
			Timestamp: time.Now().UTC(),
			Outcome:   OutcomeSuccess,
			Metadata:  make(map[string]string),
		},
	}
}

// WithID sets the entry ID.
func (b *EntryBuilder) WithID(id string) *EntryBuilder {
	b.entry.ID = id
	return b
}

// WithEntity sets the entity type and ID.
func (b *EntryBuilder) WithEntity(entityType EntityType, entityID string) *EntryBuilder {
	b.entry.Entity = entityType
	b.entry.EntityID = entityID
	return b
}

// WithAction sets the action.
func (b *EntryBuilder) WithAction(action Action) *EntryBuilder {
	b.entry.Action = action
	return b
}

// WithOutcome sets the outcome.
func (b *EntryBuilder) WithOutcome(outcome Outcome) *EntryBuilder {
	b.entry.Outcome = outcome
	return b
}

// WithActor sets the actor information.
func (b *EntryBuilder) WithActor(actorID, actorType string) *EntryBuilder {
	b.entry.Actor.ID = actorID
	b.entry.Actor.Type = actorType
	return b
}

// WithActorDetails sets additional actor information.
func (b *EntryBuilder) WithActorDetails(name, email string) *EntryBuilder {
	b.entry.Actor.Name = name
	b.entry.Actor.Email = email
	return b
}

// WithContext sets the request context.
func (b *EntryBuilder) WithContext(ctx RequestContext) *EntryBuilder {
	b.entry.Context = ctx
	return b
}

// WithIP sets the client IP address.
func (b *EntryBuilder) WithIP(ip string) *EntryBuilder {
	b.entry.Context.IPAddress = ip
	return b
}

// WithChanges sets the change details.
func (b *EntryBuilder) WithChanges(before, after interface{}) *EntryBuilder {
	changes := &ChangeSet{}

	if before != nil {
		if data, err := json.Marshal(before); err == nil {
			changes.Before = data
		}
	}

	if after != nil {
		if data, err := json.Marshal(after); err == nil {
			changes.After = data
		}
	}

	b.entry.Changes = changes
	return b
}

// AddFieldChange adds a specific field change.
func (b *EntryBuilder) AddFieldChange(field string, oldVal, newVal interface{}) *EntryBuilder {
	if b.entry.Changes == nil {
		b.entry.Changes = &ChangeSet{}
	}

	b.entry.Changes.Fields = append(b.entry.Changes.Fields, FieldChange{
		Field:    field,
		OldValue: oldVal,
		NewValue: newVal,
	})

	return b
}

// WithMetadata adds metadata.
func (b *EntryBuilder) WithMetadata(key, value string) *EntryBuilder {
	b.entry.Metadata[key] = value
	return b
}

// WithOrgID sets the organization ID.
func (b *EntryBuilder) WithOrgID(orgID string) *EntryBuilder {
	b.entry.OrgID = orgID
	return b
}

// WithWorkspaceID sets the workspace ID.
func (b *EntryBuilder) WithWorkspaceID(workspaceID string) *EntryBuilder {
	b.entry.WorkspaceID = workspaceID
	return b
}

// WithCorrelationID sets the correlation ID.
func (b *EntryBuilder) WithCorrelationID(id string) *EntryBuilder {
	b.entry.CorrelationID = id
	return b
}

// WithDuration sets the operation duration.
func (b *EntryBuilder) WithDuration(d time.Duration) *EntryBuilder {
	b.entry.DurationMs = d.Milliseconds()
	return b
}

// WithError marks the entry as failed with an error.
func (b *EntryBuilder) WithError(err error) *EntryBuilder {
	b.entry.Outcome = OutcomeFailure
	if err != nil {
		b.entry.ErrorMessage = err.Error()
	}
	return b
}

// Build returns the constructed entry after validation.
func (b *EntryBuilder) Build() (AuditEntry, error) {
	if err := b.entry.Validate(); err != nil {
		return AuditEntry{}, err
	}

	return b.entry, nil
}

// MustBuild returns the entry, panicking on validation error.
func (b *EntryBuilder) MustBuild() AuditEntry {
	e, err := b.Build()
	if err != nil {
		panic(err)
	}
	return e
}

// =============================================================================
// Query Types
// =============================================================================

// Query specifies criteria for searching audit entries.
type Query struct {
	// Entity filters by entity type.
	Entity EntityType `json:"entity,omitempty"`

	// EntityID filters by specific entity.
	EntityID string `json:"entity_id,omitempty"`

	// Action filters by action type.
	Action Action `json:"action,omitempty"`

	// ActorID filters by actor.
	ActorID string `json:"actor_id,omitempty"`

	// OrgID filters by organization.
	OrgID string `json:"org_id,omitempty"`

	// From filters entries after this time.
	From time.Time `json:"from,omitempty"`

	// To filters entries before this time.
	To time.Time `json:"to,omitempty"`

	// Outcome filters by outcome.
	Outcome Outcome `json:"outcome,omitempty"`

	// CorrelationID filters by correlation ID.
	CorrelationID string `json:"correlation_id,omitempty"`

	// Limit is the maximum entries to return.
	Limit int `json:"limit,omitempty"`

	// Offset is the number of entries to skip.
	Offset int `json:"offset,omitempty"`

	// OrderBy specifies sort order (timestamp, -timestamp).
	OrderBy string `json:"order_by,omitempty"`
}

// QueryResult contains search results.
type QueryResult struct {
	// Entries are the matching audit entries.
	Entries []AuditEntry `json:"entries"`

	// Total is the total count matching the query (ignoring limit).
	Total int `json:"total"`

	// Limit is the limit that was applied.
	Limit int `json:"limit"`

	// Offset is the offset that was applied.
	Offset int `json:"offset"`
}

// HasMore returns true if there are more results beyond this page.
func (r QueryResult) HasMore() bool {
	return r.Offset+len(r.Entries) < r.Total
}

// =============================================================================
// Summary Types
// =============================================================================

// Summary provides aggregated audit statistics.
type Summary struct {
	// TotalEntries is the total count of audit entries.
	TotalEntries int64 `json:"total_entries"`

	// ByAction counts entries per action type.
	ByAction map[Action]int64 `json:"by_action"`

	// ByEntity counts entries per entity type.
	ByEntity map[EntityType]int64 `json:"by_entity"`

	// ByOutcome counts entries per outcome.
	ByOutcome map[Outcome]int64 `json:"by_outcome"`

	// TopActors lists the most active actors.
	TopActors []ActorSummary `json:"top_actors,omitempty"`

	// TimeRange is the range of entries summarized.
	TimeRange [2]time.Time `json:"time_range"`

	// GeneratedAt is when the summary was generated.
	GeneratedAt time.Time `json:"generated_at"`
}

// ActorSummary summarizes an actor's audit activity.
type ActorSummary struct {
	ActorID    string `json:"actor_id"`
	ActorName  string `json:"actor_name,omitempty"`
	EntryCount int64  `json:"entry_count"`
}
