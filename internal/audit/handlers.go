package audit

import (
	"encoding/json"
	"net/http"

	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/auth"
)

// Handlers provides HTTP endpoints for audit operations.
type Handlers struct {
	store *Store
}

// NewHandlers creates handlers for audit endpoints.
func NewHandlers(store *Store) *Handlers {
	return &Handlers{store: store}
}

// GetCalculationLedger returns the immutable calculation ledger for a tenant.
// GET /api/audit/ledger?scope=scope2&limit=100
func (h *Handlers) GetCalculationLedger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}
	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}

	scope := r.URL.Query().Get("scope")
	limit := 200

	entries, err := h.store.GetLedgerByTenant(r.Context(), tenant.ID, scope, limit)
	if err != nil {
		responders.InternalError(w, "failed to fetch calculation ledger")
		return
	}
	if entries == nil {
		entries = []LedgerEntry{}
	}
	responders.JSON(w, http.StatusOK, map[string]any{"entries": entries, "count": len(entries)})
}

// GetApprovals returns all approval workflow records.
// GET /api/audit/approvals
func (h *Handlers) GetApprovals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}
	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}

	records, err := h.store.GetApprovalsByTenant(r.Context(), tenant.ID)
	if err != nil {
		responders.InternalError(w, "failed to fetch approvals")
		return
	}
	if records == nil {
		records = []ApprovalRecord{}
	}
	responders.JSON(w, http.StatusOK, map[string]any{"approvals": records})
}

// CreateApprovalRequest creates a new approval workflow.
// POST /api/audit/approvals {entity_type, entity_id}
func (h *Handlers) CreateApprovalRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responders.MethodNotAllowed(w, http.MethodPost)
		return
	}
	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}
	user, uOk := auth.UserFromContext(r.Context())
	if !uOk || user == nil {
		responders.Unauthorized(w, "unauthorized", "user context required")
		return
	}

	var body struct {
		EntityType string `json:"entity_type"`
		EntityID   string `json:"entity_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		responders.BadRequest(w, "invalid_request", "invalid JSON body")
		return
	}
	if body.EntityType == "" || body.EntityID == "" {
		responders.BadRequest(w, "validation_error", "entity_type and entity_id are required")
		return
	}

	rec, err := h.store.CreateApproval(r.Context(), tenant.ID, body.EntityType, body.EntityID, user.ID)
	if err != nil {
		responders.InternalError(w, "failed to create approval")
		return
	}

	_ = h.store.LogChange(r.Context(), tenant.ID, "approval", rec.ID, "create", "", "", body.EntityType+":"+body.EntityID, user.ID)
	responders.JSON(w, http.StatusCreated, rec)
}

// UpdateApproval handles status transitions (submit, review, approve, reject).
// PUT /api/audit/approvals/{id} {action: "submit"|"review"|"approve"|"reject", notes: "..."}
func (h *Handlers) UpdateApproval(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		responders.MethodNotAllowed(w, http.MethodPut)
		return
	}
	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}
	user, uOk := auth.UserFromContext(r.Context())
	if !uOk || user == nil {
		responders.Unauthorized(w, "unauthorized", "user context required")
		return
	}

	// Extract approval ID from path: /api/audit/approvals/{id}
	id := r.URL.Path[len("/api/audit/approvals/"):]
	if id == "" {
		responders.BadRequest(w, "missing_id", "approval ID is required")
		return
	}

	var body struct {
		Action string `json:"action"`
		Notes  string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		responders.BadRequest(w, "invalid_request", "invalid JSON body")
		return
	}

	var err error
	switch body.Action {
	case "submit":
		err = h.store.SubmitForReview(r.Context(), id, tenant.ID)
	case "review":
		err = h.store.Review(r.Context(), id, tenant.ID, user.ID, body.Notes)
	case "approve":
		err = h.store.Approve(r.Context(), id, tenant.ID, user.ID, body.Notes)
	case "reject":
		err = h.store.Reject(r.Context(), id, tenant.ID, user.ID, body.Notes)
	default:
		responders.BadRequest(w, "invalid_action", "action must be submit, review, approve, or reject")
		return
	}

	if err != nil {
		responders.InternalError(w, "failed to update approval: "+err.Error())
		return
	}

	_ = h.store.LogChange(r.Context(), tenant.ID, "approval", id, body.Action, "status", "", body.Action, user.ID)
	responders.JSON(w, http.StatusOK, map[string]string{"status": "ok", "action": body.Action})
}

// GetChangeLog returns the change history for the tenant.
// GET /api/audit/changelog?entity_type=report&entity_id=xxx&limit=100
func (h *Handlers) GetChangeLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}
	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}

	entityType := r.URL.Query().Get("entity_type")
	entityID := r.URL.Query().Get("entity_id")

	entries, err := h.store.GetChangeLog(r.Context(), tenant.ID, entityType, entityID, 200)
	if err != nil {
		responders.InternalError(w, "failed to fetch change log")
		return
	}
	if entries == nil {
		entries = []ChangeEntry{}
	}
	responders.JSON(w, http.StatusOK, map[string]any{"changes": entries, "count": len(entries)})
}
