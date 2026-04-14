-- OffGridFlow Diamond-Tier Audit Infrastructure Migration
-- Version: 000002
-- Description: Adds factor version locking, data quality anomaly detection,
--              alert action system, customer health scoring, and stakeholder export support.
-- Author: OffGridFlow Team
-- Date: 2026-04-13

BEGIN;

-- ============================================================================
-- Factor Snapshots (Panel 1B: Reproducibility by Version)
-- Locks emission factors to specific reporting periods so calculations
-- can be exactly reproduced at any future date.
-- ============================================================================
CREATE TABLE IF NOT EXISTS factor_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Reporting period this snapshot covers
    reporting_period_start DATE NOT NULL,
    reporting_period_end DATE NOT NULL,
    snapshot_name VARCHAR(255) NOT NULL,

    -- The locked factor set (immutable once locked)
    factors JSONB NOT NULL DEFAULT '[]'::jsonb,
    factor_count INT NOT NULL DEFAULT 0,

    -- Lock state
    status VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'locked', 'superseded')),
    locked_at TIMESTAMP WITH TIME ZONE,
    locked_by UUID REFERENCES users(id),

    -- Provenance
    source_registry_version VARCHAR(100),
    notes TEXT,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(organization_id, reporting_period_start, reporting_period_end, snapshot_name)
);

CREATE INDEX idx_factor_snapshots_org_id ON factor_snapshots(organization_id);
CREATE INDEX idx_factor_snapshots_period ON factor_snapshots(reporting_period_start, reporting_period_end);
CREATE INDEX idx_factor_snapshots_status ON factor_snapshots(status);

-- ============================================================================
-- Calculation Ledger (Panel 1B: Full Traceability)
-- Immutable record of every emissions calculation with full factor lineage.
-- ============================================================================
CREATE TABLE IF NOT EXISTS calculation_ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- What was calculated
    activity_id UUID NOT NULL REFERENCES activities(id),
    scope INT NOT NULL CHECK (scope IN (1, 2, 3)),
    category VARCHAR(100),

    -- Inputs
    input_quantity DECIMAL(20, 6) NOT NULL,
    input_unit VARCHAR(50) NOT NULL,

    -- Factor used (frozen copy)
    factor_id VARCHAR(255) NOT NULL,
    factor_value DECIMAL(20, 10) NOT NULL,
    factor_source VARCHAR(255),
    factor_region VARCHAR(100),
    factor_year INT,
    factor_snapshot_id UUID REFERENCES factor_snapshots(id),

    -- Outputs
    emissions_kg_co2e DECIMAL(20, 6) NOT NULL,
    emissions_tonnes_co2e DECIMAL(20, 6) NOT NULL,

    -- Calculation metadata
    calculation_method VARCHAR(50) NOT NULL,
    formula TEXT NOT NULL,
    data_quality VARCHAR(50) NOT NULL DEFAULT 'measured',

    -- Period
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,

    -- Immutability
    is_locked BOOLEAN NOT NULL DEFAULT FALSE,
    calculated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    calculated_by UUID REFERENCES users(id),

    -- Version tracking (for recalculations)
    version INT NOT NULL DEFAULT 1,
    supersedes_id UUID REFERENCES calculation_ledger(id),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_calc_ledger_org_id ON calculation_ledger(organization_id);
CREATE INDEX idx_calc_ledger_activity_id ON calculation_ledger(activity_id);
CREATE INDEX idx_calc_ledger_scope ON calculation_ledger(scope);
CREATE INDEX idx_calc_ledger_period ON calculation_ledger(period_start, period_end);
CREATE INDEX idx_calc_ledger_snapshot ON calculation_ledger(factor_snapshot_id);
CREATE INDEX idx_calc_ledger_locked ON calculation_ledger(is_locked) WHERE is_locked = TRUE;

-- ============================================================================
-- Data Quality Anomalies (Panel 1B: Anomaly Detection)
-- Flags unusual data patterns for review before they reach reports.
-- ============================================================================
CREATE TABLE IF NOT EXISTS data_quality_anomalies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- What triggered the anomaly
    entity_type VARCHAR(50) NOT NULL CHECK (entity_type IN ('activity', 'factor', 'report', 'connector')),
    entity_id UUID NOT NULL,

    -- Anomaly classification
    anomaly_type VARCHAR(100) NOT NULL CHECK (anomaly_type IN (
        'quantity_outlier',          -- Value is >3 std deviations from mean
        'duplicate_entry',           -- Possible duplicate activity
        'missing_period',            -- Gap in expected time series
        'unit_mismatch',            -- Unit doesn't match expected for category
        'factor_deviation',          -- Factor differs significantly from expected
        'sudden_change',            -- >50% change from prior period
        'negative_value',           -- Unexpected negative quantity
        'stale_data',               -- Data hasn't been updated in expected window
        'scope_misclassification',  -- Activity may be in wrong scope
        'completeness_gap'          -- Required data fields are missing
    )),
    severity VARCHAR(20) NOT NULL DEFAULT 'warning' CHECK (severity IN ('info', 'warning', 'critical')),

    -- Details
    description TEXT NOT NULL,
    expected_value DECIMAL(20, 6),
    actual_value DECIMAL(20, 6),
    deviation_percent DECIMAL(10, 2),
    detection_rule TEXT,

    -- Resolution
    status VARCHAR(50) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'acknowledged', 'resolved', 'dismissed')),
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by UUID REFERENCES users(id),
    resolution_notes TEXT,

    -- Assignment
    assigned_to UUID REFERENCES users(id),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_dq_anomalies_org_id ON data_quality_anomalies(organization_id);
CREATE INDEX idx_dq_anomalies_entity ON data_quality_anomalies(entity_type, entity_id);
CREATE INDEX idx_dq_anomalies_type ON data_quality_anomalies(anomaly_type);
CREATE INDEX idx_dq_anomalies_status ON data_quality_anomalies(status);
CREATE INDEX idx_dq_anomalies_severity ON data_quality_anomalies(severity);

-- ============================================================================
-- Alert Actions (Panel 2C: Built-in Next Actions)
-- Every critical alert has assign, comment, resolve, escalate capability.
-- ============================================================================
CREATE TABLE IF NOT EXISTS alert_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Source: what triggered this alert
    source_type VARCHAR(50) NOT NULL CHECK (source_type IN (
        'anomaly', 'threshold', 'deadline', 'compliance', 'connector', 'approval'
    )),
    source_id UUID NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,

    -- Classification
    priority VARCHAR(20) NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    category VARCHAR(100) NOT NULL,

    -- Ownership
    assigned_to UUID REFERENCES users(id),
    escalated_to UUID REFERENCES users(id),

    -- State
    status VARCHAR(50) NOT NULL DEFAULT 'open' CHECK (status IN (
        'open', 'in_progress', 'blocked', 'resolved', 'dismissed', 'escalated'
    )),

    -- Timestamps
    due_date TIMESTAMP WITH TIME ZONE,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    escalated_at TIMESTAMP WITH TIME ZONE,

    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_alert_actions_org_id ON alert_actions(organization_id);
CREATE INDEX idx_alert_actions_source ON alert_actions(source_type, source_id);
CREATE INDEX idx_alert_actions_assigned ON alert_actions(assigned_to);
CREATE INDEX idx_alert_actions_status ON alert_actions(status);
CREATE INDEX idx_alert_actions_priority ON alert_actions(priority);

-- ============================================================================
-- Alert Comments (Panel 2C: Comment on any alert)
-- ============================================================================
CREATE TABLE IF NOT EXISTS alert_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_id UUID NOT NULL REFERENCES alert_actions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_alert_comments_alert_id ON alert_comments(alert_id);

-- ============================================================================
-- Customer Health Scores (Panel 3E: Renewal Engineering)
-- Tracks engagement, adoption, and value realization per organization.
-- ============================================================================
CREATE TABLE IF NOT EXISTS customer_health_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Score components (0-100 each)
    overall_score INT NOT NULL DEFAULT 0 CHECK (overall_score BETWEEN 0 AND 100),
    data_freshness_score INT NOT NULL DEFAULT 0 CHECK (data_freshness_score BETWEEN 0 AND 100),
    feature_adoption_score INT NOT NULL DEFAULT 0 CHECK (feature_adoption_score BETWEEN 0 AND 100),
    report_completion_score INT NOT NULL DEFAULT 0 CHECK (report_completion_score BETWEEN 0 AND 100),
    user_engagement_score INT NOT NULL DEFAULT 0 CHECK (user_engagement_score BETWEEN 0 AND 100),
    data_quality_score INT NOT NULL DEFAULT 0 CHECK (data_quality_score BETWEEN 0 AND 100),

    -- Health status derived from score
    health_status VARCHAR(20) NOT NULL DEFAULT 'unknown' CHECK (health_status IN (
        'healthy', 'at_risk', 'critical', 'churning', 'unknown'
    )),

    -- Renewal risk
    renewal_risk_percent INT DEFAULT 0 CHECK (renewal_risk_percent BETWEEN 0 AND 100),
    next_renewal_date DATE,
    contract_value_cents BIGINT,

    -- Usage metrics (snapshot)
    active_users_count INT DEFAULT 0,
    total_activities_count INT DEFAULT 0,
    reports_generated_count INT DEFAULT 0,
    last_login_at TIMESTAMP WITH TIME ZONE,
    last_report_at TIMESTAMP WITH TIME ZONE,
    last_data_upload_at TIMESTAMP WITH TIME ZONE,

    -- Expansion signals
    expansion_ready BOOLEAN DEFAULT FALSE,
    expansion_triggers JSONB DEFAULT '[]'::jsonb,

    scored_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(organization_id, scored_at)
);

CREATE INDEX idx_health_scores_org_id ON customer_health_scores(organization_id);
CREATE INDEX idx_health_scores_status ON customer_health_scores(health_status);
CREATE INDEX idx_health_scores_renewal ON customer_health_scores(next_renewal_date);

-- ============================================================================
-- Report Exports (Panel 2E: Export Reconciliation)
-- Tracks every export to ensure on-screen truth matches exported output.
-- ============================================================================
CREATE TABLE IF NOT EXISTS report_exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Source report
    report_id UUID NOT NULL REFERENCES compliance_reports(id),
    report_type VARCHAR(50) NOT NULL,

    -- Export details
    export_format VARCHAR(20) NOT NULL CHECK (export_format IN ('pdf', 'xbrl', 'xlsx', 'csv', 'json')),
    export_purpose VARCHAR(50) NOT NULL DEFAULT 'download' CHECK (export_purpose IN (
        'download', 'board_package', 'auditor_package', 'regulator_submission', 'stakeholder_review'
    )),

    -- Reconciliation (checksum of data at export time)
    data_checksum VARCHAR(64) NOT NULL,
    scope1_at_export DECIMAL(20, 6),
    scope2_at_export DECIMAL(20, 6),
    scope3_at_export DECIMAL(20, 6),
    total_at_export DECIMAL(20, 6),

    -- File info
    file_url TEXT,
    file_size_bytes BIGINT,

    exported_by UUID REFERENCES users(id),
    exported_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_report_exports_org_id ON report_exports(organization_id);
CREATE INDEX idx_report_exports_report_id ON report_exports(report_id);

-- ============================================================================
-- Onboarding Milestones (Panel 3D: Segment-Adaptive Onboarding)
-- Tracks per-org onboarding progress with segment-specific paths.
-- ============================================================================
CREATE TABLE IF NOT EXISTS onboarding_milestones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Segment
    onboarding_segment VARCHAR(50) NOT NULL DEFAULT 'self_serve' CHECK (onboarding_segment IN (
        'self_serve', 'assisted', 'enterprise'
    )),

    -- Milestones (each tracked individually)
    account_created_at TIMESTAMP WITH TIME ZONE,
    profile_completed_at TIMESTAMP WITH TIME ZONE,
    first_data_upload_at TIMESTAMP WITH TIME ZONE,
    first_connector_at TIMESTAMP WITH TIME ZONE,
    first_calculation_at TIMESTAMP WITH TIME ZONE,
    first_report_at TIMESTAMP WITH TIME ZONE,
    first_approval_at TIMESTAMP WITH TIME ZONE,
    first_export_at TIMESTAMP WITH TIME ZONE,
    training_completed_at TIMESTAMP WITH TIME ZONE,

    -- Time to value
    time_to_first_value_hours DECIMAL(10, 2),

    -- Completion
    completed BOOLEAN DEFAULT FALSE,
    completed_at TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(organization_id)
);

CREATE INDEX idx_onboarding_org_id ON onboarding_milestones(organization_id);
CREATE INDEX idx_onboarding_segment ON onboarding_milestones(onboarding_segment);
CREATE INDEX idx_onboarding_completed ON onboarding_milestones(completed);

-- ============================================================================
-- Triggers
-- ============================================================================
CREATE TRIGGER update_factor_snapshots_updated_at BEFORE UPDATE ON factor_snapshots
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_dq_anomalies_updated_at BEFORE UPDATE ON data_quality_anomalies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_alert_actions_updated_at BEFORE UPDATE ON alert_actions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_health_scores_updated_at BEFORE UPDATE ON customer_health_scores
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_onboarding_updated_at BEFORE UPDATE ON onboarding_milestones
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMIT;
