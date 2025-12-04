package benchmarking

import (
	"context"
	"testing"
	"time"
)

func TestNewPool(t *testing.T) {
	cfg := PoolConfig{
		SecretKey: "test-secret-key",
	}
	pool := NewPool(cfg)

	if pool == nil {
		t.Fatal("NewPool returned nil")
	}
}

func TestNewService(t *testing.T) {
	pool := NewPool(PoolConfig{SecretKey: "test-secret"})
	svc := NewService(ServiceConfig{Pool: pool})

	if svc == nil {
		t.Fatal("NewService returned nil")
	}
}

func TestPool_Submit(t *testing.T) {
	pool := NewPool(PoolConfig{SecretKey: "test-secret"})

	metrics := EmissionMetrics{
		TotalCO2e:         1500.0,
		Scope1:            200.0,
		Scope2:            800.0,
		Scope3:            500.0,
		IntensityRevenue:  15.0,
		IntensityEmployee: 2.5,
		RenewablePercent:  45.0,
		YoYChange:         -5.0,
	}

	err := pool.Submit("tenant-1", IndustryTechnology, SizeMedium, RegionNorthAmerica, 2024, metrics)
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}
}

func TestPool_GetPeerGroup(t *testing.T) {
	pool := NewPool(PoolConfig{SecretKey: "test-secret"})

	// Submit some data first
	for i := 0; i < 5; i++ {
		metrics := EmissionMetrics{
			TotalCO2e:         float64(1000 + i*100),
			Scope1:            float64(100 + i*10),
			Scope2:            float64(600 + i*50),
			Scope3:            float64(300 + i*40),
			IntensityRevenue:  float64(10 + i),
			IntensityEmployee: float64(2 + float64(i)/10),
			RenewablePercent:  float64(30 + i*5),
			YoYChange:         float64(-3 + i),
		}
		tenantID := "tenant-" + string(rune('A'+i))
		if err := pool.Submit(tenantID, IndustryTechnology, SizeMedium, RegionNorthAmerica, 2024, metrics); err != nil {
			t.Fatalf("Submit failed: %v", err)
		}
	}

	// Get peer group
	peers := pool.GetPeerGroup(IndustryTechnology, SizeMedium, RegionNorthAmerica, 2024)

	if len(peers) != 5 {
		t.Errorf("Expected 5 peers, got %d", len(peers))
	}
}

func TestService_Compare(t *testing.T) {
	pool := NewPool(PoolConfig{SecretKey: "test-secret"})
	svc := NewService(ServiceConfig{Pool: pool})
	ctx := context.Background()

	// Submit varied data
	for i := 0; i < 10; i++ {
		metrics := EmissionMetrics{
			TotalCO2e:         float64(5000 + i*500),
			Scope1:            float64(500 + i*50),
			Scope2:            float64(3000 + i*300),
			Scope3:            float64(1500 + i*150),
			IntensityRevenue:  float64(10 + i),
			IntensityEmployee: float64(5 + float64(i)/2),
			RenewablePercent:  float64(20 + i*3),
			YoYChange:         float64(-2 + i),
		}
		tenantID := "tenant-" + string(rune('A'+i))
		pool.Submit(tenantID, IndustryManufacturing, SizeLarge, RegionEurope, 2024, metrics)
	}

	// Compare with your metrics
	yourMetrics := EmissionMetrics{
		TotalCO2e:         7000.0,
		Scope1:            700.0,
		Scope2:            4000.0,
		Scope3:            2300.0,
		IntensityRevenue:  14.0,
		IntensityEmployee: 7.0,
		RenewablePercent:  35.0,
		YoYChange:         -4.0,
	}

	result, err := svc.Compare(ctx, "my-tenant", IndustryManufacturing, SizeLarge, RegionEurope, 2024, yourMetrics)
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	if result == nil {
		t.Fatal("Compare returned nil")
	}

	if result.PeerCount < 1 {
		t.Errorf("Expected peer count >= 1, got %d", result.PeerCount)
	}
}

func TestAnonymizer_HashTenantID(t *testing.T) {
	anon := NewAnonymizer("test-secret")

	hash1 := anon.HashTenantID("tenant-1")
	hash2 := anon.HashTenantID("tenant-1")
	hash3 := anon.HashTenantID("tenant-2")

	if hash1 != hash2 {
		t.Error("Same tenant ID should produce same hash")
	}

	if hash1 == hash3 {
		t.Error("Different tenant IDs should produce different hashes")
	}

	if len(hash1) != 16 {
		t.Errorf("Hash should be 16 characters, got %d", len(hash1))
	}
}

func TestAnonymizer_BucketValue(t *testing.T) {
	anon := NewAnonymizer("test-secret")

	tests := []struct {
		value      float64
		bucketSize float64
		expected   float64
	}{
		{123.0, 100.0, 100.0},
		{178.0, 100.0, 200.0},
		{1234.5, 500.0, 1000.0},
		{1500.0, 500.0, 1500.0},
	}

	for _, tt := range tests {
		result := anon.BucketValue(tt.value, tt.bucketSize)
		if result != tt.expected {
			t.Errorf("BucketValue(%v, %v) = %v, expected %v", tt.value, tt.bucketSize, result, tt.expected)
		}
	}
}

func TestIndustry_Values(t *testing.T) {
	industries := []Industry{
		IndustryManufacturing,
		IndustryTechnology,
		IndustryRetail,
		IndustryFinancial,
		IndustryHealthcare,
		IndustryEnergy,
		IndustryTransportation,
		IndustryConstruction,
		IndustryAgriculture,
	}

	for _, ind := range industries {
		if string(ind) == "" {
			t.Errorf("Industry %v has empty string value", ind)
		}
	}
}

func TestCompanySize_Values(t *testing.T) {
	sizes := []CompanySize{
		SizeSmall,
		SizeMedium,
		SizeLarge,
		SizeEnterprise,
	}

	for _, size := range sizes {
		if string(size) == "" {
			t.Errorf("CompanySize %v has empty string value", size)
		}
	}
}

func TestRegion_Values(t *testing.T) {
	regions := []Region{
		RegionNorthAmerica,
		RegionEurope,
		RegionAsiaPacific,
		RegionLatinAmerica,
		RegionGlobal,
	}

	for _, region := range regions {
		if string(region) == "" {
			t.Errorf("Region %v has empty string value", region)
		}
	}
}

func TestEmissionMetrics_Fields(t *testing.T) {
	m := EmissionMetrics{
		TotalCO2e:         1000.0,
		Scope1:            100.0,
		Scope2:            600.0,
		Scope3:            300.0,
		IntensityRevenue:  10.0,
		IntensityEmployee: 2.0,
		RenewablePercent:  50.0,
		YoYChange:         -5.0,
	}

	// Verify scopes sum to total
	scopeSum := m.Scope1 + m.Scope2 + m.Scope3
	if scopeSum != m.TotalCO2e {
		t.Logf("Note: Scope sum (%.2f) != TotalCO2e (%.2f)", scopeSum, m.TotalCO2e)
	}
}

func TestAnonymousSubmission_Fields(t *testing.T) {
	sub := AnonymousSubmission{
		AnonymousID: "abc123",
		Industry:    IndustryTechnology,
		Size:        SizeMedium,
		Region:      RegionNorthAmerica,
		Year:        2024,
		Metrics: EmissionMetrics{
			TotalCO2e: 1500.0,
		},
		SubmittedAt: time.Now(),
	}

	if sub.AnonymousID == "" {
		t.Error("AnonymousID should not be empty")
	}
	if sub.Year <= 0 {
		t.Error("Year should be positive")
	}
}
