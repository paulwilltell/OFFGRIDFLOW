package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/auth"
)

// DataGovernanceHandlers provides endpoints for data export, deletion, and retention.
type DataGovernanceHandlers struct {
	db *sql.DB
}

// NewDataGovernanceHandlers creates data governance endpoints.
func NewDataGovernanceHandlers(db *sql.DB) *DataGovernanceHandlers {
	return &DataGovernanceHandlers{db: db}
}

// ExportAllData exports all tenant data as a single JSON package.
// GET /api/governance/export
func (h *DataGovernanceHandlers) ExportAllData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}
	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		responders.Unauthorized(w, "unauthorized", "authentication required")
		return
	}

	ctx := r.Context()
	export := map[string]any{
		"exported_at": time.Now().UTC().Format(time.RFC3339),
		"tenant_id":   tenant.ID,
		"tenant_name": tenant.Name,
	}

	// Export users
	if rows, err := h.db.QueryContext(ctx, `SELECT id, email, name, first_name, last_name, role, created_at FROM users WHERE tenant_id = $1`, tenant.ID); err == nil {
		defer rows.Close()
		var users []map[string]any
		for rows.Next() {
			var id, email, name, firstName, lastName, role string
			var createdAt time.Time
			rows.Scan(&id, &email, &name, &firstName, &lastName, &role, &createdAt)
			users = append(users, map[string]any{"id": id, "email": email, "name": name, "first_name": firstName, "last_name": lastName, "role": role, "created_at": createdAt})
		}
		export["users"] = users
	}

	// Export activities
	if rows, err := h.db.QueryContext(ctx, `SELECT id, source, location, quantity_kwh, period_start, period_end, created_at FROM activities WHERE org_id = $1`, tenant.ID); err == nil {
		defer rows.Close()
		var activities []map[string]any
		for rows.Next() {
			var id, source, location string
			var qty float64
			var pStart, pEnd, createdAt time.Time
			if rows.Scan(&id, &source, &location, &qty, &pStart, &pEnd, &createdAt) == nil {
				activities = append(activities, map[string]any{"id": id, "source": source, "location": location, "quantity_kwh": qty, "period_start": pStart, "period_end": pEnd, "created_at": createdAt})
			}
		}
		export["activities"] = activities
	}

	// Export calculation ledger
	if rows, err := h.db.QueryContext(ctx, `SELECT id, scope, category, quantity, unit, emission_factor_id, emission_factor_value, method, result_kg_co2e, result_tonnes_co2e, formula, calculated_at FROM calculation_ledger WHERE tenant_id = $1 ORDER BY calculated_at`, tenant.ID); err == nil {
		defer rows.Close()
		var ledger []map[string]any
		for rows.Next() {
			var id, scope, category, unit, factorID, method, formula string
			var qty, factorVal, kgCO2e, tonnesCO2e float64
			var calcAt time.Time
			if rows.Scan(&id, &scope, &category, &qty, &unit, &factorID, &factorVal, &method, &kgCO2e, &tonnesCO2e, &formula, &calcAt) == nil {
				ledger = append(ledger, map[string]any{"id": id, "scope": scope, "category": category, "quantity": qty, "unit": unit, "factor_id": factorID, "factor_value": factorVal, "method": method, "kg_co2e": kgCO2e, "tonnes_co2e": tonnesCO2e, "formula": formula, "calculated_at": calcAt})
			}
		}
		export["calculation_ledger"] = ledger
	}

	// Export change log
	if rows, err := h.db.QueryContext(ctx, `SELECT id, entity_type, entity_id, action, field_name, old_value, new_value, changed_at FROM change_log WHERE tenant_id = $1 ORDER BY changed_at`, tenant.ID); err == nil {
		defer rows.Close()
		var changes []map[string]any
		for rows.Next() {
			var id, eType, eID, action, field, oldVal, newVal string
			var changedAt time.Time
			if rows.Scan(&id, &eType, &eID, &action, &field, &oldVal, &newVal, &changedAt) == nil {
				changes = append(changes, map[string]any{"id": id, "entity_type": eType, "entity_id": eID, "action": action, "field": field, "old_value": oldVal, "new_value": newVal, "changed_at": changedAt})
			}
		}
		export["change_log"] = changes
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="offgridflow-export-%s-%s.json"`, tenant.ID, time.Now().Format("20060102")))
	json.NewEncoder(w).Encode(export)
}

// RequestDeletion marks a tenant's data for deletion after the retention period.
// POST /api/governance/delete-request
func (h *DataGovernanceHandlers) RequestDeletion(w http.ResponseWriter, r *http.Request) {
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
	if !uOk || user == nil || user.Role != "admin" {
		responders.Error(w, http.StatusForbidden, "forbidden", "only tenant admins can request data deletion")
		return
	}

	// Log the deletion request
	_, err := h.db.ExecContext(r.Context(),
		`INSERT INTO change_log (tenant_id, entity_type, entity_id, action, field_name, new_value, changed_by, changed_at)
		 VALUES ($1, 'tenant', $1, 'deletion_request', 'status', 'pending_deletion', $2, NOW())`,
		tenant.ID, user.ID)
	if err != nil {
		responders.InternalError(w, "failed to record deletion request")
		return
	}

	responders.JSON(w, http.StatusOK, map[string]any{
		"status":          "deletion_requested",
		"tenant_id":       tenant.ID,
		"retention_days":  30,
		"deletion_date":   time.Now().AddDate(0, 0, 30).Format("2006-01-02"),
		"message":         "Your data will be permanently deleted after the 30-day retention period. You can export your data before then via GET /api/governance/export.",
	})
}

// GetRetentionPolicy returns the data retention policy.
// GET /api/governance/retention
func (h *DataGovernanceHandlers) GetRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}

	policy := map[string]any{
		"emission_data": map[string]any{
			"retention_period": "Duration of subscription + 90 days",
			"basis":            "Contractual necessity for compliance reporting",
		},
		"calculation_ledger": map[string]any{
			"retention_period": "Duration of subscription + 7 years",
			"basis":            "Audit trail preservation for regulatory compliance",
		},
		"user_accounts": map[string]any{
			"retention_period": "Duration of subscription + 30 days",
			"basis":            "Service delivery",
		},
		"change_log": map[string]any{
			"retention_period": "Duration of subscription + 7 years",
			"basis":            "Audit integrity",
		},
		"evidence_files": map[string]any{
			"retention_period": "Duration of subscription + 90 days",
			"basis":            "Compliance documentation",
		},
		"deletion_process": "Data export available on request. Deletion completes within 30 days of request. Audit logs retained per regulatory requirement.",
		"data_ownership":   "All emission data, reports, and uploaded evidence remain the property of the customer. OffGridFlow processes data solely for service delivery.",
		"export_format":    "JSON (full dataset), PDF (reports), CSV (activities), XBRL (compliance)",
	}

	responders.JSON(w, http.StatusOK, policy)
}
