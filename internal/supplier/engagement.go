// Package supplier provides Scope 3 supplier engagement platform.
//
// This package enables collaboration with supply chain partners
// for comprehensive Scope 3 emissions accounting.
package supplier

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// =============================================================================
// Core Types
// =============================================================================

// Supplier represents a supply chain partner.
type Supplier struct {
	ID            string             `json:"id"`
	TenantID      string             `json:"tenantId"`
	Name          string             `json:"name"`
	ContactName   string             `json:"contactName"`
	ContactEmail  string             `json:"contactEmail"`
	Industry      string             `json:"industry,omitempty"`
	Country       string             `json:"country"`
	Tier          SupplierTier       `json:"tier"`
	Status        SupplierStatus     `json:"status"`
	Categories    []string           `json:"categories"`
	SpendAnnual   float64            `json:"spendAnnual"`
	SpendCurrency string             `json:"spendCurrency"`
	Emissions     *SupplierEmissions `json:"emissions,omitempty"`
	Engagement    *EngagementStatus  `json:"engagement,omitempty"`
	CreatedAt     time.Time          `json:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt"`
}

// SupplierTier categorizes supplier importance.
type SupplierTier string

const (
	TierStrategic SupplierTier = "strategic" // Top suppliers, high spend
	TierPreferred SupplierTier = "preferred" // Important suppliers
	TierApproved  SupplierTier = "approved"  // Standard suppliers
	TierPotential SupplierTier = "potential" // Under evaluation
)

// SupplierStatus tracks supplier engagement status.
type SupplierStatus string

const (
	StatusPending   SupplierStatus = "pending"   // Invite sent
	StatusActive    SupplierStatus = "active"    // Engaged
	StatusResponded SupplierStatus = "responded" // Data provided
	StatusVerified  SupplierStatus = "verified"  // Data verified
	StatusInactive  SupplierStatus = "inactive"  // Not responding
)

// SupplierEmissions contains supplier emissions data.
type SupplierEmissions struct {
	Year          int         `json:"year"`
	Scope1        float64     `json:"scope1"`
	Scope2        float64     `json:"scope2"`
	Scope3        float64     `json:"scope3,omitempty"`
	Total         float64     `json:"total"`
	IntensityUnit string      `json:"intensityUnit,omitempty"` // "per_unit", "per_revenue"
	Intensity     float64     `json:"intensity,omitempty"`
	Methodology   string      `json:"methodology,omitempty"`
	DataQuality   DataQuality `json:"dataQuality"`
	VerifiedBy    string      `json:"verifiedBy,omitempty"`
	SubmittedAt   time.Time   `json:"submittedAt"`
}

// DataQuality indicates the quality of emissions data.
type DataQuality string

const (
	QualityPrimary   DataQuality = "primary"   // Direct measurement
	QualitySecondary DataQuality = "secondary" // Calculated/estimated
	QualityDefault   DataQuality = "default"   // Industry averages
)

// EngagementStatus tracks supplier engagement program progress.
type EngagementStatus struct {
	InvitedAt       *time.Time `json:"invitedAt,omitempty"`
	LastContactAt   *time.Time `json:"lastContactAt,omitempty"`
	ResponseRate    float64    `json:"responseRate"` // 0-100
	DataSubmissions int        `json:"dataSubmissions"`
	TargetSet       bool       `json:"targetSet"`
	TargetReduction float64    `json:"targetReduction,omitempty"` // %
	TargetYear      int        `json:"targetYear,omitempty"`
	Progress        float64    `json:"progress,omitempty"` // % toward target
}

// =============================================================================
// Invitation System
// =============================================================================

// Invitation represents a supplier engagement invitation.
type Invitation struct {
	ID          string           `json:"id"`
	TenantID    string           `json:"tenantId"`
	SupplierID  string           `json:"supplierId"`
	Token       string           `json:"token"`
	Email       string           `json:"email"`
	Message     string           `json:"message,omitempty"`
	RequestType InvitationType   `json:"requestType"`
	Deadline    *time.Time       `json:"deadline,omitempty"`
	SentAt      time.Time        `json:"sentAt"`
	OpenedAt    *time.Time       `json:"openedAt,omitempty"`
	CompletedAt *time.Time       `json:"completedAt,omitempty"`
	Status      InvitationStatus `json:"status"`
}

// InvitationType defines what data is requested.
type InvitationType string

const (
	TypeEmissionsData    InvitationType = "emissions_data"
	TypeProductFootprint InvitationType = "product_footprint"
	TypeTarget           InvitationType = "target_setting"
	TypeQuestionnaire    InvitationType = "questionnaire"
)

// InvitationStatus tracks invitation status.
type InvitationStatus string

const (
	InvitePending   InvitationStatus = "pending"
	InviteSent      InvitationStatus = "sent"
	InviteOpened    InvitationStatus = "opened"
	InviteCompleted InvitationStatus = "completed"
	InviteExpired   InvitationStatus = "expired"
	InviteDeclined  InvitationStatus = "declined"
)

// =============================================================================
// Product Carbon Footprint
// =============================================================================

// ProductFootprint tracks product-level emissions.
type ProductFootprint struct {
	ID            string      `json:"id"`
	SupplierID    string      `json:"supplierId"`
	ProductCode   string      `json:"productCode"`
	ProductName   string      `json:"productName"`
	Year          int         `json:"year"`
	CO2ePerUnit   float64     `json:"co2ePerUnit"`
	Unit          string      `json:"unit"` // "kg", "piece", etc.
	Methodology   string      `json:"methodology"`
	Scope         []int       `json:"scope"` // Which scopes included
	CradleToGate  bool        `json:"cradleToGate"`
	CradleToGrave bool        `json:"cradleToGrave,omitempty"`
	DataQuality   DataQuality `json:"dataQuality"`
	SubmittedAt   time.Time   `json:"submittedAt"`
}

// =============================================================================
// Service
// =============================================================================

// Service manages supplier engagement.
type Service struct {
	suppliers   map[string]*Supplier          // id -> supplier
	invitations map[string]*Invitation        // id -> invitation
	footprints  map[string][]ProductFootprint // supplierID -> footprints
	emailer     EmailSender
	logger      *slog.Logger
	mu          sync.RWMutex
}

// EmailSender sends emails.
type EmailSender interface {
	Send(to, subject, body string) error
}

// ServiceConfig configures the supplier service.
type ServiceConfig struct {
	EmailSender EmailSender
	Logger      *slog.Logger
}

// NewService creates a new supplier engagement service.
func NewService(cfg ServiceConfig) *Service {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	return &Service{
		suppliers:   make(map[string]*Supplier),
		invitations: make(map[string]*Invitation),
		footprints:  make(map[string][]ProductFootprint),
		emailer:     cfg.EmailSender,
		logger:      cfg.Logger.With("component", "supplier-service"),
	}
}

// =============================================================================
// Supplier Management
// =============================================================================

// AddSupplier adds a new supplier.
func (s *Service) AddSupplier(tenantID string, supplier *Supplier) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if supplier.ID == "" {
		supplier.ID = fmt.Sprintf("supplier-%d", time.Now().UnixNano())
	}
	supplier.TenantID = tenantID
	supplier.Status = StatusPending
	supplier.CreatedAt = time.Now()
	supplier.UpdatedAt = time.Now()
	supplier.Engagement = &EngagementStatus{}

	s.suppliers[supplier.ID] = supplier

	s.logger.Info("supplier added",
		"supplierId", supplier.ID,
		"name", supplier.Name,
		"tier", supplier.Tier)

	return nil
}

// GetSupplier retrieves a supplier.
func (s *Service) GetSupplier(supplierID string) (*Supplier, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	supplier, ok := s.suppliers[supplierID]
	if !ok {
		return nil, errors.New("supplier not found")
	}
	return supplier, nil
}

// ListSuppliers returns suppliers for a tenant.
func (s *Service) ListSuppliers(tenantID string, tier *SupplierTier, status *SupplierStatus) []*Supplier {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Supplier
	for _, supplier := range s.suppliers {
		if supplier.TenantID != tenantID {
			continue
		}
		if tier != nil && supplier.Tier != *tier {
			continue
		}
		if status != nil && supplier.Status != *status {
			continue
		}
		result = append(result, supplier)
	}
	return result
}

// UpdateSupplier updates a supplier.
func (s *Service) UpdateSupplier(supplierID string, update func(*Supplier)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	supplier, ok := s.suppliers[supplierID]
	if !ok {
		return errors.New("supplier not found")
	}

	update(supplier)
	supplier.UpdatedAt = time.Now()
	return nil
}

// =============================================================================
// Invitation System
// =============================================================================

// SendInvitation sends an engagement invitation.
func (s *Service) SendInvitation(ctx context.Context, tenantID, supplierID string, requestType InvitationType, deadline *time.Time) (*Invitation, error) {
	s.mu.Lock()
	supplier, ok := s.suppliers[supplierID]
	if !ok {
		s.mu.Unlock()
		return nil, errors.New("supplier not found")
	}
	s.mu.Unlock()

	// Generate secure token
	token, err := s.generateToken()
	if err != nil {
		return nil, err
	}

	invitation := &Invitation{
		ID:          fmt.Sprintf("invite-%d", time.Now().UnixNano()),
		TenantID:    tenantID,
		SupplierID:  supplierID,
		Token:       token,
		Email:       supplier.ContactEmail,
		RequestType: requestType,
		Deadline:    deadline,
		SentAt:      time.Now(),
		Status:      InvitePending,
	}

	// Send email
	if s.emailer != nil {
		subject := s.getInvitationSubject(requestType)
		body := s.getInvitationBody(supplier.Name, requestType, token, deadline)
		if err := s.emailer.Send(supplier.ContactEmail, subject, body); err != nil {
			s.logger.Error("failed to send invitation email",
				"supplierId", supplierID,
				"error", err)
		} else {
			invitation.Status = InviteSent
		}
	}

	s.mu.Lock()
	s.invitations[invitation.ID] = invitation
	if supplier.Engagement != nil {
		now := time.Now()
		supplier.Engagement.InvitedAt = &now
		supplier.Engagement.LastContactAt = &now
	}
	s.mu.Unlock()

	s.logger.Info("invitation sent",
		"invitationId", invitation.ID,
		"supplierId", supplierID,
		"type", requestType)

	return invitation, nil
}

// generateToken creates a secure random token.
func (s *Service) generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// getInvitationSubject returns email subject.
func (s *Service) getInvitationSubject(requestType InvitationType) string {
	switch requestType {
	case TypeEmissionsData:
		return "Request for Emissions Data - Supply Chain Collaboration"
	case TypeProductFootprint:
		return "Request for Product Carbon Footprint Data"
	case TypeTarget:
		return "Invitation: Set Science-Based Emission Reduction Targets"
	case TypeQuestionnaire:
		return "Sustainability Questionnaire Request"
	default:
		return "Supplier Engagement Request"
	}
}

// getInvitationBody returns email body.
func (s *Service) getInvitationBody(supplierName string, requestType InvitationType, token string, deadline *time.Time) string {
	var deadlineStr string
	if deadline != nil {
		deadlineStr = fmt.Sprintf("Please respond by %s.", deadline.Format("January 2, 2006"))
	}

	switch requestType {
	case TypeEmissionsData:
		return fmt.Sprintf(`Dear %s,

As part of our commitment to measuring and reducing our Scope 3 emissions, we are reaching out to request your organization's emissions data.

This information will help us:
- Accurately measure our supply chain emissions
- Identify opportunities for collaborative emission reductions
- Meet our regulatory reporting requirements

Please use the following link to submit your data:
https://app.offgridflow.com/supplier/submit/%s

%s

Thank you for your partnership in our sustainability journey.

Best regards,
Your Customer's Sustainability Team`, supplierName, token, deadlineStr)
	default:
		return fmt.Sprintf(`Dear %s,

We invite you to participate in our supplier engagement program.

Please use the following link:
https://app.offgridflow.com/supplier/submit/%s

%s

Best regards`, supplierName, token, deadlineStr)
	}
}

// ValidateToken validates an invitation token.
func (s *Service) ValidateToken(token string) (*Invitation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, invitation := range s.invitations {
		if invitation.Token == token {
			// Check expiration
			if invitation.Deadline != nil && time.Now().After(*invitation.Deadline) {
				return nil, errors.New("invitation expired")
			}
			return invitation, nil
		}
	}
	return nil, errors.New("invalid token")
}

// MarkOpened marks an invitation as opened.
func (s *Service) MarkOpened(invitationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	invitation, ok := s.invitations[invitationID]
	if !ok {
		return errors.New("invitation not found")
	}

	now := time.Now()
	invitation.OpenedAt = &now
	invitation.Status = InviteOpened
	return nil
}

// =============================================================================
// Data Submission
// =============================================================================

// SubmitEmissions allows a supplier to submit emissions data.
func (s *Service) SubmitEmissions(token string, emissions SupplierEmissions) error {
	invitation, err := s.ValidateToken(token)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	supplier, ok := s.suppliers[invitation.SupplierID]
	if !ok {
		return errors.New("supplier not found")
	}

	emissions.SubmittedAt = time.Now()
	supplier.Emissions = &emissions
	supplier.Status = StatusResponded
	supplier.UpdatedAt = time.Now()

	if supplier.Engagement != nil {
		supplier.Engagement.DataSubmissions++
		now := time.Now()
		supplier.Engagement.LastContactAt = &now
	}

	// Mark invitation complete
	now := time.Now()
	invitation.CompletedAt = &now
	invitation.Status = InviteCompleted

	s.logger.Info("emissions data submitted",
		"supplierId", supplier.ID,
		"year", emissions.Year,
		"total", emissions.Total)

	return nil
}

// SubmitProductFootprint submits product-level footprint.
func (s *Service) SubmitProductFootprint(token string, footprint ProductFootprint) error {
	invitation, err := s.ValidateToken(token)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	footprint.ID = fmt.Sprintf("fp-%d", time.Now().UnixNano())
	footprint.SupplierID = invitation.SupplierID
	footprint.SubmittedAt = time.Now()

	s.footprints[invitation.SupplierID] = append(s.footprints[invitation.SupplierID], footprint)

	s.logger.Info("product footprint submitted",
		"supplierId", invitation.SupplierID,
		"productCode", footprint.ProductCode)

	return nil
}

// GetProductFootprints retrieves footprints for a supplier.
func (s *Service) GetProductFootprints(supplierID string) []ProductFootprint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.footprints[supplierID]
}

// =============================================================================
// Analytics
// =============================================================================

// EngagementMetrics provides engagement program metrics.
type EngagementMetrics struct {
	TotalSuppliers   int                      `json:"totalSuppliers"`
	EngagedSuppliers int                      `json:"engagedSuppliers"`
	ResponseRate     float64                  `json:"responseRate"`
	DataCoverage     float64                  `json:"dataCoverage"` // % of spend with data
	PrimaryDataRate  float64                  `json:"primaryDataRate"`
	TargetSetRate    float64                  `json:"targetSetRate"`
	EmissionsByTier  map[SupplierTier]float64 `json:"emissionsByTier"`
	TotalScope3      float64                  `json:"totalScope3"`
}

// GetEngagementMetrics calculates engagement metrics.
func (s *Service) GetEngagementMetrics(tenantID string) *EngagementMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics := &EngagementMetrics{
		EmissionsByTier: make(map[SupplierTier]float64),
	}

	var responded, withData, primaryData, targetSet int
	var totalSpend, spendWithData float64

	for _, supplier := range s.suppliers {
		if supplier.TenantID != tenantID {
			continue
		}

		metrics.TotalSuppliers++
		totalSpend += supplier.SpendAnnual

		if supplier.Status == StatusResponded || supplier.Status == StatusVerified {
			responded++
		}

		if supplier.Emissions != nil {
			withData++
			spendWithData += supplier.SpendAnnual

			if supplier.Emissions.DataQuality == QualityPrimary {
				primaryData++
			}

			// Allocate emissions by spend share
			emissions := supplier.Emissions.Total
			metrics.EmissionsByTier[supplier.Tier] += emissions
			metrics.TotalScope3 += emissions
		}

		if supplier.Engagement != nil && supplier.Engagement.TargetSet {
			targetSet++
		}
	}

	if metrics.TotalSuppliers > 0 {
		metrics.EngagedSuppliers = responded
		metrics.ResponseRate = float64(responded) / float64(metrics.TotalSuppliers) * 100
		metrics.PrimaryDataRate = float64(primaryData) / float64(metrics.TotalSuppliers) * 100
		metrics.TargetSetRate = float64(targetSet) / float64(metrics.TotalSuppliers) * 100
	}

	if totalSpend > 0 {
		metrics.DataCoverage = spendWithData / totalSpend * 100
	}

	return metrics
}

// IdentifyHotspots identifies high-emission suppliers.
func (s *Service) IdentifyHotspots(tenantID string, limit int) []*Supplier {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var suppliers []*Supplier
	for _, supplier := range s.suppliers {
		if supplier.TenantID != tenantID && supplier.Emissions != nil {
			suppliers = append(suppliers, supplier)
		}
	}

	// Sort by emissions descending
	for i := 0; i < len(suppliers)-1; i++ {
		for j := i + 1; j < len(suppliers); j++ {
			if suppliers[j].Emissions.Total > suppliers[i].Emissions.Total {
				suppliers[i], suppliers[j] = suppliers[j], suppliers[i]
			}
		}
	}

	if limit > 0 && limit < len(suppliers) {
		return suppliers[:limit]
	}
	return suppliers
}

// =============================================================================
// Export
// =============================================================================

// ExportSupplierData exports supplier data as JSON.
func (s *Service) ExportSupplierData(tenantID string) ([]byte, error) {
	suppliers := s.ListSuppliers(tenantID, nil, nil)
	return json.MarshalIndent(suppliers, "", "  ")
}
