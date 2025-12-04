// Package http provides handler dependency injection types.
package http

import (
	"fmt"
	"log/slog"

	"github.com/example/offgridflow/internal/api/http/handlers"
	"github.com/example/offgridflow/internal/compliance/csrd"
	"github.com/example/offgridflow/internal/emissions"
	"github.com/example/offgridflow/internal/ingestion"
	"github.com/example/offgridflow/internal/workflow"
)

// =============================================================================
// Handler Dependencies
// =============================================================================

// HandlerDependencies bundles all dependencies needed by HTTP handlers.
// This allows clean dependency injection across all handler types.
type HandlerDependencies struct {
	// Emissions-related dependencies
	Scope2          *Scope2HandlerDeps
	Compliance      *ComplianceHandlerDeps
	CSVIngestion    *CSVIngestionHandlerDeps
	UtilityBills    *UtilityBillsHandlerDeps
	Factors         *FactorsHandlerDeps
	Connectors      *ConnectorsHandlerDeps
	Workflow        *WorkflowHandlerDeps
	IngestionStatus *IngestionStatusHandlerDeps

	// Logger for all handlers
	Logger *slog.Logger
}

// Scope2HandlerDeps holds dependencies for Scope 2 handler
type Scope2HandlerDeps struct {
	ActivityStore    ingestion.ActivityStore
	Scope2Calculator *emissions.Scope2Calculator
	Logger           *slog.Logger
}

// ComplianceHandlerDeps holds dependencies for compliance handlers
type ComplianceHandlerDeps struct {
	ActivityStore    ingestion.ActivityStore
	Scope1Calculator *emissions.Scope1Calculator
	Scope2Calculator *emissions.Scope2Calculator
	Scope3Calculator *emissions.Scope3Calculator
	FactorRegistry   emissions.FactorRegistry
	CSRDMapper       csrd.CSRDMapper
	Logger           *slog.Logger
}

// CSVIngestionHandlerDeps holds dependencies for CSV ingestion handler
type CSVIngestionHandlerDeps struct {
	ActivityStore ingestion.ActivityStore
	LogStore      ingestion.LogStore
	Logger        *slog.Logger
}

// UtilityBillsHandlerDeps holds dependencies for utility bills handler
type UtilityBillsHandlerDeps struct {
	ActivityStore ingestion.ActivityStore
	LogStore      ingestion.LogStore
	Logger        *slog.Logger
}

// FactorsHandlerDeps holds dependencies for factors handler
type FactorsHandlerDeps struct {
	Registry emissions.FactorRegistry
	Logger   *slog.Logger
}

// ConnectorsHandlerDeps holds dependencies for connectors handler
type ConnectorsHandlerDeps struct {
	Store  interface{} // Will be connectors.Store, but avoid import cycle
	Logger *slog.Logger
}

// WorkflowHandlerDeps holds dependencies for workflow handler
type WorkflowHandlerDeps struct {
	Service *workflow.Service
	Logger  *slog.Logger
}

// IngestionStatusHandlerDeps holds dependencies for ingestion status handler
type IngestionStatusHandlerDeps struct {
	LogStore ingestion.LogStore
	Logger   *slog.Logger
}

// =============================================================================
// Router Config Methods
// =============================================================================

// buildHandlerDependencies constructs the HandlerDependencies from RouterConfig.
// This wires all dependencies needed by HTTP handlers.
func (rc *RouterConfig) buildHandlerDependencies() (*HandlerDependencies, error) {
	logger := rc.Logger
	if logger == nil {
		logger = slog.Default().With("component", "handlers")
	}

	// Validate critical dependencies
	if rc.ActivityStore == nil {
		return nil, fmt.Errorf("ActivityStore is required")
	}
	if rc.FactorRegistry == nil {
		return nil, fmt.Errorf("FactorRegistry is required")
	}

	deps := &HandlerDependencies{
		Logger: logger,

		// Scope 2 handler dependencies
		Scope2: &Scope2HandlerDeps{
			ActivityStore:    rc.ActivityStore,
			Scope2Calculator: rc.Scope2Calculator,
			Logger:           logger,
		},

		// Compliance handler dependencies
		Compliance: &ComplianceHandlerDeps{
			ActivityStore:    rc.ActivityStore,
			Scope1Calculator: rc.Scope1Calculator,
			Scope2Calculator: rc.Scope2Calculator,
			Scope3Calculator: rc.Scope3Calculator,
			FactorRegistry:   rc.FactorRegistry,
			CSRDMapper:       rc.CSRDMapper,
			Logger:           logger,
		},

		// CSV ingestion handler dependencies
		CSVIngestion: &CSVIngestionHandlerDeps{
			ActivityStore: rc.ActivityStore,
			LogStore:      rc.IngestionLogs,
			Logger:        logger,
		},

		// Utility bills handler dependencies
		UtilityBills: &UtilityBillsHandlerDeps{
			ActivityStore: rc.UtilityBillsStore,
			LogStore:      rc.IngestionLogs,
			Logger:        logger,
		},

		// Factors handler dependencies
		Factors: &FactorsHandlerDeps{
			Registry: rc.FactorRegistry,
			Logger:   logger,
		},

		// Connectors handler dependencies
		Connectors: &ConnectorsHandlerDeps{
			Store:  rc.ConnectorStore,
			Logger: logger,
		},

		// Workflow handler dependencies
		Workflow: &WorkflowHandlerDeps{
			Service: rc.WorkflowService,
			Logger:  logger,
		},

		// Ingestion status handler dependencies
		IngestionStatus: &IngestionStatusHandlerDeps{
			LogStore: rc.IngestionLogs,
			Logger:   logger,
		},
	}

	return deps, nil
}

// =============================================================================
// Handler Factory Functions
// =============================================================================

// NewScope2Handler creates a Scope2 handler from dependencies.
func NewScope2HandlerFromDeps(deps *Scope2HandlerDeps) *handlers.Scope2Handler {
	return handlers.NewScope2HandlerWithDeps(
		&handlers.Scope2HandlerDeps{
			ActivityStore:    deps.ActivityStore,
			Scope2Calculator: deps.Scope2Calculator,
		},
		deps.Logger,
	)
}

// Note: Add more handler factory functions as needed for other handler types
