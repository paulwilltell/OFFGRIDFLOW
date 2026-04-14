package audit

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/auth"
)

// ============================================================================
// Factor Snapshot Handlers (Panel 1B: Reproducibility)
// ============================================================================

// GetFactorSnapshots returns all factor snapshots for the organization.
// GET /api/audit/factor-snapshots
func (h *Handlers) GetFactorSnapshots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}
	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}

	snapshots, err := h.store.ListFactorSnapshots(r.Context(), tenant.ID)
	if err != nil {
		responders.InternalError(w, "failed to fetch factor snapshots")
		return
	}
	if snapshots == nil {
		snapshots = []FactorSnapshot{}
	}
	responders.JSON(w, http.StatusOK, map[string]any{"snapshots": snapshots, "count": len(snapshots)})
}

// GetFactorSnapshot returns a single factor snapshot with full factor data.
// GET /api/audit/factor-snapshots/{id}
func (h *Handlers) GetFactorSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}
	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/audit/factor-snapshots/")
	if id == "" {
		responders.BadRequest(w, "missing_id", "snapshot ID is required")
		return
	}

	snapshot, err := h.store.GetFactorSnapshot(r.Context(), id, tenant.ID)
	if err != nil {
		responders.NotFound(w, "snapshot_not_found", err.Error())
		return
	}
	responders.JSON(w, http.StatusOK, snapshot)
}

// CreateFactorSnapshot creates a new draft factor snapshot.
// POST /api/audit/factor-snapshots
func (h *Handlers) CreateFactorSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responders.MethodNotAllowed(w, http.MethodPost)
		return
	}
	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}

	var body struct {
		PeriodStart     string           `json:"period_start"`
		PeriodEnd       string           `json:"period_end"`
		Name            string           `json:"name"`
		Factors         []SnapshotFactor `json:"factors"`
		RegistryVersion string           `json:"registry_version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		responders.BadRequest(w, "invalid_request", "invalid JSON body")
		return
	}
	if body.PeriodStart == "" || body.PeriodEnd == "" || body.Name == "" {
		responders.BadRequest(w, "validation_error", "period_start, period_end, and name are required")
		return
	}

	snapshot, err := h.store.CreateFactorSnapshot(r.Context(), tenant.ID, body.PeriodStart, body.PeriodEnd, body.Name, body.Factors, body.RegistryVersion)
	if err != nil {
		responders.InternalError(w, "failed to create factor snapshot: "+err.Error())
		return
	}

	responders.JSON(w, http.StatusCreated, snapshot)
}

// LockFactorSnapshot transitions a snapshot from draft to locked (immutable).
// POST /api/audit/factor-snapshots/{id}/lock
func (h *Handlers) LockFactorSnapshot(w http.ResponseWriter, r *http.Request) {
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

	path := strings.TrimPrefix(r.URL.Path, "/api/audit/factor-snapshots/")
	id := strings.TrimSuffix(path, "/lock")
	if id == "" {
		responders.BadRequest(w, "missing_id", "snapshot ID is required")
		return
	}

	if err := h.store.LockFactorSnapshot(r.Context(), id, tenant.ID, user.ID); err != nil {
		responders.InternalError(w, "failed to lock factor snapshot: "+err.Error())
		return
	}

	responders.JSON(w, http.StatusOK, map[string]string{"status": "locked", "id": id})
}

// ============================================================================
// Data Quality Anomaly Handlers (Panel 1B: Anomaly Detection)
// ============================================================================

// GetAnomalies returns data quality anomalies for the organization.
// GET /api/audit/anomalies?status=open&severity=critical
func (h *Handlers) GetAnomalies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}
	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}

	status := r.URL.Query().Get("status")
	severity := r.URL.Query().Get("severity")

	anomalies, err := h.store.GetAnomalies(r.Context(), tenant.ID, status, severity, 100)
	if err != nil {
		responders.InternalError(w, "failed to fetch anomalies")
		return
	}
	if anomalies == nil {
		anomalies = []Anomaly{}
	}

	counts, _ := h.store.GetAnomalyCounts(r.Context(), tenant.ID)

	responders.JSON(w, http.StatusOK, map[string]any{
		"anomalies": anomalies,
		"count":     len(anomalies),
		"summary":   counts,
	})
}

// RunAnomalyDetection triggers a data quality scan.
// POST /api/audit/anomalies/scan
func (h *Handlers) RunAnomalyDetectionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responders.MethodNotAllowed(w, http.MethodPost)
		return
	}
	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}

	detected, err := h.store.RunAnomalyDetection(r.Context(), tenant.ID)
	if err != nil {
		responders.InternalError(w, "anomaly detection failed: "+err.Error())
		return
	}

	responders.JSON(w, http.StatusOK, map[string]any{
		"detected": detected,
		"status":   "scan_complete",
	})
}

// ResolveAnomaly resolves or dismisses a data quality anomaly.
// PUT /api/audit/anomalies/{id}
func (h *Handlers) ResolveAnomalyHandler(w http.ResponseWriter, r *http.Request) {
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

	id := strings.TrimPrefix(r.URL.Path, "/api/audit/anomalies/")
	if id == "" {
		responders.BadRequest(w, "missing_id", "anomaly ID is required")
		return
	}

	var body struct {
		Action string `json:"action"` // resolve, dismiss
		Notes  string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		responders.BadRequest(w, "invalid_request", "invalid JSON body")
		return
	}

	var err error
	switch body.Action {
	case "resolve":
		err = h.store.ResolveAnomaly(r.Context(), id, tenant.ID, user.ID, body.Notes)
	case "dismiss":
		err = h.store.DismissAnomaly(r.Context(), id, tenant.ID, user.ID, body.Notes)
	default:
		responders.BadRequest(w, "invalid_action", "action must be resolve or dismiss")
		return
	}

	if err != nil {
		responders.InternalError(w, "failed to update anomaly: "+err.Error())
		return
	}

	responders.JSON(w, http.StatusOK, map[string]string{"status": "ok", "action": body.Action})
}

// ============================================================================
// Alert Action Handlers (Panel 2C: Built-in Next Actions)
// ============================================================================

// GetAlertActions returns alert actions for the organization.
// GET /api/audit/alerts?status=open&priority=critical
func (h *Handlers) GetAlertActions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}
	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}

	status := r.URL.Query().Get("status")
	priority := r.URL.Query().Get("priority")

	alerts, err := h.store.GetAlertActions(r.Context(), tenant.ID, status, priority, 100)
	if err != nil {
		responders.InternalError(w, "failed to fetch alerts")
		return
	}
	if alerts == nil {
		alerts = []AlertAction{}
	}

	counts, _ := h.store.GetAlertActionCounts(r.Context(), tenant.ID)

	responders.JSON(w, http.StatusOK, map[string]any{
		"alerts":  alerts,
		"count":   len(alerts),
		"summary": counts,
	})
}

// UpdateAlertAction handles alert state changes (assign, escalate, resolve, dismiss).
// PUT /api/audit/alerts/{id}
func (h *Handlers) UpdateAlertAction(w http.ResponseWriter, r *http.Request) {
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

	id := strings.TrimPrefix(r.URL.Path, "/api/audit/alerts/")
	parts := strings.SplitN(id, "/", 2)
	id = parts[0]
	if id == "" {
		responders.BadRequest(w, "missing_id", "alert ID is required")
		return
	}

	var body struct {
		Action     string `json:"action"` // assign, escalate, resolve, dismiss, comment
		AssignedTo string `json:"assigned_to,omitempty"`
		Content    string `json:"content,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		responders.BadRequest(w, "invalid_request", "invalid JSON body")
		return
	}

	var err error
	switch body.Action {
	case "assign":
		target := body.AssignedTo
		if target == "" {
			target = user.ID
		}
		err = h.store.AssignAlertAction(r.Context(), id, tenant.ID, target)
	case "escalate":
		err = h.store.EscalateAlertAction(r.Context(), id, tenant.ID, body.AssignedTo)
	case "resolve":
		err = h.store.ResolveAlertAction(r.Context(), id, tenant.ID)
	case "dismiss":
		err = h.store.DismissAlertAction(r.Context(), id, tenant.ID)
	case "comment":
		if body.Content == "" {
			responders.BadRequest(w, "validation_error", "content is required for comments")
			return
		}
		_, err = h.store.AddAlertComment(r.Context(), id, user.ID, body.Content)
	default:
		responders.BadRequest(w, "invalid_action", "action must be assign, escalate, resolve, dismiss, or comment")
		return
	}

	if err != nil {
		responders.InternalError(w, "failed to update alert: "+err.Error())
		return
	}

	responders.JSON(w, http.StatusOK, map[string]string{"status": "ok", "action": body.Action})
}

// GetAlertComments returns comments for an alert.
// GET /api/audit/alerts/{id}/comments
func (h *Handlers) GetAlertCommentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/audit/alerts/")
	id := strings.TrimSuffix(path, "/comments")
	if id == "" {
		responders.BadRequest(w, "missing_id", "alert ID is required")
		return
	}

	comments, err := h.store.GetAlertComments(r.Context(), id)
	if err != nil {
		responders.InternalError(w, "failed to fetch comments")
		return
	}
	if comments == nil {
		comments = []AlertComment{}
	}
	responders.JSON(w, http.StatusOK, map[string]any{"comments": comments})
}

// ============================================================================
// Export Reconciliation Handlers (Panel 2E: Exports Match On-Screen Truth)
// ============================================================================

// GetExportHistory returns export records for reconciliation tracking.
// GET /api/audit/exports?report_id=xxx
func (h *Handlers) GetExportHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}
	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}

	reportID := r.URL.Query().Get("report_id")

	exports, err := h.store.GetExportHistory(r.Context(), tenant.ID, reportID)
	if err != nil {
		responders.InternalError(w, "failed to fetch export history")
		return
	}
	if exports == nil {
		exports = []ReportExport{}
	}
	responders.JSON(w, http.StatusOK, map[string]any{"exports": exports, "count": len(exports)})
}

// ReconcileExportHandler compares an export's data to current values.
// GET /api/audit/exports/{id}/reconcile
func (h *Handlers) ReconcileExportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}
	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/audit/exports/")
	id := strings.TrimSuffix(path, "/reconcile")

	recon, err := h.store.ReconcileExport(r.Context(), id, tenant.ID)
	if err != nil {
		responders.InternalError(w, "reconciliation failed: "+err.Error())
		return
	}
	responders.JSON(w, http.StatusOK, recon)
}

// GenerateStakeholderExportHandler creates a stakeholder-ready export package.
// POST /api/audit/exports/stakeholder
func (h *Handlers) GenerateStakeholderExportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responders.MethodNotAllowed(w, http.MethodPost)
		return
	}
	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}

	var body struct {
		ReportID string                  `json:"report_id"`
		Config   StakeholderExportConfig `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		responders.BadRequest(w, "invalid_request", "invalid JSON body")
		return
	}
	if body.ReportID == "" {
		responders.BadRequest(w, "validation_error", "report_id is required")
		return
	}

	data, err := h.store.GenerateStakeholderExport(r.Context(), tenant.ID, body.ReportID, body.Config)
	if err != nil {
		responders.InternalError(w, "stakeholder export failed: "+err.Error())
		return
	}
	responders.JSON(w, http.StatusOK, data)
}

// ============================================================================
// Customer Health Handlers (Panel 3E: Renewal Engineering)
// ============================================================================

// GetCustomerHealth returns the latest health score for the organization.
// GET /api/audit/health
func (h *Handlers) GetCustomerHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}
	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}

	health, err := h.store.GetLatestHealthScore(r.Context(), tenant.ID)
	if err != nil {
		responders.InternalError(w, "failed to fetch health score")
		return
	}
	if health == nil {
		responders.JSON(w, http.StatusOK, map[string]any{"health": nil, "message": "no health score available yet"})
		return
	}
	responders.JSON(w, http.StatusOK, map[string]any{"health": health})
}

// RefreshCustomerHealth recalculates the health score.
// POST /api/audit/health/refresh
func (h *Handlers) RefreshCustomerHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responders.MethodNotAllowed(w, http.MethodPost)
		return
	}
	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}

	health, err := h.store.CalculateAndStoreHealthScore(r.Context(), tenant.ID)
	if err != nil {
		responders.InternalError(w, "failed to calculate health score: "+err.Error())
		return
	}
	responders.JSON(w, http.StatusOK, map[string]any{"health": health})
}

// GetHealthHistory returns health score trend data.
// GET /api/audit/health/history?limit=30
func (h *Handlers) GetHealthHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}
	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}

	scores, err := h.store.GetHealthScoreHistory(r.Context(), tenant.ID, 30)
	if err != nil {
		responders.InternalError(w, "failed to fetch health history")
		return
	}
	if scores == nil {
		scores = []CustomerHealth{}
	}
	responders.JSON(w, http.StatusOK, map[string]any{"history": scores, "count": len(scores)})
}
