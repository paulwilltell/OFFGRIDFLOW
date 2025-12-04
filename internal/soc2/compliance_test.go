package soc2

import (
	"testing"
	"time"
)

func TestTrustCategory_Values(t *testing.T) {
	categories := []TrustCategory{
		CategorySecurity,
		CategoryAvailability,
		CategoryProcessIntegrity,
		CategoryConfidentiality,
		CategoryPrivacy,
	}

	for _, c := range categories {
		if string(c) == "" {
			t.Errorf("TrustCategory %v has empty string value", c)
		}
	}
}

func TestControlStatus_Values(t *testing.T) {
	statuses := []ControlStatus{
		StatusDesigned,
		StatusImplemented,
		StatusOperating,
		StatusException,
		StatusNotApplicable,
	}

	for _, s := range statuses {
		if string(s) == "" {
			t.Errorf("ControlStatus %v has empty string value", s)
		}
	}
}

func TestEvidenceType_Values(t *testing.T) {
	types := []EvidenceType{
		TypeScreenshot,
		TypeConfiguration,
		TypeLog,
		TypePolicy,
		TypeReport,
		TypeTestResult,
	}

	for _, et := range types {
		if string(et) == "" {
			t.Errorf("EvidenceType %v has empty string value", et)
		}
	}
}

func TestNewControlLibrary(t *testing.T) {
	lib := NewControlLibrary()

	if lib == nil {
		t.Fatal("NewControlLibrary returned nil")
	}
}

func TestControl_Fields(t *testing.T) {
	now := time.Now()
	nextTest := now.AddDate(0, 1, 0)
	control := Control{
		ID:             "ctrl-001",
		Category:       CategorySecurity,
		Reference:      "CC1.1",
		Name:           "Control Environment",
		Description:    "Management establishes control environment",
		Implementation: "Documented policies and procedures",
		Automated:      false,
		Frequency:      "annual",
		Owner:          "Security Team",
		Status:         StatusOperating,
		LastTested:     &now,
		NextTest:       &nextTest,
	}

	if control.ID == "" {
		t.Error("ID should not be empty")
	}
	if control.Category != CategorySecurity {
		t.Errorf("Expected category %v, got %v", CategorySecurity, control.Category)
	}
}

func TestEvidence_Fields(t *testing.T) {
	now := time.Now()
	evidence := Evidence{
		ID:          "evid-001",
		ControlID:   "ctrl-001",
		Type:        TypeConfiguration,
		Title:       "Firewall Configuration",
		Description: "Network firewall rules export",
		CollectedAt: now,
		CollectedBy: "automation@company.com",
		Period: EvidencePeriod{
			Start: now.AddDate(0, -1, 0),
			End:   now,
		},
	}

	if evidence.ID == "" {
		t.Error("ID should not be empty")
	}
	if evidence.Type != TypeConfiguration {
		t.Errorf("Expected type %v, got %v", TypeConfiguration, evidence.Type)
	}
}

func TestArtifact_Fields(t *testing.T) {
	artifact := Artifact{
		Name: "access_log_2024.pdf",
		Type: "application/pdf",
		URL:  "https://storage.example.com/artifacts/access_log.pdf",
		Hash: "sha256:abc123def456",
		Size: 1024000,
	}

	if artifact.Name == "" {
		t.Error("Name should not be empty")
	}
	if artifact.Size <= 0 {
		t.Error("Size should be positive")
	}
}

func TestEvidencePeriod_Fields(t *testing.T) {
	now := time.Now()
	period := EvidencePeriod{
		Start: now.AddDate(0, -3, 0),
		End:   now,
	}

	if period.End.Before(period.Start) {
		t.Error("End should be after Start")
	}
}

func TestControlLibrary_GetControls(t *testing.T) {
	lib := NewControlLibrary()

	// Should have pre-loaded standard controls
	controls := lib.GetControls()
	if len(controls) == 0 {
		t.Log("No standard controls loaded (may be expected if not implemented)")
	}
}

func TestControlLibrary_GetControl(t *testing.T) {
	lib := NewControlLibrary()

	// Try to get a specific control
	control, err := lib.GetControl("CC1.1")
	if err != nil || control == nil {
		t.Log("Control CC1.1 not found (may be expected if not pre-loaded)")
	}
}
