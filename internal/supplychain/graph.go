// Package supplychain provides Scope 3 supply chain emissions graph and management.
//
// This package enables visualization of supplier emissions networks, identification
// of high-impact suppliers, and cascading reduction targets through the value chain.
package supplychain

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Core Types
// =============================================================================

// Supplier represents a supplier in the value chain.
type Supplier struct {
	ID                string            `json:"id"`
	TenantID          string            `json:"tenantId"`
	Name              string            `json:"name"`
	Industry          string            `json:"industry,omitempty"`
	Country           string            `json:"country,omitempty"`
	Region            string            `json:"region,omitempty"`
	Tier              int               `json:"tier"` // 1 = direct, 2 = tier 2, etc.
	Category          Scope3Category    `json:"category"`
	Status            SupplierStatus    `json:"status"`
	EmissionsKgCO2e   float64           `json:"emissionsKgCo2e"`
	EmissionsFactor   float64           `json:"emissionsFactor,omitempty"` // Per unit
	SpendAmount       float64           `json:"spendAmount,omitempty"`
	SpendCurrency     string            `json:"spendCurrency,omitempty"`
	DataQuality       DataQualityLevel  `json:"dataQuality"`
	LastUpdated       time.Time         `json:"lastUpdated"`
	ContactEmail      string            `json:"contactEmail,omitempty"`
	EngagementStatus  EngagementStatus  `json:"engagementStatus"`
	ReductionTarget   float64           `json:"reductionTarget,omitempty"` // Percentage
	ReductionDeadline *time.Time        `json:"reductionDeadline,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty"`
	CreatedAt         time.Time         `json:"createdAt"`
	UpdatedAt         time.Time         `json:"updatedAt"`
}

// Scope3Category represents GHG Protocol Scope 3 categories.
type Scope3Category int

const (
	CategoryPurchasedGoods         Scope3Category = 1
	CategoryCapitalGoods           Scope3Category = 2
	CategoryFuelEnergy             Scope3Category = 3
	CategoryUpstreamTransport      Scope3Category = 4
	CategoryWasteGenerated         Scope3Category = 5
	CategoryBusinessTravel         Scope3Category = 6
	CategoryEmployeeCommuting      Scope3Category = 7
	CategoryUpstreamLeasedAssets   Scope3Category = 8
	CategoryDownstreamTransport    Scope3Category = 9
	CategoryProcessingProducts     Scope3Category = 10
	CategoryUseOfProducts          Scope3Category = 11
	CategoryEndOfLifeTreatment     Scope3Category = 12
	CategoryDownstreamLeasedAssets Scope3Category = 13
	CategoryFranchises             Scope3Category = 14
	CategoryInvestments            Scope3Category = 15
)

// String returns human-readable category name.
func (c Scope3Category) String() string {
	names := map[Scope3Category]string{
		1:  "Purchased Goods and Services",
		2:  "Capital Goods",
		3:  "Fuel and Energy Related Activities",
		4:  "Upstream Transportation and Distribution",
		5:  "Waste Generated in Operations",
		6:  "Business Travel",
		7:  "Employee Commuting",
		8:  "Upstream Leased Assets",
		9:  "Downstream Transportation and Distribution",
		10: "Processing of Sold Products",
		11: "Use of Sold Products",
		12: "End-of-Life Treatment of Sold Products",
		13: "Downstream Leased Assets",
		14: "Franchises",
		15: "Investments",
	}
	if name, ok := names[c]; ok {
		return name
	}
	return fmt.Sprintf("Category %d", c)
}

// SupplierStatus indicates data collection status.
type SupplierStatus string

const (
	StatusPending    SupplierStatus = "pending"
	StatusActive     SupplierStatus = "active"
	StatusInactive   SupplierStatus = "inactive"
	StatusOnboarding SupplierStatus = "onboarding"
)

// DataQualityLevel indicates the quality of emissions data.
type DataQualityLevel string

const (
	DataQualityPrimary   DataQualityLevel = "primary"   // Direct from supplier
	DataQualitySecondary DataQualityLevel = "secondary" // Industry average
	DataQualityEstimate  DataQualityLevel = "estimate"  // Spend-based estimate
	DataQualityUnknown   DataQualityLevel = "unknown"
)

// EngagementStatus tracks supplier engagement progress.
type EngagementStatus string

const (
	EngagementNotStarted EngagementStatus = "not_started"
	EngagementContacted  EngagementStatus = "contacted"
	EngagementResponded  EngagementStatus = "responded"
	EngagementDataShared EngagementStatus = "data_shared"
	EngagementTargetSet  EngagementStatus = "target_set"
	EngagementOnTrack    EngagementStatus = "on_track"
	EngagementOffTrack   EngagementStatus = "off_track"
)

// =============================================================================
// Supply Chain Graph
// =============================================================================

// Edge represents a relationship between entities.
type Edge struct {
	ID              string    `json:"id"`
	FromID          string    `json:"fromId"`
	FromType        string    `json:"fromType"` // "tenant" or "supplier"
	ToID            string    `json:"toId"`
	ToType          string    `json:"toType"`
	RelationType    string    `json:"relationType"` // "purchases_from", "supplies_to"
	EmissionsKgCO2e float64   `json:"emissionsKgCo2e"`
	Percentage      float64   `json:"percentage"` // % of total Scope 3
	CreatedAt       time.Time `json:"createdAt"`
}

// Graph represents the supply chain network.
type Graph struct {
	TenantID  string              `json:"tenantId"`
	Suppliers map[string]Supplier `json:"suppliers"`
	Edges     []Edge              `json:"edges"`
	Stats     GraphStats          `json:"stats"`
	UpdatedAt time.Time           `json:"updatedAt"`
}

// GraphStats provides aggregate statistics.
type GraphStats struct {
	TotalSuppliers       int                        `json:"totalSuppliers"`
	TotalEmissionsKgCO2e float64                    `json:"totalEmissionsKgCo2e"`
	ByCategory           map[Scope3Category]float64 `json:"byCategory"`
	ByTier               map[int]float64            `json:"byTier"`
	DataQualityDist      map[DataQualityLevel]int   `json:"dataQualityDist"`
	EngagementDist       map[EngagementStatus]int   `json:"engagementDist"`
	TopContributors      []SupplierContribution     `json:"topContributors"`
}

// SupplierContribution shows a supplier's emissions contribution.
type SupplierContribution struct {
	SupplierID      string  `json:"supplierId"`
	SupplierName    string  `json:"supplierName"`
	EmissionsKgCO2e float64 `json:"emissionsKgCo2e"`
	Percentage      float64 `json:"percentage"`
}

// =============================================================================
// Engagement & Outreach
// =============================================================================

// OutreachCampaign represents a supplier engagement campaign.
type OutreachCampaign struct {
	ID           string             `json:"id"`
	TenantID     string             `json:"tenantId"`
	Name         string             `json:"name"`
	Description  string             `json:"description,omitempty"`
	SupplierIDs  []string           `json:"supplierIds"`
	Status       CampaignStatus     `json:"status"`
	EmailSubject string             `json:"emailSubject"`
	EmailBody    string             `json:"emailBody"`
	SentAt       *time.Time         `json:"sentAt,omitempty"`
	ResponseRate float64            `json:"responseRate"`
	Responses    []CampaignResponse `json:"responses,omitempty"`
	CreatedAt    time.Time          `json:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt"`
}

// CampaignStatus tracks campaign progress.
type CampaignStatus string

const (
	CampaignDraft     CampaignStatus = "draft"
	CampaignScheduled CampaignStatus = "scheduled"
	CampaignSent      CampaignStatus = "sent"
	CampaignCompleted CampaignStatus = "completed"
)

// CampaignResponse tracks individual supplier responses.
type CampaignResponse struct {
	SupplierID   string    `json:"supplierId"`
	Status       string    `json:"status"` // opened, clicked, responded, submitted
	RespondedAt  time.Time `json:"respondedAt,omitempty"`
	DataProvided bool      `json:"dataProvided"`
}

// =============================================================================
// Reduction Targets
// =============================================================================

// ReductionTarget represents a cascaded reduction target.
type ReductionTarget struct {
	ID                string       `json:"id"`
	TenantID          string       `json:"tenantId"`
	SupplierID        string       `json:"supplierId"`
	BaselineYear      int          `json:"baselineYear"`
	BaselineEmissions float64      `json:"baselineEmissions"`
	TargetYear        int          `json:"targetYear"`
	TargetReduction   float64      `json:"targetReduction"` // Percentage
	CurrentProgress   float64      `json:"currentProgress"` // Percentage achieved
	Status            TargetStatus `json:"status"`
	Milestones        []Milestone  `json:"milestones,omitempty"`
	CreatedAt         time.Time    `json:"createdAt"`
	UpdatedAt         time.Time    `json:"updatedAt"`
}

// TargetStatus indicates reduction target progress.
type TargetStatus string

const (
	TargetPending  TargetStatus = "pending"
	TargetAccepted TargetStatus = "accepted"
	TargetOnTrack  TargetStatus = "on_track"
	TargetAtRisk   TargetStatus = "at_risk"
	TargetOffTrack TargetStatus = "off_track"
	TargetAchieved TargetStatus = "achieved"
)

// Milestone represents an interim target.
type Milestone struct {
	Year     int     `json:"year"`
	Target   float64 `json:"target"` // Cumulative reduction %
	Achieved float64 `json:"achieved"`
	Status   string  `json:"status"`
}

// =============================================================================
// Service
// =============================================================================

// Service manages supply chain data and operations.
type Service struct {
	db     *sql.DB
	logger *slog.Logger
	cache  map[string]*Graph
	mu     sync.RWMutex
}

// NewService creates a new supply chain service.
func NewService(db *sql.DB, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		db:     db,
		logger: logger.With("component", "supplychain"),
		cache:  make(map[string]*Graph),
	}
}

// =============================================================================
// Supplier Management
// =============================================================================

// AddSupplier adds a new supplier to the chain.
func (s *Service) AddSupplier(ctx context.Context, supplier Supplier) (*Supplier, error) {
	if supplier.ID == "" {
		supplier.ID = uuid.NewString()
	}
	supplier.CreatedAt = time.Now()
	supplier.UpdatedAt = time.Now()

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO suppliers (
			id, tenant_id, name, industry, country, region, tier,
			category, status, emissions_kg_co2e, spend_amount, spend_currency,
			data_quality, contact_email, engagement_status, reduction_target,
			metadata, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
	`,
		supplier.ID, supplier.TenantID, supplier.Name, supplier.Industry,
		supplier.Country, supplier.Region, supplier.Tier, supplier.Category,
		supplier.Status, supplier.EmissionsKgCO2e, supplier.SpendAmount,
		supplier.SpendCurrency, supplier.DataQuality, supplier.ContactEmail,
		supplier.EngagementStatus, supplier.ReductionTarget,
		mustJSON(supplier.Metadata), supplier.CreatedAt, supplier.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add supplier: %w", err)
	}

	s.invalidateCache(supplier.TenantID)
	return &supplier, nil
}

// GetSupplier retrieves a supplier by ID.
func (s *Service) GetSupplier(ctx context.Context, tenantID, supplierID string) (*Supplier, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, industry, country, region, tier,
			category, status, emissions_kg_co2e, spend_amount, spend_currency,
			data_quality, contact_email, engagement_status, reduction_target,
			metadata, created_at, updated_at
		FROM suppliers WHERE id = $1 AND tenant_id = $2
	`, supplierID, tenantID)

	var sup Supplier
	var metadata []byte
	err := row.Scan(
		&sup.ID, &sup.TenantID, &sup.Name, &sup.Industry, &sup.Country,
		&sup.Region, &sup.Tier, &sup.Category, &sup.Status,
		&sup.EmissionsKgCO2e, &sup.SpendAmount, &sup.SpendCurrency,
		&sup.DataQuality, &sup.ContactEmail, &sup.EngagementStatus,
		&sup.ReductionTarget, &metadata, &sup.CreatedAt, &sup.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(metadata, &sup.Metadata)
	return &sup, nil
}

// ListSuppliers returns all suppliers for a tenant.
func (s *Service) ListSuppliers(ctx context.Context, tenantID string, opts ListOptions) ([]Supplier, error) {
	query := `
		SELECT id, tenant_id, name, industry, country, region, tier,
			category, status, emissions_kg_co2e, spend_amount, spend_currency,
			data_quality, contact_email, engagement_status, reduction_target,
			metadata, created_at, updated_at
		FROM suppliers WHERE tenant_id = $1
		ORDER BY emissions_kg_co2e DESC
		LIMIT $2 OFFSET $3
	`
	limit := opts.Limit
	if limit == 0 {
		limit = 100
	}

	rows, err := s.db.QueryContext(ctx, query, tenantID, limit, opts.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suppliers []Supplier
	for rows.Next() {
		var sup Supplier
		var metadata []byte
		err := rows.Scan(
			&sup.ID, &sup.TenantID, &sup.Name, &sup.Industry, &sup.Country,
			&sup.Region, &sup.Tier, &sup.Category, &sup.Status,
			&sup.EmissionsKgCO2e, &sup.SpendAmount, &sup.SpendCurrency,
			&sup.DataQuality, &sup.ContactEmail, &sup.EngagementStatus,
			&sup.ReductionTarget, &metadata, &sup.CreatedAt, &sup.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		_ = json.Unmarshal(metadata, &sup.Metadata)
		suppliers = append(suppliers, sup)
	}
	return suppliers, rows.Err()
}

// ListOptions provides pagination options.
type ListOptions struct {
	Limit  int
	Offset int
}

// =============================================================================
// Graph Operations
// =============================================================================

// BuildGraph constructs the supply chain graph for a tenant.
func (s *Service) BuildGraph(ctx context.Context, tenantID string) (*Graph, error) {
	// Check cache
	s.mu.RLock()
	if cached, ok := s.cache[tenantID]; ok {
		s.mu.RUnlock()
		return cached, nil
	}
	s.mu.RUnlock()

	// Build fresh graph
	suppliers, err := s.ListSuppliers(ctx, tenantID, ListOptions{Limit: 10000})
	if err != nil {
		return nil, err
	}

	graph := &Graph{
		TenantID:  tenantID,
		Suppliers: make(map[string]Supplier),
		Edges:     make([]Edge, 0),
		UpdatedAt: time.Now(),
	}

	var totalEmissions float64
	categoryEmissions := make(map[Scope3Category]float64)
	tierEmissions := make(map[int]float64)
	dataQualityDist := make(map[DataQualityLevel]int)
	engagementDist := make(map[EngagementStatus]int)

	for _, sup := range suppliers {
		graph.Suppliers[sup.ID] = sup
		totalEmissions += sup.EmissionsKgCO2e
		categoryEmissions[sup.Category] += sup.EmissionsKgCO2e
		tierEmissions[sup.Tier] += sup.EmissionsKgCO2e
		dataQualityDist[sup.DataQuality]++
		engagementDist[sup.EngagementStatus]++

		// Create edge from tenant to supplier
		graph.Edges = append(graph.Edges, Edge{
			ID:              uuid.NewString(),
			FromID:          tenantID,
			FromType:        "tenant",
			ToID:            sup.ID,
			ToType:          "supplier",
			RelationType:    "purchases_from",
			EmissionsKgCO2e: sup.EmissionsKgCO2e,
			CreatedAt:       sup.CreatedAt,
		})
	}

	// Calculate percentages for edges
	for i := range graph.Edges {
		if totalEmissions > 0 {
			graph.Edges[i].Percentage = (graph.Edges[i].EmissionsKgCO2e / totalEmissions) * 100
		}
	}

	// Build top contributors
	var contributions []SupplierContribution
	for _, sup := range suppliers {
		pct := 0.0
		if totalEmissions > 0 {
			pct = (sup.EmissionsKgCO2e / totalEmissions) * 100
		}
		contributions = append(contributions, SupplierContribution{
			SupplierID:      sup.ID,
			SupplierName:    sup.Name,
			EmissionsKgCO2e: sup.EmissionsKgCO2e,
			Percentage:      pct,
		})
	}
	sort.Slice(contributions, func(i, j int) bool {
		return contributions[i].EmissionsKgCO2e > contributions[j].EmissionsKgCO2e
	})
	if len(contributions) > 20 {
		contributions = contributions[:20]
	}

	graph.Stats = GraphStats{
		TotalSuppliers:       len(suppliers),
		TotalEmissionsKgCO2e: totalEmissions,
		ByCategory:           categoryEmissions,
		ByTier:               tierEmissions,
		DataQualityDist:      dataQualityDist,
		EngagementDist:       engagementDist,
		TopContributors:      contributions,
	}

	// Cache the graph
	s.mu.Lock()
	s.cache[tenantID] = graph
	s.mu.Unlock()

	return graph, nil
}

// GetTopContributors returns the top N suppliers by emissions.
func (s *Service) GetTopContributors(ctx context.Context, tenantID string, n int) ([]SupplierContribution, error) {
	graph, err := s.BuildGraph(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	top := graph.Stats.TopContributors
	if n > 0 && n < len(top) {
		top = top[:n]
	}
	return top, nil
}

// =============================================================================
// Engagement Operations
// =============================================================================

// GenerateOutreachEmail creates a personalized outreach email for a supplier.
func (s *Service) GenerateOutreachEmail(ctx context.Context, tenantID string, supplier Supplier) (string, string, error) {
	subject := fmt.Sprintf("Request for Carbon Emissions Data - %s Partnership", supplier.Name)

	body := fmt.Sprintf(`Dear %s Team,

As part of our sustainability commitments and regulatory obligations under CSRD/SEC climate disclosure requirements, we are conducting a comprehensive assessment of our Scope 3 value chain emissions.

As one of our valued partners, your emissions data is important to our reporting accuracy. We kindly request that you share:

1. Your organization's annual greenhouse gas emissions (Scope 1, 2, and if available, Scope 3)
2. Any carbon reduction targets or net-zero commitments
3. Relevant certifications (ISO 14001, Science Based Targets, etc.)

To make this process easy, we've set up a secure data submission portal:
[PORTAL_LINK]

Benefits of participation:
- Strengthen our partnership through transparency
- Demonstrate your sustainability leadership
- Receive benchmarking insights (anonymized)
- Priority consideration in procurement decisions

We appreciate your participation in our sustainability journey.

Best regards,
[COMPANY_NAME] Sustainability Team

---
This request is part of our CSRD compliance program.
`, supplier.Name)

	return subject, body, nil
}

// CreateCampaign creates a new outreach campaign.
func (s *Service) CreateCampaign(ctx context.Context, campaign OutreachCampaign) (*OutreachCampaign, error) {
	if campaign.ID == "" {
		campaign.ID = uuid.NewString()
	}
	campaign.CreatedAt = time.Now()
	campaign.UpdatedAt = time.Now()
	campaign.Status = CampaignDraft

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO outreach_campaigns (
			id, tenant_id, name, description, supplier_ids, status,
			email_subject, email_body, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`,
		campaign.ID, campaign.TenantID, campaign.Name, campaign.Description,
		mustJSON(campaign.SupplierIDs), campaign.Status,
		campaign.EmailSubject, campaign.EmailBody,
		campaign.CreatedAt, campaign.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create campaign: %w", err)
	}

	return &campaign, nil
}

// =============================================================================
// Reduction Target Operations
// =============================================================================

// SetReductionTarget sets a reduction target for a supplier.
func (s *Service) SetReductionTarget(ctx context.Context, target ReductionTarget) (*ReductionTarget, error) {
	if target.ID == "" {
		target.ID = uuid.NewString()
	}
	target.CreatedAt = time.Now()
	target.UpdatedAt = time.Now()
	target.Status = TargetPending

	// Generate milestones (linear interpolation)
	yearsToTarget := target.TargetYear - target.BaselineYear
	if yearsToTarget > 0 {
		target.Milestones = make([]Milestone, yearsToTarget)
		for i := 0; i < yearsToTarget; i++ {
			year := target.BaselineYear + i + 1
			cumulative := target.TargetReduction * float64(i+1) / float64(yearsToTarget)
			target.Milestones[i] = Milestone{
				Year:   year,
				Target: cumulative,
				Status: "pending",
			}
		}
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO reduction_targets (
			id, tenant_id, supplier_id, baseline_year, baseline_emissions,
			target_year, target_reduction, current_progress, status,
			milestones, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`,
		target.ID, target.TenantID, target.SupplierID,
		target.BaselineYear, target.BaselineEmissions,
		target.TargetYear, target.TargetReduction,
		target.CurrentProgress, target.Status,
		mustJSON(target.Milestones), target.CreatedAt, target.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set reduction target: %w", err)
	}

	// Update supplier with target
	_, _ = s.db.ExecContext(ctx, `
		UPDATE suppliers 
		SET reduction_target = $1, engagement_status = $2, updated_at = NOW()
		WHERE id = $3 AND tenant_id = $4
	`, target.TargetReduction, EngagementTargetSet, target.SupplierID, target.TenantID)

	s.invalidateCache(target.TenantID)
	return &target, nil
}

// CascadeTargets distributes a parent target across top suppliers.
func (s *Service) CascadeTargets(ctx context.Context, tenantID string, totalReduction float64, targetYear int) ([]ReductionTarget, error) {
	// Get top contributors (Pareto principle: focus on top 20% for 80% impact)
	top, err := s.GetTopContributors(ctx, tenantID, 20)
	if err != nil {
		return nil, err
	}

	var targets []ReductionTarget
	currentYear := time.Now().Year()

	for _, contrib := range top {
		sup, _ := s.GetSupplier(ctx, tenantID, contrib.SupplierID)
		if sup == nil {
			continue
		}

		// Proportional target based on contribution
		supplierTarget := totalReduction * (contrib.Percentage / 100)

		target := ReductionTarget{
			TenantID:          tenantID,
			SupplierID:        contrib.SupplierID,
			BaselineYear:      currentYear,
			BaselineEmissions: contrib.EmissionsKgCO2e,
			TargetYear:        targetYear,
			TargetReduction:   supplierTarget,
		}

		created, err := s.SetReductionTarget(ctx, target)
		if err != nil {
			s.logger.Warn("failed to set target for supplier",
				"supplierId", contrib.SupplierID,
				"error", err)
			continue
		}
		targets = append(targets, *created)
	}

	return targets, nil
}

// =============================================================================
// Helpers
// =============================================================================

func (s *Service) invalidateCache(tenantID string) {
	s.mu.Lock()
	delete(s.cache, tenantID)
	s.mu.Unlock()
}

func mustJSON(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}
