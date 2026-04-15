-- OffGridFlow Database Schema (PostgreSQL)
-- Source of truth for migrations. Keep in sync with infra/db/schema.sql.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =============================================================================
-- MULTI-TENANCY TABLES
-- =============================================================================

CREATE TABLE IF NOT EXISTS tenants (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL UNIQUE,
    slug       TEXT NOT NULL UNIQUE,
    plan       TEXT NOT NULL DEFAULT 'free',
    is_active  BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Ensure slug column exists for upgrades and backfills
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS slug TEXT;
DO $$
BEGIN
    -- Backfill slug for existing tenants without one
    IF EXISTS (SELECT 1 FROM tenants WHERE slug IS NULL) THEN
        UPDATE tenants SET slug = lower(regexp_replace(name, '[^a-z0-9]+', '-', 'g')) || '-' || substring(gen_random_uuid()::text, 1, 8) WHERE slug IS NULL;
    END IF;
END$$;
CREATE UNIQUE INDEX IF NOT EXISTS idx_tenants_slug ON tenants(slug);

CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE,
    name          TEXT,
    first_name    TEXT,
    last_name     TEXT,
    job_title     TEXT,
    password_hash TEXT NOT NULL,
    tenant_id     UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    role          TEXT NOT NULL DEFAULT 'viewer',
    roles         TEXT NOT NULL DEFAULT 'viewer',
    is_active     BOOLEAN NOT NULL DEFAULT true,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    email_verification_token TEXT,
    email_verification_sent_at TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Ensure last_login_at exists for tracking user logins
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMPTZ;

-- Ensure single-role column exists for legacy compatibility
ALTER TABLE users ADD COLUMN IF NOT EXISTS role TEXT;
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM users WHERE role IS NULL) THEN
        UPDATE users SET role = split_part(roles, ',', 1) WHERE role IS NULL;
    END IF;
END$$;

-- Ensure profile fields exist
ALTER TABLE users ADD COLUMN IF NOT EXISTS first_name TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_name TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS job_title TEXT;

-- Ensure email verification fields exist and backfill existing users as verified
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN;
DO $$
BEGIN
    UPDATE users SET email_verified = true WHERE email_verified IS NULL;
END$$;
ALTER TABLE users ALTER COLUMN email_verified SET DEFAULT false;
ALTER TABLE users ALTER COLUMN email_verified SET NOT NULL;

ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verification_token TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verification_sent_at TIMESTAMPTZ;

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
    tenant_id              UUID NOT NULL UNIQUE REFERENCES tenants(id) ON DELETE CASCADE,
    stripe_customer_id     TEXT,
    stripe_subscription_id TEXT,
    status                 TEXT NOT NULL DEFAULT 'trialing',
    plan                   TEXT NOT NULL DEFAULT 'basic',
    current_period_end     TIMESTAMPTZ,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- Ensure unique constraint exists on existing DBs (idempotent)
DO $ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'subscriptions'::regclass
        AND contype = 'u'
        AND conname = 'subscriptions_tenant_id_key'
    ) THEN
        ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_tenant_id_key UNIQUE (tenant_id);
    END IF;
END $;

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

CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_email_verification_token ON users(email_verification_token);
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

-- =============================================================================
-- RISK ABATEMENT ENGINE
-- =============================================================================

CREATE TABLE IF NOT EXISTS readiness_action_items (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    framework           TEXT NOT NULL,
    compliance_check_id TEXT NOT NULL,
    title               TEXT NOT NULL,
    severity            TEXT NOT NULL CHECK (severity IN ('blocker', 'warning')),
    priority            TEXT NOT NULL CHECK (priority IN ('high', 'medium', 'low')),
    description         TEXT NOT NULL,
    acceptance_criteria TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    required_evidence_types TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    justification       TEXT,
    evidence_urls       TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    engine_status       TEXT CHECK (engine_status IN ('recommended', 'needs_clarification', 'insufficient')),
    engine_feedback     TEXT,
    criteria_checked    TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    completed           BOOLEAN NOT NULL DEFAULT FALSE,
    self_certified      BOOLEAN NOT NULL DEFAULT FALSE,
    certified_at        TIMESTAMPTZ,
    certified_by        UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by          UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, framework, compliance_check_id)
);

CREATE INDEX IF NOT EXISTS idx_readiness_action_items_tenant_framework
    ON readiness_action_items(tenant_id, framework, priority, severity);

CREATE TABLE IF NOT EXISTS abatement_evidence (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    action_item_id  UUID NOT NULL REFERENCES readiness_action_items(id) ON DELETE CASCADE,
    framework       TEXT NOT NULL,
    storage_path    TEXT NOT NULL UNIQUE,
    file_name       TEXT NOT NULL,
    mime_type       TEXT NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    file_bytes      BYTEA NOT NULL,
    created_by      UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_abatement_evidence_tenant_action
    ON abatement_evidence(tenant_id, action_item_id, created_at DESC);

ALTER TABLE readiness_action_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE abatement_evidence ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS readiness_action_items_tenant_select ON readiness_action_items;
CREATE POLICY readiness_action_items_tenant_select
    ON readiness_action_items
    FOR SELECT
    USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid);

DROP POLICY IF EXISTS readiness_action_items_tenant_modify ON readiness_action_items;
CREATE POLICY readiness_action_items_tenant_modify
    ON readiness_action_items
    FOR ALL
    USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid)
    WITH CHECK (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid);

DROP POLICY IF EXISTS abatement_evidence_tenant_select ON abatement_evidence;
CREATE POLICY abatement_evidence_tenant_select
    ON abatement_evidence
    FOR SELECT
    USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid);

DROP POLICY IF EXISTS abatement_evidence_tenant_modify ON abatement_evidence;
CREATE POLICY abatement_evidence_tenant_modify
    ON abatement_evidence
    FOR ALL
    USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid)
    WITH CHECK (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid);

-- =============================================================================
-- AUDIT INFRASTRUCTURE (Diamond-Tier Panel 1B) — minimum surface used by
-- the abatement engine to read compliance state. Tables are kept here so
-- the embedded schema migration can create them on first boot.
-- =============================================================================

CREATE TABLE IF NOT EXISTS factor_snapshots (
    id                        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id           UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    reporting_period_start    DATE NOT NULL,
    reporting_period_end      DATE NOT NULL,
    status                    TEXT NOT NULL DEFAULT 'draft'
                              CHECK (status IN ('draft', 'locked', 'archived')),
    factor_set_name           TEXT,
    factor_set_version        TEXT,
    factor_source             TEXT,
    locked_at                 TIMESTAMPTZ,
    locked_by                 UUID REFERENCES users(id) ON DELETE SET NULL,
    notes                     TEXT,
    created_at                TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_factor_snapshots_org
    ON factor_snapshots(organization_id);
CREATE INDEX IF NOT EXISTS idx_factor_snapshots_period
    ON factor_snapshots(reporting_period_start, reporting_period_end);
CREATE INDEX IF NOT EXISTS idx_factor_snapshots_status
    ON factor_snapshots(status);

CREATE TABLE IF NOT EXISTS approval_workflow (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_type   TEXT NOT NULL,
    entity_id     TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'draft'
                  CHECK (status IN ('draft', 'submitted', 'reviewed', 'approved', 'rejected')),
    prepared_by   UUID REFERENCES users(id) ON DELETE SET NULL,
    prepared_at   TIMESTAMPTZ,
    reviewed_by   UUID REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at   TIMESTAMPTZ,
    approved_by   UUID REFERENCES users(id) ON DELETE SET NULL,
    approved_at   TIMESTAMPTZ,
    rejected_by   UUID REFERENCES users(id) ON DELETE SET NULL,
    rejected_at   TIMESTAMPTZ,
    reject_reason TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_approval_workflow_tenant_entity
    ON approval_workflow(tenant_id, entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_approval_workflow_tenant_updated
    ON approval_workflow(tenant_id, updated_at DESC);
