# Sales Force Automation (SFA)

White Label SaaS platform for sales force automation — multi-tenant, offline-capable, enterprise-ready.

**Status:** P23 delivered (WebSocket live channel) on P0–P22

## Architecture

Full design pack: **[docs/architecture/README.md](docs/architecture/README.md)**

## Stack

- **Backend:** Go, Fiber, GORM, PostgreSQL (SQLite for local/dev), Redis, RabbitMQ, MinIO
- **Frontend:** Flutter (Android, iOS, Web Admin)
- **Delivery:** Docker Compose, Kubernetes-ready Dockerfile
- **Offline:** Drift SQLite (native + web Wasm) · outbox · conflict resolve UX
- **Push:** FCM HTTP v1 (server) · `firebase_messaging` client (bring-your-own Firebase project)

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

**P16 — Drift codegen · outbox tables · Firebase options**
- Drift `SfaDatabase`: `cached_entities`, `sync_meta`, `outbox_ops` (codegen via `build_runner`)
- Mobile/desktop: Drift EntityCache + Drift outbox; web keeps encrypted blob
- `lib/firebase_options.dart` + `google-services.json.example` / `GoogleService-Info.plist.example`
- Enable with `--dart-define=SFA_FIREBASE_CONFIGURED=true` or native plist/json + Gradle plugin
- Helm chart `0.16.0`

**P17 — Conflict UX**
- Resolve applies `server_wins` / `client_wins` (client win appends changelog)
- Flutter Sync center lists conflicts; resolve screen Take server / Keep mine
- Outbox marks `conflict` / `rejected` statuses; real `DeviceService` device id on sync calls
- Helm chart `0.17.0`

**P18 — Background sync worker**
- `SyncWorker`: periodic + connectivity regain + app resume → push then pull
- Sync center shows last success/error; “Run sync cycle” action
- Helm chart `0.18.0`

**P19 — Drift web (Wasm)**
- `drift_flutter` opens the same `SfaDatabase` on web (Wasm) and mobile (native)
- Web EntityCache/outbox use Drift (blob remains migration source)
- Helm chart `0.19.0`

**P20 — Firebase enablement (BYO project)**
- Placeholder `firebase_options.dart` + native `.example` configs; Gradle plugin hooks documented
- No production secrets in repo — run `flutterfire configure`, copy examples, set `SFA_FIREBASE_CONFIGURED=true`
- Helm chart `0.20.0`

## Firebase setup (production)

```bash
cd frontend/sfa_app
# 1) Generate real options (overwrites lib/firebase_options.dart)
dart pub global activate flutterfire_cli
flutterfire configure

# 2) Native files
cp android/app/google-services.json.example android/app/google-services.json
cp ios/Runner/GoogleService-Info.plist.example ios/Runner/GoogleService-Info.plist
# replace placeholders with Firebase console files

# 3) Uncomment Google Services plugin in:
#    android/settings.gradle  and  android/app/build.gradle

# 4) Run with Dart options enabled
flutter run --dart-define=SFA_FIREBASE_CONFIGURED=true
```

Without Firebase configured, the app falls back to `stub-push-*` tokens (worker skips them).

## Roadmap delivery notes

**P21 — Domain → sync fan-out · agent writes** *(delivered)*
- `RecordChange` wired from orders / customers / visits into `sync_change_log`
- Agent Flutter screens: create customer, draft order, visit check-in/out via online-first + outbox fallback
- Helm chart `0.21.0`

**P22 — Redis sync locks · offline file upload queue**
- `sync:lock:{tenant}:{device}` via Redis SET NX on push (429 when busy)
- Drift `file_uploads` table + `FileUploadQueue` (enqueue / flush with documents presign)
- SyncWorker flushes uploads after push→pull; Documents page queues demo uploads
- Helm chart `0.22.0`

**P23 — WebSocket live channel**
- `/ws/v1?token=` (Fiber websocket) with ping/pong + `ready`
- `RecordChange` publishes `sync.invalidate` to tenant peers
- Flutter `LiveChannel` + Sync center connection status
- Helm chart `0.23.0`

**P24 — Sync push → domain tables**
- `EntityApplicator` applies customer / order / visit create·update·delete into domain repos
- Push & conflict resolve write domain then changelog; `WithoutFanout` avoids double `RecordChange`
- Helm chart `0.24.0`

**P25 — Field-level conflict merge**
- Resolve API accepts `resolution=merge` + `merged_payload`
- Flutter conflict UI: per-field checkboxes (yours vs server) → Apply merge
- Helm chart `0.25.0`

**P26 — Catalog sync · draft order with lines · live pull**
- Catalog `product` / `product_price` → `RecordChange` fan-out; demo seed writes initial changelog
- Agent order compose: customer + product picker + qty/price lines (online-first + outbox)
- `LiveChannel` triggers `SyncWorker.tick` on `sync.invalidate` and domain WS events
- Domain WS: `order.updated` / `visit.updated` / `product.updated` / `notification.created`
- Conflict UI summarizes nested `lines`
- Helm chart `0.26.0`

**P27 — Returns sync · agent return compose**
- Returns `WithSync` on create / draft update / status transitions → changelog
- Sync `DomainApplicator` supports `return` create·update·delete (cancel)
- WS `return.updated`; agent More → Returns + compose with lines
- Helm chart `0.27.0`

**P28 — Notifications mark-read · unread badge**
- Mark-read / mark-all use `notifications:read` (own inbox)
- Flutter: mark one / mark all, unread filter, app-bar badge (agent + admin)
- WS `notification.created` from create → refresh unread count
- Helm chart `0.28.0`

**P29 — Persist offline upload bytes**
- Drift `file_uploads.payload` blob (schema v3) stores pending file bytes
- `FileUploadQueue` restores bytes after restart; clears blob after successful upload
- Helm chart `0.29.0`

**P30 — Redis Pub/Sub WS fan-out**
- `Hub.WithRedis` / `EventBus`: local deliver + publish on `sfa:ws:events`
- Remote instances skip self-origin and deliver to local sockets only
- In-memory bus for tests; no-op when Redis unavailable
- Helm chart `0.30.0`

**P31 — Per-line conflict merge**
- Conflict UI merges `lines` by `product_id` (yours vs server per row)
- Shared `conflict_line_merge` helper + unit tests
- Helm chart `0.31.0`

**P32 — GPS offline queue**
- Drift `gps_pending` (schema v4); `GpsQueue` enqueue/flush via `POST /gps/points`
- UploadPoints → `RecordChange` `gps_point` + WS `gps.agent.updated`
- Agent Routes: GPS button (demo coords); SyncWorker flushes GPS with uploads
- Helm chart `0.32.0`

Roadmap phases P1–P32 delivered. Optional polish: OS background sync, SQLCipher.

## Locales & themes

- Languages: Russian, Uzbek, English
- Themes: Light, Dark, Brand Color (per tenant)
