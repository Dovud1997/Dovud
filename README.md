# Sales Force Automation (SFA)

White Label SaaS platform for sales force automation — multi-tenant, offline-capable, enterprise-ready.

**Status:** P15 delivered (SQLite EntityCache + FCM client) on top of P0–P14

## Architecture

Full design pack: **[docs/architecture/README.md](docs/architecture/README.md)**

## Stack

- **Backend:** Go, Fiber, GORM, PostgreSQL (SQLite for local/dev), Redis, RabbitMQ, MinIO
- **Frontend:** Flutter (Android, iOS, Web Admin)
- **Delivery:** Docker Compose, Kubernetes-ready Dockerfile

## Quick start (Docker)

```bash
# Full stack: Postgres, Redis, RabbitMQ, MinIO, API, worker, scheduler, Web Admin
docker compose up --build -d

# Web UI:  http://localhost:3000
# API:     http://localhost:8080/api/v1
# Mailhog: http://localhost:8025
# MinIO:   http://localhost:9001  (minioadmin / minioadmin)
```

Demo login in the browser: tenant `demo`, `admin@demo.local` / `Admin123!`

On restricted VMs where container DNS/routing fails, run services with `network_mode: host` (Linux only) or use local SQLite API + `flutter run -d chrome`.

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

**P7 — Portal hardening / API security / Helm prod**
- Portal documents scoped by `documents.customer_id`; summary counts only linked customer docs
- Portal RBAC tightened to `portal:read` + `notifications:read` (no tenant-wide orders/finance/docs)
- Admin link APIs: `GET/POST /portal/links`, `DELETE /portal/links/:user_id` (`portal:write`)
- API hardening: security headers, body limit, CORS config, in-memory rate limit, login lockout (memory/Redis)
- Migration `000008_portal_hardening`
- Helm `0.7.0`: Secret, Ingress, HPA, NetworkPolicy, `values-prod.yaml`
- Flutter: portal-only redirect; portal shows receivables + documents

**P8 — Admin shell / Devices & push / SMS drivers / Encrypted offline**
- Flutter AdminShell: permission-aware NavigationRail / drawer
- `POST/GET/DELETE /auth/devices` with push token registration (stub token until FCM)
- SMS/Push drivers: `log` + `http` webhook; worker resolves push tokens from `user_devices`
- Offline cache/outbox encrypted via secure key + SharedPreferences
- Helm chart `0.8.0`

**P9 — Tenant providers / AgentShell / Portal links admin**
- `tenant_providers` table + AES-GCM encrypted configs
- API: `GET/PUT /tenant/providers`, `POST /tenant/providers/:type/test`
- Worker merges per-tenant SMTP/SMS/Push overrides over global notify config
- Flutter Providers admin page + Portal links admin
- AgentShell bottom nav for `sales_agent` (Home · Customers · Orders · Route · More)
- Helm chart `0.9.0`

**P10 — Users / Roles admin · Branding studio**
- Users API: `role_ids` + `status` on DTOs, `PATCH /users/:id` (name/phone/locale/status)
- Roles API: `permission_codes` on list/create; all `is_system` roles locked for permission edits
- Flutter: Users, Roles, Branding studio pages in AdminShell
- Branding save refreshes session theme via public branding bootstrap
- Helm chart `0.10.0`

**P11 — Domains admin · Logo upload via presign**
- Domains: host normalize/validate, primary uniqueness, `409 DOMAIN_EXISTS`
- `POST /tenant/branding/assets` attaches ready image file (`logo`/`favicon`/`icon`) with long-lived URL
- Flutter Domains page + Branding studio logo upload (presign → PUT → complete → attach)
- Helm chart `0.11.0`

**P12 — FCM HTTP v1 push driver**
- Push drivers: `log` | `http` | `fcm` (service-account JWT → OAuth → FCM v1 send)
- Env: `SFA_PUSH_FCM_PROJECT_ID`, `SFA_PUSH_FCM_CREDENTIALS` / `_FILE`
- Tenant provider config: `project_id` + encrypted `service_account_json`
- Worker skips `stub-push-*` device tokens
- Flutter `PushTokenSource` abstraction (stub default; inject real FCM later)
- Providers admin UI supports FCM fields
- Helm chart `0.12.0`

**P13 — Multi-device push fan-out · EntityCache**
- Worker fans out push to all usable device tokens (dedupe, skip stubs); partial success OK
- Flutter `EntityCache` interface; `OfflineStore` implements it (blob backend until Drift)
- Helm chart `0.13.0`

**P14 — Per-device push deliveries**
- `notification_deliveries` gains `device_id` / `platform` / `token_suffix`
- Worker records one delivery row per device after fan-out
- API: `GET /notifications/:id/deliveries`
- Flutter Notifications page expands to show per-device status
- Helm chart `0.14.0`

**P15 — SQLite EntityCache · FCM / APNs client**
- Mobile/desktop: `SqliteEntityCache` (`cached_entities` + `sync_meta`); one-shot migrate from encrypted blob
- Web Admin keeps encrypted blob `EntityCache` (no sqflite on browser)
- `FcmPushTokenSource` via `firebase_messaging` (APNs through FCM); stub fallback when Firebase is not configured
- Helm chart `0.15.0`

## Next (P16+)

Drift codegen (`build_runner`) · outbox SQLite tables · Firebase options / google-services wiring

## Locales & themes

- Languages: Russian, Uzbek, English
- Themes: Light, Dark, Brand Color (per tenant)
