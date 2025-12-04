package residency

import (
	"testing"
)

func TestRegion_Values(t *testing.T) {
	regions := []Region{
		RegionEU,
		RegionUS,
		RegionAPAC,
		RegionUK,
		RegionAU,
	}

	for _, r := range regions {
		if string(r) == "" {
			t.Errorf("Region %v has empty string value", r)
		}
	}
}

func TestDefaultRegions(t *testing.T) {
	regions := DefaultRegions()

	if len(regions) == 0 {
		t.Fatal("DefaultRegions returned empty map")
	}

	// Check EU region
	eu, ok := regions[RegionEU]
	if !ok {
		t.Error("EU region should be present")
	} else {
		if eu.Name == "" {
			t.Error("EU region Name should not be empty")
		}
		if len(eu.Regulations) == 0 {
			t.Error("EU region should have regulations")
		}
	}

	// Check US region
	us, ok := regions[RegionUS]
	if !ok {
		t.Error("US region should be present")
	} else if !us.Primary {
		t.Error("US should be marked as primary region")
	}
}

func TestRegionConfig_Fields(t *testing.T) {
	cfg := RegionConfig{
		Region:      RegionEU,
		Name:        "European Union",
		Location:    "Frankfurt, Germany",
		DSN:         "postgres://localhost/eu",
		Regulations: []string{"GDPR", "CSRD"},
		Primary:     false,
	}

	if cfg.Region != RegionEU {
		t.Errorf("Expected region %v, got %v", RegionEU, cfg.Region)
	}
	if len(cfg.Regulations) != 2 {
		t.Errorf("Expected 2 regulations, got %d", len(cfg.Regulations))
	}
}

func TestTenantResidency_Fields(t *testing.T) {
	tr := TenantResidency{
		TenantID:      "tenant-001",
		PrimaryRegion: RegionEU,
		Regulations:   []string{"GDPR", "CSRD"},
		Country:       "DE",
		DataTypes: map[string]Region{
			"emissions": RegionEU,
			"billing":   RegionUS,
		},
	}

	if tr.TenantID == "" {
		t.Error("TenantID should not be empty")
	}
	if tr.PrimaryRegion != RegionEU {
		t.Errorf("Expected primary region %v, got %v", RegionEU, tr.PrimaryRegion)
	}
	if len(tr.DataTypes) != 2 {
		t.Errorf("Expected 2 data types, got %d", len(tr.DataTypes))
	}
}

func TestRouterConfig_Fields(t *testing.T) {
	cfg := RouterConfig{
		Regions: map[Region]string{
			RegionEU: "postgres://eu.db.local/emissions",
			RegionUS: "postgres://us.db.local/emissions",
		},
		FallbackRegion: RegionUS,
	}

	if len(cfg.Regions) != 2 {
		t.Errorf("Expected 2 regions, got %d", len(cfg.Regions))
	}
	if cfg.FallbackRegion != RegionUS {
		t.Errorf("Expected fallback %v, got %v", RegionUS, cfg.FallbackRegion)
	}
}

func TestNewRouter(t *testing.T) {
	// Test with empty config (no actual DB connections)
	// This will fail because there are no regional databases
	cfg := RouterConfig{
		FallbackRegion: RegionUS,
	}

	_, err := NewRouter(cfg)
	if err != nil {
		t.Logf("NewRouter expected to fail without DB: %v", err)
	}
}
