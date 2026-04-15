-- 0004_audit_support: create audit support tables used by the abatement engine.
-- The embedded schema.sql also creates these; this migration mirrors that for
-- environments that apply migrations via golang-migrate.

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
