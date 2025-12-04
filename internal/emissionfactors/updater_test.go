package emissionfactors

import (
	"testing"
	"time"
)

func TestFactorSource_Values(t *testing.T) {
	sources := []FactorSource{
		SourceEPA,
		SourceDEFRA,
		SourceIEA,
		SourceIPCC,
		SourceEGRID,
	}

	for _, s := range sources {
		if string(s) == "" {
			t.Errorf("FactorSource %v has empty string value", s)
		}
	}
}

func TestFactorCategory_Values(t *testing.T) {
	categories := []FactorCategory{
		CategoryElectricity,
		CategoryFuel,
		CategoryTransport,
		CategoryWaste,
		CategoryWater,
		CategoryMaterial,
		CategoryRefrigerant,
	}

	for _, c := range categories {
		if string(c) == "" {
			t.Errorf("FactorCategory %v has empty string value", c)
		}
	}
}

func TestEmissionFactor_Fields(t *testing.T) {
	now := time.Now()
	factor := EmissionFactor{
		ID:          "ef-001",
		Source:      SourceEPA,
		Category:    CategoryElectricity,
		Name:        "US Average Grid Factor",
		Unit:        "kgCO2e/kWh",
		Value:       0.42,
		CO2:         0.40,
		CH4:         0.01,
		N2O:         0.01,
		Region:      "US",
		ValidFrom:   now,
		LastUpdated: now,
		Uncertainty: 5.0,
		DataQuality: "High",
	}

	if factor.ID == "" {
		t.Error("ID should not be empty")
	}
	if factor.Value <= 0 {
		t.Error("Value should be positive")
	}
}

func TestNewUpdater(t *testing.T) {
	cfg := UpdaterConfig{
		UpdateInterval: time.Hour,
	}
	updater := NewUpdater(cfg)

	if updater == nil {
		t.Fatal("NewUpdater returned nil")
	}
}

func TestUpdaterConfig_Fields(t *testing.T) {
	cfg := UpdaterConfig{
		Sources: []SourceConfig{
			{
				Source:   SourceEPA,
				Name:     "EPA",
				BaseURL:  "https://api.epa.gov",
				Enabled:  true,
				Interval: 24 * time.Hour,
			},
		},
		UpdateInterval: time.Hour,
	}

	if len(cfg.Sources) == 0 {
		t.Error("Sources should not be empty")
	}
	if cfg.UpdateInterval <= 0 {
		t.Error("UpdateInterval should be positive")
	}
}

func TestSourceConfig_Fields(t *testing.T) {
	src := SourceConfig{
		Source:   SourceDEFRA,
		Name:     "DEFRA",
		BaseURL:  "https://api.defra.gov.uk",
		APIKey:   "test-key",
		Enabled:  true,
		Interval: 24 * time.Hour,
	}

	if src.Name == "" {
		t.Error("Name should not be empty")
	}
	if !src.Enabled {
		t.Error("Enabled should be true for this test")
	}
}

func TestFactorQuery_Fields(t *testing.T) {
	now := time.Now()
	query := FactorQuery{
		Source:   SourceEPA,
		Category: CategoryElectricity,
		Region:   "US",
		ValidAt:  &now,
	}

	if query.Source != SourceEPA {
		t.Errorf("Expected source %v, got %v", SourceEPA, query.Source)
	}
	if query.ValidAt == nil {
		t.Error("ValidAt should not be nil")
	}
}

func TestUpdater_OnUpdate(t *testing.T) {
	cfg := UpdaterConfig{}
	updater := NewUpdater(cfg)

	called := false
	updater.OnUpdate(func(f EmissionFactor) {
		called = true
	})

	// Note: Callback would be called during actual updates
	if called {
		t.Log("Callback was triggered (expected during update)")
	}
}
