CREATE TABLE IF NOT EXISTS sync_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    user_id UUID NOT NULL REFERENCES users(id),
    device_id VARCHAR(128) NOT NULL,
    platform VARCHAR(64) NULL,
    app_version VARCHAR(64) NULL,
    last_pull_cursor VARCHAR(256) NOT NULL DEFAULT '',
    last_push_at TIMESTAMPTZ NULL,
    last_pull_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_sync_devices_tenant_user_device ON sync_devices(tenant_id, user_id, device_id);
CREATE INDEX IF NOT EXISTS idx_sync_devices_tenant ON sync_devices(tenant_id);

CREATE TABLE IF NOT EXISTS sync_change_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    entity_type VARCHAR(64) NOT NULL,
    entity_id VARCHAR(64) NOT NULL,
    version BIGINT NOT NULL,
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    payload_json TEXT NOT NULL DEFAULT '{}'
);
CREATE INDEX IF NOT EXISTS idx_sync_change_log_tenant_cursor ON sync_change_log(tenant_id, updated_at, id);
CREATE INDEX IF NOT EXISTS idx_sync_change_log_entity ON sync_change_log(tenant_id, entity_type, entity_id, version DESC);

CREATE TABLE IF NOT EXISTS sync_conflicts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    device_id VARCHAR(128) NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id),
    entity_type VARCHAR(64) NOT NULL,
    entity_id VARCHAR(64) NOT NULL,
    client_op_id VARCHAR(128) NOT NULL,
    base_version BIGINT NOT NULL,
    server_version BIGINT NOT NULL,
    client_payload TEXT NOT NULL DEFAULT '',
    server_payload TEXT NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'open',
    resolution VARCHAR(64) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ NULL
);
CREATE INDEX IF NOT EXISTS idx_sync_conflicts_tenant_status ON sync_conflicts(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_sync_conflicts_device ON sync_conflicts(tenant_id, device_id, status);

CREATE TABLE IF NOT EXISTS sync_applied_ops (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    op_id VARCHAR(128) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_sync_applied_ops_tenant_op ON sync_applied_ops(tenant_id, op_id);
