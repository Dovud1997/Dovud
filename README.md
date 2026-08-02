# Sales Force Automation (SFA)

White Label SaaS platform for sales force automation — multi-tenant, offline-capable, enterprise-ready.

**Status:** P6 delivered (Notify providers · Thumbnails · Customer portal · Offline cache) on top of P0–P5

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
| Portal | `portal@demo.local` / `Portal123!` |

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

**P5 — Audit / Scheduler / Offline outbox**
- Module: `audit` (`GET /audit-logs`, mutating-request middleware)
- Binary: `cmd/scheduler` — token cleanup, KPI daily snapshots, outbox nudge
- Notifications: email/push/sms enqueue to transactional outbox (`notification.*`)
- Worker: `q.audit.write` + SMS consumer; media handler logs thumbnail queue
- Compose/Helm: `scheduler` service
- Flutter: local SharedPreferences outbox + flush via `/sync/push`; Audit logs screen

**P6 — Notify providers / Thumbnails / Portal / Offline cache**
- Platform: email providers (`log` / `file` / `smtp`), SMS & push log stubs; MailHog in Compose
- Worker: real notify delivery + delivery status updates; JPEG/PNG → 256px thumbnails
- Module: `portal` — read-only customer summary/orders/receivables/documents (`portal:read`)
- Migration `000007_portal_media` (`thumbnail_key`, `customer_users`)
- Demo portal user linked to demo customer
- Flutter: Customer portal screen; OfflineStore entity cache + pullAndCache in Sync center
- Helm chart `0.6.0`

## Next (P6+)

Encrypted Drift/Isar DB · real FCM · richer SMS providers · portal hardening / scoped documents

## Locales & themes

- Languages: Russian, Uzbek, English
- Themes: Light, Dark, Brand Color (per tenant)
