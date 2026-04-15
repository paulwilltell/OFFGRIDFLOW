BEGIN;

DROP TRIGGER IF EXISTS update_readiness_items_updated_at ON readiness_action_items;
DROP TABLE IF EXISTS abatement_evaluations CASCADE;
DROP TABLE IF EXISTS readiness_action_items CASCADE;

COMMIT;
