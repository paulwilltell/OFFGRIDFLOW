-- OffGridFlow Risk Abatement & Justification Engine
-- Version: 000003
-- Description: Tables and columns for regulatory risk tracking, customer
--              justifications, engine evaluations, and self-certification.
-- Date: 2026-04-14

BEGIN;

-- ============================================================================
-- Readiness action items (the risks surfaced for each tenant + framework)
-- ============================================================================
CREATE TABLE IF NOT EXISTS readiness_action_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Which regulatory framework this item belongs to
    framework VARCHAR(50) NOT NULL CHECK (framework IN (
        'sb253', 'csrd', 'sec_climate', 'cbam', 'ifrs_s2'
    )),

    -- Stable identifier for the type of risk (maps to evaluator function)
    compliance_check_id VARCHAR(100) NOT NULL,

    -- Surface metadata
    title VARCHAR(500) NOT NULL,
    description TEXT NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'warning' CHECK (severity IN ('blocker', 'warning', 'info')),
    category VARCHAR(100), -- e.g., scope2, scope3, governance, assurance

    -- Customer-provided abatement
    justification TEXT,
    evidence_urls TEXT[] DEFAULT ARRAY[]::TEXT[],

    -- Engine verdict (see lib/abatement/evaluators.ts)
    engine_status VARCHAR(50) CHECK (engine_status IN ('recommended', 'needs_clarification', 'insufficient', NULL)),
    engine_feedback TEXT,
    engine_criteria_checked TEXT[] DEFAULT ARRAY[]::TEXT[],
    engine_evaluated_at TIMESTAMP WITH TIME ZONE,

    -- Customer self-certification override
    self_certified BOOLEAN DEFAULT FALSE,
    certified_at TIMESTAMP WITH TIME ZONE,
    certified_by UUID REFERENCES users(id),

    -- Reporting period this item applies to
    reporting_period_start DATE,
    reporting_period_end DATE,

    -- Lifecycle
    status VARCHAR(50) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'addressed', 'dismissed')),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    UNIQUE (tenant_id, framework, compliance_check_id, reporting_period_start, reporting_period_end)
);

CREATE INDEX IF NOT EXISTS idx_readiness_items_tenant_framework
    ON readiness_action_items(tenant_id, framework);
CREATE INDEX IF NOT EXISTS idx_readiness_items_status
    ON readiness_action_items(status);
CREATE INDEX IF NOT EXISTS idx_readiness_items_severity
    ON readiness_action_items(severity);

-- ============================================================================
-- Abatement evaluation history (audit trail for every engine evaluation)
-- ============================================================================
CREATE TABLE IF NOT EXISTS abatement_evaluations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    readiness_action_item_id UUID NOT NULL REFERENCES readiness_action_items(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Input snapshot
    justification_at_eval TEXT NOT NULL,
    evidence_url_count INT NOT NULL DEFAULT 0,
    evaluator_version VARCHAR(50) NOT NULL,

    -- Output
    status VARCHAR(50) NOT NULL CHECK (status IN ('recommended', 'needs_clarification', 'insufficient')),
    feedback TEXT NOT NULL,
    criteria_checked TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],

    -- Actor
    evaluated_by UUID REFERENCES users(id),
    evaluated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_abatement_evals_item
    ON abatement_evaluations(readiness_action_item_id);
CREATE INDEX IF NOT EXISTS idx_abatement_evals_tenant
    ON abatement_evaluations(tenant_id);

-- ============================================================================
-- Triggers
-- ============================================================================
CREATE TRIGGER update_readiness_items_updated_at BEFORE UPDATE ON readiness_action_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMIT;
