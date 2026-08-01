# 1. System Architecture

## 1.1 Product Vision

**Sales Force Automation (SFA)** — multi-tenant White Label SaaS for field sales teams.

Sellable to distributors, FMCG, pharma, wholesale, and retail chains without code changes per customer.

### Core value

- Agents work offline in the field (orders, visits, GPS, photos)
- Managers control routes, KPIs, receivables, and compliance in real time
- Each tenant brands the product as their own (logo, colors, domain, SMS/Push/Email)

---

## 1.2 Architectural Principles

| Principle | Application |
|-----------|-------------|
| Clean Architecture | Domain has zero infra dependencies |
| DDD | Bounded contexts = future microservices |
| SOLID | Interfaces at module boundaries |
| CQRS | Applied to Analytics, Dashboard, heavy list reads |
| Repository Pattern | Persistence abstracted from domain |
| DI | Wire via constructor injection (uber/fx or dig) |
| API Versioning | `/api/v1`, `/api/v2` |
| Event-Driven | Domain events → RabbitMQ for cross-module async |
| Multi-tenant | Every tenant-scoped row has `tenant_id` |
| Independent entities | Soft coupling via IDs + events, not hard FK cascades across contexts |

---

## 1.3 High-Level Architecture

```
                         ┌─────────────────────────────────────────┐
                         │           Edge / Ingress                │
                         │  CDN · TLS · Custom Domains · WAF       │
                         └──────────────────┬──────────────────────┘
                                            │
                    ┌───────────────────────┼───────────────────────┐
                    │                       ▼                       │
                    │            API Gateway (Fiber)                 │
                    │   Auth · Rate Limit · Tenant Resolve · RBAC    │
                    │   REST /api/v1/*  ·  WebSocket /ws/v1          │
                    └───────────────────────┬───────────────────────┘
                                            │
          ┌─────────────────────────────────┼─────────────────────────────────┐
          │                                 │                                 │
          ▼                                 ▼                                 ▼
 ┌─────────────────┐              ┌─────────────────┐              ┌─────────────────┐
 │ Identity Module │              │  Domain Modules │              │ Sync / Realtime │
 │ Auth · RBAC     │◄────────────►│ Org · Catalog   │◄────────────►│ Offline Sync    │
 │ Audit · Users   │   events     │ CRM · Orders    │   events     │ WebSocket Hub   │
 └────────┬────────┘              │ Field · Finance │              │ Notifications   │
          │                       │ Docs · Analytics│              └────────┬────────┘
          │                       └────────┬────────┘                       │
          │                                │                                │
          ▼                                ▼                                ▼
 ┌──────────────────────────────────────────────────────────────────────────────────┐
 │                         Infrastructure Layer                                      │
 │  PostgreSQL  │  Redis  │  RabbitMQ  │  MinIO  │  Workers  │  Scheduler            │
 └──────────────────────────────────────────────────────────────────────────────────┘
          │
          ▼
 ┌──────────────────────────────────────────────────────────────────────────────────┐
 │  Clients: Flutter Mobile (Agent)  ·  Flutter Web (Admin)  ·  Customer Portal     │
 └──────────────────────────────────────────────────────────────────────────────────┘
```

---

## 1.4 Deployment Shape: Modular Monolith (Microservices-Ready)

**Phase 1 (now):** Single deployable Go binary with internal modules.  
**Phase 2:** Extract high-load modules (Sync, Notifications, Analytics, GPS ingest) into separate services sharing contracts.

### Why modular monolith first

- Shared transactions where needed (order + stock reservation)
- Lower ops cost for SaaS early stage
- Clear package boundaries already equal future service boundaries
- Same Docker/K8s manifests; split is network + deploy change, not rewrite

### Module → future microservice mapping

| Module | Future Service | Split trigger |
|--------|----------------|---------------|
| `identity` | auth-service | Multi-product SSO |
| `tenant` | tenant-service | Billing/white-label platform |
| `organization` | org-service | Large org trees |
| `catalog` | catalog-service | High read QPS |
| `crm` | crm-service | CRM productization |
| `fieldforce` | field-service | GPS stream volume |
| `orders` | order-service | Order throughput |
| `finance` | finance-service | Accounting integrations |
| `documents` | media-service | Large upload traffic |
| `notifications` | notify-service | Multi-channel volume |
| `analytics` | analytics-service | Heavy OLAP |
| `sync` | sync-service | Offline conflict load |
| `audit` | audit-service | Compliance retention |

---

## 1.5 Bounded Contexts (DDD)

```
┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│   Identity   │  │    Tenant    │  │ Organization │  │   Catalog    │
│ Users Roles  │  │ White Label  │  │ Company      │  │ Products     │
│ Permissions  │  │ Branding     │  │ Branch       │  │ Prices       │
│ Sessions     │  │ Providers    │  │ Warehouse    │  │ Promotions   │
└──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘

┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│     CRM      │  │  FieldForce  │  │    Orders    │  │   Finance    │
│ Customers    │  │ Agents       │  │ Orders       │  │ Receivables  │
│ Contacts     │  │ Routes Visits│  │ Returns      │  │ Payments     │
│              │  │ GPS Photos   │  │              │  │              │
└──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘

┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  Documents   │  │ Notification │  │  Analytics   │  │    Sync      │
│ Files Docs   │  │ Push SMS     │  │ KPI Dashboard│  │ Offline      │
│ Comments     │  │ Email InApp  │  │ Reports      │  │ Conflicts    │
└──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘
                              ┌──────────────┐
                              │    Audit     │
                              │ Immutable log│
                              └──────────────┘
```

**Anti-corruption:** Cross-context communication via:
1. Synchronous application services (same process, Phase 1)
2. Domain events published to RabbitMQ
3. Read models / projections for Analytics (CQRS)

---

## 1.6 Multi-Tenancy Model

```
Platform (SaaS operator)
 └── Tenant (Company / Customer of SFA)
      ├── Branding (logo, colors, icons, app name)
      ├── Domains (custom domain → tenant)
      ├── Providers (SMTP, SMS, FCM/APNs)
      ├── Branches
      ├── Warehouses
      ├── Users (Admin, Manager, Agent, …)
      └── Business data (catalog, customers, orders, …)
```

### Tenant resolution order

1. `X-Tenant-ID` / JWT claim `tenant_id` (authenticated)
2. Custom Host header → `tenant_domains`
3. Subdomain `{slug}.sfa.platform.tld`
4. Mobile app build-time / runtime config `tenant_code`

### Isolation rules

- Every query filtered by `tenant_id`
- JWT always carries `tenant_id` + `user_id` + `roles`
- Super-admin (platform) role can cross tenants (audit-logged)
- Future: PostgreSQL RLS policies per tenant

---

## 1.7 White Label Runtime Model

Configurable **without code deploy**:

| Asset | Storage | Applied where |
|-------|---------|---------------|
| Logo / icons / favicon | MinIO + CDN URL | Mobile splash, Admin sidebar, PWA |
| Brand colors | `tenant_branding` JSON | Flutter Theme + CSS variables (web) |
| App / company name | `tenant_branding` | Titles, emails, push title prefix |
| Custom domain | `tenant_domains` | Ingress / TLS / tenant resolve |
| Email from/name | `tenant_providers.smtp` | Notification worker |
| SMS sender | `tenant_providers.sms` | Notification worker |
| Push credentials | `tenant_providers.push` | FCM/APNs per tenant |

Mobile clients fetch `/api/v1/public/branding?tenant=...` before login (or use cached branding).

---

## 1.8 Technology Stack

| Layer | Technology |
|-------|------------|
| API | Go 1.22+, Fiber |
| ORM | GORM + golang-migrate |
| DB | PostgreSQL 16 |
| Cache / tokens | Redis 7 |
| Queue | RabbitMQ 3.13 |
| Object storage | MinIO |
| Docs | Swagger (swaggo) |
| Clients | Flutter 3.x (Android, iOS, Web) |
| Containers | Docker |
| Orchestration | Kubernetes (Helm) |
| Observability | OpenTelemetry, Prometheus, Grafana, Loki |
| CI/CD | GitHub Actions |

---

## 1.9 Request Lifecycle

```
Client → Ingress → Gateway Middleware chain:
  1. Request ID / Correlation ID
  2. Tenant Resolve
  3. Auth (JWT) — except public routes
  4. RBAC permission check
  5. Rate limit (Redis)
  6. Idempotency-Key (for POST mutating)
  7. Handler → Application UseCase → Domain → Repository
  8. Domain Events → Outbox → RabbitMQ
  9. Audit Log (async)
 10. Response (+ ETag / sync version headers when applicable)
```

---

## 1.10 Non-Functional Targets

| Metric | Target |
|--------|--------|
| API p95 latency (CRUD) | < 200 ms |
| Auth login | < 300 ms |
| Sync batch (100 ops) | < 2 s |
| Availability | 99.9% |
| RPO | ≤ 15 min (WAL archive) |
| RTO | ≤ 1 h |
| Offline support | Full agent critical path |
| Concurrent tenants | 1000+ |
| Agents per tenant | 5000+ (horizontal scale) |
