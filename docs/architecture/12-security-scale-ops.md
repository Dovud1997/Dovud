# 24–27. Security, Scaling, Backup, Deployment

## 24. Security

### Authentication

- Argon2id password hashing
- JWT access tokens (short-lived, 15m)
- Refresh tokens: opaque, hashed at rest, rotation, reuse detection
- Device binding optional (`device_id` claim)
- Secure storage on mobile for tokens
- Brute-force protection via Redis counters + temporary lockout

### Authorization (RBAC)

- Permission checks on every mutating/sensitive route
- Branch-scoped roles for managers/agents
- Tenant isolation enforced in repository layer (mandatory `tenant_id` predicate)
- Platform super-admin separated and fully audit-logged

### Data protection

- TLS everywhere (ingress + internal mesh preferred)
- Secrets in K8s Secrets / Vault; tenant provider configs encrypted (AES-GCM) at rest
- PII minimization in logs
- MinIO private buckets; presigned URLs with short TTL
- SQL injection avoided via parameterized GORM/SQL
- XSS: Admin web CSP; sanitize HTML if any rich text

### Audit & compliance

- Immutable `audit_logs` for auth, RBAC changes, money docs, branding, exports
- Request ID correlation across API → workers
- Configurable retention / partitioned storage

### API hardening

- Rate limiting (user + IP + tenant)
- Idempotency keys
- Payload size limits
- CORS allowlist per tenant domains
- Security headers (HSTS, X-Content-Type-Options, …)
- Dependency scanning + image scanning in CI

### Mobile

- Certificate pinning (optional enterprise)
- Root/jailbreak detection soft-warn (policy)
- Offline DB unencrypted optional; recommend SQLCipher/Isar encryption for PII tenants

---

## 25. Scaling

### Horizontal scale points

| Component | Scale strategy |
|-----------|----------------|
| `sfa-api` | HPA on CPU/RPS; Redis for WS pubsub & sessions |
| `sfa-worker` | Compete consumers per queue; prefetch tuning |
| `sfa-scheduler` | Single leader (lease in Redis) |
| PostgreSQL | Vertical → read replicas for analytics/pull; partitioning GPS/audit |
| Redis | Cluster if needed; separate logical DBs |
| RabbitMQ | Quorum queues; mirrored in managed offering |
| MinIO | Distributed mode |

### Performance tactics

- CQRS read models for dashboard
- Redis caching for branding, RBAC, hot catalog
- Sync pull pagination + change log
- GPS batching
- Async notifications/media
- Connection pooling (PgBouncer)

### Multi-tenant isolation at scale

- Noisy-neighbor rate limits per tenant
- Optional schema-per-tenant / DB-per-tenant for enterprise tier (Phase 3)
- Per-tenant feature flags

### Future microservice split order

1. Notifications  
2. Sync  
3. Analytics  
4. GPS ingest  
5. Media  

Keep shared `events` contract and auth introspection.

---

## 26. Backup Strategy

| Asset | Method | Frequency | Retention |
|-------|--------|-----------|-----------|
| PostgreSQL | Continuous WAL archiving + daily base backup | continuous / daily | 30d daily, 12 weekly, 12 monthly |
| Redis | RDB+AOF; treat as ephemeral cache (rebuildable) | continuous | short |
| RabbitMQ | Durable queues; definitions export | daily definitions | 30d |
| MinIO | Versioning + bucket replication / mc mirror | continuous / daily | 90d+ |
| K8s config / Helm values | GitOps repo | on change | git history |
| Secrets | Vault snapshot / sealed-secrets backups | daily | 30d |

### Targets

- **RPO:** ≤ 15 minutes (DB WAL)
- **RTO:** ≤ 1 hour for primary region restore

### Tests

- Monthly restore drill to staging
- Backup freshness metric monitored by `backup_verify_ping`
- Documented runbook: DB restore → MinIO → reprocess outbox → smoke tests

### Soft-deleted data

- Soft delete default; hard purge jobs per tenant retention policy (GDPR-like requests as explicit workflow)

---

## 27. Deployment

### Environments

| Env | Purpose |
|-----|---------|
| `local` | docker-compose |
| `dev` | integration |
| `staging` | pre-prod, restore drills |
| `prod` | customer traffic |

### Local (`docker-compose`)

Services: `api`, `worker`, `scheduler`, `postgres`, `redis`, `rabbitmq`, `minio`, `mailhog` (dev).

### Container images

- `sfa-api`
- `sfa-worker`
- `sfa-scheduler`
- Distroless/alpine, non-root, multi-stage build

### Kubernetes (Helm chart `sfa`)

```
Deployments: api, worker, scheduler
Services: api
Ingress: public API + admin web + custom tenant domains
HPA: api, worker
ConfigMaps / Secrets
PVC or external: Postgres, Redis, RabbitMQ, MinIO (prefer managed)
ServiceMonitor / PodMonitor (Prometheus)
NetworkPolicies
```

### Ingress / Domains

- Platform: `api.sfa.example`, `admin.sfa.example`
- Tenant white-label: `crm.customer.com` → same ingress, host-based tenant resolve
- TLS via cert-manager

### CI/CD (GitHub Actions)

1. Lint + unit tests (Go, Dart)
2. Build & scan images
3. Migrate job (init container or CI step with lock)
4. Deploy Helm to env
5. Smoke: health/ready + login
6. OpenAPI publish

### Zero-downtime migrations

- Expand → deploy → contract
- Outbox/consumers backward compatible event versions

### Flutter delivery

| Target | Pipeline |
|--------|----------|
| Android | Play / private APK + white-label flavors optional |
| iOS | TestFlight / App Store |
| Web Admin | Static build → CDN / Nginx / Firebase Hosting |

Runtime branding avoids per-tenant store builds; optional private store builds for deep white-label (icons) using CI matrix later.

### Observability

- OpenTelemetry traces (API → DB/RMQ)
- Metrics: RPS, latency, sync ops, queue depth, error rate
- Structured logs → Loki
- Alerts: 5xx, DB lag, queue DLQ growth, backup age, cert expiry

### Deployment topology (prod target)

```
Internet → CDN (web) → Ingress Controller
                    → sfa-api (N pods)
                    → sfa-worker (M pods)
                    → sfa-scheduler (1)
                    → PostgreSQL primary (+ replica)
                    → Redis
                    → RabbitMQ
                    → MinIO / S3
```

---

## Implementation Phases (post-approval)

| Phase | Scope |
|-------|-------|
| P0 | Monorepo skeleton, platform libs, Identity, Tenant/Branding, RBAC |
| P1 | Org, Catalog, CRM |
| P2 | FieldForce (routes, visits, GPS, photos) |
| P3 | Orders, Returns, Finance |
| P4 | Offline Sync + mobile agent critical path |
| P5 | Notifications, Dashboard/KPI, Audit |
| P6 | Admin web polish, Customer portal, hardening, Helm prod |

---

## Confirmation Required

Architecture pack is complete across documents 01–12.

**Please confirm** to proceed with implementation (starting P0), or list required changes (tenancy model, stack swaps, module priority, etc.).
