package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/example/offgridflow/internal/ai"
	apihttp "github.com/example/offgridflow/internal/api/http"
	"github.com/example/offgridflow/internal/auth"
	"github.com/example/offgridflow/internal/billing"
	"github.com/example/offgridflow/internal/config"
	"github.com/example/offgridflow/internal/connectors"
	"github.com/example/offgridflow/internal/db"
	"github.com/example/offgridflow/internal/emissions"
	"github.com/example/offgridflow/internal/emissions/factors"
	"github.com/example/offgridflow/internal/ingestion"
	"github.com/example/offgridflow/internal/ingestion/sources/aws"
	"github.com/example/offgridflow/internal/ingestion/sources/azure"
	"github.com/example/offgridflow/internal/ingestion/sources/gcp"
	"github.com/example/offgridflow/internal/ingestion/sources/sap"
	"github.com/example/offgridflow/internal/offgrid"
	"github.com/example/offgridflow/internal/secrets"
	"github.com/example/offgridflow/internal/tracing"
	"github.com/example/offgridflow/internal/workflow"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("[offgridflow] fatal error: %v", err)
	}
}

func run() (err error) {
	// Set up JSON structured logging
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true, // Include file:line in logs
	})
	logger := slog.New(jsonHandler)
	slog.SetDefault(logger)

	// Recover from any panics so we can log the stack trace for debugging.
	defer func() {
		if r := recover(); r != nil {
			logger.Error("PANIC", "error", r, "stack", string(debug.Stack()))
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config load: %w", err)
	}
	log.Printf("[offgridflow] booting api (env=%s port=%d)", cfg.Server.Env, cfg.Server.Port)

	// Create background context for the process
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	secretManager := secrets.NewEnvProvider()
	resolveSecret := func(explicit, key string) string {
		return secrets.Resolve(ctx, secretManager, explicit, key)
	}

	// Initialize tracing
	tracingEnabled := os.Getenv("OFFGRIDFLOW_TRACING_ENABLED") != "false" // Enabled by default
	traceProvider, err := tracing.Setup(tracing.Config{
		ServiceName:    "offgridflow-api",
		ServiceVersion: "1.0.0",
		Environment:    cfg.Server.Env,
		OTLPEndpoint:   os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"), // Defaults to http://localhost:4318
		SamplingRate:   1.0,
		Enabled:        tracingEnabled,
	})
	if err != nil {
		log.Printf("[offgridflow] WARNING: failed to setup tracing: %v", err)
	} else if tracingEnabled {
		defer func() {
			if err := traceProvider.Shutdown(ctx); err != nil {
				log.Printf("[offgridflow] WARNING: failed to shutdown tracing: %v", err)
			}
		}()
		log.Printf("[offgridflow] tracing enabled (endpoint: %s)", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	} // Try to connect to database if DSN is provided
	var database *db.DB
	if cfg.Database.DSN != "" {
		database, err = db.Connect(ctx, db.Config{DSN: cfg.Database.DSN})
		if err != nil {
			log.Printf("[offgridflow] WARNING: failed to connect to DB: %v, falling back to in-memory stores", err)
		} else {
			log.Printf("[offgridflow] connected to Postgres")
			// Run migrations
			if err := database.RunMigrations(ctx); err != nil {
				log.Printf("[offgridflow] WARNING: failed to run migrations: %v", err)
			}
		}
	} else {
		log.Printf("[offgridflow] no OFFGRIDFLOW_DB_DSN provided, using in-memory stores")
	}

	// 1. Create mode manager (start pessimistic as OFFLINE until proven otherwise)
	modeManager := offgrid.NewModeManager(offgrid.ModeOffline)
	modeManager.OnChange(func(oldMode, newMode offgrid.Mode) {
		log.Printf("[offgridflow] mode transition: %s -> %s", oldMode, newMode)
	})
	log.Printf("[offgridflow] mode manager initialized (mode=%s)", modeManager.GetMode())

	// 2. Start connectivity watcher in background
	watcher := offgrid.NewConnectivityWatcher(modeManager, offgrid.WatcherConfig{
		Interval: 5 * time.Second,
	})
	go watcher.Start(ctx)

	// 3. Build AI router with real OpenAI provider (if configured) or offline
	var cloud ai.CloudProvider
	if cfg.OpenAI.APIKey != "" {
		openaiProvider, err := ai.NewOpenAIProvider(ai.OpenAIProviderConfig{
			APIKey: cfg.OpenAI.APIKey,
			Model:  cfg.OpenAI.Model,
		})
		if err != nil {
			log.Printf("[offgridflow] WARNING: failed to create OpenAI provider: %v, using offline mode", err)
			cloud = nil
		} else {
			cloud = openaiProvider
			log.Printf("[offgridflow] using OpenAI cloud provider (model: %s)", cfg.OpenAI.Model)
		}
	} else {
		log.Printf("[offgridflow] no OPENAI_API_KEY provided, using offline AI mode")
	}

	// 3b. Set up local AI provider - try Ollama first, fall back to simple provider
	var local ai.LocalProvider
	localAIProvider, err := ai.NewLocalOfflineProvider(ai.LocalOfflineConfig{
		BaseURL: os.Getenv("OFFGRIDFLOW_LOCAL_AI_URL"),   // Default: http://localhost:11434
		Model:   os.Getenv("OFFGRIDFLOW_LOCAL_AI_MODEL"), // Default: llama3.2:3b
		Timeout: 120 * time.Second,
	})
	if err == nil && localAIProvider.IsAvailable() {
		local = localAIProvider
		log.Printf("[offgridflow] local AI provider initialized (Ollama at %s)",
			firstNonEmpty(os.Getenv("OFFGRIDFLOW_LOCAL_AI_URL"), "http://localhost:11434"))
	} else {
		// Fall back to simple provider for basic offline responses
		local = &ai.SimpleLocalProvider{}
		if err != nil {
			log.Printf("[offgridflow] Ollama not available (%v), using simple offline provider", err)
		} else {
			log.Printf("[offgridflow] Ollama not available, using simple offline provider")
		}
	}

	aiRouter, err := ai.NewRouter(ai.RouterConfig{
		ModeManager:    modeManager,
		Cloud:          cloud,
		Local:          local,
		EnableFallback: ai.Ptr(true), // Enable fallback to local on cloud failure
		RetryConfig: &ai.RetryConfig{
			MaxAttempts:       3,
			InitialBackoff:    100 * time.Millisecond,
			MaxBackoff:        5 * time.Second,
			BackoffMultiplier: 2.0,
		},
	})
	if err != nil {
		return fmt.Errorf("ai router: %w", err)
	}

	// 4. Set up activity store (Postgres if available, otherwise in-memory)
	var activityStore ingestion.ActivityStore
	var ingestionLogs ingestion.LogStore
	var ingestionSvc *ingestion.Service
	adapters := make([]ingestion.SourceIngestionAdapter, 0, 5)
	var connectorStore connectors.Store
	if database != nil {
		activityStore = ingestion.NewPostgresActivityStore(database.DB)
		ingestionLogs = ingestion.NewPostgresLogStore(database.DB)
		connectorStore = connectors.NewPostgresStore(database.DB)
		log.Printf("[offgridflow] using PostgreSQL activity store")
	} else {
		memStore := ingestion.NewInMemoryActivityStore()
		memStore.SeedDemoData()
		activityStore = memStore
		ingestionLogs = ingestion.NewInMemoryLogStore()
		log.Printf("[offgridflow] using in-memory activity store with demo data")
	}

	// Build ingestion adapters for run-now/connectors UI (reuse worker config)
	start := time.Now().AddDate(0, 0, -cfg.Ingestion.LookbackDays)
	end := time.Now()
	defaultOrgID := firstNonEmpty(
		cfg.Ingestion.AWS.OrgID,
		cfg.Ingestion.Azure.OrgID,
		cfg.Ingestion.GCP.OrgID,
		cfg.Ingestion.SAP.OrgID,
		cfg.Ingestion.Utility.OrgID,
	)
	if defaultOrgID != "" {
		log.Printf("[offgridflow] default org/tenant context: %s", defaultOrgID)
	}
	if cfg.Ingestion.AWS.Enabled {
		if awsAdapter, err := aws.NewAdapter(aws.Config{
			AccessKeyID:     resolveSecret(cfg.Ingestion.AWS.AccessKeyID, secrets.AWSAccessKeyID),
			SecretAccessKey: resolveSecret(cfg.Ingestion.AWS.SecretAccessKey, secrets.AWSSecretAccessKey),
			Region:          cfg.Ingestion.AWS.Region,
			RoleARN:         cfg.Ingestion.AWS.RoleARN,
			AccountID:       cfg.Ingestion.AWS.AccountID,
			OrgID:           cfg.Ingestion.AWS.OrgID,
			StartDate:       start,
			EndDate:         end,
		}); err == nil {
			adapters = append(adapters, awsAdapter)
		} else {
			log.Printf("[offgridflow] aws adapter disabled: %v", err)
		}
	}

	if cfg.Ingestion.Azure.Enabled {
		if azureAdapter, err := azure.NewAdapter(azure.Config{
			TenantID:       cfg.Ingestion.Azure.TenantID,
			ClientID:       cfg.Ingestion.Azure.ClientID,
			ClientSecret:   resolveSecret(cfg.Ingestion.Azure.ClientSecret, secrets.AzureClientSecret),
			SubscriptionID: cfg.Ingestion.Azure.SubscriptionID,
			OrgID:          cfg.Ingestion.Azure.OrgID,
			StartDate:      start,
			EndDate:        end,
		}); err == nil {
			adapters = append(adapters, azureAdapter)
		} else {
			log.Printf("[offgridflow] azure adapter disabled: %v", err)
		}
	}

	if cfg.Ingestion.GCP.Enabled {
		if gcpAdapter, err := gcp.NewAdapter(gcp.Config{
			ProjectID:         cfg.Ingestion.GCP.ProjectID,
			BillingAccountID:  cfg.Ingestion.GCP.BillingAccountID,
			BigQueryDataset:   cfg.Ingestion.GCP.BigQueryDataset,
			BigQueryTable:     cfg.Ingestion.GCP.BigQueryTable,
			ServiceAccountKey: resolveSecret(cfg.Ingestion.GCP.ServiceAccountKey, secrets.GCPServiceAccountKey),
			OrgID:             cfg.Ingestion.GCP.OrgID,
			StartDate:         start,
			EndDate:           end,
		}); err == nil {
			adapters = append(adapters, gcpAdapter)
		} else {
			log.Printf("[offgridflow] gcp adapter disabled: %v", err)
		}
	}

	if cfg.Ingestion.SAP.Enabled {
		if sapAdapter, err := sap.NewAdapter(sap.Config{
			BaseURL:      cfg.Ingestion.SAP.BaseURL,
			ClientID:     cfg.Ingestion.SAP.ClientID,
			ClientSecret: resolveSecret(cfg.Ingestion.SAP.ClientSecret, secrets.SAPClientSecret),
			Company:      cfg.Ingestion.SAP.Company,
			Plant:        cfg.Ingestion.SAP.Plant,
			OrgID:        cfg.Ingestion.SAP.OrgID,
			StartDate:    start,
			EndDate:      end,
		}); err == nil {
			adapters = append(adapters, sapAdapter)
		} else {
			log.Printf("[offgridflow] sap adapter disabled: %v", err)
		}
	}

	ingestionSvc = &ingestion.Service{
		Adapters:       adapters,
		Store:          activityStore,
		Logger:         slog.Default(),
		Logs:           ingestionLogs,
		ConnectorStore: connectorStore,
		OrgID:          defaultOrgID,
	}

	orchestrator := &ingestion.Orchestrator{
		Service:        ingestionSvc,
		Attempts:       4,
		InitialBackoff: 3 * time.Second,
		Logger:         slog.Default(),
	}
	scheduler := &ingestion.Scheduler{
		Orchestrator: orchestrator,
		Interval:     cfg.Ingestion.ScheduleInterval,
		Logger:       slog.Default(),
	}
	if scheduler.Interval > 0 {
		go scheduler.Start(ctx)
	}

	// 5. Set up emission factor registry (Postgres if available, otherwise in-memory)
	var factorRegistry emissions.FactorRegistry
	if database != nil {
		factorRegistry = factors.NewPostgresRegistrySimple(database.DB)
		// Seed default factors if DB is empty
		if pgReg, ok := factorRegistry.(*factors.PostgresRegistry); ok {
			if err := pgReg.SeedDefaultFactors(ctx); err != nil {
				log.Printf("[offgridflow] WARNING: failed to seed emission factors: %v", err)
			}
		}
		log.Printf("[offgridflow] using PostgreSQL emission factor registry")
	} else {
		factorRegistry = factors.NewInMemoryRegistry(factors.RegistryConfig{
			PreloadDefaults:    true,
			ValidateOnRegister: true,
		})
		log.Printf("[offgridflow] using in-memory emission factor registry")
	}

	// 6. Set up Scope 2 calculator
	scope2Calculator := emissions.NewScope2Calculator(emissions.Scope2Config{
		Registry: factorRegistry,
	})
	scope1Calculator := emissions.NewScope1Calculator(emissions.Scope1Config{
		Registry: factorRegistry,
	})
	scope3Calculator := emissions.NewScope3Calculator(emissions.Scope3Config{
		Registry: factorRegistry,
	})

	// 7. Set up auth store (Postgres if available, otherwise in-memory)
	var authStore auth.Store
	if database != nil {
		authStore = auth.NewPostgresStore(database.DB)
		log.Printf("[offgridflow] using PostgreSQL auth store")
	} else {
		authStore = auth.NewInMemoryStore()
		log.Printf("[offgridflow] using in-memory auth store")
	}

	// 8. Set up session manager for JWT auth
	var sessionManager *auth.SessionManager
	jwtSecret := cfg.Auth.JWTSecret
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-in-production"
		log.Printf("[offgridflow] WARNING: using dev JWT secret, set OFFGRIDFLOW_JWT_SECRET in production")
	}
	sessionManager, err = auth.NewSessionManager(jwtSecret)
	if err != nil {
		return fmt.Errorf("session manager: %w", err)
	}
	sessionManager.SetTTL(7 * 24 * time.Hour) // 7 day sessions
	log.Printf("[offgridflow] session manager initialized (7 day TTL)")

	// 9. Set up billing service (Stripe if configured)
	var billingService *billing.Service
	if cfg.Stripe.SecretKey != "" {
		stripeClient, err := billing.NewStripeClient(
			cfg.Stripe.SecretKey,
			cfg.Stripe.WebhookSecret,
			cfg.Stripe.PriceFree,
			cfg.Stripe.PriceBasic,
			cfg.Stripe.PricePro,
			cfg.Stripe.PriceEnterprise,
		)
		if err != nil {
			log.Printf("[offgridflow] WARNING: failed to create Stripe client: %v", err)
		} else {
			var billingStore billing.Store
			if database != nil {
				billingStore = billing.NewPostgresStore(database.DB)
			} else {
				billingStore = billing.NewInMemoryStore()
			}
			billingService = billing.NewService(stripeClient, billingStore)
			log.Printf("[offgridflow] billing service initialized with Stripe")
		}
	} else {
		log.Printf("[offgridflow] no OFFGRIDFLOW_STRIPE_SECRET_KEY provided, billing disabled")
	}

	// Check if auth should be required
	requireAuth := os.Getenv("OFFGRIDFLOW_REQUIRE_AUTH") != "false"

	// Determine cookie settings based on environment
	cookieSecure := cfg.Server.Env == "production"
	cookieDomain := os.Getenv("OFFGRIDFLOW_COOKIE_DOMAIN") // e.g., ".offgridflow.com"

	// 10. Create HTTP router with all dependencies including auth and billing
	routerDeps := &apihttp.RouterDeps{
		ModeManager:           modeManager,
		AIRouter:              aiRouter,
		ActivityStore:         activityStore,
		Scope1Calculator:      scope1Calculator,
		Scope2Calculator:      scope2Calculator,
		Scope3Calculator:      scope3Calculator,
		FactorRegistry:        factorRegistry,
		IngestionLogs:         ingestionLogs,
		IngestionSvc:          ingestionSvc,
		ConnectorStore:        connectorStore,
		IngestionOrchestrator: orchestrator,
		IngestionScheduler:    scheduler,
		AuthStore:             authStore,
		SessionManager:        sessionManager,
		BillingService:        billingService,
		DB:                    database,
		RequireAuth:           requireAuth,
		CookieDomain:          cookieDomain,
		CookieSecure:          cookieSecure,
		WorkflowService:       workflow.NewService(slog.Default(), nil),
	}
	router := apihttp.NewRouterWithDeps(routerDeps)

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("[offgridflow] starting api server on %s (env=%s)\n", addr, cfg.Server.Env)
	if requireAuth {
		log.Printf("[offgridflow] authentication REQUIRED (set OFFGRIDFLOW_REQUIRE_AUTH=false to disable)")
	} else {
		log.Printf("[offgridflow] authentication OPTIONAL (set OFFGRIDFLOW_REQUIRE_AUTH=true to enforce)")
	}

	if err := http.ListenAndServe(addr, router); err != nil {
		return fmt.Errorf("http server: %w", err)
	}

	return nil
}

// firstNonEmpty returns the first non-empty, trimmed string from the provided list.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
