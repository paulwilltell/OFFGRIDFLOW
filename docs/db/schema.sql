-- OffGridFlow Database Schema (PostgreSQL)
-- Version: 1.0.0
-- 
-- This schema is automatically applied via RunMigrations() in internal/db/db.go.
-- In production, consider using a migration tool like golang-migrate for versioning.

-- =============================================================================
-- MULTI-TENANCY TABLES
-- =============================================================================

-- Tenants: Organizations using OffGridFlow
CREATE TABLE IF NOT EXISTS tenants (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    plan       TEXT NOT NULL DEFAULT 'free',  -- free, pro, enterprise
    is_active  BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Users: Individuals within tenants
CREATE TABLE IF NOT EXISTS users (
    id         TEXT PRIMARY KEY,
    email      TEXT NOT NULL UNIQUE,
    name       TEXT,
    tenant_id  TEXT NOT NULL REFERENCES tenants(id),
    roles      TEXT NOT NULL DEFAULT 'viewer',  -- viewer, editor, admin (comma-separated)
    is_active  BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- API Keys: Authentication tokens for API access
CREATE TABLE IF NOT EXISTS api_keys (
    id           TEXT PRIMARY KEY,
    key_hash     TEXT NOT NULL UNIQUE,  -- SHA-256 hash of the actual key
    key_prefix   TEXT NOT NULL,          -- First 12 chars for identification (e.g., "ogf_live_abc")
    name         TEXT NOT NULL,
    tenant_id    TEXT NOT NULL REFERENCES tenants(id),
    user_id      TEXT REFERENCES users(id),  -- Optional: tie key to specific user
    scopes       TEXT NOT NULL DEFAULT '*',  -- Comma-separated: read, write, admin, * (all)
    expires_at   TIMESTAMPTZ,            -- NULL = never expires
    last_used_at TIMESTAMPTZ,
    is_active    BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- =============================================================================
-- CARBON ACCOUNTING TABLES
-- =============================================================================

-- Activities: Raw input data from various sources (electricity, fuel, travel, etc.)
CREATE TABLE IF NOT EXISTS activities (
    id           TEXT PRIMARY KEY,
    source       TEXT NOT NULL,           -- Data source: manual, utility_api, erp_sync, csv_import
    category     TEXT,                    -- Activity type: electricity, natural_gas, fleet_fuel, etc.
    meter_id     TEXT,                    -- Optional: specific meter/asset identifier
    location     TEXT,                    -- Region for emission factor lookup (e.g., US-WEST, EU-CENTRAL)
    period_start TIMESTAMPTZ,
    period_end   TIMESTAMPTZ,
    quantity     DOUBLE PRECISION,        -- Amount consumed
    unit         TEXT,                    -- kWh, therms, gallons, miles, etc.
    org_id       TEXT REFERENCES tenants(id),  -- Tenant that owns this data
    metadata     JSONB,                   -- Flexible storage for source-specific data
    workspace_id TEXT,                    -- Optional: sub-organization grouping
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Emission Factors: Conversion rates from activities to CO2e
CREATE TABLE IF NOT EXISTS emission_factors (
    id         TEXT PRIMARY KEY,
    scope      TEXT NOT NULL,           -- scope1, scope2_location, scope2_market, scope3
    category   TEXT,                    -- electricity, natural_gas, diesel, etc.
    region     TEXT NOT NULL,           -- Geographic applicability: US-WEST, EU, GLOBAL
    unit       TEXT NOT NULL,           -- Input unit: kWh, therms, gallons
    value      DOUBLE PRECISION NOT NULL, -- kg CO2e per unit
    source     TEXT,                    -- EPA eGRID, IEA, DEFRA, etc.
    valid_from DATE,                    -- NULL = always valid
    valid_to   DATE                     -- NULL = still current
);

-- =============================================================================
-- PERFORMANCE INDEXES
-- =============================================================================

-- User lookups
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- API key validation (hot path)
CREATE INDEX IF NOT EXISTS idx_api_keys_tenant_id ON api_keys(tenant_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);

-- Activity queries (common filters)
CREATE INDEX IF NOT EXISTS idx_activities_source ON activities(source);
CREATE INDEX IF NOT EXISTS idx_activities_org_id ON activities(org_id);
CREATE INDEX IF NOT EXISTS idx_activities_created_at ON activities(created_at);
CREATE INDEX IF NOT EXISTS idx_activities_category ON activities(category);
CREATE INDEX IF NOT EXISTS idx_activities_period ON activities(period_start, period_end);

-- Emission factor lookups
CREATE INDEX IF NOT EXISTS idx_emission_factors_region ON emission_factors(region);
CREATE INDEX IF NOT EXISTS idx_emission_factors_scope ON emission_factors(scope);
CREATE INDEX IF NOT EXISTS idx_emission_factors_category ON emission_factors(scope, category);

-- =============================================================================
-- EXAMPLE SEED DATA (for development)
-- =============================================================================

-- Demo tenant
-- INSERT INTO tenants (id, name, plan, is_active)
-- VALUES ('tenant-demo', 'Demo Organization', 'pro', true)
-- ON CONFLICT (id) DO NOTHING;

-- Demo user
-- INSERT INTO users (id, email, name, tenant_id, roles, is_active)
-- VALUES ('user-demo', 'demo@offgridflow.com', 'Demo User', 'tenant-demo', 'admin', true)
-- ON CONFLICT (id) DO NOTHING;

-- Sample emission factors (EPA eGRID 2023)
-- INSERT INTO emission_factors (id, scope, category, region, unit, value, source)
-- VALUES 
--   ('ef-electricity-us-west', 'scope2_location', 'electricity', 'US-WEST', 'kWh', 0.0004, 'EPA eGRID 2023'),
--   ('ef-electricity-us-east', 'scope2_location', 'electricity', 'US-EAST', 'kWh', 0.0005, 'EPA eGRID 2023'),
--   ('ef-electricity-eu', 'scope2_location', 'electricity', 'EU-CENTRAL', 'kWh', 0.00035, 'IEA 2023'),
--   ('ef-natural-gas', 'scope1', 'natural_gas', 'GLOBAL', 'therms', 0.0053, 'EPA GHG Factors 2023')
-- ON CONFLICT (id) DO NOTHING;
