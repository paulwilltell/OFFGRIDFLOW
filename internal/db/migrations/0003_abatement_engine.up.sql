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
