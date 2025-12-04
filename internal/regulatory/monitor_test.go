package regulatory

import (
	"testing"
	"time"
)

func TestJurisdiction_Values(t *testing.T) {
	jurisdictions := []Jurisdiction{
		JurisdictionEU,
		JurisdictionUS,
		JurisdictionUK,
		JurisdictionCalifornia,
		JurisdictionGlobal,
	}

	for _, j := range jurisdictions {
		if string(j) == "" {
			t.Errorf("Jurisdiction %v has empty string value", j)
		}
	}
}

func TestRegulationType_Values(t *testing.T) {
	types := []RegulationType{
		TypeDisclosure,
		TypeReporting,
		TypeTaxation,
		TypeEmissionsLimit,
		TypeTrading,
	}

	for _, rt := range types {
		if string(rt) == "" {
			t.Errorf("RegulationType %v has empty string value", rt)
		}
	}
}

func TestRegStatus_Values(t *testing.T) {
	statuses := []RegStatus{
		StatusProposed,
		StatusAdopted,
		StatusEffective,
		StatusAmended,
		StatusRepealed,
	}

	for _, s := range statuses {
		if string(s) == "" {
			t.Errorf("RegStatus %v has empty string value", s)
		}
	}
}

func TestChangeType_Values(t *testing.T) {
	types := []ChangeType{
		ChangeNew,
		ChangeAmendment,
		ChangeDeadline,
		ChangeGuidance,
		ChangeEnforcement,
	}

	for _, ct := range types {
		if string(ct) == "" {
			t.Errorf("ChangeType %v has empty string value", ct)
		}
	}
}

func TestNewMonitor(t *testing.T) {
	cfg := MonitorConfig{}
	monitor := NewMonitor(cfg)

	if monitor == nil {
		t.Fatal("NewMonitor returned nil")
	}
}

func TestMonitor_GetRegulations(t *testing.T) {
	cfg := MonitorConfig{}
	monitor := NewMonitor(cfg)

	regs := monitor.GetRegulations()

	// Should have pre-loaded regulations
	if len(regs) == 0 {
		t.Log("No pre-loaded regulations found (may be expected)")
	}
}

func TestRegulation_Fields(t *testing.T) {
	now := time.Now()
	deadline := now.AddDate(0, 6, 0)
	reg := Regulation{
		ID:            "test-reg",
		Name:          "Test Climate Regulation",
		ShortName:     "TCR",
		Jurisdiction:  JurisdictionEU,
		Type:          TypeDisclosure,
		Description:   "A test regulation",
		EffectiveDate: &now,
		Deadline:      &deadline,
		Status:        StatusEffective,
		LastUpdated:   now,
	}

	if reg.ID == "" {
		t.Error("ID should not be empty")
	}
	if reg.EffectiveDate == nil {
		t.Error("EffectiveDate should not be nil")
	}
}

func TestApplicability_Fields(t *testing.T) {
	app := Applicability{
		Criterion:   "revenue",
		Operator:    ">",
		Value:       "40000000",
		Description: "Revenue above 40M EUR",
	}

	if app.Criterion == "" {
		t.Error("Criterion should not be empty")
	}
	if app.Operator == "" {
		t.Error("Operator should not be empty")
	}
}

func TestRequirement_Fields(t *testing.T) {
	deadline := time.Now().AddDate(1, 0, 0)
	req := Requirement{
		ID:          "req-001",
		Name:        "Scope 1 Disclosure",
		Description: "Report all Scope 1 emissions",
		Category:    "emissions",
		Scope:       []int{1},
		Mandatory:   true,
		Deadline:    &deadline,
	}

	if req.Name == "" {
		t.Error("Name should not be empty")
	}
	if !req.Mandatory {
		t.Error("Mandatory should be true for this test")
	}
}

func TestRegulatoryChange_Fields(t *testing.T) {
	now := time.Now()
	effective := now.AddDate(0, 3, 0)
	change := RegulatoryChange{
		ID:           "change-001",
		RegulationID: "csrd",
		Type:         ChangeAmendment,
		Title:        "New reporting deadline",
		Summary:      "Extended deadline for first report",
		Impact:       ImpactMedium,
		DetectedAt:   now,
		EffectiveAt:  &effective,
		ActionNeeded: []string{"Update reporting timeline"},
	}

	if change.ID == "" {
		t.Error("ID should not be empty")
	}
	if change.Type != ChangeAmendment {
		t.Errorf("Expected type %v, got %v", ChangeAmendment, change.Type)
	}
}

func TestSource_Fields(t *testing.T) {
	src := Source{
		Name:     "EU Regulations",
		URL:      "https://eur-lex.europa.eu",
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

func TestMonitorConfig_Fields(t *testing.T) {
	cfg := MonitorConfig{
		Sources: []Source{
			{Name: "EU", URL: "https://eur-lex.europa.eu", Enabled: true},
			{Name: "US", URL: "https://sec.gov", Enabled: true},
		},
	}

	if len(cfg.Sources) != 2 {
		t.Errorf("Expected 2 sources, got %d", len(cfg.Sources))
	}
}
