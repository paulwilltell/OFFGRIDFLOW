package graph

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/example/offgridflow/internal/compliance/csrd"
	"github.com/example/offgridflow/internal/emissions"
	"github.com/example/offgridflow/internal/ingestion"
)

// ProductionQueryResolver implements QueryResolver with real service integrations.
// This resolver replaces DefaultQueryResolver for production use.
type ProductionQueryResolver struct {
	activityStore    ingestion.ActivityStore
	scope1Calculator *emissions.Scope1Calculator
	scope2Calculator *emissions.Scope2Calculator
	scope3Calculator *emissions.Scope3Calculator
	csrdMapper       csrd.CSRDMapper
	logger           *slog.Logger
}

// ProductionQueryResolverConfig holds dependencies for ProductionQueryResolver.
type ProductionQueryResolverConfig struct {
	ActivityStore    ingestion.ActivityStore
	Scope1Calculator *emissions.Scope1Calculator
	Scope2Calculator *emissions.Scope2Calculator
	Scope3Calculator *emissions.Scope3Calculator
	CSRDMapper       csrd.CSRDMapper
	Logger           *slog.Logger
}

// NewProductionQueryResolver creates a production-ready query resolver.
func NewProductionQueryResolver(cfg ProductionQueryResolverConfig) (*ProductionQueryResolver, error) {
	if cfg.ActivityStore == nil {
		return nil, fmt.Errorf("production resolver requires ActivityStore")
	}
	if cfg.Scope1Calculator == nil {
		return nil, fmt.Errorf("production resolver requires Scope1Calculator")
	}
	if cfg.Scope2Calculator == nil {
		return nil, fmt.Errorf("production resolver requires Scope2Calculator")
	}
	if cfg.Scope3Calculator == nil {
		return nil, fmt.Errorf("production resolver requires Scope3Calculator")
	}

	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default().With("component", "production-query-resolver")
	}

	return &ProductionQueryResolver{
		activityStore:    cfg.ActivityStore,
		scope1Calculator: cfg.Scope1Calculator,
		scope2Calculator: cfg.Scope2Calculator,
		scope3Calculator: cfg.Scope3Calculator,
		csrdMapper:       cfg.CSRDMapper,
		logger:           logger,
	}, nil
}

// Placeholder implements the health check placeholder query.
func (r *ProductionQueryResolver) Placeholder(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("%w: %v", ErrContextCanceled, err)
	}
	return "OffGridFlow GraphQL API - Production", nil
}

// Emissions resolves emissions data queries with real data from calculators.
func (r *ProductionQueryResolver) Emissions(ctx context.Context, filter EmissionsFilter) (*EmissionsConnection, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrContextCanceled, err)
	}

	// Load activities from store
	activities, err := r.activityStore.ListBySource(ctx, "utility_bill")
	if err != nil {
		r.logger.Error("failed to load activities", "error", err)
		return nil, fmt.Errorf("failed to load activities: %w", err)
	}

	// Convert to emissions.Activity interface
	emissionsActivities := make([]emissions.Activity, 0, len(activities))
	for _, act := range activities {
		emissionsActivities = append(emissionsActivities, act)
	}

	// Calculate emissions for all scopes
	scope1Records, err := r.scope1Calculator.CalculateBatch(ctx, emissionsActivities)
	if err != nil {
		r.logger.Error("scope1 calculation failed", "error", err)
		return nil, fmt.Errorf("scope1 calculation failed: %w", err)
	}

	scope2Records, err := r.scope2Calculator.CalculateBatch(ctx, emissionsActivities)
	if err != nil {
		r.logger.Error("scope2 calculation failed", "error", err)
		return nil, fmt.Errorf("scope2 calculation failed: %w", err)
	}

	scope3Records, err := r.scope3Calculator.CalculateBatch(ctx, emissionsActivities)
	if err != nil {
		r.logger.Error("scope3 calculation failed", "error", err)
		return nil, fmt.Errorf("scope3 calculation failed: %w", err)
	}

	// Combine all records
	allRecords := make([]emissions.EmissionRecord, 0, len(scope1Records)+len(scope2Records)+len(scope3Records))
	allRecords = append(allRecords, scope1Records...)
	allRecords = append(allRecords, scope2Records...)
	allRecords = append(allRecords, scope3Records...)

	// Apply scope filter if provided
	var filteredRecords []emissions.EmissionRecord
	if filter.Scope != nil && *filter.Scope != "" {
		for _, rec := range allRecords {
			if rec.Scope.String() == *filter.Scope {
				filteredRecords = append(filteredRecords, rec)
			}
		}
	} else {
		filteredRecords = allRecords
	}

	// Convert to GraphQL edges
	edges := make([]*EmissionsEdge, 0, len(filteredRecords))
	for _, rec := range filteredRecords {
		edges = append(edges, &EmissionsEdge{
			Node: &Emission{
				ID:               rec.ID,
				Scope:            rec.Scope.String(),
				Category:         rec.FactorID,
				Source:           rec.ActivityID,
				AmountKgCO2e:     rec.EmissionsKgCO2e,
				AmountTonnesCO2e: rec.EmissionsTonnesCO2e,
				Region:           rec.Region,
				PeriodStart:      rec.PeriodStart,
				PeriodEnd:        rec.PeriodEnd,
				CalculatedAt:     time.Now(),
				EmissionFactor:   rec.EmissionFactor,
				DataQuality:      string(rec.DataQuality),
			},
			Cursor: rec.ID,
		})
	}

	// Create cursor pointers
	var startCursor, endCursor *string
	if len(edges) > 0 {
		start := edges[0].Cursor
		end := edges[len(edges)-1].Cursor
		startCursor = &start
		endCursor = &end
	}

	return &EmissionsConnection{
		Edges: edges,
		PageInfo: &PageInfo{
			HasNextPage:     false,
			HasPreviousPage: false,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
		TotalCount: len(filteredRecords),
	}, nil
}

// EmissionsSummary resolves the emissions summary with real calculated totals.
func (r *ProductionQueryResolver) EmissionsSummary(ctx context.Context, year int) (*EmissionsSummary, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrContextCanceled, err)
	}

	// Load activities from store
	activities, err := r.activityStore.ListBySource(ctx, "utility_bill")
	if err != nil {
		r.logger.Error("failed to load activities", "error", err)
		return nil, fmt.Errorf("failed to load activities: %w", err)
	}

	// Convert to emissions.Activity interface
	emissionsActivities := make([]emissions.Activity, 0, len(activities))
	for _, act := range activities {
		emissionsActivities = append(emissionsActivities, act)
	}

	// Calculate all scopes
	scope1Records, _ := r.scope1Calculator.CalculateBatch(ctx, emissionsActivities)
	scope2Records, _ := r.scope2Calculator.CalculateBatch(ctx, emissionsActivities)
	scope3Records, _ := r.scope3Calculator.CalculateBatch(ctx, emissionsActivities)

	// Aggregate totals
	var scope1Total, scope2Total, scope3Total float64
	for _, rec := range scope1Records {
		scope1Total += rec.EmissionsTonnesCO2e
	}
	for _, rec := range scope2Records {
		scope2Total += rec.EmissionsTonnesCO2e
	}
	for _, rec := range scope3Records {
		scope3Total += rec.EmissionsTonnesCO2e
	}

	return &EmissionsSummary{
		Scope1TonnesCO2e: scope1Total,
		Scope2TonnesCO2e: scope2Total,
		Scope3TonnesCO2e: scope3Total,
		TotalTonnesCO2e:  scope1Total + scope2Total + scope3Total,
		Year:             year,
		Timestamp:        time.Now(),
	}, nil
}

// ComplianceStatus resolves compliance status for a specific framework with real data.
func (r *ProductionQueryResolver) ComplianceStatus(ctx context.Context, framework string) (*ComplianceStatus, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrContextCanceled, err)
	}

	// Load activities for compliance calculation
	activities, err := r.activityStore.ListBySource(ctx, "utility_bill")
	if err != nil {
		r.logger.Error("failed to load activities", "error", err)
		return nil, fmt.Errorf("failed to load activities: %w", err)
	}

	// Convert to emissions.Activity interface
	emissionsActivities := make([]emissions.Activity, 0, len(activities))
	for _, act := range activities {
		emissionsActivities = append(emissionsActivities, act)
	}

	// Calculate all scopes for compliance
	scope1Records, _ := r.scope1Calculator.CalculateBatch(ctx, emissionsActivities)
	scope2Records, _ := r.scope2Calculator.CalculateBatch(ctx, emissionsActivities)
	scope3Records, _ := r.scope3Calculator.CalculateBatch(ctx, emissionsActivities)

	// Aggregate totals
	var scope1Total, scope2Total, scope3Total float64
	for _, rec := range scope1Records {
		scope1Total += rec.EmissionsTonnesCO2e
	}
	for _, rec := range scope2Records {
		scope2Total += rec.EmissionsTonnesCO2e
	}
	for _, rec := range scope3Records {
		scope3Total += rec.EmissionsTonnesCO2e
	}

	// Determine compliance status based on data completeness
	status := "not_started"
	score := 0.0

	if scope2Total > 0 {
		status = "in_progress"
		score = 33.3 // Has Scope 2
	}
	if scope1Total > 0 {
		score = 66.6 // Has Scope 1 and 2
	}
	if scope3Total > 0 {
		status = "complete"
		score = 100.0 // Has all scopes
	}

	// Define required metrics based on framework
	requiredMetrics := []string{"scope1_emissions", "scope2_emissions", "scope3_emissions"}
	completedMetrics := []string{}
	missingMetrics := []string{}

	if scope1Total > 0 {
		completedMetrics = append(completedMetrics, "scope1_emissions")
	} else {
		missingMetrics = append(missingMetrics, "scope1_emissions")
	}

	if scope2Total > 0 {
		completedMetrics = append(completedMetrics, "scope2_emissions")
	} else {
		missingMetrics = append(missingMetrics, "scope2_emissions")
	}

	if scope3Total > 0 {
		completedMetrics = append(completedMetrics, "scope3_emissions")
	} else {
		missingMetrics = append(missingMetrics, "scope3_emissions")
	}

	return &ComplianceStatus{
		Framework:        framework,
		Status:           status,
		RequiredMetrics:  requiredMetrics,
		CompletedMetrics: completedMetrics,
		MissingMetrics:   missingMetrics,
		Score:            score,
		LastUpdated:      time.Now(),
	}, nil
}
