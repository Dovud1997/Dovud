# Sales Force Automation (SFA)

White Label SaaS platform for sales force automation — multi-tenant, offline-capable, enterprise-ready.

**Status:** P0 implementation in progress (Identity · Tenant/Branding · RBAC · Flutter shell)

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

## P0 delivered

**Backend**

- Clean Architecture modules: `identity`, `tenant`
- JWT access + rotating refresh tokens
- RBAC permissions/roles/users APIs
- White-label public branding + admin branding/domains
- Seed data, AutoMigrate, SQL migrations for Postgres
- Unit/integration tests for auth

**Frontend**

- Flutter app skeleton under `frontend/sfa_app`
- Login, session, branding theme, dashboard shell
- i18n ARB: Russian, Uzbek, English

## Next (P1)

Organization (companies/branches/warehouses) · Catalog · CRM

## Locales & themes

- Languages: Russian, Uzbek, English
- Themes: Light, Dark, Brand Color (per tenant)
