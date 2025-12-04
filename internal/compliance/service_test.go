package compliance

import (
	"context"
	"testing"
	"time"

	"github.com/example/offgridflow/internal/emissions"
	"github.com/example/offgridflow/internal/emissions/factors"
	"github.com/example/offgridflow/internal/ingestion"
	"github.com/stretchr/testify/require"
)

func TestGenerateCSRDReportIncludesEnergyData(t *testing.T) {
	ctx := context.Background()
	store := ingestion.NewInMemoryActivityStore()
	now := time.Now().UTC()

	renewable := ingestion.NewActivityBuilder().
		WithID("act-renew").
		WithSource(string(ingestion.SourceUtilityBill)).
		WithCategory("solar").
		WithLocation("EU-CENTRAL").
		WithQuantity(2500, string(ingestion.UnitKWh)).
		WithPeriod(now.AddDate(0, -1, 0), now).
		WithOrgID("org-test").
		WithMetadata("energy_source", "renewable").
		MustBuild()

	fossil := ingestion.NewActivityBuilder().
		WithID("act-fossil").
		WithSource(string(ingestion.SourceUtilityBill)).
		WithCategory("natural_gas").
		WithLocation("US-WEST").
		WithQuantity(4000, string(ingestion.UnitKWh)).
		WithPeriod(now.AddDate(0, -1, 0), now).
		WithOrgID("org-test").
		WithMetadata("energy_source", "fossil").
		MustBuild()

	scope3 := ingestion.NewActivityBuilder().
		WithID("act-scope3").
		WithSource(string(ingestion.SourceTravel)).
		WithCategory("business_travel").
		WithLocation("US-WEST").
		WithQuantity(1, string(ingestion.UnitTonne)).
		WithPeriod(now.AddDate(0, -1, 0), now).
		WithOrgID("org-test").
		WithMetadata("scope3_category", "travel").
		MustBuild()

	require.NoError(t, store.SaveBatch(ctx, []ingestion.Activity{renewable, fossil, scope3}))

	registry := factors.NewInMemoryRegistry(factors.RegistryConfig{
		PreloadDefaults:    true,
		ValidateOnRegister: true,
	})

	scope1 := emissions.NewScope1Calculator(emissions.Scope1Config{Registry: registry})
	scope2 := emissions.NewScope2Calculator(emissions.Scope2Config{Registry: registry})
	scope3Calc := emissions.NewScope3Calculator(emissions.Scope3Config{Registry: registry})

	service := NewService(store, scope1, scope2, scope3Calc)

	report, err := service.GenerateCSRDReport(ctx, "org-test", now.Year())
	require.NoError(t, err)
	require.NotNil(t, report)

	energyMetrics, ok := report.Metrics["E1-5_energy"].(map[string]interface{})
	require.True(t, ok)

	mix, ok := energyMetrics["energyMix"].(map[string]interface{})
	require.True(t, ok)

	renewableData, ok := mix["renewable"].(map[string]interface{})
	require.True(t, ok)
	require.True(t, renewableData["value"].(float64) > 0)

	fossilData, ok := mix["fossilFuel"].(map[string]interface{})
	require.True(t, ok)
	require.True(t, fossilData["value"].(float64) > 0)

	emissionsMetrics, ok := report.Metrics["E1-6_ghgEmissions"].(map[string]interface{})
	require.True(t, ok)
	scope3Data, ok := emissionsMetrics["scope3"].(map[string]interface{})
	require.True(t, ok)
	if breakdown, has := scope3Data["categoryBreakdown"]; has {
		require.NotEmpty(t, breakdown)
	}
}
