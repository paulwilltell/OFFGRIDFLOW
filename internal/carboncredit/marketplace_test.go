package carboncredit

import (
	"testing"
	"time"
)

func TestRegistry_Values(t *testing.T) {
	registries := []Registry{
		RegistryVerra,
		RegistryGoldStandard,
		RegistryACR,
		RegistryCAR,
		RegistryPlan,
	}

	for _, r := range registries {
		if string(r) == "" {
			t.Errorf("Registry %v has empty string value", r)
		}
	}
}

func TestProjectType_Values(t *testing.T) {
	types := []ProjectType{
		ProjectTypeReforestation,
		ProjectTypeAREDD,
		ProjectTypeRenewable,
		ProjectTypeCookstoves,
		ProjectTypeMethaneCapture,
		ProjectTypeBluCarbon,
		ProjectTypeDACCS,
		ProjectTypeBiochar,
		ProjectTypeEnhancedWeathering,
	}

	for _, pt := range types {
		if string(pt) == "" {
			t.Errorf("ProjectType %v has empty string value", pt)
		}
	}
}

func TestCreditStatus_Values(t *testing.T) {
	statuses := []CreditStatus{
		CreditStatusAvailable,
		CreditStatusReserved,
		CreditStatusPurchased,
		CreditStatusRetired,
		CreditStatusCancelled,
	}

	for _, s := range statuses {
		if string(s) == "" {
			t.Errorf("CreditStatus %v has empty string value", s)
		}
	}
}

func TestProject_Struct(t *testing.T) {
	project := Project{
		ID:               "proj-1",
		Registry:         RegistryVerra,
		RegistryID:       "VCS-1234",
		Name:             "Test Reforestation Project",
		Type:             ProjectTypeReforestation,
		Country:          "Brazil",
		Vintage:          2024,
		AvailableCredits: 10000.0,
		PricePerTonne:    15.50,
		Currency:         "USD",
		Verified:         true,
		SDGGoals:         []int{13, 15},
	}

	if project.ID == "" {
		t.Error("Project ID should not be empty")
	}
	if project.PricePerTonne <= 0 {
		t.Error("PricePerTonne should be positive")
	}
}

func TestCredit_Struct(t *testing.T) {
	now := time.Now()
	credit := Credit{
		ID:               "cred-1",
		TenantID:         "tenant-1",
		ProjectID:        "proj-1",
		Registry:         RegistryVerra,
		SerialNumber:     "VCS-1234-001",
		Vintage:          2024,
		Quantity:         100.0,
		Status:           CreditStatusAvailable,
		PurchasePrice:    1550.0,
		PurchaseCurrency: "USD",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if credit.Quantity <= 0 {
		t.Error("Quantity should be positive")
	}
	if credit.Status != CreditStatusAvailable {
		t.Errorf("Expected status %v, got %v", CreditStatusAvailable, credit.Status)
	}
}

func TestCertificate_Struct(t *testing.T) {
	cert := Certificate{
		ID:              "cert-1",
		CreditID:        "cred-1",
		Quantity:        100.0,
		RetirementDate:  time.Now(),
		BeneficiaryName: "Test Company",
		ProjectName:     "Amazon Reforestation",
		ProjectCountry:  "Brazil",
		Vintage:         2024,
		IssuedAt:        time.Now(),
	}

	if cert.ID == "" {
		t.Error("Certificate ID should not be empty")
	}
	if cert.Quantity <= 0 {
		t.Error("Quantity should be positive")
	}
}

func TestSearchCriteria_Struct(t *testing.T) {
	verified := true
	criteria := SearchCriteria{
		Registries:   []Registry{RegistryVerra, RegistryGoldStandard},
		ProjectTypes: []ProjectType{ProjectTypeReforestation},
		Countries:    []string{"Brazil", "Indonesia"},
		MinVintage:   2020,
		MaxVintage:   2024,
		MinPrice:     5.0,
		MaxPrice:     25.0,
		MinQuantity:  100.0,
		SDGGoals:     []int{13, 15},
		Verified:     &verified,
		SortBy:       "price",
		Limit:        10,
	}

	if len(criteria.Registries) != 2 {
		t.Errorf("Expected 2 registries, got %d", len(criteria.Registries))
	}
	if criteria.MinVintage >= criteria.MaxVintage {
		t.Error("MinVintage should be less than MaxVintage")
	}
}

func TestDocument_Struct(t *testing.T) {
	doc := Document{
		Type: "pdd",
		Name: "Project Design Document",
		URL:  "https://registry.example.com/docs/pdd.pdf",
	}

	if doc.Type == "" {
		t.Error("Document Type should not be empty")
	}
	if doc.URL == "" {
		t.Error("Document URL should not be empty")
	}
}

func TestConfig_Struct(t *testing.T) {
	cfg := Config{
		CacheExpiry: 1 * time.Hour,
		VerraAPIKey: "test-key",
	}

	if cfg.CacheExpiry <= 0 {
		t.Error("CacheExpiry should be positive")
	}
}
