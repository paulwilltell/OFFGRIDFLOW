-- Connector state and status tracking

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
