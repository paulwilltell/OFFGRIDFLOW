// Package http provides HTTP routing and handler configuration for the OffGridFlow API.
//
// The router supports multiple operational modes:
//   - Authenticated mode: Full authentication with session management
//   - Demo mode: Limited functionality without authentication
//   - Off-grid mode: Graceful degradation when external services unavailable
//
// Route Structure:
//
//	/health                       - Health check (public)
//	/livez                        - Kubernetes liveness probe (public)
//	/readyz                       - Kubernetes readiness probe (public)
//	/api/offgrid/mode             - Off-grid mode status (public)
//	/api/auth/*                   - Authentication endpoints
//	/api/billing/*                - Subscription management
//	/api/ai/chat                  - AI assistant (requires subscription)
//	/api/emissions/*              - Emissions data and calculations
//	/api/compliance/*             - Regulatory compliance reports
package http

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/example/offgridflow/internal/ai"
	"github.com/example/offgridflow/internal/api/http/handlers"
	"github.com/example/offgridflow/internal/api/http/middleware"
	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/auth"
	"github.com/example/offgridflow/internal/billing"
	"github.com/example/offgridflow/internal/compliance"
	"github.com/example/offgridflow/internal/compliance/csrd"
	"github.com/example/offgridflow/internal/connectors"
	"github.com/example/offgridflow/internal/db"
	"github.com/example/offgridflow/internal/emissions"
	"github.com/example/offgridflow/internal/emissions/factors"
	"github.com/example/offgridflow/internal/ingestion"
	"github.com/example/offgridflow/internal/ingestion/sources/utility_bills"
	"github.com/example/offgridflow/internal/offgrid"
	"github.com/example/offgridflow/internal/workflow"
)

// -----------------------------------------------------------------------------
// Configuration Types
// -----------------------------------------------------------------------------

// RouterConfig holds configuration for the HTTP router.
type RouterConfig struct {
	// Service dependencies
	ModeManager           *offgrid.ModeManager
	AIRouter              *ai.Router
	ActivityStore         ingestion.ActivityStore
	Scope1Calculator      *emissions.Scope1Calculator
	Scope2Calculator      *emissions.Scope2Calculator
	Scope3Calculator      *emissions.Scope3Calculator
	FactorRegistry        emissions.FactorRegistry
	CSRDMapper            csrd.CSRDMapper
	IngestionLogs         ingestion.LogStore
	IngestionSvc          *ingestion.Service
	ConnectorStore        connectors.Store
	UtilityBillsStore     ingestion.ActivityStore
	DB                    *db.DB
	WorkflowService       *workflow.Service
	IngestionOrchestrator *ingestion.Orchestrator
	IngestionScheduler    *ingestion.Scheduler

	// Auth configuration
	AuthStore      auth.Store
	SessionManager *auth.SessionManager
	RequireAuth    bool
	CookieDomain   string
	CookieSecure   bool

	// LockoutManager enforces login throttling for auth handlers.
	LockoutManager *auth.LockoutManager

	// CSRFMiddleware wraps routes to enforce CSRF protection.
	CSRFMiddleware *middleware.CSRFMiddleware

	// Billing configuration
	BillingService *billing.Service

	// Logging
	Logger *slog.Logger

	// CORS configuration
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string

	// ObservabilityMiddleware optionally wraps handlers with tracing/metrics.
	// If nil, no observability middleware is applied.
	ObservabilityMiddleware func(http.Handler) http.Handler
}

// RouterDeps holds dependencies for HTTP handlers (legacy interface for compatibility).
type RouterDeps = RouterConfig

type readyzDependencies struct {
	DB               *db.DB
	ModeManager      *offgrid.ModeManager
	BillingService   *billing.Service
	Scope1Calculator *emissions.Scope1Calculator
	Scope2Calculator *emissions.Scope2Calculator
	Scope3Calculator *emissions.Scope3Calculator
}

// -----------------------------------------------------------------------------
// Router Factory Functions
// -----------------------------------------------------------------------------

// NewRouter creates a new HTTP router with minimal dependencies.
// This is suitable for demo/development mode without authentication.
func NewRouter(mm *offgrid.ModeManager, aiRouter *ai.Router, activityStore ingestion.ActivityStore, scope2Calc *emissions.Scope2Calculator) http.Handler {
	if mm == nil {
		mm = offgrid.NewModeManager(offgrid.ModeOffline)
	}
	if activityStore == nil {
		mem := ingestion.NewInMemoryActivityStore()
		mem.SeedDemoData()
		activityStore = mem
	}
	if scope2Calc == nil {
		scope2Calc = emissions.NewScope2Calculator(emissions.Scope2Config{})
	}
	if aiRouter == nil {
		if fallback, err := ai.NewRouter(ai.RouterConfig{
			ModeManager: mm,
			Local:       &ai.SimpleLocalProvider{},
		}); err == nil {
			aiRouter = fallback
		}
	}

	scope1Calc := emissions.NewScope1Calculator(emissions.Scope1Config{})
	scope3Calc := emissions.NewScope3Calculator(emissions.Scope3Config{})
	logStore := ingestion.NewInMemoryLogStore()
	factorRegistry := factors.NewInMemoryRegistry(factors.DefaultRegistryConfig())
	ingestionSvc := &ingestion.Service{
		Adapters: []ingestion.SourceIngestionAdapter{},
		Store:    activityStore,
		Logs:     logStore,
	}
	authStore := auth.NewInMemoryStore()
	sessionManager, err := auth.NewSessionManager("dev-router-secret")
	if err != nil {
		panic(fmt.Sprintf("failed to create session manager: %v", err))
	}

	cfg := &RouterConfig{
		ModeManager:      mm,
		AIRouter:         aiRouter,
		ActivityStore:    activityStore,
		Scope1Calculator: scope1Calc,
		Scope2Calculator: scope2Calc,
		Scope3Calculator: scope3Calc,
		FactorRegistry:   factorRegistry,
		CSRDMapper:       csrd.NewDefaultCSRDMapper(),
		IngestionLogs:    logStore,
		IngestionSvc:     ingestionSvc,
		AuthStore:        authStore,
		SessionManager:   sessionManager,
		RequireAuth:      false,
		WorkflowService:  workflow.NewService(nil, nil),
	}
	return NewRouterWithConfig(cfg)
}

// NewRouterWithAuth creates a new HTTP router with authentication enabled.
func NewRouterWithAuth(deps *RouterConfig) http.Handler {
	return NewRouterWithConfig(deps)
}

// NewRouterWithDeps creates a new HTTP router with full dependency injection.
// This is an alias for NewRouterWithConfig for backward compatibility.
func NewRouterWithDeps(deps *RouterConfig) http.Handler {
	return NewRouterWithConfig(deps)
}

// NewRouterWithConfig creates a new HTTP router with full configuration.
func NewRouterWithConfig(cfg *RouterConfig) http.Handler {
	r := &router{
		cfg:    cfg,
		logger: cfg.Logger,
	}

	if r.logger == nil {
		r.logger = slog.Default().With("component", "http-router")
	}

	if cfg.LockoutManager == nil {
		cfg.LockoutManager = auth.NewLockoutManager(5, 15*time.Minute, 5*time.Minute, 5*time.Minute)
	}
	if cfg.CSRFMiddleware == nil {
		cfg.CSRFMiddleware = middleware.NewCSRFMiddleware(middleware.CSRFMiddlewareConfig{
			TokenTTL:        24 * time.Hour,
			CleanupInterval: 10 * time.Minute,
			ExemptPaths: []string{
				"/api/auth/login",
				"/api/auth/register",
				"/api/auth/password/forgot",
				"/api/auth/password/reset",
				"/api/auth/csrf-token",
				"/api/billing/webhook",
			},
		})
	}

	handlerDeps, err := cfg.buildHandlerDependencies()
	if err != nil {
		panic(fmt.Sprintf("invalid router dependencies: %v", err))
	}

	r.deps = handlerDeps
	r.readyz = readyzDependencies{
		DB:               cfg.DB,
		ModeManager:      cfg.ModeManager,
		BillingService:   cfg.BillingService,
		Scope1Calculator: cfg.Scope1Calculator,
		Scope2Calculator: cfg.Scope2Calculator,
		Scope3Calculator: cfg.Scope3Calculator,
	}
	if cfg.AuthStore != nil {
		r.authSvc = auth.NewService(cfg.AuthStore, nil)
	}

	return r.build()
}

// -----------------------------------------------------------------------------
// Router Implementation
// -----------------------------------------------------------------------------

// router encapsulates route registration logic.
type router struct {
	cfg     *RouterConfig
	logger  *slog.Logger
	deps    *HandlerDependencies
	readyz  readyzDependencies
	authSvc *auth.Service
}

// build constructs the HTTP handler tree.
func (r *router) build() http.Handler {
	mux := http.NewServeMux()

	// Register public routes
	r.registerPublicRoutes(mux)

	// Register protected routes
	r.registerProtectedRoutes(mux)

	handler := http.Handler(mux)
	if r.cfg != nil && r.cfg.CSRFMiddleware != nil {
		handler = r.cfg.CSRFMiddleware.Wrap(handler)
	}

	return handler
}

// registerPublicRoutes adds routes that don't require authentication.
func (r *router) registerPublicRoutes(mux *http.ServeMux) {
	// Health and liveness probes (Kubernetes compatible)
	mux.HandleFunc("/health", r.healthHandler)
	mux.HandleFunc("/livez", handlers.LivezHandler)
	mux.HandleFunc("/readyz", r.readyzHandler())

	// Off-grid mode status (system diagnostics)
	mux.Handle("/api/offgrid/mode", r.offgridModeHandler())

	// Public authentication endpoints
	if r.cfg.AuthStore != nil && r.cfg.SessionManager != nil {
		authHandlers := handlers.NewAuthHandlers(handlers.AuthHandlersConfig{
			AuthStore:      r.cfg.AuthStore,
			SessionManager: r.cfg.SessionManager,
			CookieDomain:   r.cfg.CookieDomain,
			CookieSecure:   r.cfg.CookieSecure,
			LockoutManager: r.cfg.LockoutManager,
		})

		mux.HandleFunc("/api/auth/register", authHandlers.Register)
		mux.HandleFunc("/api/auth/login", authHandlers.Login)
		if r.cfg.CSRFMiddleware != nil {
			mux.HandleFunc("/api/auth/csrf-token", r.csrfTokenHandler())
		}
		mux.HandleFunc("/api/auth/logout", authHandlers.Logout)
		if r.authSvc != nil {
			mux.HandleFunc("/api/auth/password/forgot", handlers.NewPasswordForgotHandler(r.authSvc, r.logger))
			mux.HandleFunc("/api/auth/password/reset", handlers.NewPasswordResetHandler(r.authSvc, r.logger))
		}
	}

	// Billing webhook (receives Stripe webhooks)
	if r.cfg.BillingService != nil {
		billingHandlers := NewBillingHandlers(r.cfg.BillingService)
		mux.HandleFunc("/api/billing/webhook", billingHandlers.HandleWebhook)
	}
}

// registerProtectedRoutes adds routes that require authentication.
func (r *router) registerProtectedRoutes(mux *http.ServeMux) {
	protectedMux := http.NewServeMux()

	// Auth management endpoints
	if r.cfg.AuthStore != nil && r.cfg.SessionManager != nil {
		authHandlers := handlers.NewAuthHandlers(handlers.AuthHandlersConfig{
			AuthStore:      r.cfg.AuthStore,
			SessionManager: r.cfg.SessionManager,
			CookieDomain:   r.cfg.CookieDomain,
			CookieSecure:   r.cfg.CookieSecure,
			LockoutManager: r.cfg.LockoutManager,
		})

		protectedMux.HandleFunc("/api/auth/me", authHandlers.Me)
		protectedMux.HandleFunc("/api/auth/change-password", authHandlers.ChangePassword)

		// API key management
		authService := auth.NewService(r.cfg.AuthStore, nil)
		protectedMux.HandleFunc("/api/auth/keys", handlers.NewAPIKeyHandler(authService))
	}

	// Billing management endpoints
	if r.cfg.BillingService != nil {
		billingHandlers := NewBillingHandlers(r.cfg.BillingService)
		protectedMux.HandleFunc("/api/billing/checkout", billingHandlers.CreateCheckoutSession)
		protectedMux.HandleFunc("/api/billing/status", billingHandlers.GetStatus)
		protectedMux.HandleFunc("/api/billing/portal", billingHandlers.CreatePortalSession)
	}

	// AI chat endpoint (subscription required)
	protectedMux.Handle("/api/ai/chat", r.aiChatHandler())

	// Emissions endpoints
	if scope2Handler := r.buildScope2Handler(); scope2Handler != nil {
		protectedMux.HandleFunc("/api/emissions/scope2", scope2Handler.List)
		protectedMux.HandleFunc("/api/emissions/scope2/summary", scope2Handler.Summary)
	}

	// Ingestion status endpoints
	if ingestionLogsHandler := r.buildIngestionStatusHandler(); ingestionLogsHandler != nil {
		protectedMux.HandleFunc("/api/ingestion/logs", ingestionLogsHandler)
	}

	// CSV ingestion upload endpoint
	if csvHandler := r.buildCSVIngestionHandler(); csvHandler != nil {
		protectedMux.HandleFunc("/api/ingestion/upload/csv", csvHandler)
		protectedMux.HandleFunc("/api/ingestion/csv", csvHandler) // alias
	}

	// Utility Bills ingestion endpoints (multi-format: CSV, JSON, PDF, Excel)
	if cfg := r.buildUtilityBillsHandlerConfig(); cfg != nil {
		protectedMux.HandleFunc("/api/ingestion/utility-bills/upload", handlers.NewUtilityBillUploadHandler(*cfg))
		protectedMux.HandleFunc("/api/ingestion/utility-bills/batch-upload", handlers.NewUtilityBillBatchUploadHandler(*cfg))
		protectedMux.HandleFunc("/api/ingestion/utility-bills", handlers.NewUtilityBillListHandler(*cfg))
		protectedMux.HandleFunc("/api/ingestion/utility-bills/list", handlers.NewUtilityBillListHandler(*cfg)) // alias
	}

	// Connectors endpoints (backend wiring for UI)
	if connectorsHandler := r.buildConnectorsHandler(); connectorsHandler != nil {
		protectedMux.Handle("/api/connectors/", connectorsHandler)
	}

	// Factors endpoints
	if registry := r.cfg.FactorRegistry; registry != nil {
		protectedMux.HandleFunc("/api/factors", handlers.NewFactorsHandler(handlers.FactorsHandlerConfig{
			Registry: registry,
		}))
	}

	// Workflow tasks endpoints
	if workflowCfg := r.buildWorkflowHandlerConfig(); workflowCfg != nil {
		protectedMux.HandleFunc("/api/workflow/tasks", handlers.NewWorkflowHandler(*workflowCfg))
	}

	// Users endpoints (Phase 2 start for user management)
	if r.cfg.AuthStore != nil {
		usersHandler := handlers.NewUsersHandler(handlers.UsersHandlerConfig{
			Store: r.cfg.AuthStore,
		})
		protectedMux.Handle("/api/users", usersHandler)
	}

	// Compliance endpoints
	if complianceDeps := r.buildComplianceHandlerDeps(); complianceDeps != nil {
		protectedMux.HandleFunc("/api/compliance/csrd", handlers.NewCSRDComplianceHandler(complianceDeps))
		protectedMux.HandleFunc("/api/compliance/sec", handlers.NewSECComplianceHandler(complianceDeps))
		protectedMux.HandleFunc("/api/compliance/california", handlers.NewCaliforniaComplianceHandler(complianceDeps))
		protectedMux.HandleFunc("/api/compliance/cbam", handlers.NewCBAMComplianceHandler(complianceDeps))
		protectedMux.HandleFunc("/api/compliance/ifrs", handlers.NewIFRSComplianceHandler(complianceDeps))
		protectedMux.HandleFunc("/api/compliance/summary", handlers.NewComplianceSummaryHandler(complianceDeps))
		protectedMux.HandleFunc("/api/compliance/export", handlers.NewComplianceExportHandler(complianceDeps))
	}

	// GraphQL endpoint
	graphqlHandler := handlers.NewGraphQLHandler(handlers.GraphQLHandlerConfig{
		ActivityStore:    r.cfg.ActivityStore,
		Scope1Calculator: r.cfg.Scope1Calculator,
		Scope2Calculator: r.cfg.Scope2Calculator,
		Scope3Calculator: r.cfg.Scope3Calculator,
		CSRDMapper:       r.cfg.CSRDMapper,
		Logger:           r.logger,
	})
	protectedMux.HandleFunc("/api/graphql", graphqlHandler)
	protectedMux.HandleFunc("/graphql", graphqlHandler) // Alias for convenience

	// Apply middleware chain to protected routes
	protectedHandler := r.applyProtectedMiddleware(protectedMux)

	// Mount protected routes to main mux
	r.mountProtectedRoutes(mux, protectedHandler)
}

// applyProtectedMiddleware wraps the protected handler with auth and subscription middleware.
func (r *router) applyProtectedMiddleware(handler http.Handler) http.Handler {
	result := handler

	// Observability middleware (tracing/metrics)
	if mw := r.buildObservabilityMiddleware(); mw != nil {
		result = mw(result)
	}

	// Apply subscription enforcement for premium features
	if r.cfg.BillingService != nil {
		subscriptionMiddleware := middleware.NewSubscriptionMiddleware(middleware.SubscriptionMiddlewareConfig{
			BillingService: r.cfg.BillingService,
			FreePaths: []string{
				"/api/auth/me",
				"/api/auth/change-password",
				"/api/auth/keys",
				"/api/billing/checkout",
				"/api/billing/status",
				"/api/billing/portal",
			},
		})
		result = subscriptionMiddleware.Wrap(result)
	}

	// Apply authentication middleware
	if r.cfg.AuthStore != nil {
		authMiddleware := middleware.NewAuthMiddleware(middleware.AuthMiddlewareConfig{
			AuthStore:      r.cfg.AuthStore,
			SessionManager: r.cfg.SessionManager,
			RequireAuth:    true,
			AllowedPaths: []string{
				"/health",
				"/livez",
				"/readyz",
				"/api/auth/",
				"/api/billing/webhook",
				"/api/offgrid/mode",
			}, // All other /api paths require auth
		})
		result = authMiddleware.Wrap(result)
	}

	return result
}

// buildObservabilityMiddleware wires OpenTelemetry HTTP middleware if tracer/meter configured.
func (r *router) buildObservabilityMiddleware() func(http.Handler) http.Handler {
	if r.cfg != nil && r.cfg.ObservabilityMiddleware != nil {
		return r.cfg.ObservabilityMiddleware
	}
	return nil
}

// mountProtectedRoutes registers protected routes on the main mux.
func (r *router) mountProtectedRoutes(mux *http.ServeMux, handler http.Handler) {
	mux.Handle("/api/", handler)
}

// -----------------------------------------------------------------------------
// Handler Factories
// -----------------------------------------------------------------------------

// offgridModeHandler creates a handler for the off-grid mode endpoint.
func (r *router) offgridModeHandler() http.Handler {
	cfg := r.cfg
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if cfg.ModeManager == nil {
			responders.JSON(w, http.StatusOK, map[string]string{
				"mode": "normal",
			})
			return
		}

		mode := cfg.ModeManager.GetMode()
		responders.JSON(w, http.StatusOK, map[string]string{
			"mode": string(mode),
		})
	})
}

func (r *router) csrfTokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			responders.MethodNotAllowed(w, http.MethodGet)
			return
		}

		if r.cfg == nil || r.cfg.CSRFMiddleware == nil {
			responders.InternalError(w, "CSRF middleware not configured")
			return
		}

		token, err := r.cfg.CSRFMiddleware.GenerateToken()
		if err != nil {
			r.logger.Error("failed to generate csrf token", "error", err)
			responders.InternalError(w, "failed to generate CSRF token")
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     r.cfg.CSRFMiddleware.CookieName(),
			Value:    token,
			Path:     "/",
			Domain:   r.cfg.CookieDomain,
			HttpOnly: true,
			Secure:   r.cfg.CookieSecure,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   int(r.cfg.CSRFMiddleware.TokenTTL().Seconds()),
		})

		responders.JSON(w, http.StatusOK, map[string]string{
			"csrf_token": token,
		})
	}
}

// aiChatHandler creates a handler for the AI chat endpoint.
func (r *router) aiChatHandler() http.Handler {
	cfg := r.cfg
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			responders.MethodNotAllowed(w, http.MethodPost)
			return
		}

		if cfg.AIRouter == nil {
			responders.Error(w, http.StatusServiceUnavailable, "ai_unavailable", "AI service not configured")
			return
		}

		var chatReq ai.ChatRequest
		if err := json.NewDecoder(req.Body).Decode(&chatReq); err != nil {
			responders.BadRequest(w, "invalid_request", "invalid JSON payload")
			return
		}

		resp, err := cfg.AIRouter.Chat(req.Context(), chatReq)
		if err != nil {
			r.logger.Error("AI chat failed",
				"error", err.Error(),
			)
			responders.InternalError(w, "AI chat request failed")
			return
		}

		responders.JSON(w, http.StatusOK, resp)
	})
}

// healthHandler is a simple health check endpoint.
func (r *router) healthHandler(w http.ResponseWriter, req *http.Request) {
	responders.JSON(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "offgridflow-api",
	})
}

// readyzHandler validates dependencies that the app needs to serve traffic.
func (r *router) readyzHandler() http.HandlerFunc {
	deps := r.readyz
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		checks := map[string]string{}
		status := "ready"
		httpStatus := http.StatusOK

		if deps.DB != nil {
			if err := deps.DB.HealthCheck(ctx); err != nil {
				status = "degraded"
				httpStatus = http.StatusServiceUnavailable
				checks["database"] = "unreachable: " + err.Error()
			} else {
				checks["database"] = "ok"
			}
		} else {
			checks["database"] = "not_configured"
		}

		if deps.BillingService != nil {
			if err := deps.BillingService.Ready(); err != nil {
				status = "degraded"
				httpStatus = http.StatusServiceUnavailable
				checks["stripe"] = "not_ready: " + err.Error()
			} else {
				checks["stripe"] = "ok"
			}
		} else {
			checks["stripe"] = "disabled"
		}

		if deps.ModeManager == nil {
			status = "degraded"
			checks["offgrid"] = "unknown"
		} else {
			mode := deps.ModeManager.GetMode()
			checks["offgrid"] = string(mode)
			if mode.IsOffline() {
				status = "degraded"
				httpStatus = http.StatusServiceUnavailable
			}
		}

		if deps.Scope1Calculator == nil && deps.Scope2Calculator == nil && deps.Scope3Calculator == nil {
			status = "degraded"
			httpStatus = http.StatusServiceUnavailable
			checks["emissions"] = "no calculators configured"
		} else {
			checks["emissions"] = "calculators_available"
		}

		responders.JSON(w, httpStatus, map[string]any{
			"status":    status,
			"checks":    checks,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func (r *router) buildScope2Handler() *handlers.Scope2Handler {
	if r.deps == nil || r.deps.Scope2 == nil {
		return nil
	}
	return handlers.NewScope2HandlerWithDeps(
		&handlers.Scope2HandlerDeps{
			ActivityStore:    r.deps.Scope2.ActivityStore,
			Scope2Calculator: r.deps.Scope2.Scope2Calculator,
		},
		r.deps.Scope2.Logger,
	)
}

func (r *router) buildIngestionStatusHandler() http.HandlerFunc {
	if r.deps == nil || r.deps.IngestionStatus == nil {
		return nil
	}
	return handlers.NewIngestionStatusHandler(handlers.IngestionHandlerConfig{
		LogStore: r.deps.IngestionStatus.LogStore,
	})
}

func (r *router) buildCSVIngestionHandler() http.HandlerFunc {
	if r.cfg == nil || r.cfg.ActivityStore == nil {
		return nil
	}
	return handlers.NewCSVIngestionHandler(handlers.CSVIngestionHandlerConfig{
		Store: r.cfg.ActivityStore,
	})
}

func (r *router) buildUtilityBillsHandlerConfig() *handlers.UtilityBillsHandlerConfig {
	if r.cfg == nil || r.cfg.UtilityBillsStore == nil {
		return nil
	}

	config := utility_bills.DefaultConfig("")
	config.Store = r.cfg.UtilityBillsStore
	if r.deps != nil && r.deps.UtilityBills != nil && r.deps.UtilityBills.Logger != nil {
		config.Logger = r.deps.UtilityBills.Logger
	}
	adapter := utility_bills.NewAdapter(config)
	cfg := handlers.DefaultUtilityBillsConfig(adapter, r.cfg.UtilityBillsStore)
	if r.deps != nil && r.deps.UtilityBills != nil && r.deps.UtilityBills.Logger != nil {
		cfg.Logger = r.deps.UtilityBills.Logger
	}
	return &cfg
}

func (r *router) buildConnectorsHandler() http.Handler {
	if r.cfg == nil {
		return nil
	}
	return handlers.NewConnectorsHandler(handlers.ConnectorsHandlerConfig{
		IngestionSvc:   r.cfg.IngestionSvc,
		ConnectorStore: r.cfg.ConnectorStore,
		Orchestrator:   r.cfg.IngestionOrchestrator,
		Scheduler:      r.cfg.IngestionScheduler,
	})
}

func (r *router) buildWorkflowHandlerConfig() *handlers.WorkflowHandlerConfig {
	if r.cfg == nil || r.cfg.WorkflowService == nil {
		return nil
	}
	return &handlers.WorkflowHandlerConfig{
		Service: r.cfg.WorkflowService,
	}
}

func (r *router) buildComplianceHandlerDeps() *handlers.ComplianceHandlerDeps {
	if r.cfg == nil || r.cfg.ActivityStore == nil {
		return nil
	}

	return &handlers.ComplianceHandlerDeps{
		ComplianceService: compliance.NewService(
			r.cfg.ActivityStore,
			r.cfg.Scope1Calculator,
			r.cfg.Scope2Calculator,
			r.cfg.Scope3Calculator,
		),
	}
}

// -----------------------------------------------------------------------------
// Legacy API (backward compatibility)
// -----------------------------------------------------------------------------

// Router builds HTTP mux for the API server (legacy version - kept for compatibility).
func Router(handlers map[string]http.Handler) http.Handler {
	mux := http.NewServeMux()
	for path, handler := range handlers {
		mux.Handle(path, handler)
	}
	return mux
}
