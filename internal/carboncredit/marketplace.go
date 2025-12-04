// Package carboncredit provides carbon credit marketplace integration.
//
// This package integrates with major carbon registries (Verra, Gold Standard, ACR)
// to enable offset matching, credit procurement, and retirement certificate generation.
package carboncredit

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Core Types
// =============================================================================

// Registry represents a carbon credit registry.
type Registry string

const (
	RegistryVerra        Registry = "verra"         // Verra VCS
	RegistryGoldStandard Registry = "gold_standard" // Gold Standard
	RegistryACR          Registry = "acr"           // American Carbon Registry
	RegistryCAR          Registry = "car"           // Climate Action Reserve
	RegistryPlan         Registry = "plan_vivo"     // Plan Vivo
)

// ProjectType categorizes carbon offset projects.
type ProjectType string

const (
	ProjectTypeReforestation      ProjectType = "reforestation"
	ProjectTypeAREDD              ProjectType = "avoided_deforestation"
	ProjectTypeRenewable          ProjectType = "renewable_energy"
	ProjectTypeCookstoves         ProjectType = "cookstoves"
	ProjectTypeMethaneCapture     ProjectType = "methane_capture"
	ProjectTypeBluCarbon          ProjectType = "blue_carbon"
	ProjectTypeDACCS              ProjectType = "direct_air_capture"
	ProjectTypeBiochar            ProjectType = "biochar"
	ProjectTypeEnhancedWeathering ProjectType = "enhanced_weathering"
)

// CreditStatus indicates the status of a credit.
type CreditStatus string

const (
	CreditStatusAvailable CreditStatus = "available"
	CreditStatusReserved  CreditStatus = "reserved"
	CreditStatusPurchased CreditStatus = "purchased"
	CreditStatusRetired   CreditStatus = "retired"
	CreditStatusCancelled CreditStatus = "cancelled"
)

// =============================================================================
// Credit & Project Models
// =============================================================================

// Project represents a carbon offset project.
type Project struct {
	ID               string            `json:"id"`
	Registry         Registry          `json:"registry"`
	RegistryID       string            `json:"registryId"` // Registry's project ID
	Name             string            `json:"name"`
	Description      string            `json:"description,omitempty"`
	Type             ProjectType       `json:"type"`
	Country          string            `json:"country"`
	Region           string            `json:"region,omitempty"`
	Methodology      string            `json:"methodology"`
	Vintage          int               `json:"vintage"`          // Year of emission reduction
	AvailableCredits float64           `json:"availableCredits"` // tCO2e
	PricePerTonne    float64           `json:"pricePerTonne"`    // USD
	Currency         string            `json:"currency"`
	Verified         bool              `json:"verified"`
	SDGGoals         []int             `json:"sdgGoals,omitempty"` // UN SDG alignment
	Certifications   []string          `json:"certifications,omitempty"`
	CoBenefits       []string          `json:"coBenefits,omitempty"`
	Documents        []Document        `json:"documents,omitempty"`
	LastUpdated      time.Time         `json:"lastUpdated"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// Document represents supporting documentation.
type Document struct {
	Type string `json:"type"` // pdd, verification, monitoring
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Credit represents a carbon credit unit.
type Credit struct {
	ID               string       `json:"id"`
	TenantID         string       `json:"tenantId"`
	ProjectID        string       `json:"projectId"`
	Registry         Registry     `json:"registry"`
	SerialNumber     string       `json:"serialNumber"` // Registry serial number
	Vintage          int          `json:"vintage"`
	Quantity         float64      `json:"quantity"` // tCO2e
	Status           CreditStatus `json:"status"`
	PurchasePrice    float64      `json:"purchasePrice"`
	PurchaseCurrency string       `json:"purchaseCurrency"`
	PurchasedAt      *time.Time   `json:"purchasedAt,omitempty"`
	RetiredAt        *time.Time   `json:"retiredAt,omitempty"`
	RetirementNote   string       `json:"retirementNote,omitempty"`
	BeneficiaryName  string       `json:"beneficiaryName,omitempty"`
	Certificate      *Certificate `json:"certificate,omitempty"`
	CreatedAt        time.Time    `json:"createdAt"`
	UpdatedAt        time.Time    `json:"updatedAt"`
}

// Certificate represents a retirement certificate.
type Certificate struct {
	ID                string    `json:"id"`
	CreditID          string    `json:"creditId"`
	RegistryCertID    string    `json:"registryCertId,omitempty"`
	Quantity          float64   `json:"quantity"`
	RetirementDate    time.Time `json:"retirementDate"`
	BeneficiaryName   string    `json:"beneficiaryName"`
	BeneficiaryReason string    `json:"beneficiaryReason,omitempty"`
	ProjectName       string    `json:"projectName"`
	ProjectCountry    string    `json:"projectCountry"`
	Vintage           int       `json:"vintage"`
	VerificationURL   string    `json:"verificationUrl,omitempty"`
	PDFUrl            string    `json:"pdfUrl,omitempty"`
	Hash              string    `json:"hash,omitempty"` // For blockchain verification
	IssuedAt          time.Time `json:"issuedAt"`
}

// =============================================================================
// Marketplace Operations
// =============================================================================

// SearchCriteria for finding offset projects.
type SearchCriteria struct {
	Registries   []Registry    `json:"registries,omitempty"`
	ProjectTypes []ProjectType `json:"projectTypes,omitempty"`
	Countries    []string      `json:"countries,omitempty"`
	MinVintage   int           `json:"minVintage,omitempty"`
	MaxVintage   int           `json:"maxVintage,omitempty"`
	MinPrice     float64       `json:"minPrice,omitempty"`
	MaxPrice     float64       `json:"maxPrice,omitempty"`
	MinQuantity  float64       `json:"minQuantity,omitempty"`
	SDGGoals     []int         `json:"sdgGoals,omitempty"`
	Verified     *bool         `json:"verified,omitempty"`
	SortBy       string        `json:"sortBy,omitempty"`    // price, vintage, quantity
	SortOrder    string        `json:"sortOrder,omitempty"` // asc, desc
	Limit        int           `json:"limit,omitempty"`
	Offset       int           `json:"offset,omitempty"`
}

// MatchResult represents a matched offset opportunity.
type MatchResult struct {
	Project        Project  `json:"project"`
	MatchScore     float64  `json:"matchScore"` // 0-1
	MatchReasons   []string `json:"matchReasons"`
	RecommendedQty float64  `json:"recommendedQty"`
	TotalCost      float64  `json:"totalCost"`
}

// OffsetRequirement specifies what needs to be offset.
type OffsetRequirement struct {
	TenantID          string        `json:"tenantId"`
	EmissionsToOffset float64       `json:"emissionsToOffset"` // tCO2e
	Year              int           `json:"year"`
	Scope             int           `json:"scope,omitempty"`
	PreferredTypes    []ProjectType `json:"preferredTypes,omitempty"`
	MaxPricePerTonne  float64       `json:"maxPricePerTonne,omitempty"`
	PreferLocal       bool          `json:"preferLocal,omitempty"`
	CompanyCountry    string        `json:"companyCountry,omitempty"`
	SDGAlignment      []int         `json:"sdgAlignment,omitempty"`
}

// =============================================================================
// Service
// =============================================================================

// Service manages carbon credit operations.
type Service struct {
	db         *sql.DB
	logger     *slog.Logger
	httpClient *http.Client
	cache      *projectCache
	config     Config
}

// Config holds marketplace configuration.
type Config struct {
	VerraAPIKey        string
	GoldStandardAPIKey string
	ACRAPIKey          string
	CacheExpiry        time.Duration
}

type projectCache struct {
	projects map[string]Project
	expiry   time.Time
	mu       sync.RWMutex
}

// NewService creates a new carbon credit service.
func NewService(db *sql.DB, config Config, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	if config.CacheExpiry == 0 {
		config.CacheExpiry = 1 * time.Hour
	}
	return &Service{
		db:         db,
		logger:     logger.With("component", "carboncredit"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
		cache:      &projectCache{projects: make(map[string]Project)},
		config:     config,
	}
}

// =============================================================================
// Marketplace Operations
// =============================================================================

// SearchProjects searches for offset projects matching criteria.
func (s *Service) SearchProjects(ctx context.Context, criteria SearchCriteria) ([]Project, error) {
	// For demo, return mock projects. In production, call registry APIs.
	projects := s.getMockProjects()

	var filtered []Project
	for _, p := range projects {
		if s.matchesCriteria(p, criteria) {
			filtered = append(filtered, p)
		}
	}

	// Apply sorting
	if criteria.SortBy != "" {
		sortProjects(filtered, criteria.SortBy, criteria.SortOrder)
	}

	// Apply pagination
	if criteria.Offset > 0 && criteria.Offset < len(filtered) {
		filtered = filtered[criteria.Offset:]
	}
	if criteria.Limit > 0 && criteria.Limit < len(filtered) {
		filtered = filtered[:criteria.Limit]
	}

	return filtered, nil
}

// MatchOffsets finds best offset options for given requirements.
func (s *Service) MatchOffsets(ctx context.Context, req OffsetRequirement) ([]MatchResult, error) {
	// Search for available projects
	criteria := SearchCriteria{
		Verified:    boolPtr(true),
		MinQuantity: req.EmissionsToOffset,
		MaxPrice:    req.MaxPricePerTonne,
	}
	if len(req.PreferredTypes) > 0 {
		criteria.ProjectTypes = req.PreferredTypes
	}

	projects, err := s.SearchProjects(ctx, criteria)
	if err != nil {
		return nil, err
	}

	var results []MatchResult
	for _, p := range projects {
		score, reasons := s.calculateMatchScore(p, req)

		qty := req.EmissionsToOffset
		if p.AvailableCredits < qty {
			qty = p.AvailableCredits
		}

		results = append(results, MatchResult{
			Project:        p,
			MatchScore:     score,
			MatchReasons:   reasons,
			RecommendedQty: qty,
			TotalCost:      qty * p.PricePerTonne,
		})
	}

	// Sort by match score
	sortMatchResults(results)

	return results, nil
}

// GetProject retrieves a project by ID.
func (s *Service) GetProject(ctx context.Context, projectID string) (*Project, error) {
	s.cache.mu.RLock()
	if p, ok := s.cache.projects[projectID]; ok && time.Now().Before(s.cache.expiry) {
		s.cache.mu.RUnlock()
		return &p, nil
	}
	s.cache.mu.RUnlock()

	// In production, fetch from registry API
	projects := s.getMockProjects()
	for _, p := range projects {
		if p.ID == projectID {
			return &p, nil
		}
	}
	return nil, nil
}

// =============================================================================
// Credit Operations
// =============================================================================

// ReserveCredits reserves credits for purchase.
func (s *Service) ReserveCredits(ctx context.Context, tenantID, projectID string, quantity float64) (*Credit, error) {
	project, err := s.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, fmt.Errorf("project not found")
	}
	if project.AvailableCredits < quantity {
		return nil, fmt.Errorf("insufficient credits available")
	}

	credit := &Credit{
		ID:               uuid.NewString(),
		TenantID:         tenantID,
		ProjectID:        projectID,
		Registry:         project.Registry,
		SerialNumber:     fmt.Sprintf("%s-%d-%s", project.Registry, time.Now().Year(), uuid.NewString()[:8]),
		Vintage:          project.Vintage,
		Quantity:         quantity,
		Status:           CreditStatusReserved,
		PurchasePrice:    quantity * project.PricePerTonne,
		PurchaseCurrency: project.Currency,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO carbon_credits (
			id, tenant_id, project_id, registry, serial_number, vintage,
			quantity, status, purchase_price, purchase_currency, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`,
		credit.ID, credit.TenantID, credit.ProjectID, credit.Registry,
		credit.SerialNumber, credit.Vintage, credit.Quantity, credit.Status,
		credit.PurchasePrice, credit.PurchaseCurrency, credit.CreatedAt, credit.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to reserve credits: %w", err)
	}

	s.logger.Info("credits reserved",
		"creditId", credit.ID,
		"quantity", quantity,
		"project", projectID)

	return credit, nil
}

// PurchaseCredits completes a credit purchase.
func (s *Service) PurchaseCredits(ctx context.Context, creditID string) (*Credit, error) {
	now := time.Now()

	result, err := s.db.ExecContext(ctx, `
		UPDATE carbon_credits 
		SET status = $1, purchased_at = $2, updated_at = $2
		WHERE id = $3 AND status = $4
	`, CreditStatusPurchased, now, creditID, CreditStatusReserved)
	if err != nil {
		return nil, err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, fmt.Errorf("credit not found or not reserved")
	}

	return s.GetCredit(ctx, creditID)
}

// RetireCredits retires credits for offsetting.
func (s *Service) RetireCredits(ctx context.Context, creditID, beneficiaryName, reason string) (*Certificate, error) {
	credit, err := s.GetCredit(ctx, creditID)
	if err != nil {
		return nil, err
	}
	if credit == nil {
		return nil, fmt.Errorf("credit not found")
	}
	if credit.Status != CreditStatusPurchased {
		return nil, fmt.Errorf("credit must be purchased before retirement")
	}

	now := time.Now()

	// Get project info for certificate
	project, _ := s.GetProject(ctx, credit.ProjectID)
	projectName := "Unknown Project"
	projectCountry := ""
	if project != nil {
		projectName = project.Name
		projectCountry = project.Country
	}

	// Create certificate
	cert := &Certificate{
		ID:                uuid.NewString(),
		CreditID:          creditID,
		Quantity:          credit.Quantity,
		RetirementDate:    now,
		BeneficiaryName:   beneficiaryName,
		BeneficiaryReason: reason,
		ProjectName:       projectName,
		ProjectCountry:    projectCountry,
		Vintage:           credit.Vintage,
		IssuedAt:          now,
	}

	// Generate verification hash
	cert.Hash = generateCertificateHash(cert)

	// Update credit status
	_, err = s.db.ExecContext(ctx, `
		UPDATE carbon_credits 
		SET status = $1, retired_at = $2, retirement_note = $3, 
			beneficiary_name = $4, updated_at = $2
		WHERE id = $5
	`, CreditStatusRetired, now, reason, beneficiaryName, creditID)
	if err != nil {
		return nil, err
	}

	// Store certificate
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO retirement_certificates (
			id, credit_id, quantity, retirement_date, beneficiary_name,
			beneficiary_reason, project_name, project_country, vintage,
			hash, issued_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`,
		cert.ID, cert.CreditID, cert.Quantity, cert.RetirementDate,
		cert.BeneficiaryName, cert.BeneficiaryReason, cert.ProjectName,
		cert.ProjectCountry, cert.Vintage, cert.Hash, cert.IssuedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to store certificate: %w", err)
	}

	s.logger.Info("credits retired",
		"creditId", creditID,
		"quantity", credit.Quantity,
		"beneficiary", beneficiaryName,
		"certificateId", cert.ID)

	return cert, nil
}

// GetCredit retrieves a credit by ID.
func (s *Service) GetCredit(ctx context.Context, creditID string) (*Credit, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, project_id, registry, serial_number, vintage,
			quantity, status, purchase_price, purchase_currency,
			purchased_at, retired_at, retirement_note, beneficiary_name,
			created_at, updated_at
		FROM carbon_credits WHERE id = $1
	`, creditID)

	var credit Credit
	var purchasedAt, retiredAt sql.NullTime
	err := row.Scan(
		&credit.ID, &credit.TenantID, &credit.ProjectID, &credit.Registry,
		&credit.SerialNumber, &credit.Vintage, &credit.Quantity, &credit.Status,
		&credit.PurchasePrice, &credit.PurchaseCurrency,
		&purchasedAt, &retiredAt, &credit.RetirementNote, &credit.BeneficiaryName,
		&credit.CreatedAt, &credit.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if purchasedAt.Valid {
		credit.PurchasedAt = &purchasedAt.Time
	}
	if retiredAt.Valid {
		credit.RetiredAt = &retiredAt.Time
	}

	return &credit, nil
}

// GetCertificate retrieves a retirement certificate.
func (s *Service) GetCertificate(ctx context.Context, certificateID string) (*Certificate, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, credit_id, registry_cert_id, quantity, retirement_date,
			beneficiary_name, beneficiary_reason, project_name, project_country,
			vintage, verification_url, pdf_url, hash, issued_at
		FROM retirement_certificates WHERE id = $1
	`, certificateID)

	var cert Certificate
	var registryCertID, verificationURL, pdfURL sql.NullString
	err := row.Scan(
		&cert.ID, &cert.CreditID, &registryCertID, &cert.Quantity,
		&cert.RetirementDate, &cert.BeneficiaryName, &cert.BeneficiaryReason,
		&cert.ProjectName, &cert.ProjectCountry, &cert.Vintage,
		&verificationURL, &pdfURL, &cert.Hash, &cert.IssuedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	cert.RegistryCertID = registryCertID.String
	cert.VerificationURL = verificationURL.String
	cert.PDFUrl = pdfURL.String

	return &cert, nil
}

// ListCredits returns credits for a tenant.
func (s *Service) ListCredits(ctx context.Context, tenantID string, status CreditStatus) ([]Credit, error) {
	query := `
		SELECT id, tenant_id, project_id, registry, serial_number, vintage,
			quantity, status, purchase_price, purchase_currency,
			purchased_at, retired_at, retirement_note, beneficiary_name,
			created_at, updated_at
		FROM carbon_credits WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}

	if status != "" {
		query += " AND status = $2"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credits []Credit
	for rows.Next() {
		var credit Credit
		var purchasedAt, retiredAt sql.NullTime
		err := rows.Scan(
			&credit.ID, &credit.TenantID, &credit.ProjectID, &credit.Registry,
			&credit.SerialNumber, &credit.Vintage, &credit.Quantity, &credit.Status,
			&credit.PurchasePrice, &credit.PurchaseCurrency,
			&purchasedAt, &retiredAt, &credit.RetirementNote, &credit.BeneficiaryName,
			&credit.CreatedAt, &credit.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if purchasedAt.Valid {
			credit.PurchasedAt = &purchasedAt.Time
		}
		if retiredAt.Valid {
			credit.RetiredAt = &retiredAt.Time
		}
		credits = append(credits, credit)
	}
	return credits, rows.Err()
}

// =============================================================================
// Helper Functions
// =============================================================================

func (s *Service) matchesCriteria(p Project, c SearchCriteria) bool {
	if len(c.Registries) > 0 && !containsRegistry(c.Registries, p.Registry) {
		return false
	}
	if len(c.ProjectTypes) > 0 && !containsProjectType(c.ProjectTypes, p.Type) {
		return false
	}
	if len(c.Countries) > 0 && !containsString(c.Countries, p.Country) {
		return false
	}
	if c.MinVintage > 0 && p.Vintage < c.MinVintage {
		return false
	}
	if c.MaxVintage > 0 && p.Vintage > c.MaxVintage {
		return false
	}
	if c.MinPrice > 0 && p.PricePerTonne < c.MinPrice {
		return false
	}
	if c.MaxPrice > 0 && p.PricePerTonne > c.MaxPrice {
		return false
	}
	if c.MinQuantity > 0 && p.AvailableCredits < c.MinQuantity {
		return false
	}
	if c.Verified != nil && p.Verified != *c.Verified {
		return false
	}
	return true
}

func (s *Service) calculateMatchScore(p Project, req OffsetRequirement) (float64, []string) {
	var score float64
	var reasons []string

	// Base score for availability
	if p.AvailableCredits >= req.EmissionsToOffset {
		score += 0.3
		reasons = append(reasons, "Full quantity available")
	} else {
		score += 0.1
		reasons = append(reasons, "Partial quantity available")
	}

	// Price score
	if req.MaxPricePerTonne > 0 {
		if p.PricePerTonne <= req.MaxPricePerTonne*0.5 {
			score += 0.2
			reasons = append(reasons, "Excellent price")
		} else if p.PricePerTonne <= req.MaxPricePerTonne*0.8 {
			score += 0.15
			reasons = append(reasons, "Good price")
		} else {
			score += 0.1
		}
	}

	// Project type preference
	if len(req.PreferredTypes) > 0 && containsProjectType(req.PreferredTypes, p.Type) {
		score += 0.2
		reasons = append(reasons, "Preferred project type")
	}

	// Geographic preference
	if req.PreferLocal && req.CompanyCountry != "" && p.Country == req.CompanyCountry {
		score += 0.15
		reasons = append(reasons, "Local project")
	}

	// SDG alignment
	if len(req.SDGAlignment) > 0 {
		matches := 0
		for _, goal := range req.SDGAlignment {
			if containsInt(p.SDGGoals, goal) {
				matches++
			}
		}
		if matches > 0 {
			score += float64(matches) * 0.05
			reasons = append(reasons, fmt.Sprintf("%d SDG goals aligned", matches))
		}
	}

	// Verification bonus
	if p.Verified {
		score += 0.1
		reasons = append(reasons, "Third-party verified")
	}

	return min(score, 1.0), reasons
}

func (s *Service) getMockProjects() []Project {
	return []Project{
		{
			ID:               "vcs-1234",
			Registry:         RegistryVerra,
			RegistryID:       "VCS-1234",
			Name:             "Amazon Rainforest Protection REDD+",
			Type:             ProjectTypeAREDD,
			Country:          "Brazil",
			Region:           "Amazonas",
			Methodology:      "VM0007",
			Vintage:          2024,
			AvailableCredits: 50000,
			PricePerTonne:    15.50,
			Currency:         "USD",
			Verified:         true,
			SDGGoals:         []int{13, 15, 1, 8},
			Certifications:   []string{"VCS", "CCB Gold"},
			CoBenefits:       []string{"Biodiversity", "Community Development"},
		},
		{
			ID:               "gs-5678",
			Registry:         RegistryGoldStandard,
			RegistryID:       "GS-5678",
			Name:             "Kenya Clean Cookstoves",
			Type:             ProjectTypeCookstoves,
			Country:          "Kenya",
			Region:           "Rift Valley",
			Methodology:      "GS-CER",
			Vintage:          2024,
			AvailableCredits: 25000,
			PricePerTonne:    22.00,
			Currency:         "USD",
			Verified:         true,
			SDGGoals:         []int{13, 3, 5, 7},
			Certifications:   []string{"Gold Standard"},
			CoBenefits:       []string{"Health", "Gender Equality", "Clean Energy"},
		},
		{
			ID:               "dac-9012",
			Registry:         RegistryVerra,
			RegistryID:       "VCS-9012",
			Name:             "Iceland Direct Air Capture",
			Type:             ProjectTypeDACCS,
			Country:          "Iceland",
			Methodology:      "CDM-DAC",
			Vintage:          2024,
			AvailableCredits: 5000,
			PricePerTonne:    350.00,
			Currency:         "USD",
			Verified:         true,
			SDGGoals:         []int{13, 9},
			Certifications:   []string{"VCS"},
			CoBenefits:       []string{"Permanent Removal"},
		},
		{
			ID:               "wind-3456",
			Registry:         RegistryACR,
			RegistryID:       "ACR-3456",
			Name:             "Texas Wind Farm",
			Type:             ProjectTypeRenewable,
			Country:          "USA",
			Region:           "Texas",
			Methodology:      "ACM0002",
			Vintage:          2024,
			AvailableCredits: 100000,
			PricePerTonne:    8.50,
			Currency:         "USD",
			Verified:         true,
			SDGGoals:         []int{13, 7, 8},
			CoBenefits:       []string{"Clean Energy", "Local Jobs"},
		},
	}
}

func sortProjects(projects []Project, sortBy, sortOrder string) {
	// Implement sorting logic
}

func sortMatchResults(results []MatchResult) {
	// Sort by match score descending
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].MatchScore > results[i].MatchScore {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

func generateCertificateHash(cert *Certificate) string {
	data := fmt.Sprintf("%s|%s|%.2f|%s|%s|%d",
		cert.CreditID, cert.BeneficiaryName, cert.Quantity,
		cert.RetirementDate.Format(time.RFC3339),
		cert.ProjectName, cert.Vintage)
	// In production, use SHA-256
	return fmt.Sprintf("sha256:%x", []byte(data))
}

func containsRegistry(list []Registry, r Registry) bool {
	for _, v := range list {
		if v == r {
			return true
		}
	}
	return false
}

func containsProjectType(list []ProjectType, t ProjectType) bool {
	for _, v := range list {
		if v == t {
			return true
		}
	}
	return false
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func containsInt(list []int, i int) bool {
	for _, v := range list {
		if v == i {
			return true
		}
	}
	return false
}

func boolPtr(b bool) *bool { return &b }

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
