package supplier

import (
	"testing"
	"time"
)

func TestSupplierTier_Values(t *testing.T) {
	tiers := []SupplierTier{
		TierStrategic,
		TierPreferred,
		TierApproved,
		TierPotential,
	}

	for _, tier := range tiers {
		if string(tier) == "" {
			t.Errorf("SupplierTier %v has empty string value", tier)
		}
	}
}

func TestSupplierStatus_Values(t *testing.T) {
	statuses := []SupplierStatus{
		StatusPending,
		StatusActive,
		StatusResponded,
		StatusVerified,
		StatusInactive,
	}

	for _, s := range statuses {
		if string(s) == "" {
			t.Errorf("SupplierStatus %v has empty string value", s)
		}
	}
}

func TestDataQuality_Values(t *testing.T) {
	qualities := []DataQuality{
		QualityPrimary,
		QualitySecondary,
		QualityDefault,
	}

	for _, q := range qualities {
		if string(q) == "" {
			t.Errorf("DataQuality %v has empty string value", q)
		}
	}
}

func TestNewService(t *testing.T) {
	cfg := ServiceConfig{}
	svc := NewService(cfg)

	if svc == nil {
		t.Fatal("NewService returned nil")
	}
}

func TestSupplier_Fields(t *testing.T) {
	now := time.Now()
	supplier := Supplier{
		ID:            "supplier-001",
		TenantID:      "tenant-1",
		Name:          "Acme Supplies",
		ContactName:   "John Doe",
		ContactEmail:  "john@acme.com",
		Industry:      "Manufacturing",
		Country:       "US",
		Tier:          TierStrategic,
		Status:        StatusActive,
		Categories:    []string{"components", "materials"},
		SpendAnnual:   1000000.0,
		SpendCurrency: "USD",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Verify all fields are set correctly
	if supplier.ID == "" {
		t.Error("ID should not be empty")
	}
	if supplier.TenantID != "tenant-1" {
		t.Errorf("Expected TenantID tenant-1, got %s", supplier.TenantID)
	}
	if supplier.Name != "Acme Supplies" {
		t.Errorf("Expected Name Acme Supplies, got %s", supplier.Name)
	}
	if supplier.ContactName != "John Doe" {
		t.Errorf("Expected ContactName John Doe, got %s", supplier.ContactName)
	}
	if supplier.ContactEmail != "john@acme.com" {
		t.Errorf("Expected ContactEmail john@acme.com, got %s", supplier.ContactEmail)
	}
	if supplier.Industry != "Manufacturing" {
		t.Errorf("Expected Industry Manufacturing, got %s", supplier.Industry)
	}
	if supplier.Country != "US" {
		t.Errorf("Expected Country US, got %s", supplier.Country)
	}
	if supplier.Tier != TierStrategic {
		t.Errorf("Expected tier %v, got %v", TierStrategic, supplier.Tier)
	}
	if supplier.Status != StatusActive {
		t.Errorf("Expected status %v, got %v", StatusActive, supplier.Status)
	}
	if len(supplier.Categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(supplier.Categories))
	}
	if supplier.SpendAnnual != 1000000.0 {
		t.Errorf("Expected SpendAnnual 1000000.0, got %f", supplier.SpendAnnual)
	}
	if supplier.SpendCurrency != "USD" {
		t.Errorf("Expected SpendCurrency USD, got %s", supplier.SpendCurrency)
	}
	if supplier.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if supplier.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestSupplierEmissions_Fields(t *testing.T) {
	now := time.Now()
	emissions := SupplierEmissions{
		Year:          2024,
		Scope1:        500.0,
		Scope2:        1000.0,
		Scope3:        2000.0,
		Total:         3500.0,
		IntensityUnit: "per_revenue",
		Intensity:     35.0,
		Methodology:   "GHG Protocol",
		DataQuality:   QualityPrimary,
		SubmittedAt:   now,
	}

	// Verify all fields are set correctly
	if emissions.Year != 2024 {
		t.Errorf("Expected Year 2024, got %d", emissions.Year)
	}
	if emissions.Scope1 != 500.0 {
		t.Errorf("Expected Scope1 500.0, got %f", emissions.Scope1)
	}
	if emissions.Scope2 != 1000.0 {
		t.Errorf("Expected Scope2 1000.0, got %f", emissions.Scope2)
	}
	if emissions.Scope3 != 2000.0 {
		t.Errorf("Expected Scope3 2000.0, got %f", emissions.Scope3)
	}
	if emissions.Total != 3500.0 {
		t.Errorf("Expected Total 3500.0, got %f", emissions.Total)
	}
	if emissions.IntensityUnit != "per_revenue" {
		t.Errorf("Expected IntensityUnit per_revenue, got %s", emissions.IntensityUnit)
	}
	if emissions.Intensity != 35.0 {
		t.Errorf("Expected Intensity 35.0, got %f", emissions.Intensity)
	}
	if emissions.Methodology != "GHG Protocol" {
		t.Errorf("Expected Methodology GHG Protocol, got %s", emissions.Methodology)
	}
	if emissions.DataQuality != QualityPrimary {
		t.Errorf("Expected quality %v, got %v", QualityPrimary, emissions.DataQuality)
	}
	if emissions.SubmittedAt.IsZero() {
		t.Error("SubmittedAt should not be zero")
	}
}

func TestEngagementStatus_Fields(t *testing.T) {
	now := time.Now()
	status := EngagementStatus{
		InvitedAt:       &now,
		LastContactAt:   &now,
		ResponseRate:    75.0,
		DataSubmissions: 3,
		TargetSet:       true,
		TargetReduction: 30.0,
		TargetYear:      2030,
		Progress:        10.0,
	}

	// Verify all fields are set correctly
	if status.InvitedAt == nil || status.InvitedAt.IsZero() {
		t.Error("InvitedAt should be set")
	}
	if status.LastContactAt == nil || status.LastContactAt.IsZero() {
		t.Error("LastContactAt should be set")
	}
	if status.ResponseRate != 75.0 {
		t.Errorf("Expected ResponseRate 75.0, got %f", status.ResponseRate)
	}
	if status.DataSubmissions != 3 {
		t.Errorf("Expected DataSubmissions 3, got %d", status.DataSubmissions)
	}
	if !status.TargetSet {
		t.Error("TargetSet should be true")
	}
	if status.TargetReduction != 30.0 {
		t.Errorf("Expected TargetReduction 30.0, got %f", status.TargetReduction)
	}
	if status.TargetYear != 2030 {
		t.Errorf("Expected TargetYear 2030, got %d", status.TargetYear)
	}
	if status.Progress != 10.0 {
		t.Errorf("Expected Progress 10.0, got %f", status.Progress)
	}
}

func TestInvitation_Fields(t *testing.T) {
	now := time.Now()
	deadline := now.AddDate(0, 0, 30)
	invitation := Invitation{
		ID:         "inv-001",
		TenantID:   "tenant-1",
		SupplierID: "supplier-001",
		Token:      "abc123token",
		Email:      "supplier@example.com",
		Message:    "Please provide emissions data",
		Deadline:   &deadline,
		SentAt:     now,
	}

	// Verify all fields are set correctly
	if invitation.ID == "" {
		t.Error("ID should not be empty")
	}
	if invitation.TenantID != "tenant-1" {
		t.Errorf("Expected TenantID tenant-1, got %s", invitation.TenantID)
	}
	if invitation.SupplierID != "supplier-001" {
		t.Errorf("Expected SupplierID supplier-001, got %s", invitation.SupplierID)
	}
	if invitation.Token == "" {
		t.Error("Token should not be empty")
	}
	if invitation.Email != "supplier@example.com" {
		t.Errorf("Expected Email supplier@example.com, got %s", invitation.Email)
	}
	if invitation.Message != "Please provide emissions data" {
		t.Errorf("Expected Message 'Please provide emissions data', got %s", invitation.Message)
	}
	if invitation.Deadline == nil || invitation.Deadline.IsZero() {
		t.Error("Deadline should be set")
	}
	if invitation.SentAt.IsZero() {
		t.Error("SentAt should not be zero")
	}
}

func TestService_AddSupplier(t *testing.T) {
	cfg := ServiceConfig{}
	svc := NewService(cfg)

	supplier := &Supplier{
		ID:       "supplier-001",
		TenantID: "tenant-1",
		Name:     "Test Supplier",
		Tier:     TierApproved,
		Status:   StatusPending,
	}

	err := svc.AddSupplier("tenant-1", supplier)
	if err != nil {
		t.Fatalf("AddSupplier failed: %v", err)
	}
}

func TestServiceConfig_Fields(t *testing.T) {
	cfg := ServiceConfig{
		// EmailSender would be set in production
	}

	// Config with nil EmailSender should still work
	svc := NewService(cfg)
	if svc == nil {
		t.Fatal("Service should be created even without EmailSender")
	}
}
