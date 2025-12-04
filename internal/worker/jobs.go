package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/example/offgridflow/internal/emissions"
	"github.com/example/offgridflow/internal/events"
	"github.com/example/offgridflow/internal/ingestion"
)

// IngestionJob triggers the ingestion service to pull new activities.
type IngestionJob struct {
	Service *ingestion.Service
	Logger  *slog.Logger
}

func (j IngestionJob) Name() string { return "ingestion_sync" }

func (j IngestionJob) Run(ctx context.Context) error {
	if j.Service == nil {
		return fmt.Errorf("ingestion service is nil")
	}
	activities, err := j.Service.Run(ctx)
	if err != nil {
		return err
	}
	if j.Logger != nil {
		j.Logger.Info("ingestion sync completed", "activities", len(activities))
	}
	return nil
}

// RecalculationJob recomputes emissions for all stored activities.
type RecalculationJob struct {
	Store  ingestion.ActivityStore
	Engine *emissions.Engine
	Bus    events.Bus
	Logger *slog.Logger
}

func (j RecalculationJob) Name() string { return "emissions_recalc" }

func (j RecalculationJob) Run(ctx context.Context) error {
	if j.Store == nil {
		return fmt.Errorf("activity store is nil")
	}
	if j.Engine == nil {
		return fmt.Errorf("emissions engine is nil")
	}

	activities, err := j.Store.List(ctx)
	if err != nil {
		return fmt.Errorf("list activities: %w", err)
	}
	if len(activities) == 0 {
		if j.Logger != nil {
			j.Logger.Info("recalc skipped; no activities available")
		}
		return nil
	}

	activityInterfaces := make([]emissions.Activity, 0, len(activities))
	for i := range activities {
		activityInterfaces = append(activityInterfaces, activities[i])
	}

	result, err := j.Engine.CalculateBatch(ctx, activityInterfaces)
	if err != nil {
		return fmt.Errorf("calculate batch: %w", err)
	}

	if j.Bus != nil {
		_ = j.Bus.Publish(ctx, events.Event{
			Type:      "emissions.recalculated",
			Timestamp: time.Now().UTC(),
			Payload: map[string]interface{}{
				"records":     len(result.Records),
				"successes":   result.SuccessCount,
				"errors":      result.ErrorCount,
				"total_kgco2": result.TotalEmissionsKgCO2e,
			},
		})
	}

	if j.Logger != nil {
		j.Logger.Info("emissions recalculated",
			"records", len(result.Records),
			"successes", result.SuccessCount,
			"errors", result.ErrorCount,
			"kg_co2e", result.TotalEmissionsKgCO2e)
	}

	return nil
}

// AlertJob scans for recent failures and emits alerts.
type AlertJob struct {
	Bus    events.Bus
	Logger *slog.Logger
}

func (j AlertJob) Name() string { return "alerts" }

func (j AlertJob) Run(ctx context.Context) error {
	// In a full implementation this would read from a durable queue / DB.
	// For now we emit a heartbeat event to prove alerting is wired.
	if j.Bus != nil {
		_ = j.Bus.Publish(ctx, events.Event{
			Type:      "worker.heartbeat",
			Timestamp: time.Now().UTC(),
			Payload:   map[string]string{"service": "worker", "status": "ok"},
		})
	}
	if j.Logger != nil {
		j.Logger.Info("alert heartbeat emitted")
	}
	return nil
}
