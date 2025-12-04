package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/example/offgridflow/internal/config"
	"github.com/example/offgridflow/internal/db"
	"github.com/example/offgridflow/internal/emissions"
	"github.com/example/offgridflow/internal/emissions/factors"
	"github.com/example/offgridflow/internal/events"
	"github.com/example/offgridflow/internal/ingestion"
	"github.com/example/offgridflow/internal/ingestion/sources/aws"
	"github.com/example/offgridflow/internal/ingestion/sources/azure"
	"github.com/example/offgridflow/internal/ingestion/sources/gcp"
	"github.com/example/offgridflow/internal/ingestion/sources/sap"
	"github.com/example/offgridflow/internal/ingestion/sources/utility_bills"
	"github.com/example/offgridflow/internal/logging"
	"github.com/example/offgridflow/internal/worker"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func main() {
	logger := logging.New(logging.Config{
		Level:  slog.LevelInfo,
		Format: logging.FormatText,
		Output: os.Stdout,
	})

	cfg, err := config.Load()
	if err != nil {
		logger.Error("configuration error", "error", err)
		os.Exit(1)
	}

	if err := initMetricsProvider(logger); err != nil {
		logger.Warn("metrics exporter not initialized", "error", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	dbConn, err := db.Connect(ctx, db.Config{
		DSN:             cfg.Database.DSN,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer dbConn.Close()

	if err := dbConn.RunMigrations(ctx); err != nil {
		logger.Error("database migrations failed", "error", err)
		os.Exit(1)
	}

	// Wiring: activity store -> factor registry -> emissions engine -> jobs.
	activityStore := ingestion.NewPostgresActivityStore(dbConn.DB)
	ingestionLogs := ingestion.NewPostgresLogStore(dbConn.DB)
	connectorStore := ingestion.NewPostgresConnectorStatusStore(dbConn.DB)
	factorRegistry, err := factors.NewPostgresRegistry(factors.DefaultPostgresConfig(dbConn.DB))
	if err != nil {
		logger.Error("failed to initialize factor registry", "error", err)
		os.Exit(1)
	}
	engine := emissions.NewEngine(emissions.EngineConfig{
		Registry: factorRegistry,
		Logger:   logger,
	})

	eventBus := events.NewInMemoryBus()
	workerCfg := worker.FromEnv()
	metrics := worker.NewMetricsRecorder()
	alerts := worker.NewAlertQueue(eventBus, logger, 256)

	adapters := make([]ingestion.SourceIngestionAdapter, 0, 5)
	start := time.Now().AddDate(0, 0, -cfg.Ingestion.LookbackDays)
	end := time.Now()

	if cfg.Ingestion.AWS.Enabled {
		awsAdapter, err := aws.NewAdapter(aws.Config{
			AccessKeyID:     cfg.Ingestion.AWS.AccessKeyID,
			SecretAccessKey: cfg.Ingestion.AWS.SecretAccessKey,
			Region:          cfg.Ingestion.AWS.Region,
			RoleARN:         cfg.Ingestion.AWS.RoleARN,
			AccountID:       cfg.Ingestion.AWS.AccountID,
			OrgID:           cfg.Ingestion.AWS.OrgID,
			StartDate:       start,
			EndDate:         end,
		})
		if err != nil {
			logger.Warn("aws adapter disabled due to config error", "error", err)
		} else {
			adapters = append(adapters, awsAdapter)
		}
	}

	if cfg.Ingestion.Azure.Enabled {
		azureAdapter, err := azure.NewAdapter(azure.Config{
			TenantID:       cfg.Ingestion.Azure.TenantID,
			ClientID:       cfg.Ingestion.Azure.ClientID,
			ClientSecret:   cfg.Ingestion.Azure.ClientSecret,
			SubscriptionID: cfg.Ingestion.Azure.SubscriptionID,
			OrgID:          cfg.Ingestion.Azure.OrgID,
			StartDate:      start,
			EndDate:        end,
		})
		if err != nil {
			logger.Warn("azure adapter disabled due to config error", "error", err)
		} else {
			adapters = append(adapters, azureAdapter)
		}
	}

	if cfg.Ingestion.GCP.Enabled {
		gcpAdapter, err := gcp.NewAdapter(gcp.Config{
			ProjectID:         cfg.Ingestion.GCP.ProjectID,
			BillingAccountID:  cfg.Ingestion.GCP.BillingAccountID,
			BigQueryDataset:   cfg.Ingestion.GCP.BigQueryDataset,
			BigQueryTable:     cfg.Ingestion.GCP.BigQueryTable,
			ServiceAccountKey: cfg.Ingestion.GCP.ServiceAccountKey,
			OrgID:             cfg.Ingestion.GCP.OrgID,
			StartDate:         start,
			EndDate:           end,
		})
		if err != nil {
			logger.Warn("gcp adapter disabled due to config error", "error", err)
		} else {
			adapters = append(adapters, gcpAdapter)
		}
	}

	if cfg.Ingestion.SAP.Enabled {
		adapters = append(adapters, &sap.Adapter{})
	}

	if cfg.Ingestion.Utility.Enabled {
		adapters = append(adapters, &utility_bills.Adapter{})
	}

	if len(adapters) == 0 {
		logger.Warn("no ingestion adapters enabled; worker ingestion will be idle")
	}

	ingestionService := &ingestion.Service{
		Adapters:       adapters,
		Store:          activityStore,
		Logs:           ingestionLogs,
		ConnectorStore: connectorStore,
	}

	logger.Info("worker starting",
		"ingestion_every", workerCfg.IngestionInterval.String(),
		"recalc_every", workerCfg.RecalcInterval.String(),
		"alert_every", workerCfg.AlertInterval.String(),
	)

	alerts.Start(ctx)

	runner := worker.NewRunner(logger, []worker.JobSpec{
		{
			Job:            worker.IngestionJob{Service: ingestionService, Logger: logger},
			Every:          workerCfg.IngestionInterval,
			Timeout:        workerCfg.DefaultTimeout,
			RetryLimit:     workerCfg.DefaultRetryLimit,
			BackoffInitial: workerCfg.DefaultBackoff,
			BackoffMax:     workerCfg.DefaultBackoffMax,
			Jitter:         workerCfg.DefaultJitter,
		},
		{
			Job:            worker.RecalculationJob{Store: activityStore, Engine: engine, Bus: eventBus, Logger: logger},
			Every:          workerCfg.RecalcInterval,
			Timeout:        workerCfg.DefaultTimeout,
			RetryLimit:     workerCfg.DefaultRetryLimit,
			BackoffInitial: workerCfg.DefaultBackoff,
			BackoffMax:     workerCfg.DefaultBackoffMax,
			Jitter:         workerCfg.DefaultJitter,
		},
		{
			Job:            worker.AlertJob{Bus: eventBus, Logger: logger},
			Every:          workerCfg.AlertInterval,
			Timeout:        15 * time.Second,
			RetryLimit:     1,
			BackoffInitial: 1 * time.Second,
			BackoffMax:     5 * time.Second,
			Jitter:         workerCfg.DefaultJitter,
		},
	}, metrics, alerts)

	runner.Start(ctx)
	runner.Wait()
	logger.Info("worker shutdown complete")
}

// initMetricsProvider configures an OTLP metrics exporter if OTEL_EXPORTER_OTLP_ENDPOINT is set.
func initMetricsProvider(logger *slog.Logger) error {
	endpoint := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	if endpoint == "" {
		// Graceful no-op if not configured.
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(stripScheme(endpoint)),
		otlpmetrichttp.WithInsecure(), // assume internal collector unless TLS configured
	)
	if err != nil {
		return err
	}

	res, err := sdkresource.Merge(
		sdkresource.Default(),
		sdkresource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("offgridflow-worker"),
			semconv.DeploymentEnvironment(os.Getenv("APP_ENV")),
		),
	)
	if err != nil {
		return err
	}

	provider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter)),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(provider)
	logger.Info("metrics exporter initialized", "endpoint", endpoint)
	return nil
}

func stripScheme(endpoint string) string {
	if strings.HasPrefix(endpoint, "http://") {
		return strings.TrimPrefix(endpoint, "http://")
	}
	if strings.HasPrefix(endpoint, "https://") {
		return strings.TrimPrefix(endpoint, "https://")
	}
	return endpoint
}
