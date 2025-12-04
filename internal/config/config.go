// Package config provides centralized configuration loading for OffGridFlow.
// It reads configuration from environment variables with sensible defaults
// and validation to fail fast on misconfiguration.
//
// Environment variable naming convention:
//   - OFFGRIDFLOW_* prefix for application-specific settings
//   - Standard names (PORT, APP_ENV) for platform conventions
//   - STRIPE_* prefix for Stripe-related settings
//
// Usage:
//
//	cfg, err := config.Load()
//	if err != nil {
//	    log.Fatalf("configuration error: %v", err)
//	}
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// =============================================================================
// Environment Constants
// =============================================================================

const (
	// EnvDevelopment is the development environment.
	EnvDevelopment = "development"

	// EnvStaging is the staging/preview environment.
	EnvStaging = "staging"

	// EnvProduction is the production environment.
	EnvProduction = "production"

	// EnvTest is the test environment.
	EnvTest = "test"
)

// =============================================================================
// Default Values
// =============================================================================

const (
	defaultHTTPPort               = 8090 // Avoids conflict with common services (80, 8080)
	defaultEnv                    = EnvDevelopment
	defaultOpenAIModel            = "gpt-4o-mini"
	defaultReadTimeout            = 30 * time.Second
	defaultWriteTimeout           = 30 * time.Second
	defaultIdleTimeout            = 120 * time.Second
	defaultIngestLookbackDays     = 30
	defaultIngestScheduleInterval = 30 * time.Minute
)

// =============================================================================
// Environment Variable Keys
// =============================================================================

const (
	// Server configuration
	envHTTPPort       = "OFFGRIDFLOW_HTTP_PORT"
	envPortFallback   = "PORT" // Platform convention fallback
	envAppEnv         = "OFFGRIDFLOW_APP_ENV"
	envAppEnvLegacy   = "APP_ENV"
	envReadTimeout    = "OFFGRIDFLOW_READ_TIMEOUT"
	envWriteTimeout   = "OFFGRIDFLOW_WRITE_TIMEOUT"
	envIdleTimeout    = "OFFGRIDFLOW_IDLE_TIMEOUT"
	envTrustedProxies = "OFFGRIDFLOW_TRUSTED_PROXIES"

	// Database configuration
	envDBDSN             = "OFFGRIDFLOW_DB_DSN"
	envDBMaxOpenConns    = "OFFGRIDFLOW_DB_MAX_OPEN_CONNS"
	envDBMaxIdleConns    = "OFFGRIDFLOW_DB_MAX_IDLE_CONNS"
	envDBConnMaxLifetime = "OFFGRIDFLOW_DB_CONN_MAX_LIFETIME"

	// Authentication
	envAPIKey    = "OFFGRIDFLOW_API_KEY"
	envJWTSecret = "OFFGRIDFLOW_JWT_SECRET"

	// OpenAI configuration
	envOpenAIKey     = "OFFGRIDFLOW_OPENAI_API_KEY"
	envOpenAIModel   = "OFFGRIDFLOW_OPENAI_MODEL"
	envOpenAIBaseURL = "OFFGRIDFLOW_OPENAI_BASE_URL"

	// Stripe configuration
	envStripeSecretKey       = "STRIPE_SECRET_KEY"
	envStripeWebhookSecret   = "STRIPE_WEBHOOK_SECRET"
	envStripePriceFree       = "STRIPE_PRICE_FREE"
	envStripePriceBasic      = "STRIPE_PRICE_BASIC"
	envStripePricePro        = "STRIPE_PRICE_PRO"
	envStripePriceEnterprise = "STRIPE_PRICE_ENTERPRISE"

	// Feature flags
	envEnableAuditLog  = "OFFGRIDFLOW_ENABLE_AUDIT_LOG"
	envEnableMetrics   = "OFFGRIDFLOW_ENABLE_METRICS"
	envEnableGraphQL   = "OFFGRIDFLOW_ENABLE_GRAPHQL"
	envEnableOfflineAI = "OFFGRIDFLOW_ENABLE_OFFLINE_AI"

	// Ingestion configuration
	envIngestLookbackDays     = "OFFGRIDFLOW_INGEST_LOOKBACK_DAYS"
	envIngestScheduleInterval = "OFFGRIDFLOW_INGESTION_SCHEDULE_INTERVAL"

	envAWSIngestEnabled   = "OFFGRIDFLOW_AWS_INGEST_ENABLED"
	envAWSAccessKeyID     = "OFFGRIDFLOW_AWS_ACCESS_KEY_ID"
	envAWSSecretAccessKey = "OFFGRIDFLOW_AWS_SECRET_ACCESS_KEY"
	envAWSRegion          = "OFFGRIDFLOW_AWS_REGION"
	envAWSRoleARN         = "OFFGRIDFLOW_AWS_ROLE_ARN"
	envAWSAccountID       = "OFFGRIDFLOW_AWS_ACCOUNT_ID"
	envAWSBucket          = "OFFGRIDFLOW_AWS_S3_BUCKET"
	envAWSPrefix          = "OFFGRIDFLOW_AWS_S3_PREFIX"
	envAWSOrgID           = "OFFGRIDFLOW_AWS_ORG_ID"

	envAzureIngestEnabled  = "OFFGRIDFLOW_AZURE_INGEST_ENABLED"
	envAzureTenantID       = "OFFGRIDFLOW_AZURE_TENANT_ID"
	envAzureClientID       = "OFFGRIDFLOW_AZURE_CLIENT_ID"
	envAzureClientSecret   = "OFFGRIDFLOW_AZURE_CLIENT_SECRET"
	envAzureSubscriptionID = "OFFGRIDFLOW_AZURE_SUBSCRIPTION_ID"
	envAzureOrgID          = "OFFGRIDFLOW_AZURE_ORG_ID"

	envGCPIngestEnabled     = "OFFGRIDFLOW_GCP_INGEST_ENABLED"
	envGCPProjectID         = "OFFGRIDFLOW_GCP_PROJECT_ID"
	envGCPBillingAccountID  = "OFFGRIDFLOW_GCP_BILLING_ACCOUNT_ID"
	envGCPBigQueryDataset   = "OFFGRIDFLOW_GCP_BIGQUERY_DATASET"
	envGCPBigQueryTable     = "OFFGRIDFLOW_GCP_BIGQUERY_TABLE"
	envGCPServiceAccountKey = "OFFGRIDFLOW_GCP_SERVICE_ACCOUNT_KEY"
	envGCPOrgID             = "OFFGRIDFLOW_GCP_ORG_ID"

	envSAPIngestEnabled = "OFFGRIDFLOW_SAP_INGEST_ENABLED"
	envSAPOrgID         = "OFFGRIDFLOW_SAP_ORG_ID"
	envSAPBaseURL       = "OFFGRIDFLOW_SAP_BASE_URL"
	envSAPClientID      = "OFFGRIDFLOW_SAP_CLIENT_ID"
	envSAPClientSecret  = "OFFGRIDFLOW_SAP_CLIENT_SECRET"
	envSAPCompany       = "OFFGRIDFLOW_SAP_COMPANY"
	envSAPPlant         = "OFFGRIDFLOW_SAP_PLANT"

	envUtilityIngestEnabled = "OFFGRIDFLOW_UTILITY_INGEST_ENABLED"
	envUtilityOrgID         = "OFFGRIDFLOW_UTILITY_ORG_ID"
)

// =============================================================================
// Configuration Structs
// =============================================================================

// Config holds all application configuration.
// Fields are grouped by domain for clarity.
type Config struct {
	// Server holds HTTP server configuration.
	Server ServerConfig

	// Database holds PostgreSQL connection configuration.
	Database DatabaseConfig

	// Auth holds authentication configuration.
	Auth AuthConfig

	// OpenAI holds OpenAI API configuration.
	OpenAI OpenAIConfig

	// Stripe holds Stripe payment configuration.
	Stripe StripeConfig

	// Ingestion holds ingestion connector configuration.
	Ingestion IngestionConfig

	// Features holds feature flag configuration.
	Features FeatureConfig
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	// Port is the HTTP server listen port.
	Port int `json:"port"`

	// Env is the application environment (development, staging, production).
	Env string `json:"env"`

	// ReadTimeout is the maximum duration for reading the entire request.
	ReadTimeout time.Duration `json:"read_timeout"`

	// WriteTimeout is the maximum duration for writing the response.
	WriteTimeout time.Duration `json:"write_timeout"`

	// IdleTimeout is the maximum time to wait for the next request.
	IdleTimeout time.Duration `json:"idle_timeout"`

	// TrustedProxies is a list of trusted proxy IP addresses/CIDRs.
	TrustedProxies []string `json:"trusted_proxies,omitempty"`
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	// DSN is the PostgreSQL connection string.
	// Format: postgres://user:pass@host:port/database?sslmode=disable
	DSN string `json:"-"` // Excluded from JSON to prevent logging

	// MaxOpenConns is the maximum number of open connections.
	MaxOpenConns int `json:"max_open_conns"`

	// MaxIdleConns is the maximum number of idle connections.
	MaxIdleConns int `json:"max_idle_conns"`

	// ConnMaxLifetime is the maximum lifetime of a connection.
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
}

// AuthConfig holds authentication settings.
type AuthConfig struct {
	// APIKey is the static API key for basic authentication.
	// Used for service-to-service communication.
	APIKey string `json:"-"` // Excluded from JSON

	// JWTSecret is the secret key for signing JWT tokens.
	JWTSecret string `json:"-"` // Excluded from JSON

	// HasAPIKey returns true if an API key is configured.
	HasAPIKey bool `json:"has_api_key"`

	// HasJWTSecret returns true if a JWT secret is configured.
	HasJWTSecret bool `json:"has_jwt_secret"`
}

// OpenAIConfig holds OpenAI API settings.
type OpenAIConfig struct {
	// APIKey is the OpenAI API key.
	APIKey string `json:"-"` // Excluded from JSON

	// Model is the model to use (e.g., gpt-4o-mini, gpt-4o).
	Model string `json:"model"`

	// BaseURL allows overriding the API endpoint (for Azure OpenAI, proxies).
	BaseURL string `json:"base_url,omitempty"`

	// IsConfigured returns true if an API key is set.
	IsConfigured bool `json:"is_configured"`
}

// StripeConfig holds Stripe payment settings.
type StripeConfig struct {
	// SecretKey is the Stripe secret API key.
	SecretKey string `json:"-"` // Excluded from JSON

	// WebhookSecret is used to verify Stripe webhook signatures.
	WebhookSecret string `json:"-"` // Excluded from JSON

	// PriceFree is the Stripe price ID for the free plan.
	PriceFree string `json:"price_free,omitempty"`

	// PriceBasic is the Stripe price ID for the basic plan.
	PriceBasic string `json:"price_basic,omitempty"`

	// PricePro is the Stripe price ID for the pro plan.
	PricePro string `json:"price_pro,omitempty"`

	// PriceEnterprise is the Stripe price ID for the enterprise plan.
	PriceEnterprise string `json:"price_enterprise,omitempty"`

	// IsConfigured returns true if Stripe keys are set.
	IsConfigured bool `json:"is_configured"`
}

// FeatureConfig holds feature flag settings.
type FeatureConfig struct {
	// EnableAuditLog enables detailed audit logging.
	EnableAuditLog bool `json:"enable_audit_log"`

	// EnableMetrics enables Prometheus metrics endpoint.
	EnableMetrics bool `json:"enable_metrics"`

	// EnableGraphQL enables the GraphQL API.
	EnableGraphQL bool `json:"enable_graphql"`

	// EnableOfflineAI enables local AI inference fallback.
	EnableOfflineAI bool `json:"enable_offline_ai"`
}

// IngestionConfig groups all ingestion adapter settings.
type IngestionConfig struct {
	LookbackDays     int           `json:"lookback_days"`
	ScheduleInterval time.Duration `json:"schedule_interval"`

	AWS     AWSIngestionConfig   `json:"aws"`
	Azure   AzureIngestionConfig `json:"azure"`
	GCP     GCPIngestionConfig   `json:"gcp"`
	SAP     SAPIngestionConfig   `json:"sap"`
	Utility UtilityIngestConfig  `json:"utility"`
}

// AWSIngestionConfig configures the AWS adapter.
type AWSIngestionConfig struct {
	Enabled         bool   `json:"enabled"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"-"` // Excluded from JSON
	Region          string `json:"region"`
	RoleARN         string `json:"role_arn,omitempty"`
	AccountID       string `json:"account_id,omitempty"`
	Bucket          string `json:"bucket,omitempty"`
	Prefix          string `json:"prefix,omitempty"`
	OrgID           string `json:"org_id"`
}

// AzureIngestionConfig configures the Azure adapter.
type AzureIngestionConfig struct {
	Enabled        bool   `json:"enabled"`
	TenantID       string `json:"tenant_id"`
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	SubscriptionID string `json:"subscription_id"`
	OrgID          string `json:"org_id"`
}

// GCPIngestionConfig configures the GCP adapter.
type GCPIngestionConfig struct {
	Enabled           bool   `json:"enabled"`
	ProjectID         string `json:"project_id"`
	BillingAccountID  string `json:"billing_account_id"`
	BigQueryDataset   string `json:"bigquery_dataset"`
	BigQueryTable     string `json:"bigquery_table"`
	ServiceAccountKey string `json:"service_account_key"`
	OrgID             string `json:"org_id"`
}

// SAPIngestionConfig configures the SAP adapter.
type SAPIngestionConfig struct {
	Enabled      bool   `json:"enabled"`
	OrgID        string `json:"org_id"`
	BaseURL      string `json:"base_url"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"-"` // Excluded from JSON
	Company      string `json:"company"`
	Plant        string `json:"plant,omitempty"`
}

// UtilityIngestConfig configures the utility bills adapter.
type UtilityIngestConfig struct {
	Enabled bool   `json:"enabled"`
	OrgID   string `json:"org_id"`
}

// =============================================================================
// Configuration Loading
// =============================================================================

// Load reads configuration from environment variables and returns a validated Config.
// Returns an error if required configuration is missing or invalid in production.
func Load() (Config, error) {
	cfg := Config{
		Server:    loadServerConfig(),
		Database:  loadDatabaseConfig(),
		Auth:      loadAuthConfig(),
		OpenAI:    loadOpenAIConfig(),
		Stripe:    loadStripeConfig(),
		Ingestion: loadIngestionConfig(),
		Features:  loadFeatureConfig(),
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// MustLoad is like Load but panics on error.
// Use only in main() or initialization code where panicking is appropriate.
func MustLoad() Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("config: failed to load: %v", err))
	}
	return cfg
}

// =============================================================================
// Section Loaders
// =============================================================================

func loadServerConfig() ServerConfig {
	port := defaultHTTPPort
	if raw := getEnvWithFallback(envHTTPPort, envPortFallback); raw != "" {
		if p, err := strconv.Atoi(raw); err == nil && p > 0 && p < 65536 {
			port = p
		}
	}

	env := getEnvWithFallback(envAppEnv, envAppEnvLegacy)
	if env == "" {
		env = defaultEnv
	}

	return ServerConfig{
		Port:           port,
		Env:            normalizeEnv(env),
		ReadTimeout:    getDurationEnv(envReadTimeout, defaultReadTimeout),
		WriteTimeout:   getDurationEnv(envWriteTimeout, defaultWriteTimeout),
		IdleTimeout:    getDurationEnv(envIdleTimeout, defaultIdleTimeout),
		TrustedProxies: getStringSliceEnv(envTrustedProxies),
	}
}

func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		DSN:             strings.TrimSpace(os.Getenv(envDBDSN)),
		MaxOpenConns:    getIntEnv(envDBMaxOpenConns, 25),
		MaxIdleConns:    getIntEnv(envDBMaxIdleConns, 10),
		ConnMaxLifetime: getDurationEnv(envDBConnMaxLifetime, 45*time.Minute),
	}
}

func loadAuthConfig() AuthConfig {
	apiKey := strings.TrimSpace(os.Getenv(envAPIKey))
	jwtSecret := strings.TrimSpace(os.Getenv(envJWTSecret))

	return AuthConfig{
		APIKey:       apiKey,
		JWTSecret:    jwtSecret,
		HasAPIKey:    apiKey != "",
		HasJWTSecret: jwtSecret != "",
	}
}

func loadOpenAIConfig() OpenAIConfig {
	apiKey := strings.TrimSpace(os.Getenv(envOpenAIKey))
	model := strings.TrimSpace(os.Getenv(envOpenAIModel))
	if model == "" {
		model = defaultOpenAIModel
	}

	return OpenAIConfig{
		APIKey:       apiKey,
		Model:        model,
		BaseURL:      strings.TrimSpace(os.Getenv(envOpenAIBaseURL)),
		IsConfigured: apiKey != "",
	}
}

func loadStripeConfig() StripeConfig {
	secretKey := strings.TrimSpace(os.Getenv(envStripeSecretKey))

	return StripeConfig{
		SecretKey:       secretKey,
		WebhookSecret:   strings.TrimSpace(os.Getenv(envStripeWebhookSecret)),
		PriceFree:       strings.TrimSpace(os.Getenv(envStripePriceFree)),
		PriceBasic:      strings.TrimSpace(os.Getenv(envStripePriceBasic)),
		PricePro:        strings.TrimSpace(os.Getenv(envStripePricePro)),
		PriceEnterprise: strings.TrimSpace(os.Getenv(envStripePriceEnterprise)),
		IsConfigured:    secretKey != "",
	}
}

func loadFeatureConfig() FeatureConfig {
	return FeatureConfig{
		EnableAuditLog:  getBoolEnv(envEnableAuditLog, false),
		EnableMetrics:   getBoolEnv(envEnableMetrics, true),
		EnableGraphQL:   getBoolEnv(envEnableGraphQL, true),
		EnableOfflineAI: getBoolEnv(envEnableOfflineAI, true),
	}
}

func loadIngestionConfig() IngestionConfig {
	lookback := getIntEnv(envIngestLookbackDays, defaultIngestLookbackDays)
	schedule := getDurationEnv(envIngestScheduleInterval, defaultIngestScheduleInterval)
	return IngestionConfig{
		LookbackDays:     lookback,
		ScheduleInterval: schedule,
		AWS: AWSIngestionConfig{
			Enabled:         getBoolEnv(envAWSIngestEnabled, false),
			AccessKeyID:     strings.TrimSpace(os.Getenv(envAWSAccessKeyID)),
			SecretAccessKey: strings.TrimSpace(os.Getenv(envAWSSecretAccessKey)),
			Region:          strings.TrimSpace(os.Getenv(envAWSRegion)),
			RoleARN:         strings.TrimSpace(os.Getenv(envAWSRoleARN)),
			AccountID:       strings.TrimSpace(os.Getenv(envAWSAccountID)),
			Bucket:          strings.TrimSpace(os.Getenv(envAWSBucket)),
			Prefix:          strings.TrimSpace(os.Getenv(envAWSPrefix)),
			OrgID:           strings.TrimSpace(os.Getenv(envAWSOrgID)),
		},
		Azure: AzureIngestionConfig{
			Enabled:        getBoolEnv(envAzureIngestEnabled, false),
			TenantID:       strings.TrimSpace(os.Getenv(envAzureTenantID)),
			ClientID:       strings.TrimSpace(os.Getenv(envAzureClientID)),
			ClientSecret:   strings.TrimSpace(os.Getenv(envAzureClientSecret)),
			SubscriptionID: strings.TrimSpace(os.Getenv(envAzureSubscriptionID)),
			OrgID:          strings.TrimSpace(os.Getenv(envAzureOrgID)),
		},
		GCP: GCPIngestionConfig{
			Enabled:           getBoolEnv(envGCPIngestEnabled, false),
			ProjectID:         strings.TrimSpace(os.Getenv(envGCPProjectID)),
			BillingAccountID:  strings.TrimSpace(os.Getenv(envGCPBillingAccountID)),
			BigQueryDataset:   strings.TrimSpace(os.Getenv(envGCPBigQueryDataset)),
			BigQueryTable:     strings.TrimSpace(os.Getenv(envGCPBigQueryTable)),
			ServiceAccountKey: strings.TrimSpace(os.Getenv(envGCPServiceAccountKey)),
			OrgID:             strings.TrimSpace(os.Getenv(envGCPOrgID)),
		},
		SAP: SAPIngestionConfig{
			Enabled:      getBoolEnv(envSAPIngestEnabled, false),
			OrgID:        strings.TrimSpace(os.Getenv(envSAPOrgID)),
			BaseURL:      strings.TrimSpace(os.Getenv(envSAPBaseURL)),
			ClientID:     strings.TrimSpace(os.Getenv(envSAPClientID)),
			ClientSecret: strings.TrimSpace(os.Getenv(envSAPClientSecret)),
			Company:      strings.TrimSpace(os.Getenv(envSAPCompany)),
			Plant:        strings.TrimSpace(os.Getenv(envSAPPlant)),
		},
		Utility: UtilityIngestConfig{
			Enabled: getBoolEnv(envUtilityIngestEnabled, false),
			OrgID:   strings.TrimSpace(os.Getenv(envUtilityOrgID)),
		},
	}
}

// =============================================================================
// Validation
// =============================================================================

// Validate checks that the configuration is valid.
// In production, this enforces stricter requirements.
func (c Config) Validate() error {
	var errs []error

	// Port validation
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		errs = append(errs, fmt.Errorf("invalid port: %d", c.Server.Port))
	}

	// Production-only validations
	if c.IsProduction() {
		if c.Database.DSN == "" {
			errs = append(errs, errors.New("database DSN required in production"))
		}
		if !c.Auth.HasJWTSecret {
			errs = append(errs, errors.New("JWT secret required in production"))
		}
		if len(c.Auth.JWTSecret) < 32 {
			errs = append(errs, errors.New("JWT secret must be at least 32 characters"))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed: %w", errors.Join(errs...))
	}

	return nil
}

// =============================================================================
// Helper Methods
// =============================================================================

// IsProduction returns true if running in production environment.
func (c Config) IsProduction() bool {
	return c.Server.Env == EnvProduction
}

// IsDevelopment returns true if running in development environment.
func (c Config) IsDevelopment() bool {
	return c.Server.Env == EnvDevelopment
}

// IsTest returns true if running in test environment.
func (c Config) IsTest() bool {
	return c.Server.Env == EnvTest
}

// ServerAddress returns the full server address (e.g., ":8090").
func (c Config) ServerAddress() string {
	return fmt.Sprintf(":%d", c.Server.Port)
}

// =============================================================================
// Environment Variable Helpers
// =============================================================================

// getEnvWithFallback returns the first non-empty environment variable value.
func getEnvWithFallback(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

// getIntEnv returns an integer from an environment variable, or the default.
func getIntEnv(key string, defaultVal int) int {
	if raw := strings.TrimSpace(os.Getenv(key)); raw != "" {
		if val, err := strconv.Atoi(raw); err == nil {
			return val
		}
	}
	return defaultVal
}

// getBoolEnv returns a boolean from an environment variable, or the default.
// Accepts: true, false, 1, 0, yes, no (case-insensitive).
func getBoolEnv(key string, defaultVal bool) bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch raw {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultVal
	}
}

// getDurationEnv returns a duration from an environment variable, or the default.
// Accepts Go duration strings (e.g., "30s", "5m", "1h").
func getDurationEnv(key string, defaultVal time.Duration) time.Duration {
	if raw := strings.TrimSpace(os.Getenv(key)); raw != "" {
		if val, err := time.ParseDuration(raw); err == nil {
			return val
		}
	}
	return defaultVal
}

// getStringSliceEnv returns a string slice from a comma-separated env var.
func getStringSliceEnv(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// normalizeEnv ensures the environment string is a known value.
func normalizeEnv(env string) string {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "production", "prod":
		return EnvProduction
	case "staging", "stage", "preview":
		return EnvStaging
	case "test", "testing":
		return EnvTest
	default:
		return EnvDevelopment
	}
}

// =============================================================================
// Legacy Compatibility
// =============================================================================

// ToLegacy converts the new Config to the old flat structure for backward compatibility.
// Deprecated: Use the new Config structure directly.
func (c Config) ToLegacy() LegacyConfig {
	return LegacyConfig{
		HTTPPort:            c.Server.Port,
		Env:                 c.Server.Env,
		AppEnv:              c.Server.Env,
		DBDSN:               c.Database.DSN,
		APIKey:              c.Auth.APIKey,
		JWTSecret:           c.Auth.JWTSecret,
		OpenAIAPIKey:        c.OpenAI.APIKey,
		OpenAIModel:         c.OpenAI.Model,
		StripeSecretKey:     c.Stripe.SecretKey,
		StripeWebhookSecret: c.Stripe.WebhookSecret,
		StripePriceBasic:    c.Stripe.PriceBasic,
		StripePricePro:      c.Stripe.PricePro,
	}
}

// LegacyConfig is the old flat configuration structure.
// Deprecated: Use Config instead.
type LegacyConfig struct {
	HTTPPort            int
	Env                 string
	AppEnv              string
	DBDSN               string
	APIKey              string
	JWTSecret           string
	OpenAIAPIKey        string
	OpenAIModel         string
	StripeSecretKey     string
	StripeWebhookSecret string
	StripePriceBasic    string
	StripePricePro      string
}
