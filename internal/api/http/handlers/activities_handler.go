package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/auth"
	"github.com/example/offgridflow/internal/emissions"
	"github.com/example/offgridflow/internal/ingestion"
)

// ActivitiesHandler provides endpoints to create and list activities for emissions.
type ActivitiesHandler struct {
	store      ingestion.ActivityStore
	scope1Calc *emissions.Scope1Calculator
	scope2Calc *emissions.Scope2Calculator
	scope3Calc *emissions.Scope3Calculator
	defaultOrg string
}

// ActivitiesHandlerConfig configures the activities handler.
type ActivitiesHandlerConfig struct {
	Store        ingestion.ActivityStore
	Scope1Calc   *emissions.Scope1Calculator
	Scope2Calc   *emissions.Scope2Calculator
	Scope3Calc   *emissions.Scope3Calculator
	DefaultOrgID string
}

// NewActivitiesHandler creates a new handler instance.
func NewActivitiesHandler(cfg ActivitiesHandlerConfig) *ActivitiesHandler {
	return &ActivitiesHandler{
		store:      cfg.Store,
		scope1Calc: cfg.Scope1Calc,
		scope2Calc: cfg.Scope2Calc,
		scope3Calc: cfg.Scope3Calc,
		defaultOrg: cfg.DefaultOrgID,
	}
}

// ServeHTTP handles GET (list) and POST (create) on /api/emissions/activities
func (h *ActivitiesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.list(w, r)
	case http.MethodPost:
		h.create(w, r)
	default:
		responders.MethodNotAllowed(w, http.MethodGet)
	}
}

func (h *ActivitiesHandler) list(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		responders.ServiceUnavailable(w, "activity store not configured", 0)
		return
	}
	ctx := r.Context()
	orgID := h.defaultOrg
	if tenant, ok := auth.TenantFromContext(ctx); ok && tenant != nil {
		orgID = tenant.ID
	}

	activities, err := h.store.ListByOrg(ctx, orgID)
	if err != nil {
		responders.InternalError(w, "failed to load activities")
		return
	}

	// Debug output: log count and org
	if activities == nil {
		activities = make([]ingestion.Activity, 0)
	}
	payload := map[string]interface{}{"activities": activities}
	responders.JSON(w, http.StatusOK, payload)
}

func (h *ActivitiesHandler) create(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		responders.ServiceUnavailable(w, "activity store not configured", 0)
		return
	}
	defer r.Body.Close()

	var payload struct {
		Name  string  `json:"name"`
		Type  string  `json:"type"`
		Value float64 `json:"value"`
		Unit  string  `json:"unit"`
		Date  string  `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		responders.BadRequest(w, "invalid_json", "invalid request body")
		return
	}

	// Parse date (YYYY-MM-DD) as period start
	var periodStart time.Time
	if payload.Date != "" {
		if t, err := time.Parse("2006-01-02", payload.Date); err == nil {
			periodStart = t
		}
	}
	if periodStart.IsZero() {
		periodStart = time.Now()
	}

	src := string(ingestion.SourceManual)
	if payload.Type != "" {
		// Map common payload types to ingestion.Source where appropriate
		if payload.Type == "electricity" {
			src = string(ingestion.SourceUtilityBill)
		} else {
			src = payload.Type
		}
	}

	act := ingestion.Activity{
		ID:          uuid.New().String(),
		Source:      src,
		Category:    payload.Type,
		Location:    "",
		PeriodStart: periodStart,
		PeriodEnd:   periodStart.Add(24 * time.Hour),
		Quantity:    payload.Value,
		Unit:        payload.Unit,
		OrgID:       h.defaultOrg,
		CreatedAt:   time.Now(),
	}

	if tenant, ok := auth.TenantFromContext(r.Context()); ok && tenant != nil {
		act.OrgID = tenant.ID
	}

	if err := h.store.Save(r.Context(), act); err != nil {
		responders.InternalError(w, "failed to save activity")
		return
	}

	// Attempt to compute emissions for the created activity (scope 2 if kWh)
	var emissionsResp interface{}
	if h.scope2Calc != nil && strings.EqualFold(act.Unit, "kWh") {
		records, err := h.scope2Calc.CalculateBatch(r.Context(), []emissions.Activity{act})
		if err == nil && len(records) > 0 {
			rec := records[0]
			emissionsResp = map[string]interface{}{
				"emissions_kg":     rec.EmissionsKgCO2e,
				"emissions_tonnes": rec.EmissionsTonnesCO2e,
				"emission_factor":  rec.EmissionFactor,
				"method":           rec.Method,
			}
		}
	}

	// Marshal activity to generic map so we can include computed emissions inside the activity
	var actMap map[string]interface{}
	if b, err := json.Marshal(act); err == nil {
		_ = json.Unmarshal(b, &actMap)
	} else {
		actMap = map[string]interface{}{"id": act.ID}
	}

	// Attach emissions into the activity object for compatibility with tests
	if emissionsResp != nil {
		actMap["emissions"] = emissionsResp
	} else {
		actMap["emissions"] = nil
	}

	responders.Created(w, map[string]interface{}{
		"activity": actMap,
	})

}
