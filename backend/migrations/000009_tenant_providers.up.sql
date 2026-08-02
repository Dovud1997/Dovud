CREATE TABLE IF NOT EXISTS tenant_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    type VARCHAR(32) NOT NULL,
    driver VARCHAR(32) NOT NULL DEFAULT 'log',
    config_enc TEXT NOT NULL,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, type)
);
CREATE INDEX IF NOT EXISTS idx_tenant_providers_tenant ON tenant_providers(tenant_id);
