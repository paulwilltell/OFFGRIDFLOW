package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/auth"
	"github.com/example/offgridflow/internal/emissions"
	"github.com/example/offgridflow/internal/ingestion"
)

// EmissionsHandler aggregates Scope 1/2/3 emissions and returns a simple summary.
type EmissionsHandler struct {
	store        ingestion.ActivityStore
	scope1Calc   *emissions.Scope1Calculator
	scope2Calc   *emissions.Scope2Calculator
	scope3Calc   *emissions.Scope3Calculator
	defaultOrgID string
}

// EmissionsHandlerConfig configures the emissions handler.
type EmissionsHandlerConfig struct {
	Store        ingestion.ActivityStore
	Scope1Calc   *emissions.Scope1Calculator
	Scope2Calc   *emissions.Scope2Calculator
	Scope3Calc   *emissions.Scope3Calculator
	DefaultOrgID string
}

// NewEmissionsHandler creates a configured emissions handler.
func NewEmissionsHandler(cfg EmissionsHandlerConfig) http.Handler {
	return &EmissionsHandler{
		store:        cfg.Store,
		scope1Calc:   cfg.Scope1Calc,
		scope2Calc:   cfg.Scope2Calc,
		scope3Calc:   cfg.Scope3Calc,
		defaultOrgID: cfg.DefaultOrgID,
	}
}

// ServeHTTP handles GET /api/emissions - returns aggregate totals.
func (h *EmissionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}
	if h.store == nil {
		responders.ServiceUnavailable(w, "activity store not configured", 0)
		return
	}

	ctx := r.Context()
	orgID := orgFromContext(r, h.defaultOrgID)

	var activities []ingestion.Activity
	var err error
	if orgID != "" {
		activities, err = h.store.ListByOrg(ctx, orgID)
	} else {
		activities, err = h.store.List(ctx)
	}
	if err != nil {
		responders.InternalError(w, "failed to load activities")
		return
	}

	emActs := make([]emissions.Activity, 0, len(activities))
	for _, a := range activities {
		emActs = append(emActs, a)
	}

	scope1 := h.calcScope1(ctx, emActs)
	scope2 := h.calcScope2(ctx, emActs)
	scope3 := h.calcScope3(ctx, emActs)

	responders.JSON(w, http.StatusOK, map[string]interface{}{
		"orgId":      orgID,
		"scope1Tons": scope1,
		"scope2Tons": scope2,
		"scope3Tons": scope3,
		"totalTons":  scope1 + scope2 + scope3,
		"timestamp":  time.Now().Format(time.RFC3339),
	})
}

func (h *EmissionsHandler) calcScope1(ctx context.Context, acts []emissions.Activity) float64 {
	if h.scope1Calc == nil {
		return 0
	}
	records, err := h.scope1Calc.CalculateBatch(ctx, acts)
	if err != nil {
		return 0
	}
	var total float64
	for _, rec := range records {
		total += rec.EmissionsTonnesCO2e
	}
	return total
}

func (h *EmissionsHandler) calcScope2(ctx context.Context, acts []emissions.Activity) float64 {
	if h.scope2Calc == nil {
		return 0
	}
	records, err := h.scope2Calc.CalculateBatch(ctx, acts)
	if err != nil {
		return 0
	}
	var total float64
	for _, rec := range records {
		total += rec.EmissionsTonnesCO2e
	}
	return total
}

func (h *EmissionsHandler) calcScope3(ctx context.Context, acts []emissions.Activity) float64 {
	if h.scope3Calc == nil {
		return 0
	}
	records, err := h.scope3Calc.CalculateBatch(ctx, acts)
	if err != nil {
		return 0
	}
	var total float64
	for _, rec := range records {
		total += rec.EmissionsTonnesCO2e
	}
	return total
}

func orgFromContext(r *http.Request, fallback string) string {
	if tenant, ok := auth.TenantFromContext(r.Context()); ok && tenant != nil && tenant.ID != "" {
		return tenant.ID
	}
	if org := r.URL.Query().Get("org_id"); org != "" {
		return org
	}
	return fallback
}
