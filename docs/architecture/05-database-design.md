# 5. Database Design (PostgreSQL)

## 5.1 Conventions

| Rule | Value |
|------|-------|
| PK | UUID v7 (time-sortable) preferred, or UUID v4 |
| Tenancy | `tenant_id UUID NOT NULL` on tenant-scoped tables |
| Audit columns | `created_at`, `updated_at`, `created_by`, `updated_by` |
| Soft delete | `deleted_at TIMESTAMPTZ NULL` |
| Concurrency | `version BIGINT NOT NULL DEFAULT 1` on syncable rows |
| Money | `NUMERIC(18,2)` + `currency CHAR(3)` |
| Geo | `lat DOUBLE PRECISION`, `lng DOUBLE PRECISION` (+ PostGIS optional later) |
| JSON | `JSONB` for flexible configs/rules |
| Naming | `snake_case` tables & columns |
| Enums | DB enums or CHECK + app constants |

---

## 5.2 Table Inventory (13. Tables)

### Platform / Tenant

| Table | Purpose |
|-------|---------|
| `tenants` | SaaS customer / company account |
| `tenant_branding` | Logo, colors, icons, app name |
| `tenant_domains` | Custom domains / subdomains |
| `tenant_providers` | SMTP, SMS, Push credentials (encrypted) |
| `tenant_settings` | Feature flags, locale defaults, modules |

### Identity / RBAC

| Table | Purpose |
|-------|---------|
| `users` | User accounts |
| `roles` | Role definitions (system + tenant) |
| `permissions` | Permission catalog |
| `role_permissions` | Role ↔ Permission |
| `user_roles` | User ↔ Role (+ optional branch scope) |
| `refresh_tokens` | Refresh token hashes |
| `user_devices` | Device / push tokens |
| `password_reset_tokens` | Reset flow |
| `login_attempts` | Brute-force tracking |

### Organization

| Table | Purpose |
|-------|---------|
| `companies` | Legal entities under tenant (optional multi-company) |
| `branches` | Branches / depots |
| `warehouses` | Warehouses |
| `warehouse_stocks` | Stock balances per product |

### Catalog

| Table | Purpose |
|-------|---------|
| `manufacturers` | Brands / manufacturers |
| `categories` | Product categories (tree) |
| `products` | SKU master |
| `product_images` | Product media refs |
| `price_lists` | Price list headers |
| `product_prices` | Prices per list/product |
| `promotions` | Campaigns |
| `promotion_items` | Promo scope |
| `promotion_rules` | Discount rules JSON/columns |

### CRM

| Table | Purpose |
|-------|---------|
| `customers` | Outlets / buyers |
| `customer_contacts` | People at customer |
| `customer_addresses` | Addresses + geo |
| `customer_categories` | Customer segmentation |
| `customer_warehouse_links` | Default warehouse/branch |

### Field Force

| Table | Purpose |
|-------|---------|
| `sales_agents` | Agent profile linked to user |
| `agent_territories` | Territory assignments |
| `routes` | Daily/weekly routes |
| `route_stops` | Ordered customer stops |
| `visits` | Visit records |
| `visit_photos` | Photo reports |
| `visit_comments` | Comments on visits |
| `gps_tracks` | GPS breadcrumbs / check-ins |
| `gps_events` | Significant GPS events |

### Orders / Returns

| Table | Purpose |
|-------|---------|
| `orders` | Sales orders |
| `order_lines` | Line items |
| `order_status_history` | Status audit |
| `returns` | Return headers |
| `return_lines` | Return lines |

### Finance

| Table | Purpose |
|-------|---------|
| `receivables` | AR documents / balances |
| `receivable_payments` | Payments / allocations |
| `credit_limits` | Per-customer credit |

### Documents / Files / Comments

| Table | Purpose |
|-------|---------|
| `documents` | Logical documents |
| `files` | MinIO object metadata |
| `document_files` | Document ↔ File |
| `comments` | Polymorphic comments |
| `entity_files` | Polymorphic file links |

### Notifications / KPI / Analytics

| Table | Purpose |
|-------|---------|
| `notifications` | In-app notifications |
| `notification_deliveries` | Channel delivery log |
| `kpi_definitions` | KPI catalog |
| `kpi_snapshots` | Computed KPI values |
| `dashboard_layouts` | Per-role dashboard config |

### Sync / Audit / Jobs

| Table | Purpose |
|-------|---------|
| `sync_devices` | Device sync state |
| `sync_change_log` | Server change feed (optional CDC-like) |
| `sync_conflicts` | Conflict records |
| `audit_logs` | Immutable audit |
| `outbox_events` | Transactional outbox |
| `idempotency_keys` | Durable idempotency |
| `background_job_runs` | Scheduler run history |

---

## 5.3 Core Table Schemas (abbreviated)

### tenants
`id`, `code`, `name`, `status`, `default_locale`, `default_currency`, `timezone`, timestamps, `deleted_at`

### tenant_branding
`id`, `tenant_id`, `app_name`, `logo_file_id`, `favicon_file_id`, `icon_file_id`,  
`primary_color`, `secondary_color`, `accent_color`, `theme_mode_default`,  
`custom_css_json`, `branding_version`, timestamps

### users
`id`, `tenant_id`, `email`, `phone`, `password_hash`, `full_name`, `status`,  
`locale`, `theme_preference`, `last_login_at`, `is_platform_admin`, timestamps, `deleted_at`, `version`

### roles / permissions / role_permissions / user_roles
Standard RBAC M2M; `roles.tenant_id NULL` for system roles.

### branches
`id`, `tenant_id`, `company_id`, `code`, `name`, `address`, `lat`, `lng`, `status`, timestamps, `deleted_at`

### warehouses
`id`, `tenant_id`, `branch_id`, `code`, `name`, `type`, `status`, timestamps, `deleted_at`

### products
`id`, `tenant_id`, `sku`, `barcode`, `name`, `description`, `category_id`, `manufacturer_id`,  
`unit`, `vat_rate`, `is_active`, `attributes_json`, timestamps, `deleted_at`, `version`

### product_prices
`id`, `tenant_id`, `price_list_id`, `product_id`, `amount`, `currency`, `valid_from`, `valid_to`, timestamps, `version`

### customers
`id`, `tenant_id`, `branch_id`, `code`, `name`, `type`, `inn`, `status`,  
`credit_limit`, `balance_cached`, `lat`, `lng`, `address`, timestamps, `deleted_at`, `version`

### sales_agents
`id`, `tenant_id`, `user_id`, `branch_id`, `employee_code`, `manager_id`, `status`, timestamps

### routes / route_stops
`routes`: `id`, `tenant_id`, `agent_id`, `date`, `name`, `status`, timestamps, `version`  
`route_stops`: `id`, `route_id`, `customer_id`, `sequence`, `planned_arrival`, `status`

### visits
`id`, `tenant_id`, `agent_id`, `customer_id`, `route_stop_id`, `started_at`, `ended_at`,  
`checkin_lat`, `checkin_lng`, `checkout_lat`, `checkout_lng`, `result`, `notes`, timestamps, `version`

### orders
`id`, `tenant_id`, `number`, `customer_id`, `agent_id`, `branch_id`, `warehouse_id`,  
`visit_id`, `status`, `currency`, `subtotal`, `discount_total`, `tax_total`, `grand_total`,  
`ordered_at`, `delivery_date`, `price_list_id`, `promotion_id`, `comment`,  
`client_request_id`, timestamps, `deleted_at`, `version`

### order_lines
`id`, `order_id`, `product_id`, `qty`, `unit_price`, `discount`, `tax`, `line_total`, `promotion_item_id`

### returns
`id`, `tenant_id`, `number`, `order_id`, `customer_id`, `agent_id`, `status`, `reason`, totals, timestamps, `version`

### receivables
`id`, `tenant_id`, `customer_id`, `document_type`, `document_id`, `amount`, `paid_amount`,  
`balance`, `due_date`, `status`, currency, timestamps, `version`

### files
`id`, `tenant_id`, `bucket`, `object_key`, `mime`, `size`, `checksum`, `uploaded_by`, timestamps

### notifications
`id`, `tenant_id`, `user_id`, `type`, `title`, `body`, `payload_json`, `read_at`, `channel`, timestamps

### audit_logs
`id`, `tenant_id`, `actor_user_id`, `action`, `entity_type`, `entity_id`,  
`before_json`, `after_json`, `ip`, `user_agent`, `request_id`, `created_at`  
(**no updated_at / no soft delete**)

### outbox_events
`id`, `tenant_id`, `aggregate_type`, `aggregate_id`, `event_type`, `payload_json`,  
`status`, `attempts`, `next_attempt_at`, `created_at`, `published_at`

### sync_devices
`id`, `tenant_id`, `user_id`, `device_id`, `platform`, `app_version`,  
`last_pull_cursor`, `last_push_at`, `last_pull_at`, timestamps

---

## 5.4 Indexes (14. Required Indexes)

### Tenancy & common filters

```sql
-- Pattern on almost all tenant tables
CREATE INDEX idx_{table}_tenant ON {table}(tenant_id) WHERE deleted_at IS NULL;

-- Users
CREATE UNIQUE INDEX uq_users_tenant_email ON users(tenant_id, lower(email)) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_tenant_phone ON users(tenant_id, phone) WHERE deleted_at IS NULL;

-- Roles
CREATE UNIQUE INDEX uq_roles_tenant_code ON roles(tenant_id, code) WHERE deleted_at IS NULL;

-- Domains
CREATE UNIQUE INDEX uq_tenant_domains_host ON tenant_domains(host) WHERE deleted_at IS NULL;

-- Branches / Warehouses
CREATE UNIQUE INDEX uq_branches_tenant_code ON branches(tenant_id, code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX uq_warehouses_tenant_code ON warehouses(tenant_id, code) WHERE deleted_at IS NULL;

-- Catalog
CREATE UNIQUE INDEX uq_products_tenant_sku ON products(tenant_id, sku) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_tenant_category ON products(tenant_id, category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_tenant_manufacturer ON products(tenant_id, manufacturer_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_tenant_name_trgm ON products USING GIN (name gin_trgm_ops); -- optional pg_trgm
CREATE INDEX idx_product_prices_list_product ON product_prices(price_list_id, product_id);
CREATE INDEX idx_promotions_tenant_dates ON promotions(tenant_id, starts_at, ends_at) WHERE deleted_at IS NULL;

-- CRM
CREATE UNIQUE INDEX uq_customers_tenant_code ON customers(tenant_id, code) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_tenant_branch ON customers(tenant_id, branch_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_geo ON customers(tenant_id, lat, lng) WHERE deleted_at IS NULL;
CREATE INDEX idx_customer_contacts_customer ON customer_contacts(customer_id) WHERE deleted_at IS NULL;

-- Field
CREATE INDEX idx_routes_agent_date ON routes(tenant_id, agent_id, date);
CREATE INDEX idx_route_stops_route_seq ON route_stops(route_id, sequence);
CREATE INDEX idx_visits_tenant_agent_time ON visits(tenant_id, agent_id, started_at DESC);
CREATE INDEX idx_visits_customer ON visits(tenant_id, customer_id, started_at DESC);
CREATE INDEX idx_gps_tracks_agent_time ON gps_tracks(tenant_id, agent_id, recorded_at DESC);
CREATE INDEX idx_gps_tracks_visit ON gps_tracks(visit_id, recorded_at);

-- Orders
CREATE UNIQUE INDEX uq_orders_tenant_number ON orders(tenant_id, number) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX uq_orders_client_request ON orders(tenant_id, client_request_id) WHERE client_request_id IS NOT NULL;
CREATE INDEX idx_orders_tenant_status_time ON orders(tenant_id, status, ordered_at DESC);
CREATE INDEX idx_orders_customer ON orders(tenant_id, customer_id, ordered_at DESC);
CREATE INDEX idx_orders_agent ON orders(tenant_id, agent_id, ordered_at DESC);
CREATE INDEX idx_order_lines_order ON order_lines(order_id);

-- Returns & Finance
CREATE INDEX idx_returns_tenant_status ON returns(tenant_id, status, created_at DESC);
CREATE INDEX idx_receivables_customer_status ON receivables(tenant_id, customer_id, status);
CREATE INDEX idx_receivables_due ON receivables(tenant_id, due_date) WHERE status != 'closed';

-- Stock
CREATE UNIQUE INDEX uq_warehouse_stocks ON warehouse_stocks(warehouse_id, product_id);
CREATE INDEX idx_warehouse_stocks_product ON warehouse_stocks(tenant_id, product_id);

-- Files / Docs
CREATE INDEX idx_files_tenant_created ON files(tenant_id, created_at DESC);
CREATE INDEX idx_entity_files_ref ON entity_files(entity_type, entity_id);

-- Notifications
CREATE INDEX idx_notifications_user_unread ON notifications(user_id, created_at DESC) WHERE read_at IS NULL;

-- Sync / Audit / Outbox
CREATE INDEX idx_sync_devices_user ON sync_devices(tenant_id, user_id, device_id);
CREATE INDEX idx_sync_change_log_cursor ON sync_change_log(tenant_id, entity_type, updated_at, id);
CREATE INDEX idx_audit_logs_tenant_time ON audit_logs(tenant_id, created_at DESC);
CREATE INDEX idx_audit_logs_entity ON audit_logs(tenant_id, entity_type, entity_id, created_at DESC);
CREATE INDEX idx_outbox_pending ON outbox_events(status, next_attempt_at) WHERE status = 'pending';
CREATE UNIQUE INDEX uq_idempotency ON idempotency_keys(tenant_id, key);

-- Tokens
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id) WHERE revoked_at IS NULL;
CREATE INDEX idx_user_devices_push ON user_devices(tenant_id, user_id) WHERE push_token IS NOT NULL;
```

---

## 5.5 Partitioning (scale phase)

| Table | Strategy |
|-------|----------|
| `audit_logs` | RANGE by `created_at` (monthly) |
| `gps_tracks` | RANGE by `recorded_at` (weekly/monthly) |
| `sync_change_log` | RANGE by `updated_at` |
| `outbox_events` | archive published > N days |

---

## 5.6 Referential Integrity Policy

- **Within context:** hard FK where lifecycle is coupled (order → order_lines)
- **Across contexts:** store UUID references; validate in application; optional FK if same DB Phase 1
- **No cascading deletes across contexts** — soft delete + domain rules
- Independent entities remain independently versioned for sync

---

## 5.7 Migrations

- Tool: `golang-migrate` SQL files in `backend/migrations`
- Expand/contract for zero-downtime
- Seeds: permissions, system roles, demo tenant (dev only)
