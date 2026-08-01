CREATE TABLE IF NOT EXISTS returns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    number VARCHAR(64) NOT NULL,
    order_id UUID NULL REFERENCES orders(id),
    customer_id UUID NOT NULL REFERENCES customers(id),
    agent_id UUID NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'draft',
    reason TEXT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'UZS',
    subtotal NUMERIC(18,2) NOT NULL DEFAULT 0,
    tax_total NUMERIC(18,2) NOT NULL DEFAULT 0,
    grand_total NUMERIC(18,2) NOT NULL DEFAULT 0,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_returns_tenant_number ON returns(tenant_id, number) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS return_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    return_id UUID NOT NULL REFERENCES returns(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    qty NUMERIC(18,3) NOT NULL,
    unit_price NUMERIC(18,2) NOT NULL,
    line_total NUMERIC(18,2) NOT NULL,
    reason TEXT NULL
);

CREATE TABLE IF NOT EXISTS receivables (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    customer_id UUID NOT NULL REFERENCES customers(id),
    document_type VARCHAR(32) NOT NULL DEFAULT 'manual',
    document_id UUID NULL,
    amount NUMERIC(18,2) NOT NULL,
    paid_amount NUMERIC(18,2) NOT NULL DEFAULT 0,
    balance NUMERIC(18,2) NOT NULL,
    due_date DATE NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'open',
    currency CHAR(3) NOT NULL DEFAULT 'UZS',
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);
CREATE INDEX IF NOT EXISTS idx_receivables_customer_status ON receivables(tenant_id, customer_id, status);

CREATE TABLE IF NOT EXISTS receivable_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    receivable_id UUID NOT NULL REFERENCES receivables(id) ON DELETE CASCADE,
    amount NUMERIC(18,2) NOT NULL,
    paid_at TIMESTAMPTZ NOT NULL,
    method VARCHAR(32) NOT NULL DEFAULT 'cash',
    reference TEXT NULL,
    created_by UUID NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS credit_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    customer_id UUID NOT NULL REFERENCES customers(id),
    amount NUMERIC(18,2) NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'UZS',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_credit_limits_customer ON credit_limits(tenant_id, customer_id);

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    user_id UUID NOT NULL REFERENCES users(id),
    type VARCHAR(64) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    payload_json TEXT NULL,
    read_at TIMESTAMPTZ NULL,
    channel VARCHAR(32) NOT NULL DEFAULT 'in_app',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS notification_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    channel VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    error TEXT NULL,
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS kpi_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NULL,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NULL,
    unit VARCHAR(32) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS kpi_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    kpi_code VARCHAR(64) NOT NULL,
    period VARCHAR(32) NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    scope_type VARCHAR(32) NOT NULL DEFAULT 'tenant',
    scope_id UUID NULL,
    value DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_kpi_snapshots_lookup ON kpi_snapshots(tenant_id, kpi_code, period, period_start);
