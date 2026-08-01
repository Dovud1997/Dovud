# 3. Backend Architecture

## 3.1 Goals

- Clean Architecture + DDD per bounded context
- CQRS where reads diverge from writes (analytics, dashboards, sync pull)
- Repository Pattern + DI
- Horizontal scale of API and workers independently
- Extractable modules without rewriting domain

---

## 3.2 Layer Responsibilities

### Domain

- Entities, Aggregates, Value Objects
- Domain services (pure business rules)
- Repository **interfaces**
- Domain events
- No Fiber, GORM, Redis, RabbitMQ imports

### Application

- Use cases / Command & Query handlers
- Transaction boundaries (Unit of Work)
- Authorization context checks (tenant + permission)
- DTO mapping
- Orchestrates repositories + publishers

### Infrastructure

- GORM models & repositories
- Redis caches
- RabbitMQ publishers/consumers
- MinIO client
- External SMS/Email/Push providers
- Outbox relay

### Interfaces

- Fiber route registration
- Request validation (go-playground/validator)
- Response envelopes
- Swagger annotations
- WebSocket handlers

---

## 3.3 Dependency Injection

Composition root: `internal/app` using **uber/fx** (or dig).

```
fx.New(
  platform.Module,
  identity.Module,
  tenant.Module,
  organization.Module,
  catalog.Module,
  crm.Module,
  fieldforce.Module,
  orders.Module,
  finance.Module,
  documents.Module,
  notifications.Module,
  analytics.Module,
  sync.Module,
  audit.Module,
  gateway.Module,
)
```

Each module exports: repositories, use cases, route registrar, event handlers.

---

## 3.4 CQRS Strategy

| Area | Pattern | Notes |
|------|---------|-------|
| Auth, CRUD masters | Simple CRUD / single model | No forced CQRS |
| Dashboard / KPI | Query side + read models | Materialized views / projections |
| Analytics reports | Query handlers + SQL views | Redis cache warm |
| Offline Sync pull | Dedicated query models | Delta by `version` / `updated_at` |
| Order create | Command side | Emits events for stock, finance, notify |

Commands mutate aggregates; Queries never mutate.

---

## 3.5 Aggregate Roots (primary)

| Context | Aggregate Root | Children / VO |
|---------|----------------|---------------|
| Identity | User | RoleAssignment, Session |
| Identity | Role | Permission bindings |
| Tenant | Tenant | Branding, Domain, ProviderConfig |
| Organization | Branch | Address |
| Organization | Warehouse | WarehouseStock (or separate) |
| Catalog | Product | ProductPrice, ProductImage, CategoryLink |
| Catalog | Promotion | PromotionRule, PromotionItem |
| CRM | Customer | Contact, CustomerAddress, CreditLimit |
| FieldForce | Route | RouteStop |
| FieldForce | Visit | VisitPhoto, VisitComment, GpsPoint |
| Orders | Order | OrderLine, OrderStatusHistory |
| Orders | Return | ReturnLine |
| Finance | Receivable | ReceivablePayment |
| Documents | Document | DocumentFile |
| Notifications | Notification | DeliveryAttempt |
| Sync | SyncSession | SyncOperation, Conflict |
| Audit | AuditLog | (immutable append-only) |

---

## 3.6 Cross-Cutting Concerns

### Error model

```
AppError { code, message_key, details, http_status, cause }
```

i18n message keys resolved on client or via `Accept-Language`.

### Response envelope

```json
{
  "success": true,
  "data": {},
  "meta": { "page": 1, "per_page": 20, "total": 100 },
  "error": null,
  "request_id": "uuid"
}
```

### Outbox Pattern

1. Business TX writes aggregate + `outbox_events` row
2. Relay publishes to RabbitMQ
3. Marks published / moves failures to DLQ after retries

### Idempotency

- Header `Idempotency-Key` for POST/PUT critical ops (orders, payments, sync push)
- Stored in Redis (TTL 24h) + optional DB for durability

### Soft delete

- `deleted_at` on most entities
- Unique indexes include partial `WHERE deleted_at IS NULL`

### Optimistic concurrency

- `version` bigint on syncable entities
- Conflict → 409 + sync conflict payload

---

## 3.7 Auth & Session Flow (Backend)

```
Login → validate credentials → issue:
  access_token  (JWT, short-lived, 15m)
  refresh_token (opaque, Redis + DB, 30d, rotatable)

Access JWT claims:
  sub, tenant_id, roles[], permissions_hash, device_id, jti, exp, iat

Refresh:
  rotate refresh token; blacklist old jti in Redis
Logout:
  revoke refresh + blacklist access jti until exp
```

Password hashing: Argon2id.  
Optional: MFA later (TOTP) — interface reserved.

---

## 3.8 RBAC Model

```
Permission = resource:action   e.g. orders:create, customers:read
Role = named set of permissions (tenant-scoped + system roles)
User ⟷ Role (many-to-many, optionally scoped to branch)
```

### System roles (seed)

| Role | Scope |
|------|-------|
| `platform_superadmin` | Cross-tenant |
| `tenant_owner` | Full tenant |
| `tenant_admin` | Admin without billing |
| `sales_manager` | Branch/team oversight |
| `sales_agent` | Field operations |
| `warehouse_clerk` | Stock / fulfillment |
| `finance_manager` | AR / payments |
| `viewer` | Read-only |
| `customer_portal` | External customer (limited) |

Permission checks in middleware + use-case guards.

---

## 3.9 WebSocket Architecture

Endpoint: `GET /ws/v1?token=...`

Channels (logical):

| Channel | Purpose |
|---------|---------|
| `user:{userId}` | Personal notifications |
| `tenant:{tenantId}:orders` | Order status (role-filtered) |
| `tenant:{tenantId}:visits` | Visit updates for managers |
| `agent:{userId}:sync` | Sync hints / invalidate |
| `tenant:{tenantId}:kpi` | Live dashboard ticks (optional) |

Hub in API process; scale via Redis Pub/Sub adapter.

---

## 3.10 Worker Processes

| Process | Role |
|---------|------|
| `api` | REST + WS |
| `worker` | RabbitMQ consumers (notify, audit, projections, media) |
| `scheduler` | Cron: GPS stale check, KPI recompute, token cleanup, outbox sweep |

---

## 3.11 Swagger

- Annotations on handlers
- Generated OpenAPI 3 at `/api/v1/swagger/*`
- Versioned per major API release

---

## 3.12 Testing Strategy

| Level | Scope |
|-------|-------|
| Unit | Domain services, pure use cases |
| Integration | Repositories + Postgres testcontainers |
| Contract | OpenAPI validation |
| E2E | Critical flows (login, order, sync) |
| Load | Sync push, GPS ingest, dashboard queries |
