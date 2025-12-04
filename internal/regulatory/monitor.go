// Package regulatory provides AI-powered regulatory change monitoring.
//
// This package monitors for changes in climate regulations and compliance
// requirements, alerting tenants to relevant updates.
package regulatory

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

// =============================================================================
// Core Types
// =============================================================================

// Jurisdiction represents a regulatory jurisdiction.
type Jurisdiction string

const (
	JurisdictionEU         Jurisdiction = "eu"
	JurisdictionUS         Jurisdiction = "us"
	JurisdictionUK         Jurisdiction = "uk"
	JurisdictionCalifornia Jurisdiction = "us-ca"
	JurisdictionGlobal     Jurisdiction = "global"
)

// RegulationType categorizes regulations.
type RegulationType string

const (
	TypeDisclosure     RegulationType = "disclosure"
	TypeReporting      RegulationType = "reporting"
	TypeTaxation       RegulationType = "taxation"
	TypeEmissionsLimit RegulationType = "emissions_limit"
	TypeTrading        RegulationType = "trading"
)

// Regulation represents a climate regulation.
type Regulation struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	ShortName     string          `json:"shortName"`
	Jurisdiction  Jurisdiction    `json:"jurisdiction"`
	Type          RegulationType  `json:"type"`
	Description   string          `json:"description"`
	EffectiveDate *time.Time      `json:"effectiveDate,omitempty"`
	Deadline      *time.Time      `json:"deadline,omitempty"`
	Applies       []Applicability `json:"applies"`
	Requirements  []Requirement   `json:"requirements"`
	Status        RegStatus       `json:"status"`
	LastUpdated   time.Time       `json:"lastUpdated"`
	Source        string          `json:"source,omitempty"`
	URL           string          `json:"url,omitempty"`
}

// RegStatus tracks regulation status.
type RegStatus string

const (
	StatusProposed  RegStatus = "proposed"
	StatusAdopted   RegStatus = "adopted"
	StatusEffective RegStatus = "effective"
	StatusAmended   RegStatus = "amended"
	StatusRepealed  RegStatus = "repealed"
)

// Applicability defines who must comply.
type Applicability struct {
	Criterion   string `json:"criterion"` // "revenue", "employees", "public"
	Operator    string `json:"operator"`  // ">", ">=", "="
	Value       string `json:"value"`
	Description string `json:"description"`
}

// Requirement defines a specific compliance requirement.
type Requirement struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	Scope       []int      `json:"scope,omitempty"` // 1, 2, 3
	Mandatory   bool       `json:"mandatory"`
	Deadline    *time.Time `json:"deadline,omitempty"`
}

// =============================================================================
// Change Detection
// =============================================================================

// RegulatoryChange represents a detected change.
type RegulatoryChange struct {
	ID           string      `json:"id"`
	RegulationID string      `json:"regulationId"`
	Type         ChangeType  `json:"type"`
	Title        string      `json:"title"`
	Summary      string      `json:"summary"`
	Details      string      `json:"details"`
	Impact       ImpactLevel `json:"impact"`
	DetectedAt   time.Time   `json:"detectedAt"`
	EffectiveAt  *time.Time  `json:"effectiveAt,omitempty"`
	ActionNeeded []string    `json:"actionNeeded,omitempty"`
	Source       string      `json:"source"`
	URL          string      `json:"url,omitempty"`
}

// ChangeType categorizes the type of change.
type ChangeType string

const (
	ChangeNew         ChangeType = "new"
	ChangeAmendment   ChangeType = "amendment"
	ChangeDeadline    ChangeType = "deadline"
	ChangeGuidance    ChangeType = "guidance"
	ChangeEnforcement ChangeType = "enforcement"
)

// ImpactLevel indicates the severity of impact.
type ImpactLevel string

const (
	ImpactCritical ImpactLevel = "critical" // Immediate action required
	ImpactHigh     ImpactLevel = "high"     // Significant changes needed
	ImpactMedium   ImpactLevel = "medium"   // Notable updates
	ImpactLow      ImpactLevel = "low"      // Minor adjustments
	ImpactInfo     ImpactLevel = "info"     // Informational only
)

// =============================================================================
// Alert System
// =============================================================================

// Alert is a notification to a tenant about regulatory changes.
type Alert struct {
	ID            string         `json:"id"`
	TenantID      string         `json:"tenantId"`
	ChangeID      string         `json:"changeId"`
	Title         string         `json:"title"`
	Message       string         `json:"message"`
	Impact        ImpactLevel    `json:"impact"`
	Regulations   []string       `json:"regulations"`
	Jurisdictions []Jurisdiction `json:"jurisdictions"`
	Actions       []string       `json:"actions"`
	CreatedAt     time.Time      `json:"createdAt"`
	ReadAt        *time.Time     `json:"readAt,omitempty"`
	DismissedAt   *time.Time     `json:"dismissedAt,omitempty"`
}

// AlertPreferences defines tenant notification preferences.
type AlertPreferences struct {
	TenantID      string           `json:"tenantId"`
	Jurisdictions []Jurisdiction   `json:"jurisdictions"`
	RegTypes      []RegulationType `json:"regTypes"`
	MinImpact     ImpactLevel      `json:"minImpact"`
	EmailEnabled  bool             `json:"emailEnabled"`
	Emails        []string         `json:"emails,omitempty"`
	WebhookURL    string           `json:"webhookUrl,omitempty"`
}

// =============================================================================
// Monitor
// =============================================================================

// Monitor watches for regulatory changes.
type Monitor struct {
	regulations map[string]Regulation
	changes     []RegulatoryChange
	sources     []Source
	logger      *slog.Logger
	mu          sync.RWMutex

	// Callbacks
	onChangeDetected func(change RegulatoryChange)
}

// Source defines a regulatory data source.
type Source struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Enabled  bool   `json:"enabled"`
	Interval time.Duration
}

// MonitorConfig configures the regulatory monitor.
type MonitorConfig struct {
	Sources []Source
	Logger  *slog.Logger
}

// NewMonitor creates a new regulatory monitor.
func NewMonitor(cfg MonitorConfig) *Monitor {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	m := &Monitor{
		regulations: make(map[string]Regulation),
		changes:     make([]RegulatoryChange, 0),
		sources:     cfg.Sources,
		logger:      cfg.Logger.With("component", "regulatory-monitor"),
	}

	// Load initial regulations
	m.loadInitialRegulations()

	return m
}

// loadInitialRegulations loads known regulations.
func (m *Monitor) loadInitialRegulations() {
	// CSRD
	csrdEffective := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	m.regulations["csrd"] = Regulation{
		ID:            "csrd",
		Name:          "Corporate Sustainability Reporting Directive",
		ShortName:     "CSRD",
		Jurisdiction:  JurisdictionEU,
		Type:          TypeDisclosure,
		Description:   "EU directive requiring sustainability reporting using ESRS standards",
		EffectiveDate: &csrdEffective,
		Applies: []Applicability{
			{Criterion: "employees", Operator: ">", Value: "250", Description: "Large undertakings with >250 employees"},
			{Criterion: "revenue", Operator: ">", Value: "40000000", Description: "Net turnover >€40M"},
			{Criterion: "public", Operator: "=", Value: "true", Description: "All listed companies"},
		},
		Requirements: []Requirement{
			{ID: "csrd-1", Name: "E1 Climate Change", Category: "environment", Scope: []int{1, 2, 3}, Mandatory: true},
			{ID: "csrd-2", Name: "E2 Pollution", Category: "environment", Mandatory: true},
			{ID: "csrd-3", Name: "Double Materiality Assessment", Category: "general", Mandatory: true},
		},
		Status:      StatusEffective,
		LastUpdated: time.Now(),
		URL:         "https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX:32022L2464",
	}

	// SEC Climate Disclosure
	secEffective := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	m.regulations["sec-climate"] = Regulation{
		ID:            "sec-climate",
		Name:          "SEC Climate Disclosure Rules",
		ShortName:     "SEC Climate",
		Jurisdiction:  JurisdictionUS,
		Type:          TypeDisclosure,
		Description:   "US SEC requirements for climate-related disclosures in SEC filings",
		EffectiveDate: &secEffective,
		Applies: []Applicability{
			{Criterion: "public", Operator: "=", Value: "true", Description: "SEC registrants"},
		},
		Requirements: []Requirement{
			{ID: "sec-1", Name: "Scope 1 & 2 Emissions", Category: "emissions", Scope: []int{1, 2}, Mandatory: true},
			{ID: "sec-2", Name: "Climate Risk Disclosure", Category: "risk", Mandatory: true},
			{ID: "sec-3", Name: "Financial Impact Assessment", Category: "financial", Mandatory: true},
		},
		Status:      StatusAdopted,
		LastUpdated: time.Now(),
		URL:         "https://www.sec.gov/rules/final/2024/33-11275.pdf",
	}

	// CBAM
	cbamEffective := time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)
	m.regulations["cbam"] = Regulation{
		ID:            "cbam",
		Name:          "Carbon Border Adjustment Mechanism",
		ShortName:     "CBAM",
		Jurisdiction:  JurisdictionEU,
		Type:          TypeTaxation,
		Description:   "EU carbon levy on imports of carbon-intensive goods",
		EffectiveDate: &cbamEffective,
		Applies: []Applicability{
			{Criterion: "imports", Operator: "=", Value: "covered_goods", Description: "Importers of steel, cement, aluminum, fertilizers, electricity, hydrogen"},
		},
		Requirements: []Requirement{
			{ID: "cbam-1", Name: "Quarterly CBAM Reports", Category: "reporting", Mandatory: true},
			{ID: "cbam-2", Name: "Embedded Emissions Calculation", Category: "emissions", Scope: []int{1, 2}, Mandatory: true},
		},
		Status:      StatusEffective,
		LastUpdated: time.Now(),
		URL:         "https://taxation-customs.ec.europa.eu/carbon-border-adjustment-mechanism_en",
	}

	// California Climate Disclosure
	caEffective := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	m.regulations["california-climate"] = Regulation{
		ID:            "california-climate",
		Name:          "California Climate Corporate Data Accountability Act",
		ShortName:     "SB 253",
		Jurisdiction:  JurisdictionCalifornia,
		Type:          TypeDisclosure,
		Description:   "California requirements for Scope 1, 2, and 3 emissions disclosure",
		EffectiveDate: &caEffective,
		Applies: []Applicability{
			{Criterion: "revenue", Operator: ">", Value: "1000000000", Description: "Revenue >$1B doing business in CA"},
		},
		Requirements: []Requirement{
			{ID: "ca-1", Name: "Scope 1 & 2 Emissions", Category: "emissions", Scope: []int{1, 2}, Mandatory: true},
			{ID: "ca-2", Name: "Scope 3 Emissions", Category: "emissions", Scope: []int{3}, Mandatory: true},
			{ID: "ca-3", Name: "Third-Party Assurance", Category: "assurance", Mandatory: true},
		},
		Status:      StatusAdopted,
		LastUpdated: time.Now(),
	}
}

// OnChangeDetected sets the callback for changes.
func (m *Monitor) OnChangeDetected(fn func(RegulatoryChange)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onChangeDetected = fn
}

// Start begins monitoring for changes.
func (m *Monitor) Start(ctx context.Context) {
	m.logger.Info("starting regulatory monitor")

	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	// Initial check
	m.checkForUpdates(ctx)

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("stopping regulatory monitor")
			return
		case <-ticker.C:
			m.checkForUpdates(ctx)
		}
	}
}

// checkForUpdates polls sources for regulatory changes.
func (m *Monitor) checkForUpdates(ctx context.Context) {
	m.logger.Debug("checking for regulatory updates")

	// In production, this would poll real regulatory sources
	// For now, we simulate by checking for date-based triggers

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for upcoming deadlines
	for _, reg := range m.regulations {
		if reg.EffectiveDate != nil {
			daysUntil := int(time.Until(*reg.EffectiveDate).Hours() / 24)

			// Alert at 90, 60, 30, 7 days
			milestones := []int{90, 60, 30, 7}
			for _, milestone := range milestones {
				if daysUntil == milestone {
					change := RegulatoryChange{
						ID:           fmt.Sprintf("deadline-%s-%d", reg.ID, milestone),
						RegulationID: reg.ID,
						Type:         ChangeDeadline,
						Title:        fmt.Sprintf("%s: %d days until effective date", reg.ShortName, milestone),
						Summary:      fmt.Sprintf("The %s becomes effective in %d days.", reg.Name, milestone),
						Impact:       m.calculateDeadlineImpact(milestone),
						DetectedAt:   time.Now(),
						EffectiveAt:  reg.EffectiveDate,
						ActionNeeded: []string{
							"Review compliance status",
							"Ensure reporting systems are ready",
							"Verify data collection processes",
						},
						Source: "Internal Monitoring",
					}

					m.changes = append(m.changes, change)

					if m.onChangeDetected != nil {
						m.onChangeDetected(change)
					}
				}
			}
		}
	}
}

// calculateDeadlineImpact determines impact based on days until deadline.
func (m *Monitor) calculateDeadlineImpact(days int) ImpactLevel {
	if days <= 7 {
		return ImpactCritical
	}
	if days <= 30 {
		return ImpactHigh
	}
	if days <= 60 {
		return ImpactMedium
	}
	return ImpactLow
}

// GetRegulations returns all regulations.
func (m *Monitor) GetRegulations() []Regulation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	regs := make([]Regulation, 0, len(m.regulations))
	for _, r := range m.regulations {
		regs = append(regs, r)
	}
	return regs
}

// GetRegulation returns a specific regulation.
func (m *Monitor) GetRegulation(id string) (*Regulation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	reg, ok := m.regulations[id]
	if !ok {
		return nil, fmt.Errorf("regulation not found: %s", id)
	}
	return &reg, nil
}

// GetChanges returns recent changes.
func (m *Monitor) GetChanges(since time.Time) []RegulatoryChange {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var changes []RegulatoryChange
	for _, c := range m.changes {
		if c.DetectedAt.After(since) {
			changes = append(changes, c)
		}
	}
	return changes
}

// =============================================================================
// Alert Service
// =============================================================================

// AlertService manages regulatory alerts.
type AlertService struct {
	monitor     *Monitor
	preferences map[string]AlertPreferences
	alerts      map[string][]Alert
	httpClient  *http.Client
	logger      *slog.Logger
	mu          sync.RWMutex
}

// AlertServiceConfig configures the alert service.
type AlertServiceConfig struct {
	Monitor *Monitor
	Logger  *slog.Logger
}

// NewAlertService creates a new alert service.
func NewAlertService(cfg AlertServiceConfig) *AlertService {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	as := &AlertService{
		monitor:     cfg.Monitor,
		preferences: make(map[string]AlertPreferences),
		alerts:      make(map[string][]Alert),
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		logger:      cfg.Logger.With("component", "alert-service"),
	}

	// Register for changes
	cfg.Monitor.OnChangeDetected(func(change RegulatoryChange) {
		as.processChange(change)
	})

	return as
}

// SetPreferences updates tenant alert preferences.
func (as *AlertService) SetPreferences(prefs AlertPreferences) {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.preferences[prefs.TenantID] = prefs
}

// GetPreferences retrieves tenant preferences.
func (as *AlertService) GetPreferences(tenantID string) (*AlertPreferences, error) {
	as.mu.RLock()
	defer as.mu.RUnlock()

	prefs, ok := as.preferences[tenantID]
	if !ok {
		return nil, fmt.Errorf("preferences not found for tenant: %s", tenantID)
	}
	return &prefs, nil
}

// processChange handles a new regulatory change.
func (as *AlertService) processChange(change RegulatoryChange) {
	as.mu.RLock()
	prefs := as.preferences
	as.mu.RUnlock()

	// Find applicable regulations
	reg, _ := as.monitor.GetRegulation(change.RegulationID)

	for tenantID, pref := range prefs {
		if !as.shouldAlert(pref, change, reg) {
			continue
		}

		alert := Alert{
			ID:            fmt.Sprintf("alert-%s-%d", change.ID, time.Now().UnixNano()),
			TenantID:      tenantID,
			ChangeID:      change.ID,
			Title:         change.Title,
			Message:       change.Summary,
			Impact:        change.Impact,
			Regulations:   []string{change.RegulationID},
			Jurisdictions: []Jurisdiction{},
			Actions:       change.ActionNeeded,
			CreatedAt:     time.Now(),
		}

		if reg != nil {
			alert.Jurisdictions = []Jurisdiction{reg.Jurisdiction}
		}

		as.mu.Lock()
		as.alerts[tenantID] = append(as.alerts[tenantID], alert)
		as.mu.Unlock()

		// Send notifications
		if pref.WebhookURL != "" {
			go as.sendWebhook(pref.WebhookURL, alert)
		}

		as.logger.Info("alert created",
			"alertId", alert.ID,
			"tenantId", tenantID,
			"impact", alert.Impact)
	}
}

// shouldAlert determines if a tenant should receive this alert.
func (as *AlertService) shouldAlert(pref AlertPreferences, change RegulatoryChange, reg *Regulation) bool {
	// Check impact level
	if !as.impactMeetsThreshold(change.Impact, pref.MinImpact) {
		return false
	}

	// Check jurisdiction
	if len(pref.Jurisdictions) > 0 && reg != nil {
		found := false
		for _, j := range pref.Jurisdictions {
			if j == reg.Jurisdiction {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// impactMeetsThreshold checks if impact meets minimum threshold.
func (as *AlertService) impactMeetsThreshold(impact, threshold ImpactLevel) bool {
	levels := map[ImpactLevel]int{
		ImpactInfo:     0,
		ImpactLow:      1,
		ImpactMedium:   2,
		ImpactHigh:     3,
		ImpactCritical: 4,
	}
	return levels[impact] >= levels[threshold]
}

// sendWebhook sends an alert via webhook.
func (as *AlertService) sendWebhook(url string, alert Alert) {
	data, _ := json.Marshal(alert)
	resp, err := as.httpClient.Post(url, "application/json", strings.NewReader(string(data)))
	if err != nil {
		as.logger.Error("webhook failed", "url", url, "error", err)
		return
	}
	resp.Body.Close()
}

// GetAlerts retrieves alerts for a tenant.
func (as *AlertService) GetAlerts(tenantID string, unreadOnly bool) []Alert {
	as.mu.RLock()
	defer as.mu.RUnlock()

	all := as.alerts[tenantID]
	if !unreadOnly {
		return all
	}

	var unread []Alert
	for _, a := range all {
		if a.ReadAt == nil && a.DismissedAt == nil {
			unread = append(unread, a)
		}
	}
	return unread
}

// MarkRead marks an alert as read.
func (as *AlertService) MarkRead(tenantID, alertID string) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	for i, a := range as.alerts[tenantID] {
		if a.ID == alertID {
			now := time.Now()
			as.alerts[tenantID][i].ReadAt = &now
			return nil
		}
	}
	return fmt.Errorf("alert not found: %s", alertID)
}

// =============================================================================
// Compliance Checker
// =============================================================================

// ComplianceStatus represents overall compliance status.
type ComplianceStatus struct {
	TenantID      string                 `json:"tenantId"`
	Regulations   []RegulationCompliance `json:"regulations"`
	OverallScore  float64                `json:"overallScore"` // 0-100
	Gaps          []ComplianceGap        `json:"gaps"`
	NextDeadlines []UpcomingDeadline     `json:"nextDeadlines"`
	LastAssessed  time.Time              `json:"lastAssessed"`
}

// RegulationCompliance tracks compliance with a specific regulation.
type RegulationCompliance struct {
	RegulationID   string              `json:"regulationId"`
	Name           string              `json:"name"`
	Applicable     bool                `json:"applicable"`
	ComplianceRate float64             `json:"complianceRate"` // 0-100
	Requirements   []RequirementStatus `json:"requirements"`
}

// RequirementStatus tracks individual requirement compliance.
type RequirementStatus struct {
	RequirementID string `json:"requirementId"`
	Name          string `json:"name"`
	Compliant     bool   `json:"compliant"`
	Notes         string `json:"notes,omitempty"`
}

// ComplianceGap identifies a compliance gap.
type ComplianceGap struct {
	RegulationID string   `json:"regulationId"`
	Requirement  string   `json:"requirement"`
	Description  string   `json:"description"`
	Priority     string   `json:"priority"` // "high", "medium", "low"
	Remediation  []string `json:"remediation"`
}

// UpcomingDeadline tracks approaching deadlines.
type UpcomingDeadline struct {
	RegulationID string    `json:"regulationId"`
	Name         string    `json:"name"`
	Deadline     time.Time `json:"deadline"`
	DaysUntil    int       `json:"daysUntil"`
	Requirement  string    `json:"requirement,omitempty"`
}

// CheckCompliance assesses tenant compliance status.
func (m *Monitor) CheckCompliance(ctx context.Context, tenantID string, tenantProfile TenantProfile) (*ComplianceStatus, error) {
	m.mu.RLock()
	regulations := m.regulations
	m.mu.RUnlock()

	status := &ComplianceStatus{
		TenantID:      tenantID,
		Regulations:   make([]RegulationCompliance, 0),
		Gaps:          make([]ComplianceGap, 0),
		NextDeadlines: make([]UpcomingDeadline, 0),
		LastAssessed:  time.Now(),
	}

	var totalCompliance float64
	var applicableCount int

	for _, reg := range regulations {
		regCompliance := RegulationCompliance{
			RegulationID: reg.ID,
			Name:         reg.Name,
			Applicable:   m.isApplicable(reg, tenantProfile),
			Requirements: make([]RequirementStatus, len(reg.Requirements)),
		}

		if !regCompliance.Applicable {
			status.Regulations = append(status.Regulations, regCompliance)
			continue
		}

		applicableCount++

		// Check each requirement
		var compliantCount int
		for i, req := range reg.Requirements {
			compliant := m.checkRequirement(tenantProfile, req)
			regCompliance.Requirements[i] = RequirementStatus{
				RequirementID: req.ID,
				Name:          req.Name,
				Compliant:     compliant,
			}
			if compliant {
				compliantCount++
			} else {
				// Add gap
				status.Gaps = append(status.Gaps, ComplianceGap{
					RegulationID: reg.ID,
					Requirement:  req.Name,
					Description:  fmt.Sprintf("Missing compliance with %s requirement: %s", reg.ShortName, req.Name),
					Priority:     "high",
					Remediation:  []string{"Implement required reporting", "Collect necessary data"},
				})
			}
		}

		if len(reg.Requirements) > 0 {
			regCompliance.ComplianceRate = float64(compliantCount) / float64(len(reg.Requirements)) * 100
		}
		totalCompliance += regCompliance.ComplianceRate

		// Check for upcoming deadlines
		if reg.EffectiveDate != nil && reg.EffectiveDate.After(time.Now()) {
			days := int(time.Until(*reg.EffectiveDate).Hours() / 24)
			if days <= 180 {
				status.NextDeadlines = append(status.NextDeadlines, UpcomingDeadline{
					RegulationID: reg.ID,
					Name:         reg.Name,
					Deadline:     *reg.EffectiveDate,
					DaysUntil:    days,
				})
			}
		}

		status.Regulations = append(status.Regulations, regCompliance)
	}

	if applicableCount > 0 {
		status.OverallScore = totalCompliance / float64(applicableCount)
	}

	return status, nil
}

// TenantProfile contains information for applicability checks.
type TenantProfile struct {
	Employees  int      `json:"employees"`
	Revenue    float64  `json:"revenue"`
	IsPublic   bool     `json:"isPublic"`
	Countries  []string `json:"countries"`
	Industries []string `json:"industries"`
	HasScope1  bool     `json:"hasScope1"`
	HasScope2  bool     `json:"hasScope2"`
	HasScope3  bool     `json:"hasScope3"`
}

// isApplicable checks if a regulation applies to the tenant.
func (m *Monitor) isApplicable(reg Regulation, profile TenantProfile) bool {
	for _, app := range reg.Applies {
		switch app.Criterion {
		case "employees":
			// Parse operator and value
			// Simplified check
			if app.Operator == ">" && fmt.Sprintf("%d", profile.Employees) > app.Value {
				return true
			}
		case "revenue":
			// Simplified check
			return profile.Revenue > 40000000
		case "public":
			if app.Value == "true" && profile.IsPublic {
				return true
			}
		}
	}
	return false
}

// checkRequirement verifies if a requirement is met.
func (m *Monitor) checkRequirement(profile TenantProfile, req Requirement) bool {
	// Check if required scopes have data
	for _, scope := range req.Scope {
		switch scope {
		case 1:
			if !profile.HasScope1 {
				return false
			}
		case 2:
			if !profile.HasScope2 {
				return false
			}
		case 3:
			if !profile.HasScope3 {
				return false
			}
		}
	}
	return true
}
