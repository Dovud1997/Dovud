CREATE TABLE IF NOT EXISTS sales_agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    user_id UUID NOT NULL REFERENCES users(id),
    branch_id UUID NOT NULL REFERENCES branches(id),
    employee_code VARCHAR(64) NOT NULL,
    manager_id UUID NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);
CREATE INDEX IF NOT EXISTS idx_sales_agents_tenant ON sales_agents(tenant_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_sales_agents_user ON sales_agents(tenant_id, user_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    agent_id UUID NOT NULL REFERENCES sales_agents(id),
    date DATE NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'planned',
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);
CREATE INDEX IF NOT EXISTS idx_routes_agent_date ON routes(tenant_id, agent_id, date);

CREATE TABLE IF NOT EXISTS route_stops (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL REFERENCES customers(id),
    sequence INT NOT NULL,
    planned_arrival TIMESTAMPTZ NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending'
);
CREATE INDEX IF NOT EXISTS idx_route_stops_route_seq ON route_stops(route_id, sequence);

CREATE TABLE IF NOT EXISTS visits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    agent_id UUID NOT NULL REFERENCES sales_agents(id),
    customer_id UUID NOT NULL REFERENCES customers(id),
    route_stop_id UUID NULL REFERENCES route_stops(id),
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ NULL,
    checkin_lat DOUBLE PRECISION NULL,
    checkin_lng DOUBLE PRECISION NULL,
    checkout_lat DOUBLE PRECISION NULL,
    checkout_lng DOUBLE PRECISION NULL,
    result VARCHAR(32) NOT NULL DEFAULT '',
    notes TEXT NULL,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);
CREATE INDEX IF NOT EXISTS idx_visits_tenant_agent_time ON visits(tenant_id, agent_id, started_at DESC);

CREATE TABLE IF NOT EXISTS visit_photos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    visit_id UUID NOT NULL REFERENCES visits(id) ON DELETE CASCADE,
    file_url TEXT NOT NULL,
    caption TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS visit_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    visit_id UUID NOT NULL REFERENCES visits(id) ON DELETE CASCADE,
    author_user_id UUID NOT NULL REFERENCES users(id),
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS gps_tracks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    agent_id UUID NOT NULL REFERENCES sales_agents(id),
    visit_id UUID NULL REFERENCES visits(id),
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    accuracy DOUBLE PRECISION NULL,
    recorded_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_gps_tracks_agent_time ON gps_tracks(tenant_id, agent_id, recorded_at DESC);

CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    number VARCHAR(64) NOT NULL,
    customer_id UUID NOT NULL REFERENCES customers(id),
    agent_id UUID NULL REFERENCES sales_agents(id),
    branch_id UUID NULL REFERENCES branches(id),
    warehouse_id UUID NULL REFERENCES warehouses(id),
    visit_id UUID NULL REFERENCES visits(id),
    status VARCHAR(32) NOT NULL DEFAULT 'draft',
    currency CHAR(3) NOT NULL DEFAULT 'UZS',
    subtotal NUMERIC(18,2) NOT NULL DEFAULT 0,
    discount_total NUMERIC(18,2) NOT NULL DEFAULT 0,
    tax_total NUMERIC(18,2) NOT NULL DEFAULT 0,
    grand_total NUMERIC(18,2) NOT NULL DEFAULT 0,
    ordered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    delivery_date DATE NULL,
    price_list_id UUID NULL,
    promotion_id UUID NULL,
    comment TEXT NULL,
    client_request_id VARCHAR(128) NULL,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_orders_tenant_number ON orders(tenant_id, number) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_orders_client_request ON orders(tenant_id, client_request_id) WHERE client_request_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_orders_tenant_status_time ON orders(tenant_id, status, ordered_at DESC);

CREATE TABLE IF NOT EXISTS order_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    qty NUMERIC(18,3) NOT NULL,
    unit_price NUMERIC(18,2) NOT NULL,
    discount NUMERIC(18,2) NOT NULL DEFAULT 0,
    tax NUMERIC(18,2) NOT NULL DEFAULT 0,
    line_total NUMERIC(18,2) NOT NULL,
    promotion_item_id UUID NULL
);

CREATE TABLE IF NOT EXISTS order_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    from_status VARCHAR(32) NOT NULL,
    to_status VARCHAR(32) NOT NULL,
    changed_by UUID NULL,
    comment TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
