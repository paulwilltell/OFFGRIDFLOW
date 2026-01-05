-- OffGridFlow Initial Schema Migration
-- Version: 000001
-- Description: Creates core tables for multi-tenant carbon accounting platform
-- Author: OffGridFlow Team
-- Date: 2025-12-27

BEGIN;

-- ============================================================================
-- Organizations (Multi-tenancy root)
-- ============================================================================
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    
    -- Subscription & Billing
    subscription_tier VARCHAR(50) NOT NULL DEFAULT 'free' CHECK (subscription_tier IN ('free', 'basic', 'pro', 'enterprise')),
    subscription_status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (subscription_status IN ('active', 'inactive', 'cancelled', 'past_due')),
    stripe_customer_id VARCHAR(255),
    stripe_subscription_id VARCHAR(255),
    
    -- Settings
    industry VARCHAR(100),
    country_code VARCHAR(3),
    timezone VARCHAR(100) DEFAULT 'UTC',
    fiscal_year_start_month INT DEFAULT 1 CHECK (fiscal_year_start_month BETWEEN 1 AND 12),
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_organizations_slug ON organizations(slug);
CREATE INDEX idx_organizations_subscription_tier ON organizations(subscription_tier);
CREATE INDEX idx_organizations_deleted_at ON organizations(deleted_at) WHERE deleted_at IS NULL;

-- ============================================================================
-- Users
-- ============================================================================
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    
    -- Authentication
    email_verified BOOLEAN DEFAULT FALSE,
    email_verification_token VARCHAR(255),
    email_verification_expires_at TIMESTAMP WITH TIME ZONE,
    password_reset_token VARCHAR(255),
    password_reset_expires_at TIMESTAMP WITH TIME ZONE,
    
    -- Security
    role VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (role IN ('superadmin', 'admin', 'user', 'viewer')),
    mfa_enabled BOOLEAN DEFAULT FALSE,
    mfa_secret VARCHAR(255),
    failed_login_attempts INT DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    last_login_ip VARCHAR(45),
    
    -- Preferences
    language VARCHAR(10) DEFAULT 'en',
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    UNIQUE(organization_id, email)
);

CREATE INDEX idx_users_organization_id ON users(organization_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_email_verification_token ON users(email_verification_token);
CREATE INDEX idx_users_password_reset_token ON users(password_reset_token);
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NULL;

-- ============================================================================
-- Emission Activities (Core data model)
-- ============================================================================
CREATE TABLE IF NOT EXISTS activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Activity Details
    name VARCHAR(255) NOT NULL,
    description TEXT,
    activity_type VARCHAR(100) NOT NULL, -- e.g., 'electricity', 'natural_gas', 'vehicle_fuel', 'flight'
    scope INT NOT NULL CHECK (scope IN (1, 2, 3)),
    
    -- Calculation Inputs
    quantity DECIMAL(20, 6) NOT NULL,
    unit VARCHAR(50) NOT NULL, -- e.g., 'kWh', 'm3', 'liters', 'km'
    emission_factor_id UUID,
    
    -- Calculated Emissions (in kg CO2e)
    emissions_co2 DECIMAL(20, 6),
    emissions_ch4 DECIMAL(20, 6),
    emissions_n2o DECIMAL(20, 6),
    emissions_total_co2e DECIMAL(20, 6),
    
    -- Time & Location
    activity_date DATE NOT NULL,
    location VARCHAR(255),
    country_code VARCHAR(3),
    facility_id UUID,
    
    -- Classification
    category VARCHAR(100), -- GHG Protocol category
    source_type VARCHAR(100), -- 'direct', 'indirect', 'purchased_electricity'
    verification_status VARCHAR(50) DEFAULT 'unverified' CHECK (verification_status IN ('unverified', 'verified', 'assured')),
    
    -- Data Source
    data_source VARCHAR(100), -- 'manual', 'csv', 'api', 'aws', 'azure', 'gcp', 'sap'
    source_reference VARCHAR(255),
    
    -- Metadata
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_activities_organization_id ON activities(organization_id);
CREATE INDEX idx_activities_activity_date ON activities(activity_date);
CREATE INDEX idx_activities_scope ON activities(scope);
CREATE INDEX idx_activities_activity_type ON activities(activity_type);
CREATE INDEX idx_activities_deleted_at ON activities(deleted_at) WHERE deleted_at IS NULL;

-- ============================================================================
-- Emission Factors Database
-- ============================================================================
CREATE TABLE IF NOT EXISTS emission_factors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Factor Details
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL,
    scope INT CHECK (scope IN (1, 2, 3)),
    
    -- Factor Values (kg CO2e per unit)
    co2_factor DECIMAL(20, 10) NOT NULL,
    ch4_factor DECIMAL(20, 10) DEFAULT 0,
    n2o_factor DECIMAL(20, 10) DEFAULT 0,
    total_co2e_factor DECIMAL(20, 10) NOT NULL,
    
    unit VARCHAR(50) NOT NULL,
    
    -- Geographic & Temporal Scope
    region VARCHAR(100) NOT NULL DEFAULT 'global',
    country_code VARCHAR(3),
    year INT NOT NULL,
    valid_from DATE,
    valid_to DATE,
    
    -- Source & Reliability
    source VARCHAR(255), -- e.g., 'EPA', 'DEFRA', 'IEA', 'IPCC'
    source_url TEXT,
    uncertainty_percentage DECIMAL(5, 2),
    quality_rating VARCHAR(20), -- 'high', 'medium', 'low'
    
    -- Versioning
    version INT DEFAULT 1,
    superseded_by UUID REFERENCES emission_factors(id),
    
    -- Metadata
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_emission_factors_category ON emission_factors(category);
CREATE INDEX idx_emission_factors_region ON emission_factors(region);
CREATE INDEX idx_emission_factors_country_code ON emission_factors(country_code);
CREATE INDEX idx_emission_factors_year ON emission_factors(year);
CREATE INDEX idx_emission_factors_is_active ON emission_factors(is_active) WHERE is_active = TRUE;

-- ============================================================================
-- Cloud Connectors Configuration
-- ============================================================================
CREATE TABLE IF NOT EXISTS cloud_connectors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    provider VARCHAR(50) NOT NULL CHECK (provider IN ('aws', 'azure', 'gcp', 'sap', 'utility')),
    name VARCHAR(255) NOT NULL,
    
    -- Provider-specific encrypted credentials (JSON)
    credentials_encrypted TEXT NOT NULL,
    
    -- Configuration
    enabled BOOLEAN DEFAULT TRUE,
    auto_sync BOOLEAN DEFAULT TRUE,
    sync_interval_hours INT DEFAULT 24,
    last_sync_at TIMESTAMP WITH TIME ZONE,
    last_sync_status VARCHAR(50),
    last_sync_error TEXT,
    
    -- Metadata
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(organization_id, provider, name)
);

CREATE INDEX idx_cloud_connectors_organization_id ON cloud_connectors(organization_id);
CREATE INDEX idx_cloud_connectors_provider ON cloud_connectors(provider);
CREATE INDEX idx_cloud_connectors_enabled ON cloud_connectors(enabled) WHERE enabled = TRUE;

-- ============================================================================
-- Compliance Reports
-- ============================================================================
CREATE TABLE IF NOT EXISTS compliance_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    report_type VARCHAR(50) NOT NULL CHECK (report_type IN ('csrd', 'sec', 'cbam', 'california', 'ifrs_s2', 'gri', 'cdp')),
    report_year INT NOT NULL,
    reporting_period_start DATE NOT NULL,
    reporting_period_end DATE NOT NULL,
    
    -- Report Status
    status VARCHAR(50) DEFAULT 'draft' CHECK (status IN ('draft', 'in_review', 'approved', 'submitted', 'archived')),
    
    -- Generated Outputs
    pdf_url TEXT,
    xbrl_url TEXT,
    json_data JSONB,
    
    -- Summary Metrics
    total_emissions_co2e DECIMAL(20, 6),
    scope1_emissions DECIMAL(20, 6),
    scope2_emissions DECIMAL(20, 6),
    scope3_emissions DECIMAL(20, 6),
    
    -- Metadata
    generated_by UUID REFERENCES users(id),
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_compliance_reports_organization_id ON compliance_reports(organization_id);
CREATE INDEX idx_compliance_reports_report_type ON compliance_reports(report_type);
CREATE INDEX idx_compliance_reports_report_year ON compliance_reports(report_year);
CREATE INDEX idx_compliance_reports_status ON compliance_reports(status);

-- ============================================================================
-- Audit Logs (Compliance & Security)
-- ============================================================================
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    
    -- Event Details
    event_type VARCHAR(100) NOT NULL, -- 'user.login', 'activity.created', 'report.generated'
    event_category VARCHAR(50) NOT NULL, -- 'auth', 'data', 'compliance', 'admin'
    action VARCHAR(50) NOT NULL, -- 'create', 'read', 'update', 'delete', 'export'
    resource_type VARCHAR(100), -- 'activity', 'report', 'user'
    resource_id UUID,
    
    -- Details
    description TEXT,
    changes JSONB, -- Before/after for updates
    metadata JSONB, -- Additional context
    
    -- Network
    ip_address VARCHAR(45),
    user_agent TEXT,
    
    -- Status
    status VARCHAR(50) DEFAULT 'success', -- 'success', 'failure', 'error'
    error_message TEXT,
    
    -- Timestamp
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_organization_id ON audit_logs(organization_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_event_type ON audit_logs(event_type);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);

-- ============================================================================
-- Sessions (JWT token tracking for revocation)
-- ============================================================================
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    refresh_token_hash VARCHAR(255),
    
    -- Session Details
    ip_address VARCHAR(45),
    user_agent TEXT,
    
    -- Expiry
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked_at TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- ============================================================================
-- Function: Update updated_at timestamp
-- ============================================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Triggers for updated_at
-- ============================================================================
CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_activities_updated_at BEFORE UPDATE ON activities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_emission_factors_updated_at BEFORE UPDATE ON emission_factors
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cloud_connectors_updated_at BEFORE UPDATE ON cloud_connectors
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_compliance_reports_updated_at BEFORE UPDATE ON compliance_reports
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMIT;
