-- Rollback Diamond-Tier Audit Infrastructure
BEGIN;

DROP TRIGGER IF EXISTS update_onboarding_updated_at ON onboarding_milestones;
DROP TRIGGER IF EXISTS update_health_scores_updated_at ON customer_health_scores;
DROP TRIGGER IF EXISTS update_alert_actions_updated_at ON alert_actions;
DROP TRIGGER IF EXISTS update_dq_anomalies_updated_at ON data_quality_anomalies;
DROP TRIGGER IF EXISTS update_factor_snapshots_updated_at ON factor_snapshots;

DROP TABLE IF EXISTS onboarding_milestones CASCADE;
DROP TABLE IF EXISTS report_exports CASCADE;
DROP TABLE IF EXISTS customer_health_scores CASCADE;
DROP TABLE IF EXISTS alert_comments CASCADE;
DROP TABLE IF EXISTS alert_actions CASCADE;
DROP TABLE IF EXISTS data_quality_anomalies CASCADE;
DROP TABLE IF EXISTS calculation_ledger CASCADE;
DROP TABLE IF EXISTS factor_snapshots CASCADE;

COMMIT;
