package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/example/offgridflow/internal/compliance"
	"github.com/example/offgridflow/internal/config"
	"github.com/example/offgridflow/internal/db"
	"github.com/example/offgridflow/internal/demo"
	"github.com/example/offgridflow/internal/emissions"
	"github.com/example/offgridflow/internal/emissions/factors"
	"github.com/example/offgridflow/internal/ingestion"
	"github.com/example/offgridflow/internal/logging"
)

func main() {
	logger := logging.New(logging.Config{
		Level:  slog.LevelInfo,
		Format: logging.FormatText,
		Output: os.Stdout,
	})

	if len(os.Args) < 2 {
		fmt.Println("usage: offgridflow <command> [args]")
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "ingest-demo-data":
		if err := ingestDemoData(logger); err != nil {
			logger.Error("ingest demo data failed", "error", err)
			os.Exit(1)
		}
	case "recalc-emissions":
		if err := recalcEmissions(logger); err != nil {
			logger.Error("recalc failed", "error", err)
			os.Exit(1)
		}
	case "generate-csrd-report":
		if err := generateCSRDReport(logger, os.Args[2:]); err != nil {
			logger.Error("csrd report generation failed", "error", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("unknown command: %s\n", command)
		os.Exit(1)
	}
}

type runtime struct {
	ctx           context.Context
	cancel        context.CancelFunc
	db            *db.DB
	activityStore ingestion.ActivityStore
	registry      emissions.FactorRegistry
	logger        *slog.Logger
}

func buildRuntime(logger *slog.Logger) (*runtime, error) {
	if logger == nil {
		logger = slog.Default()
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	var database *db.DB
	if cfg.Database.DSN != "" {
		database, err = db.Connect(ctx, db.Config{DSN: cfg.Database.DSN})
		if err != nil {
			cancel()
			return nil, fmt.Errorf("connect db: %w", err)
		}
		if err := database.RunMigrations(ctx); err != nil {
			cancel()
			return nil, fmt.Errorf("run migrations: %w", err)
		}
	}

	var store ingestion.ActivityStore
	if database != nil {
		store = ingestion.NewPostgresActivityStore(database.DB)
	} else {
		store = ingestion.NewInMemoryActivityStore()
	}

	var registry emissions.FactorRegistry
	if database != nil {
		registry, err = factors.NewPostgresRegistry(factors.DefaultPostgresConfig(database.DB))
		if err != nil {
			cancel()
			return nil, fmt.Errorf("init factor registry: %w", err)
		}
	} else {
		registry = factors.NewInMemoryRegistry(factors.DefaultRegistryConfig())
	}

	return &runtime{
		ctx:           ctx,
		cancel:        cancel,
		db:            database,
		activityStore: store,
		registry:      registry,
		logger:        logger,
	}, nil
}

func ingestDemoData(logger *slog.Logger) error {
	rt, err := buildRuntime(logger)
	if err != nil {
		return err
	}
	defer rt.cancel()
	defer func() {
		if rt.db != nil {
			_ = rt.db.Close()
		}
	}()

	// Seed sample utility bill activities
	if mem, ok := rt.activityStore.(*ingestion.InMemoryActivityStore); ok {
		mem.SeedDemoData()
	} else {
		now := time.Now().UTC()
		activities := []ingestion.Activity{
			ingestion.NewActivityBuilder().
				WithID("demo-utility-1").
				WithSource(ingestion.SourceUtilityBill.String()).
				WithCategory("electricity").
				WithMeterID("METER-DEMO-1").
				WithLocation("US-WEST").
				WithPeriod(now.AddDate(0, -1, 0), now).
				WithQuantity(12000, ingestion.UnitKWh.String()).
				WithOrgID("org-demo").
				MustBuild(),
			ingestion.NewActivityBuilder().
				WithID("demo-utility-2").
				WithSource(ingestion.SourceUtilityBill.String()).
				WithCategory("electricity").
				WithMeterID("METER-DEMO-2").
				WithLocation("EU-CENTRAL").
				WithPeriod(now.AddDate(0, -1, 0), now).
				WithQuantity(8500, ingestion.UnitKWh.String()).
				WithOrgID("org-demo").
				MustBuild(),
		}
		if err := rt.activityStore.SaveBatch(rt.ctx, activities); err != nil {
			return fmt.Errorf("persist demo activities: %w", err)
		}
	}

	// Cache a demo compliance/emissions snapshot for offline demos
	_ = demo.GenerateDemoData(rt.ctx, demo.DemoConfig{Enabled: true})

	logger.Info("ingested demo data")
	return nil
}

func recalcEmissions(logger *slog.Logger) error {
	rt, err := buildRuntime(logger)
	if err != nil {
		return err
	}
	defer rt.cancel()
	defer func() {
		if rt.db != nil {
			_ = rt.db.Close()
		}
	}()

	engine := emissions.NewEngine(emissions.EngineConfig{
		Registry:            rt.registry,
		Logger:              rt.logger,
		DefaultRegion:       "US-AVERAGE",
		EnableParallelBatch: true,
	})

	records, err := rt.activityStore.List(rt.ctx)
	if err != nil {
		return fmt.Errorf("list activities: %w", err)
	}

	acts := make([]emissions.Activity, 0, len(records))
	for _, a := range records {
		acts = append(acts, a)
	}

	result, err := engine.CalculateBatch(rt.ctx, acts)
	if err != nil {
		return fmt.Errorf("recalculate: %w", err)
	}

	logger.Info("recalculated emissions",
		"records", len(result.Records),
		"successes", result.SuccessCount,
		"errors", result.ErrorCount,
		"total_kgco2e", result.TotalEmissionsKgCO2e)
	return nil
}

func generateCSRDReport(logger *slog.Logger, args []string) error {
	fs := flag.NewFlagSet("generate-csrd-report", flag.ExitOnError)
	year := fs.Int("year", time.Now().Year(), "reporting year")
	orgID := fs.String("org", "org-demo", "organization id")
	_ = fs.Parse(args)

	rt, err := buildRuntime(logger)
	if err != nil {
		return err
	}
	defer rt.cancel()
	defer func() {
		if rt.db != nil {
			_ = rt.db.Close()
		}
	}()

	scope1 := emissions.NewScope1Calculator(emissions.Scope1Config{Registry: rt.registry})
	scope2 := emissions.NewScope2Calculator(emissions.Scope2Config{Registry: rt.registry})
	scope3 := emissions.NewScope3Calculator(emissions.Scope3Config{Registry: rt.registry})

	complianceSvc := compliance.NewService(rt.activityStore, scope1, scope2, scope3)
	report, err := complianceSvc.GenerateCSRDReport(rt.ctx, *orgID, *year)
	if err != nil {
		return fmt.Errorf("generate csrd report: %w", err)
	}

	scope1Tons, scope2Tons, scope3Tons, totalTons, calcErr := computeTotals(rt.ctx, rt.activityStore, scope1, scope2, scope3, *orgID, *year)
	if calcErr != nil {
		return calcErr
	}

	logger.Info("generated CSRD report",
		"org_id", *orgID,
		"year", *year,
		"scope1_tons", scope1Tons,
		"scope2_tons", scope2Tons,
		"scope3_tons", scope3Tons,
		"total_tons", totalTons,
		"completeness", report.CompletenessScore,
	)
	return nil
}

func computeTotals(ctx context.Context, store ingestion.ActivityStore, scope1 *emissions.Scope1Calculator, scope2 *emissions.Scope2Calculator, scope3 *emissions.Scope3Calculator, orgID string, year int) (float64, float64, float64, float64, error) {
	activities, err := store.List(ctx)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("list activities: %w", err)
	}

	var filtered []emissions.Activity
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)
	for _, act := range activities {
		if orgID != "" && act.OrgID != "" && act.OrgID != orgID {
			continue
		}
		if act.PeriodStart.Before(start) || !act.PeriodStart.Before(end) {
			continue
		}
		filtered = append(filtered, act)
	}

	var scope1Total, scope2Total, scope3Total float64

	if scope1 != nil {
		records, err := scope1.CalculateBatch(ctx, filtered)
		if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("scope1 calc: %w", err)
		}
		for _, rec := range records {
			scope1Total += rec.EmissionsTonnesCO2e
		}
	}
	if scope2 != nil {
		records, err := scope2.CalculateBatch(ctx, filtered)
		if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("scope2 calc: %w", err)
		}
		for _, rec := range records {
			scope2Total += rec.EmissionsTonnesCO2e
		}
	}
	if scope3 != nil {
		records, err := scope3.CalculateBatch(ctx, filtered)
		if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("scope3 calc: %w", err)
		}
		for _, rec := range records {
			scope3Total += rec.EmissionsTonnesCO2e
		}
	}

	total := scope1Total + scope2Total + scope3Total
	return scope1Total, scope2Total, scope3Total, total, nil
}
