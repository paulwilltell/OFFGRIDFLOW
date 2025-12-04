// Package whitelabel provides embedded carbon accounting API for partners.
//
// This package enables partners to offer carbon accounting functionality
// within their own applications via a white-label API.
package whitelabel

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
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

// Partner represents a white-label partner.
type Partner struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	ContactEmail string          `json:"contactEmail"`
	APIKey       string          `json:"apiKey"`
	APISecret    string          `json:"-"` // Not exposed
	Status       PartnerStatus   `json:"status"`
	Plan         PartnerPlan     `json:"plan"`
	Branding     *BrandingConfig `json:"branding,omitempty"`
	Permissions  []Permission    `json:"permissions"`
	RateLimits   RateLimitConfig `json:"rateLimits"`
	Webhooks     []WebhookConfig `json:"webhooks,omitempty"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
}

// PartnerStatus tracks partner status.
type PartnerStatus string

const (
	StatusPending    PartnerStatus = "pending"
	StatusActive     PartnerStatus = "active"
	StatusSuspended  PartnerStatus = "suspended"
	StatusTerminated PartnerStatus = "terminated"
)

// PartnerPlan defines service tiers.
type PartnerPlan string

const (
	PlanBasic      PartnerPlan = "basic"
	PlanPro        PartnerPlan = "pro"
	PlanEnterprise PartnerPlan = "enterprise"
)

// Permission defines what a partner can do.
type Permission string

const (
	PermEmissionsRead   Permission = "emissions:read"
	PermEmissionsWrite  Permission = "emissions:write"
	PermReportsGenerate Permission = "reports:generate"
	PermBenchmarking    Permission = "benchmarking"
	PermScenarios       Permission = "scenarios"
	PermCredits         Permission = "credits"
)

// BrandingConfig allows partners to customize the UI.
type BrandingConfig struct {
	CompanyName    string            `json:"companyName"`
	LogoURL        string            `json:"logoUrl,omitempty"`
	PrimaryColor   string            `json:"primaryColor,omitempty"`
	SecondaryColor string            `json:"secondaryColor,omitempty"`
	CustomDomain   string            `json:"customDomain,omitempty"`
	CustomCSS      string            `json:"customCss,omitempty"`
	FooterText     string            `json:"footerText,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// RateLimitConfig defines API rate limits.
type RateLimitConfig struct {
	RequestsPerMinute int `json:"requestsPerMinute"`
	RequestsPerDay    int `json:"requestsPerDay"`
	BurstLimit        int `json:"burstLimit"`
}

// WebhookConfig configures partner webhooks.
type WebhookConfig struct {
	ID      string   `json:"id"`
	URL     string   `json:"url"`
	Events  []string `json:"events"`
	Secret  string   `json:"-"`
	Enabled bool     `json:"enabled"`
}

// =============================================================================
// Partner Client
// =============================================================================

// PartnerClient is a client context for partner API calls.
type PartnerClient struct {
	PartnerID   string `json:"partnerId"`
	CustomerID  string `json:"customerId"` // Partner's customer identifier
	TenantID    string `json:"tenantId"`   // Internal tenant mapping
	Permissions []Permission
}

// =============================================================================
// API Service
// =============================================================================

// APIService provides the white-label API.
type APIService struct {
	partners map[string]*Partner
	clients  map[string]*PartnerClient // APIKey -> Client
	logger   *slog.Logger
	mu       sync.RWMutex

	// Backend services
	emissionsService EmissionsService
	reportsService   ReportsService
}

// EmissionsService defines the emissions backend interface.
type EmissionsService interface {
	Calculate(ctx context.Context, tenantID string, input EmissionsInput) (*EmissionsResult, error)
	Get(ctx context.Context, tenantID string, query EmissionsQuery) (*EmissionsData, error)
}

// ReportsService defines the reports backend interface.
type ReportsService interface {
	Generate(ctx context.Context, tenantID string, reportType string) ([]byte, error)
}

// APIServiceConfig configures the API service.
type APIServiceConfig struct {
	EmissionsService EmissionsService
	ReportsService   ReportsService
	Logger           *slog.Logger
}

// NewAPIService creates a new white-label API service.
func NewAPIService(cfg APIServiceConfig) *APIService {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	return &APIService{
		partners:         make(map[string]*Partner),
		clients:          make(map[string]*PartnerClient),
		logger:           cfg.Logger.With("component", "whitelabel-api"),
		emissionsService: cfg.EmissionsService,
		reportsService:   cfg.ReportsService,
	}
}

// =============================================================================
// Partner Management
// =============================================================================

// RegisterPartner creates a new partner account.
func (s *APIService) RegisterPartner(name, email string, plan PartnerPlan) (*Partner, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("partner-%d", time.Now().UnixNano())
	apiKey := s.generateAPIKey()
	apiSecret := s.generateSecret()

	partner := &Partner{
		ID:           id,
		Name:         name,
		ContactEmail: email,
		APIKey:       apiKey,
		APISecret:    apiSecret,
		Status:       StatusPending,
		Plan:         plan,
		Permissions:  s.getDefaultPermissions(plan),
		RateLimits:   s.getRateLimits(plan),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	s.partners[id] = partner
	s.clients[apiKey] = &PartnerClient{
		PartnerID:   id,
		Permissions: partner.Permissions,
	}

	s.logger.Info("partner registered",
		"partnerId", id,
		"name", name,
		"plan", plan)

	return partner, nil
}

// generateAPIKey creates a new API key.
func (s *APIService) generateAPIKey() string {
	// In production, use crypto/rand
	h := sha256.Sum256([]byte(fmt.Sprintf("apikey-%d", time.Now().UnixNano())))
	return "ogf_" + hex.EncodeToString(h[:])[:32]
}

// generateSecret creates a new secret.
func (s *APIService) generateSecret() string {
	h := sha256.Sum256([]byte(fmt.Sprintf("secret-%d", time.Now().UnixNano())))
	return hex.EncodeToString(h[:])
}

// getDefaultPermissions returns permissions for a plan.
func (s *APIService) getDefaultPermissions(plan PartnerPlan) []Permission {
	switch plan {
	case PlanEnterprise:
		return []Permission{
			PermEmissionsRead,
			PermEmissionsWrite,
			PermReportsGenerate,
			PermBenchmarking,
			PermScenarios,
			PermCredits,
		}
	case PlanPro:
		return []Permission{
			PermEmissionsRead,
			PermEmissionsWrite,
			PermReportsGenerate,
			PermBenchmarking,
		}
	default:
		return []Permission{
			PermEmissionsRead,
			PermEmissionsWrite,
			PermReportsGenerate,
		}
	}
}

// getRateLimits returns rate limits for a plan.
func (s *APIService) getRateLimits(plan PartnerPlan) RateLimitConfig {
	switch plan {
	case PlanEnterprise:
		return RateLimitConfig{
			RequestsPerMinute: 1000,
			RequestsPerDay:    100000,
			BurstLimit:        200,
		}
	case PlanPro:
		return RateLimitConfig{
			RequestsPerMinute: 300,
			RequestsPerDay:    30000,
			BurstLimit:        50,
		}
	default:
		return RateLimitConfig{
			RequestsPerMinute: 60,
			RequestsPerDay:    5000,
			BurstLimit:        10,
		}
	}
}

// ActivatePartner activates a pending partner.
func (s *APIService) ActivatePartner(partnerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	partner, ok := s.partners[partnerID]
	if !ok {
		return errors.New("partner not found")
	}

	partner.Status = StatusActive
	partner.UpdatedAt = time.Now()

	s.logger.Info("partner activated", "partnerId", partnerID)
	return nil
}

// GetPartner retrieves a partner.
func (s *APIService) GetPartner(partnerID string) (*Partner, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	partner, ok := s.partners[partnerID]
	if !ok {
		return nil, errors.New("partner not found")
	}
	return partner, nil
}

// UpdateBranding updates partner branding.
func (s *APIService) UpdateBranding(partnerID string, branding BrandingConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	partner, ok := s.partners[partnerID]
	if !ok {
		return errors.New("partner not found")
	}

	partner.Branding = &branding
	partner.UpdatedAt = time.Now()

	return nil
}

// =============================================================================
// Authentication
// =============================================================================

// AuthenticateRequest authenticates a partner API request.
func (s *APIService) AuthenticateRequest(apiKey, signature, timestamp string) (*PartnerClient, error) {
	s.mu.RLock()
	client, ok := s.clients[apiKey]
	s.mu.RUnlock()

	if !ok {
		return nil, errors.New("invalid API key")
	}

	// Get partner for secret
	s.mu.RLock()
	partner := s.partners[client.PartnerID]
	s.mu.RUnlock()

	if partner == nil || partner.Status != StatusActive {
		return nil, errors.New("partner not active")
	}

	// Verify signature (HMAC-SHA256)
	if !s.verifySignature(signature, timestamp, partner.APISecret) {
		return nil, errors.New("invalid signature")
	}

	return client, nil
}

// verifySignature verifies HMAC signature.
func (s *APIService) verifySignature(signature, timestamp, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}

// =============================================================================
// Customer Management
// =============================================================================

// CreateCustomer creates a customer for a partner.
func (s *APIService) CreateCustomer(partnerID, customerID, name string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	partner, ok := s.partners[partnerID]
	if !ok {
		return "", errors.New("partner not found")
	}

	if partner.Status != StatusActive {
		return "", errors.New("partner not active")
	}

	// Create internal tenant for this customer
	tenantID := fmt.Sprintf("tenant-%s-%s", partnerID[:8], customerID)

	// Store customer mapping
	key := fmt.Sprintf("%s:%s", partnerID, customerID)
	s.clients[key] = &PartnerClient{
		PartnerID:   partnerID,
		CustomerID:  customerID,
		TenantID:    tenantID,
		Permissions: partner.Permissions,
	}

	s.logger.Info("customer created",
		"partnerId", partnerID,
		"customerId", customerID,
		"tenantId", tenantID)

	return tenantID, nil
}

// GetCustomer retrieves a customer context.
func (s *APIService) GetCustomer(partnerID, customerID string) (*PartnerClient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", partnerID, customerID)
	client, ok := s.clients[key]
	if !ok {
		return nil, errors.New("customer not found")
	}
	return client, nil
}

// =============================================================================
// API Endpoints
// =============================================================================

// EmissionsInput is input for emissions calculation.
type EmissionsInput struct {
	Source   string                 `json:"source"`
	Type     string                 `json:"type"`
	Quantity float64                `json:"quantity"`
	Unit     string                 `json:"unit"`
	Date     time.Time              `json:"date"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// EmissionsResult is the calculation result.
type EmissionsResult struct {
	ID          string    `json:"id"`
	CO2e        float64   `json:"co2e"`
	CO2         float64   `json:"co2"`
	CH4         float64   `json:"ch4,omitempty"`
	N2O         float64   `json:"n2o,omitempty"`
	Scope       int       `json:"scope"`
	Category    string    `json:"category"`
	Methodology string    `json:"methodology"`
	Timestamp   time.Time `json:"timestamp"`
}

// EmissionsQuery queries emissions data.
type EmissionsQuery struct {
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	Scope     *int      `json:"scope,omitempty"`
	Category  string    `json:"category,omitempty"`
}

// EmissionsData contains emissions data.
type EmissionsData struct {
	TotalCO2e float64            `json:"totalCo2e"`
	Scope1    float64            `json:"scope1"`
	Scope2    float64            `json:"scope2"`
	Scope3    float64            `json:"scope3"`
	ByMonth   []MonthlyEmissions `json:"byMonth,omitempty"`
}

// MonthlyEmissions contains monthly breakdown.
type MonthlyEmissions struct {
	Month string  `json:"month"`
	CO2e  float64 `json:"co2e"`
}

// CalculateEmissions calculates emissions for a partner customer.
func (s *APIService) CalculateEmissions(ctx context.Context, client *PartnerClient, input EmissionsInput) (*EmissionsResult, error) {
	if !s.hasPermission(client, PermEmissionsWrite) {
		return nil, errors.New("permission denied")
	}

	return s.emissionsService.Calculate(ctx, client.TenantID, input)
}

// GetEmissions retrieves emissions for a partner customer.
func (s *APIService) GetEmissions(ctx context.Context, client *PartnerClient, query EmissionsQuery) (*EmissionsData, error) {
	if !s.hasPermission(client, PermEmissionsRead) {
		return nil, errors.New("permission denied")
	}

	return s.emissionsService.Get(ctx, client.TenantID, query)
}

// GenerateReport generates a report for a partner customer.
func (s *APIService) GenerateReport(ctx context.Context, client *PartnerClient, reportType string) ([]byte, error) {
	if !s.hasPermission(client, PermReportsGenerate) {
		return nil, errors.New("permission denied")
	}

	return s.reportsService.Generate(ctx, client.TenantID, reportType)
}

// hasPermission checks if client has a permission.
func (s *APIService) hasPermission(client *PartnerClient, perm Permission) bool {
	for _, p := range client.Permissions {
		if p == perm {
			return true
		}
	}
	return false
}

// =============================================================================
// HTTP Handler
// =============================================================================

// Handler wraps the API service for HTTP.
type Handler struct {
	service *APIService
	logger  *slog.Logger
}

// NewHandler creates a new HTTP handler.
func NewHandler(service *APIService, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// ServeHTTP handles API requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract authentication
	apiKey := r.Header.Get("X-API-Key")
	signature := r.Header.Get("X-Signature")
	timestamp := r.Header.Get("X-Timestamp")
	customerID := r.Header.Get("X-Customer-ID")

	if apiKey == "" {
		h.error(w, "missing API key", http.StatusUnauthorized)
		return
	}

	client, err := h.service.AuthenticateRequest(apiKey, signature, timestamp)
	if err != nil {
		h.error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Get customer context if provided
	if customerID != "" {
		client, err = h.service.GetCustomer(client.PartnerID, customerID)
		if err != nil {
			h.error(w, err.Error(), http.StatusNotFound)
			return
		}
	}

	// Route request
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/partner")
	switch {
	case strings.HasPrefix(path, "/emissions") && r.Method == http.MethodPost:
		h.handleCalculateEmissions(w, r, client)
	case strings.HasPrefix(path, "/emissions") && r.Method == http.MethodGet:
		h.handleGetEmissions(w, r, client)
	case strings.HasPrefix(path, "/reports"):
		h.handleGenerateReport(w, r, client)
	case strings.HasPrefix(path, "/customers") && r.Method == http.MethodPost:
		h.handleCreateCustomer(w, r, client)
	default:
		h.error(w, "not found", http.StatusNotFound)
	}
}

func (h *Handler) handleCalculateEmissions(w http.ResponseWriter, r *http.Request, client *PartnerClient) {
	var input EmissionsInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.error(w, "invalid input", http.StatusBadRequest)
		return
	}

	result, err := h.service.CalculateEmissions(r.Context(), client, input)
	if err != nil {
		h.error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, result)
}

func (h *Handler) handleGetEmissions(w http.ResponseWriter, r *http.Request, client *PartnerClient) {
	query := EmissionsQuery{
		StartDate: time.Now().AddDate(-1, 0, 0),
		EndDate:   time.Now(),
	}

	data, err := h.service.GetEmissions(r.Context(), client, query)
	if err != nil {
		h.error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, data)
}

func (h *Handler) handleGenerateReport(w http.ResponseWriter, r *http.Request, client *PartnerClient) {
	reportType := r.URL.Query().Get("type")
	if reportType == "" {
		reportType = "summary"
	}

	data, err := h.service.GenerateReport(r.Context(), client, reportType)
	if err != nil {
		h.error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Write(data)
}

func (h *Handler) handleCreateCustomer(w http.ResponseWriter, r *http.Request, client *PartnerClient) {
	var req struct {
		CustomerID string `json:"customerId"`
		Name       string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, "invalid input", http.StatusBadRequest)
		return
	}

	tenantID, err := h.service.CreateCustomer(client.PartnerID, req.CustomerID, req.Name)
	if err != nil {
		h.error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]string{
		"customerId": req.CustomerID,
		"tenantId":   tenantID,
	})
}

func (h *Handler) json(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) error(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// =============================================================================
// Webhook Delivery
// =============================================================================

// WebhookEvent represents a webhook event.
type WebhookEvent struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	PartnerID string      `json:"partnerId"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// WebhookDelivery delivers webhooks to partners.
type WebhookDelivery struct {
	httpClient *http.Client
	logger     *slog.Logger
}

// NewWebhookDelivery creates a new webhook delivery service.
func NewWebhookDelivery(logger *slog.Logger) *WebhookDelivery {
	return &WebhookDelivery{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		logger:     logger.With("component", "webhook-delivery"),
	}
}

// Deliver sends a webhook event.
func (wd *WebhookDelivery) Deliver(webhook WebhookConfig, event WebhookEvent) error {
	if !webhook.Enabled {
		return nil
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, webhook.URL, strings.NewReader(string(data)))
	if err != nil {
		return err
	}

	// Sign the request
	signature := wd.sign(data, webhook.Secret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("X-Webhook-ID", event.ID)

	resp, err := wd.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook failed: %d", resp.StatusCode)
	}

	wd.logger.Info("webhook delivered",
		"webhookId", webhook.ID,
		"eventType", event.Type)

	return nil
}

func (wd *WebhookDelivery) sign(data []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}
