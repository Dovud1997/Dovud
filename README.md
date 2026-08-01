# Sales Force Automation (SFA)

White Label SaaS platform for sales force automation — multi-tenant, offline-capable, enterprise-ready.

**Status:** P4 delivered (Documents/MinIO · Outbox/RabbitMQ worker · Helm) on top of P0–P3

## Architecture

Full design pack: **[docs/architecture/README.md](docs/architecture/README.md)**

## Stack

- **Backend:** Go, Fiber, GORM, PostgreSQL (SQLite for local/dev), Redis, RabbitMQ, MinIO
- **Frontend:** Flutter (Android, iOS, Web Admin)
- **Delivery:** Docker Compose, Kubernetes-ready Dockerfile

## Quick start (local API)

```bash
# Option A — SQLite (no Docker required)
cd backend
go mod tidy
SFA_DATABASE_DSN='sqlite:file:./sfa_dev.db?cache=shared&mode=rwc' go run ./cmd/api -config configs/config.yaml

# Option B — Postgres + infra
docker compose up -d postgres redis rabbitmq minio
cd backend && go run ./cmd/api -config configs/config.yaml
```

API: `http://localhost:8080/api/v1`

### Demo tenant

| Field | Value |
|-------|-------|
| Tenant code | `demo` |
| Admin | `admin@demo.local` / `Admin123!` |
| Agent | `agent@demo.local` / `Agent123!` |

### Smoke

```bash
curl -s http://localhost:8080/api/v1/public/health
curl -s 'http://localhost:8080/api/v1/public/branding?tenant=demo'
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"tenant_code":"demo","email":"admin@demo.local","password":"Admin123!"}'
```

## Delivered

**P0 — Identity / White Label / RBAC**
- Modules: `identity`, `tenant`
- JWT + refresh, RBAC, branding, seed demo users

**P1 — Organization / Catalog / CRM**
- Modules: `organization`, `catalog`, `crm`
- Companies, branches, warehouses, catalog, customers

**P2 — FieldForce / Orders**
- Modules: `fieldforce`, `orders`
- Agents, routes/stops, visit check-in/out, photos, comments, GPS batch/live
- Orders with lines, status machine (draft→…→delivered), idempotent client_request_id
- Demo seed: sales agent, today route, submitted order
- Flutter lists: Routes, Orders (+ prior screens)

**P3 — Returns / Finance / Sync / Notifications / Analytics**
- Modules: `returns`, `finance`, `sync`, `notifications`, `analytics`
- Returns workflow (draft→submit→approve/reject→complete)
- Receivables, payments, credit limits, aging report
- Offline sync: bootstrap / pull / push / conflicts (op_id idempotency)
- In-app notifications + dashboard KPI summary
- Migration `000005_returns_finance_notify_analytics` (+ `000004_sync`)
- Demo seed: return, receivable, credit limit, notification, KPI definitions
- Flutter: Returns, Receivables, Notifications, Sync center, dashboard KPIs

**P4 — Documents / Media / Workers / Helm**
- Module: `documents` (files presign/complete, documents CRUD, attach)
- Platform: MinIO/local object storage, RabbitMQ topology, transactional outbox
- Binary: `cmd/worker` (outbox relay + notify/media consumers)
- Migration `000006_outbox_documents`
- Compose: `worker` service; storage driver `auto|minio|local`
- Helm chart skeleton: `deploy/helm/sfa`
- Flutter: Documents list; Sync center bootstrap/pull/push actions

## Next (P5+)

Full offline Flutter client (local DB + outbox) · audit module · scheduler · notify providers (FCM/email/SMS) · media thumbnails

## Locales & themes

- Languages: Russian, Uzbek, English
- Themes: Light, Dark, Brand Color (per tenant)
