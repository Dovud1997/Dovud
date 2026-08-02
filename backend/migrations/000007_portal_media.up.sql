ALTER TABLE files ADD COLUMN IF NOT EXISTS thumbnail_key VARCHAR(512) NULL;
ALTER TABLE files ADD COLUMN IF NOT EXISTS meta_json TEXT NULL;

CREATE TABLE IF NOT EXISTS customer_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL REFERENCES customers(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_customer_users_user ON customer_users(user_id);
CREATE INDEX IF NOT EXISTS idx_customer_users_customer ON customer_users(tenant_id, customer_id);
