package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/example/offgridflow/internal/api/http/middleware"
	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/auth"
	"github.com/example/offgridflow/internal/emissions"
	"github.com/example/offgridflow/internal/ingestion"
)

// -----------------------------------------------------------------------------
// Scope 2 Handler Types
// -----------------------------------------------------------------------------

// Scope2Response represents a single Scope 2 emission record in API responses.
type Scope2Response struct {
	ID                string  `json:"id"`                     // Unique identifier
	MeterID           string  `json:"meterId"`                // Meter or facility ID
	Location          string  `json:"location"`               // Geographic location
	Region            string  `json:"region"`                 // Grid region for factor lookup
	QuantityKWh       float64 `json:"quantityKWh"`            // Electricity consumed
	EmissionsKgCO2e   float64 `json:"emissionsKgCO2e"`        // Emissions in kg
	EmissionsTonsCO2e float64 `json:"emissionsTonsCO2e"`      // Emissions in metric tons
	EmissionFactor    float64 `json:"emissionFactor"`         // Factor used (kg CO2e/kWh)
	Methodology       string  `json:"methodology"`            // "location-based" or "market-based"
	DataSource        string  `json:"dataSource"`             // Source of activity data
	DataQuality       string  `json:"dataQuality"`            // Quality indicator
	PeriodStart       string  `json:"periodStart"`            // ISO 8601 date
	PeriodEnd         string  `json:"periodEnd"`              // ISO 8601 date
	CalculatedAt      string  `json:"calculatedAt,omitempty"` // When calculated
}

// Scope2SummaryResponse provides aggregated Scope 2 data.
type Scope2SummaryResponse struct {
	Scope                  string             `json:"scope"`
	TotalKWh               float64            `json:"totalKWh"`
	TotalEmissionsKgCO2e   float64            `json:"totalEmissionsKgCO2e"`
	TotalEmissionsTonsCO2e float64            `json:"totalEmissionsTonsCO2e"`
	AverageEmissionFactor  float64            `json:"averageEmissionFactor"`
	ActivityCount          int                `json:"activityCount"`
	RegionBreakdown        map[string]float64 `json:"regionBreakdown,omitempty"`
	PeriodStart            string             `json:"periodStart,omitempty"`
	PeriodEnd              string             `json:"periodEnd,omitempty"`
	Timestamp              string             `json:"timestamp"`
}

// -----------------------------------------------------------------------------
// Scope 2 Handler
// -----------------------------------------------------------------------------

// Scope2Handler provides HTTP handlers for Scope 2 emissions endpoints.
type Scope2Handler struct {
	activityStore ingestion.ActivityStore
	calculator    *emissions.Scope2Calculator
	logger        *slog.Logger
}

// Scope2HandlerDeps holds dependencies for the Scope 2 handler (legacy interface).
type Scope2HandlerDeps struct {
	ActivityStore    ingestion.ActivityStore
	Scope2Calculator *emissions.Scope2Calculator
}

// NewScope2HandlerWithDeps creates a Scope 2 handler with full configuration.
func NewScope2HandlerWithDeps(deps *Scope2HandlerDeps, logger *slog.Logger) *Scope2Handler {
	if logger == nil {
		logger = slog.Default().With("component", "scope2-handler")
	}

	return &Scope2Handler{
		activityStore: deps.ActivityStore,
		calculator:    deps.Scope2Calculator,
		logger:        logger,
	}
}

// NewScope2Handler creates an HTTP handler for GET /api/emissions/scope2.
func NewScope2Handler(deps *Scope2HandlerDeps) http.HandlerFunc {
	handler := NewScope2HandlerWithDeps(deps, nil)
	return handler.List
}

// List handles GET /api/emissions/scope2 - returns detailed Scope 2 records.
func (h *Scope2Handler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}

	if h.activityStore == nil || h.calculator == nil {
		responders.ServiceUnavailable(w, "scope2 handler not configured", 0)
		return
	}

	tenantID, ok := middleware.MustGetTenantID(w, r)
	if !ok {
		return
	}
	tenant, ok := auth.TenantFromContext(r.Context())
	if !ok || tenant == nil {
		tenant = &auth.Tenant{ID: tenantID}
	}

	ctx := r.Context()

	// Parse pagination
	page, perPage := responders.ParsePagination(r, 50)

	// Parse optional filters
	region := r.URL.Query().Get("region")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	// Get activities (in production, filter by tenant and date range)
	activities, err := h.activityStore.ListByOrgAndSource(ctx, tenantID, "utility_bill")
	if err != nil {
		h.logger.Error("failed to load activities",
			"tenantId", tenant.ID,
			"error", err.Error(),
		)
		responders.InternalError(w, "failed to load activities")
		return
	}

	// Convert activities to emissions.Activity interface slice
	emissionsActivities := make([]emissions.Activity, 0, len(activities))
	for _, act := range activities {
		emissionsActivities = append(emissionsActivities, act)
	}

	// Calculate Scope 2 emissions
	records, err := h.calculator.CalculateBatch(ctx, emissionsActivities)
	if err != nil {
		h.logger.Error("failed to calculate scope 2 emissions",
			"tenantId", tenant.ID,
			"error", err.Error(),
		)
		responders.InternalError(w, "failed to calculate emissions")
		return
	}

	// Debug log calculated records
	if h.logger != nil {
		h.logger.Info("scope2 calculated records", "count", len(records))
		for i, rec := range records {
			h.logger.Info("scope2 record", "idx", i, "activity_id", rec.ActivityID, "period_start", rec.PeriodStart.Format(time.RFC3339))
		}
	}

	if h.logger != nil {
		h.logger.Info("scope2 calculated records", "count", len(records))
		for i, rec := range records {
			h.logger.Info("scope2 record", "idx", i, "activity_id", rec.ActivityID, "period_start", rec.PeriodStart)
		}
	}

	// Filter results
	filtered := h.filterRecords(records, region, startDate, endDate)

	// Apply pagination
	total := len(filtered)
	start := (page - 1) * perPage
	end := start + perPage

	if start >= total {
		filtered = []emissions.EmissionRecord{}
	} else if end > total {
		filtered = filtered[start:]
	} else {
		filtered = filtered[start:end]
	}

	// Convert to API response format
	now := time.Now().Format(time.RFC3339)
	response := make([]Scope2Response, 0, len(filtered))
	for _, rec := range filtered {
		response = append(response, Scope2Response{
			ID:                rec.ID,
			MeterID:           rec.ActivityID,
			Location:          rec.Region,
			Region:            rec.Region,
			QuantityKWh:       rec.InputQuantity,
			EmissionsKgCO2e:   rec.EmissionsKgCO2e,
			EmissionsTonsCO2e: rec.EmissionsTonnesCO2e,
			EmissionFactor:    rec.EmissionFactor,
			Methodology:       string(rec.Method),
			DataSource:        "utility_bill",
			DataQuality:       string(rec.DataQuality),
			PeriodStart:       rec.PeriodStart.Format(time.RFC3339),
			PeriodEnd:         rec.PeriodEnd.Format(time.RFC3339),
			CalculatedAt:      now,
		})
	}

	pageInfo := responders.NewPageInfo(page, perPage, total)
	responders.Paged(w, http.StatusOK, response, pageInfo)
}

// filterRecords applies optional filters to the emission records.
func (h *Scope2Handler) filterRecords(records []emissions.EmissionRecord, region, startDate, endDate string) []emissions.EmissionRecord {
	if region == "" && startDate == "" && endDate == "" {
		return records
	}

	var start, end time.Time
	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			start = t
		}
	}
	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			end = t.Add(24 * time.Hour) // Include end date
		}
	}

	filtered := make([]emissions.EmissionRecord, 0, len(records))
	for _, rec := range records {
		// Region filter
		if region != "" && rec.Region != region {
			continue
		}

		// Date range filter
		if !start.IsZero() && rec.PeriodStart.Before(start) {
			continue
		}
		if !end.IsZero() && rec.PeriodEnd.After(end) {
			continue
		}

		filtered = append(filtered, rec)
	}

	return filtered
}

// Scope2SummaryHandler returns aggregated Scope 2 emissions.
func Scope2SummaryHandler(deps *Scope2HandlerDeps) http.HandlerFunc {
	handler := NewScope2HandlerWithDeps(deps, nil)
	return handler.Summary
}

// Summary handles GET /api/emissions/scope2/summary - returns aggregated data.
func (h *Scope2Handler) Summary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responders.MethodNotAllowed(w, http.MethodGet)
		return
	}

	if h.activityStore == nil || h.calculator == nil {
		responders.ServiceUnavailable(w, "scope2 handler not configured", 0)
		return
	}

	tenantID, ok := middleware.MustGetTenantID(w, r)
	if !ok {
		return
	}
	tenant, _ := auth.TenantFromContext(r.Context())
	if tenant == nil {
		tenant = &auth.Tenant{ID: tenantID}
	}

	ctx := r.Context()

	// Parse optional year filter
	year := time.Now().Year()
	if yearStr := r.URL.Query().Get("year"); yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil && y > 2000 && y <= time.Now().Year()+1 {
			year = y
		}
	}

	// Log tenant for debugging
	if h.logger != nil {
		h.logger.Info("scope2 summary request", "tenant", tenantID)
	}

	// Get activities
	activities, err := h.activityStore.ListByOrgAndSource(ctx, tenantID, "utility_bill")
	if err != nil {
		h.logger.Error("failed to load activities",
			"tenantId", tenant.ID,
			"error", err.Error(),
		)
		responders.InternalError(w, "failed to load activities")
		return
	}

	// Convert activities to emissions.Activity interface slice
	emissionsActivities := make([]emissions.Activity, 0, len(activities))
	for _, act := range activities {
		emissionsActivities = append(emissionsActivities, act)
	}

	// Calculate Scope 2 emissions
	records, err := h.calculator.CalculateBatch(ctx, emissionsActivities)
	if err != nil {
		h.logger.Error("failed to calculate scope 2 emissions",
			"tenantId", tenant.ID,
			"error", err.Error(),
		)
		responders.InternalError(w, "failed to calculate emissions")
		return
	}

	// Aggregate totals
	var totalKWh, totalKg, totalFactor float64
	regionBreakdown := make(map[string]float64)
	var minDate, maxDate time.Time
	count := 0

	for _, rec := range records {
		// Filter by year if specified (use UTC to avoid timezone shift excluding UTC-midnight records)
		if rec.PeriodStart.UTC().Year() != year {
			continue
		}
		count++

		totalKWh += rec.InputQuantity
		totalKg += rec.EmissionsKgCO2e
		totalFactor += rec.EmissionFactor
		regionBreakdown[rec.Region] += rec.EmissionsTonnesCO2e

		if minDate.IsZero() || rec.PeriodStart.Before(minDate) {
			minDate = rec.PeriodStart
		}
		if maxDate.IsZero() || rec.PeriodEnd.After(maxDate) {
			maxDate = rec.PeriodEnd
		}
	}

	avgFactor := 0.0
	if count > 0 {
		avgFactor = totalFactor / float64(count)
	}

	response := Scope2SummaryResponse{
		Scope:                  "SCOPE2",
		TotalKWh:               totalKWh,
		TotalEmissionsKgCO2e:   totalKg,
		TotalEmissionsTonsCO2e: totalKg / 1000.0,
		AverageEmissionFactor:  avgFactor,
		ActivityCount:          count,
		RegionBreakdown:        regionBreakdown,
		Timestamp:              time.Now().Format(time.RFC3339),
	}

	if !minDate.IsZero() {
		response.PeriodStart = minDate.Format("2006-01-02")
	}
	if !maxDate.IsZero() {
		response.PeriodEnd = maxDate.Format("2006-01-02")
	}

	// Build compatibility map matching test expectations
	summaryMap := map[string]interface{}{
		"scope":                   response.Scope,
		"total_emissions":         response.TotalEmissionsKgCO2e,
		"total_activities":        float64(response.ActivityCount),
		"total_kwh":               response.TotalKWh,
		"average_emission_factor": response.AverageEmissionFactor,
		"region_breakdown":        response.RegionBreakdown,
		"period_start":            response.PeriodStart,
		"period_end":              response.PeriodEnd,
		"timestamp":               response.Timestamp,
	}

	// Log summary for debugging
	if h.logger != nil {
		h.logger.Info("scope2 summary", "summary", summaryMap)
	}

	// Cache this response for 5 minutes
	responders.SetCacheControl(w, 5*time.Minute, false)
	responders.JSON(w, http.StatusOK, map[string]interface{}{"summary": summaryMap})
}
