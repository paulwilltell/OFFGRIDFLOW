-- OffGridFlow Initial Schema Rollback
-- Version: 000001
-- Description: Drops all tables created in 000001_initial_schema.up.sql
-- Author: OffGridFlow Team
-- Date: 2025-12-27

BEGIN;

-- Drop triggers
DROP TRIGGER IF EXISTS update_compliance_reports_updated_at ON compliance_reports;
DROP TRIGGER IF EXISTS update_cloud_connectors_updated_at ON cloud_connectors;
DROP TRIGGER IF EXISTS update_emission_factors_updated_at ON emission_factors;
DROP TRIGGER IF EXISTS update_activities_updated_at ON activities;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_organizations_updated_at ON organizations;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS compliance_reports CASCADE;
DROP TABLE IF EXISTS cloud_connectors CASCADE;
DROP TABLE IF EXISTS emission_factors CASCADE;
DROP TABLE IF EXISTS activities CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS organizations CASCADE;

COMMIT;
