package allocation

import (
	"context"
	"testing"
	"time"

	"github.com/example/offgridflow/internal/emissions"
)

func TestServiceMetrics_Clone(t *testing.T) {
	// Create original metrics
	original := NewServiceMetrics()
	original.RecordAllocation(DimensionDepartment, "test-rule-1", true)
	original.RecordAllocation(DimensionProject, "test-rule-2", true)
	original.RecordAllocation(DimensionCostCenter, "test-rule-1", false)

	// Clone the metrics
	cloned := original.Clone()

	// Verify cloned values match
	if cloned.TotalAllocations != original.TotalAllocations {
		t.Errorf("TotalAllocations = %v, want %v", cloned.TotalAllocations, original.TotalAllocations)
	}
	if cloned.SuccessfulAllocations != original.SuccessfulAllocations {
		t.Errorf("SuccessfulAllocations = %v, want %v", cloned.SuccessfulAllocations, original.SuccessfulAllocations)
	}
	if cloned.FailedAllocations != original.FailedAllocations {
		t.Errorf("FailedAllocations = %v, want %v", cloned.FailedAllocations, original.FailedAllocations)
	}

	// Verify maps are independent
	cloned.ByDimension[DimensionDepartment] = 999
	if original.ByDimension[DimensionDepartment] == 999 {
		t.Error("Clone shares map with original (not a deep copy)")
	}

	// Verify clone is safe to modify concurrently while original is locked
	done := make(chan bool)
	go func() {
		// Modify clone
		cloned.RecordAllocation(DimensionProduct, "test-rule-3", true)
		done <- true
	}()

	// Modify original at same time
	original.RecordAllocation(DimensionDepartment, "test-rule-4", true)

	<-done
}

func TestServiceMetrics_RecordAllocation(t *testing.T) {
	metrics := NewServiceMetrics()

	// Record successful allocation
	metrics.RecordAllocation(DimensionDepartment, "rule1", true)
	if metrics.TotalAllocations != 1 {
		t.Errorf("TotalAllocations = %v, want 1", metrics.TotalAllocations)
	}
	if metrics.SuccessfulAllocations != 1 {
		t.Errorf("SuccessfulAllocations = %v, want 1", metrics.SuccessfulAllocations)
	}
	if metrics.FailedAllocations != 0 {
		t.Errorf("FailedAllocations = %v, want 0", metrics.FailedAllocations)
	}

	// Record failed allocation
	metrics.RecordAllocation(DimensionProject, "rule2", false)
	if metrics.TotalAllocations != 2 {
		t.Errorf("TotalAllocations = %v, want 2", metrics.TotalAllocations)
	}
	if metrics.SuccessfulAllocations != 1 {
		t.Errorf("SuccessfulAllocations = %v, want 1", metrics.SuccessfulAllocations)
	}
	if metrics.FailedAllocations != 1 {
		t.Errorf("FailedAllocations = %v, want 1", metrics.FailedAllocations)
	}

	// Verify dimension tracking
	if metrics.ByDimension[DimensionDepartment] != 1 {
		t.Errorf("ByDimension[Department] = %v, want 1", metrics.ByDimension[DimensionDepartment])
	}
	if metrics.ByDimension[DimensionProject] != 1 {
		t.Errorf("ByDimension[Project] = %v, want 1", metrics.ByDimension[DimensionProject])
	}

	// Verify rule tracking
	if metrics.ByRule["rule1"] != 1 {
		t.Errorf("ByRule[rule1] = %v, want 1", metrics.ByRule["rule1"])
	}
	if metrics.ByRule["rule2"] != 1 {
		t.Errorf("ByRule[rule2] = %v, want 1", metrics.ByRule["rule2"])
	}

	// Verify timestamp was set
	if metrics.LastAllocationAt.IsZero() {
		t.Error("LastAllocationAt should not be zero")
	}
	if time.Since(metrics.LastAllocationAt) > time.Second {
		t.Error("LastAllocationAt should be recent")
	}
}

func TestAllocate_FixedRule(t *testing.T) {
	svc, err := NewService(ServiceConfig{
		Rules: []Rule{
			{
				ID:        "r1",
				Dimension: DimensionDepartment,
				Method:    MethodFixed,
				Allocations: []AllocationTarget{
					{TargetID: "eng", TargetName: "Engineering", Percentage: 60},
					{TargetID: "sales", TargetName: "Sales", Percentage: 40},
				},
				Filters: []RuleFilter{{Field: "scope", Operator: "eq", Value: emissions.Scope1.String()}},
				Enabled: true,
			},
		},
		EnableMetrics: true,
	})
	if err != nil {
		t.Fatalf("NewService error: %v", err)
	}

	record := emissions.EmissionRecord{ID: "rec-1", Scope: emissions.Scope1, EmissionsKgCO2e: 100}
	res, err := svc.Allocate(context.Background(), record, DimensionDepartment)
	if err != nil {
		t.Fatalf("Allocate returned error: %v", err)
	}
	if len(res.AllocatedEmissions) != 2 {
		t.Fatalf("expected 2 allocations, got %d", len(res.AllocatedEmissions))
	}
	if res.AllocatedEmissions[0].EmissionsKgCO2e != 60 {
		t.Fatalf("expected 60kg for first allocation, got %.1f", res.AllocatedEmissions[0].EmissionsKgCO2e)
	}
	if metrics := svc.Metrics(); metrics == nil || metrics.TotalAllocations != 1 {
		t.Fatalf("expected metrics to record allocation")
	}
}

func TestAllocate_NoMatchingRule(t *testing.T) {
	svc, err := NewService(ServiceConfig{
		Rules: []Rule{
			{
				ID:        "r1",
				Dimension: DimensionDepartment,
				Method:    MethodFixed,
				Allocations: []AllocationTarget{
					{TargetID: "eng", Percentage: 100},
				},
				Filters: []RuleFilter{{Field: "scope", Operator: "eq", Value: emissions.Scope2.String()}},
				Enabled: true,
			},
		},
	})
	if err != nil {
		t.Fatalf("NewService error: %v", err)
	}

	record := emissions.EmissionRecord{ID: "rec-1", Scope: emissions.Scope1, EmissionsKgCO2e: 100}
	_, err = svc.Allocate(context.Background(), record, DimensionDepartment)
	if err == nil {
		t.Fatalf("expected error for no applicable rule")
	}
}

func TestAllocate_DriverAllocation(t *testing.T) {
	svc, err := NewService(ServiceConfig{
		Rules: []Rule{
			{
				ID:        "driver-rule",
				Dimension: DimensionProject,
				Method:    MethodRevenue,
				Allocations: []AllocationTarget{
					{TargetID: "p1"},
					{TargetID: "p2"},
				},
				Enabled: true,
			},
		},
		EnableMetrics: true,
	})
	if err != nil {
		t.Fatalf("NewService error: %v", err)
	}
	// Set driver data: p1 has twice p2
	svc.SetDriverData(DimensionProject, "p1", 200)
	svc.SetDriverData(DimensionProject, "p2", 100)

	record := emissions.EmissionRecord{ID: "rec-1", Scope: emissions.Scope1, EmissionsKgCO2e: 90}
	res, err := svc.Allocate(context.Background(), record, DimensionProject)
	if err != nil {
		t.Fatalf("Allocate returned error: %v", err)
	}
	if len(res.AllocatedEmissions) != 2 {
		t.Fatalf("expected 2 allocations, got %d", len(res.AllocatedEmissions))
	}
	total := res.AllocatedEmissions[0].EmissionsKgCO2e + res.AllocatedEmissions[1].EmissionsKgCO2e
	if diff := total - 90; diff > 0.001 || diff < -0.001 {
		t.Fatalf("allocated totals do not sum to input, got %.3f", total)
	}
}

func TestAllocate_ExpressionFallback(t *testing.T) {
	svc, err := NewService(ServiceConfig{
		Rules: []Rule{
			{
				ID:         "expr",
				Dimension:  DimensionProduct,
				Method:     MethodExpression,
				Expression: "any()", // currently unused
				Allocations: []AllocationTarget{
					{TargetID: "a"},
					{TargetID: "b"},
					{TargetID: "c"},
				},
				Enabled: true,
			},
		},
		EnableMetrics: true,
	})
	if err != nil {
		t.Fatalf("NewService error: %v", err)
	}

	record := emissions.EmissionRecord{ID: "rec-1", Scope: emissions.Scope1, EmissionsKgCO2e: 120}
	res, err := svc.Allocate(context.Background(), record, DimensionProduct)
	if err != nil {
		t.Fatalf("Allocate returned error: %v", err)
	}
	if len(res.AllocatedEmissions) != 3 {
		t.Fatalf("expected 3 allocations, got %d", len(res.AllocatedEmissions))
	}
	expected := 40.0
	for _, ae := range res.AllocatedEmissions {
		if diff := ae.EmissionsKgCO2e - expected; diff < -0.001 || diff > 0.001 {
			t.Fatalf("expected equal allocation of %.1f, got %.3f", expected, ae.EmissionsKgCO2e)
		}
	}
}
