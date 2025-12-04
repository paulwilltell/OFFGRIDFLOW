// Package residency provides multi-region data residency for compliance.
//
// This package enables tenant data to be stored in appropriate geographic regions
// for GDPR, SEC, and other regulatory compliance requirements.
package residency

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

// =============================================================================
// Region Types
// =============================================================================

// Region represents a geographic data region.
type Region string

const (
	RegionEU   Region = "eu"   // European Union (Frankfurt)
	RegionUS   Region = "us"   // United States (Virginia)
	RegionAPAC Region = "apac" // Asia-Pacific (Singapore)
	RegionUK   Region = "uk"   // United Kingdom (London)
	RegionAU   Region = "au"   // Australia (Sydney)
)

// RegionConfig contains region-specific configuration.
type RegionConfig struct {
	Region      Region   `json:"region"`
	Name        string   `json:"name"`
	Location    string   `json:"location"`
	DSN         string   `json:"-"`           // Database connection string
	Regulations []string `json:"regulations"` // GDPR, SEC, etc.
	Primary     bool     `json:"primary"`     // Is this the primary region?
}

// DefaultRegions returns standard region configurations.
func DefaultRegions() map[Region]RegionConfig {
	return map[Region]RegionConfig{
		RegionEU: {
			Region:      RegionEU,
			Name:        "European Union",
			Location:    "Frankfurt, Germany",
			Regulations: []string{"GDPR", "CSRD", "CBAM"},
			Primary:     false,
		},
		RegionUS: {
			Region:      RegionUS,
			Name:        "United States",
			Location:    "Virginia, USA",
			Regulations: []string{"SEC", "California Climate"},
			Primary:     true,
		},
		RegionAPAC: {
			Region:      RegionAPAC,
			Name:        "Asia-Pacific",
			Location:    "Singapore",
			Regulations: []string{"PDPA"},
			Primary:     false,
		},
		RegionUK: {
			Region:      RegionUK,
			Name:        "United Kingdom",
			Location:    "London, UK",
			Regulations: []string{"UK GDPR", "TCFD"},
			Primary:     false,
		},
	}
}

// =============================================================================
// Tenant Residency
// =============================================================================

// TenantResidency defines where a tenant's data should reside.
type TenantResidency struct {
	TenantID      string            `json:"tenantId"`
	PrimaryRegion Region            `json:"primaryRegion"`
	Regulations   []string          `json:"regulations,omitempty"`
	Country       string            `json:"country,omitempty"`
	DataTypes     map[string]Region `json:"dataTypes,omitempty"` // Per-data-type regions
}

// =============================================================================
// Router
// =============================================================================

// Router directs database operations to the correct regional cluster.
type Router struct {
	regions  map[Region]*sql.DB
	tenants  map[string]TenantResidency
	fallback Region
	logger   *slog.Logger
	mu       sync.RWMutex
}

// RouterConfig configures the regional router.
type RouterConfig struct {
	Regions        map[Region]string // Region -> DSN mapping
	FallbackRegion Region
	Logger         *slog.Logger
}

// NewRouter creates a new regional database router.
func NewRouter(cfg RouterConfig) (*Router, error) {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	r := &Router{
		regions:  make(map[Region]*sql.DB),
		tenants:  make(map[string]TenantResidency),
		fallback: cfg.FallbackRegion,
		logger:   cfg.Logger.With("component", "residency-router"),
	}

	// Connect to each region
	for region, dsn := range cfg.Regions {
		db, err := sql.Open("pgx", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to %s: %w", region, err)
		}
		if err := db.Ping(); err != nil {
			r.logger.Warn("region not available", "region", region, "error", err)
			continue
		}
		r.regions[region] = db
		r.logger.Info("connected to regional cluster", "region", region)
	}

	if len(r.regions) == 0 {
		return nil, errors.New("no regional databases available")
	}

	return r, nil
}

// RegisterTenant sets the residency configuration for a tenant.
func (r *Router) RegisterTenant(residency TenantResidency) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tenants[residency.TenantID] = residency
	r.logger.Info("tenant residency registered",
		"tenantId", residency.TenantID,
		"region", residency.PrimaryRegion)
}

// GetDB returns the database connection for a tenant.
func (r *Router) GetDB(tenantID string) (*sql.DB, Region, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check if tenant has specific residency
	if residency, ok := r.tenants[tenantID]; ok {
		if db, ok := r.regions[residency.PrimaryRegion]; ok {
			return db, residency.PrimaryRegion, nil
		}
	}

	// Fall back to default region
	if db, ok := r.regions[r.fallback]; ok {
		return db, r.fallback, nil
	}

	// Use any available region
	for region, db := range r.regions {
		return db, region, nil
	}

	return nil, "", errors.New("no database available")
}

// GetDBForDataType returns the database for a specific data type.
func (r *Router) GetDBForDataType(tenantID, dataType string) (*sql.DB, Region, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check for data-type-specific routing
	if residency, ok := r.tenants[tenantID]; ok {
		if region, ok := residency.DataTypes[dataType]; ok {
			if db, ok := r.regions[region]; ok {
				return db, region, nil
			}
		}
	}

	// Fall back to primary
	return r.GetDB(tenantID)
}

// GetAllDBs returns all regional database connections.
func (r *Router) GetAllDBs() map[Region]*sql.DB {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[Region]*sql.DB)
	for k, v := range r.regions {
		result[k] = v
	}
	return result
}

// Close closes all database connections.
func (r *Router) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error
	for region, db := range r.regions {
		if err := db.Close(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", region, err))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// =============================================================================
// Regional Query Executor
// =============================================================================

// Executor provides region-aware query execution.
type Executor struct {
	router *Router
	logger *slog.Logger
}

// NewExecutor creates a new regional query executor.
func NewExecutor(router *Router, logger *slog.Logger) *Executor {
	return &Executor{router: router, logger: logger}
}

// Query executes a query in the tenant's primary region.
func (e *Executor) Query(ctx context.Context, tenantID, query string, args ...interface{}) (*sql.Rows, error) {
	db, region, err := e.router.GetDB(tenantID)
	if err != nil {
		return nil, err
	}

	e.logger.Debug("executing query",
		"tenantId", tenantID,
		"region", region)

	return db.QueryContext(ctx, query, args...)
}

// Exec executes a write operation in the tenant's primary region.
func (e *Executor) Exec(ctx context.Context, tenantID, query string, args ...interface{}) (sql.Result, error) {
	db, region, err := e.router.GetDB(tenantID)
	if err != nil {
		return nil, err
	}

	e.logger.Debug("executing write",
		"tenantId", tenantID,
		"region", region)

	return db.ExecContext(ctx, query, args...)
}

// QueryAll executes a query across all regions and combines results.
func (e *Executor) QueryAll(ctx context.Context, query string, args ...interface{}) (map[Region]*sql.Rows, error) {
	dbs := e.router.GetAllDBs()
	results := make(map[Region]*sql.Rows)

	for region, db := range dbs {
		rows, err := db.QueryContext(ctx, query, args...)
		if err != nil {
			e.logger.Warn("query failed in region",
				"region", region,
				"error", err)
			continue
		}
		results[region] = rows
	}

	return results, nil
}

// =============================================================================
// Region Determination
// =============================================================================

// DetermineRegion selects the appropriate region based on country/regulations.
func DetermineRegion(country string, regulations []string) Region {
	// EU countries
	euCountries := map[string]bool{
		"DE": true, "FR": true, "IT": true, "ES": true, "NL": true,
		"BE": true, "AT": true, "PL": true, "SE": true, "DK": true,
		"FI": true, "IE": true, "PT": true, "GR": true, "CZ": true,
	}

	// Check for explicit GDPR requirement
	for _, reg := range regulations {
		if reg == "GDPR" || reg == "CSRD" {
			return RegionEU
		}
	}

	// Check country
	if euCountries[country] {
		return RegionEU
	}
	if country == "GB" || country == "UK" {
		return RegionUK
	}
	if country == "US" {
		return RegionUS
	}
	if country == "AU" || country == "NZ" {
		return RegionAU
	}

	// APAC countries
	apacCountries := map[string]bool{
		"SG": true, "JP": true, "KR": true, "HK": true, "CN": true,
		"IN": true, "TH": true, "MY": true, "ID": true, "PH": true,
	}
	if apacCountries[country] {
		return RegionAPAC
	}

	// Default to US
	return RegionUS
}

// =============================================================================
// Compliance Helpers
// =============================================================================

// DataTransferCheck verifies if data transfer between regions is allowed.
type DataTransferCheck struct {
	SourceRegion      Region
	DestinationRegion Region
	DataType          string
	Allowed           bool
	Reason            string
	Requirements      []string
}

// CheckDataTransfer evaluates if data can be transferred between regions.
func CheckDataTransfer(source, dest Region, dataType string) DataTransferCheck {
	check := DataTransferCheck{
		SourceRegion:      source,
		DestinationRegion: dest,
		DataType:          dataType,
	}

	// Same region is always allowed
	if source == dest {
		check.Allowed = true
		check.Reason = "Same region transfer"
		return check
	}

	// EU -> Non-EU requires safeguards
	if source == RegionEU && dest != RegionEU && dest != RegionUK {
		check.Allowed = true // With conditions
		check.Reason = "GDPR cross-border transfer rules apply"
		check.Requirements = []string{
			"Standard Contractual Clauses (SCCs)",
			"Data Processing Agreement",
			"Transfer Impact Assessment",
		}
		return check
	}

	// UK has adequacy with EU
	if (source == RegionEU && dest == RegionUK) || (source == RegionUK && dest == RegionEU) {
		check.Allowed = true
		check.Reason = "EU-UK adequacy decision"
		return check
	}

	// General case - allowed with standard agreements
	check.Allowed = true
	check.Reason = "Standard data transfer"
	check.Requirements = []string{"Data Processing Agreement"}

	return check
}

// ValidateResidency checks if tenant residency meets regulatory requirements.
func ValidateResidency(residency TenantResidency) []string {
	var issues []string

	// Check GDPR compliance
	for _, reg := range residency.Regulations {
		if reg == "GDPR" && residency.PrimaryRegion != RegionEU && residency.PrimaryRegion != RegionUK {
			issues = append(issues, "GDPR requires EU data residency")
		}
		if reg == "CSRD" && residency.PrimaryRegion != RegionEU {
			issues = append(issues, "CSRD reporting data should be in EU region")
		}
	}

	return issues
}
