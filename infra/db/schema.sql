-- OffGridFlow Database Schema (PostgreSQL)
-- Source of truth for migrations. Keep in sync with internal/db/schema.sql.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =============================================================================
-- MULTI-TENANCY TABLES
-- =============================================================================

CREATE TABLE IF NOT EXISTS tenants (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL UNIQUE,
    plan       TEXT NOT NULL DEFAULT 'free',
    is_active  BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE,
    name          TEXT,
    password_hash TEXT NOT NULL,
    tenant_id     UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    roles         TEXT NOT NULL DEFAULT 'viewer',
    is_active     BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS api_keys (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_hash     TEXT NOT NULL UNIQUE,
    key_prefix   TEXT NOT NULL,
    label        TEXT NOT NULL,
    tenant_id    UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id      UUID REFERENCES users(id) ON DELETE SET NULL,
    scopes       TEXT NOT NULL DEFAULT '*',
    expires_at   TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    is_active    BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS subscriptions (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id              UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    stripe_customer_id     TEXT,
    stripe_subscription_id TEXT,
    status                 TEXT NOT NULL DEFAULT 'trialing',
    plan                   TEXT NOT NULL DEFAULT 'basic',
    current_period_end     TIMESTAMPTZ,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS activities (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source       TEXT NOT NULL,
    category     TEXT,
    meter_id     TEXT,
    location     TEXT,
    period_start TIMESTAMPTZ,
    period_end   TIMESTAMPTZ,
    quantity     DOUBLE PRECISION,
    unit         TEXT,
    org_id       UUID REFERENCES tenants(id) ON DELETE SET NULL,
    metadata     JSONB,
    workspace_id TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS emission_factors (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scope      TEXT NOT NULL,
    category   TEXT,
    region     TEXT NOT NULL,
    unit       TEXT NOT NULL,
    value      DOUBLE PRECISION NOT NULL,
    source     TEXT,
    valid_from DATE,
    valid_to   DATE
);

CREATE TABLE IF NOT EXISTS emissions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id    UUID REFERENCES activities(id) ON DELETE SET NULL,
    factor_id      UUID,
    scope          TEXT NOT NULL,
    emissions_kg   DOUBLE PRECISION NOT NULL,
    emissions_tonnes DOUBLE PRECISION NOT NULL,
    method         TEXT NOT NULL,
    region         TEXT,
    org_id         UUID REFERENCES tenants(id) ON DELETE SET NULL,
    workspace_id   TEXT,
    period_start   TIMESTAMPTZ,
    period_end     TIMESTAMPTZ,
    calculated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID REFERENCES tenants(id) ON DELETE CASCADE,
    user_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    action      TEXT NOT NULL,
    entity_type TEXT,
    entity_id   TEXT,
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS ingestion_logs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source        TEXT NOT NULL,
    status        TEXT NOT NULL,
    processed     INT NOT NULL DEFAULT 0,
    succeeded     INT NOT NULL DEFAULT 0,
    failed        INT NOT NULL DEFAULT 0,
    errors        JSONB,
    started_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at  TIMESTAMPTZ,
    org_id        UUID REFERENCES tenants(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS compliance_cache (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID REFERENCES tenants(id) ON DELETE CASCADE,
    framework    TEXT NOT NULL,
    key          TEXT NOT NULL,
    payload      JSONB NOT NULL,
    computed_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at   TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS workflow_tasks (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID REFERENCES tenants(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'pending',
    assignee_id   UUID REFERENCES users(id) ON DELETE SET NULL,
    due_date      TIMESTAMPTZ,
    metadata      JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS billing_state (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      UUID REFERENCES tenants(id) ON DELETE CASCADE,
    usage_month    DATE NOT NULL,
    usage_quantity DOUBLE PRECISION NOT NULL DEFAULT 0,
    usage_unit     TEXT NOT NULL DEFAULT 'unit',
    invoice_total  DOUBLE PRECISION NOT NULL DEFAULT 0,
    currency       TEXT NOT NULL DEFAULT 'USD',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, usage_month)
);

-- =============================================================================
-- JOB QUEUE TABLES
-- =============================================================================

CREATE TABLE IF NOT EXISTS jobs (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    payload JSONB,
    result JSONB,
    error TEXT,
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    scheduled_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_jobs_type_status ON jobs(type, status, scheduled_at);
CREATE INDEX IF NOT EXISTS idx_jobs_tenant ON jobs(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at DESC);

-- =============================================================================
-- INDEXES
-- =============================================================================

CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_api_keys_tenant_id ON api_keys(tenant_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_subscriptions_tenant_id ON subscriptions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_activities_source ON activities(source);
CREATE INDEX IF NOT EXISTS idx_activities_org_id ON activities(org_id);
CREATE INDEX IF NOT EXISTS idx_activities_created_at ON activities(created_at);
CREATE INDEX IF NOT EXISTS idx_emission_factors_region ON emission_factors(region);
CREATE INDEX IF NOT EXISTS idx_emissions_activity_id ON emissions(activity_id);
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_source ON ingestion_logs(source);
CREATE INDEX IF NOT EXISTS idx_workflow_tasks_status ON workflow_tasks(status);

CREATE TABLE IF NOT EXISTS connectors (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    config      JSONB,
    status      TEXT NOT NULL DEFAULT 'disconnected',
    last_run_at TIMESTAMPTZ,
    last_error  TEXT,
    org_id      UUID REFERENCES tenants(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (name, org_id)
);
CREATE INDEX IF NOT EXISTS idx_connectors_org ON connectors(org_id);
