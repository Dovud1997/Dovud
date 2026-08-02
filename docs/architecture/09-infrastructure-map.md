# 12 / 15 / 16 / 17. Services, RabbitMQ, Redis, Background Jobs

## 12. Services (Logical & Deployable)

### Phase 1 deployables

| Service | Binary | Responsibility |
|---------|--------|----------------|
| `sfa-api` | `cmd/api` | REST, WS, sync endpoints |
| `sfa-worker` | `cmd/worker` | Async consumers |
| `sfa-scheduler` | `cmd/scheduler` | Cron jobs |
| PostgreSQL | managed/pod | Primary data |
| Redis | managed/pod | Cache, tokens, WS pubsub, rate limit |
| RabbitMQ | managed/pod | Events / jobs |
| MinIO | managed/pod | Files, photos, branding assets |

### Logical modules inside API (future microservices)

Identity · Tenant · Organization · Catalog · CRM · FieldForce · Orders · Finance · Documents · Notifications · Analytics · Sync · Audit · Gateway

### External providers (per tenant)

| Provider | Use |
|----------|-----|
| SMTP / SES / Mailgun | Email |
| SMS gateway (e.g. Eskiz, Twilio) | SMS OTP/alerts |
| FCM + APNs | Push |

---

## 15. RabbitMQ Topology

### Exchanges

| Exchange | Type | Purpose |
|----------|------|---------|
| `sfa.events` | topic | Domain events |
| `sfa.jobs` | direct | Explicit jobs |
| `sfa.delay` | topic/plugin | Delayed retry |
| `sfa.dlx` | fanout | Dead letters |

### Routing keys / Queues

| Queue | Binding key | Consumer | Purpose |
|-------|-------------|----------|---------|
| `q.notify.push` | `notification.push` | worker | FCM/APNs |
| `q.notify.email` | `notification.email` | worker | Email send |
| `q.notify.sms` | `notification.sms` | worker | SMS send |
| `q.audit.write` | `audit.#` | worker | Persist/enrich audit |
| `q.search.index` | `catalog.#`, `crm.customer.#` | worker | Future search index |
| `q.analytics.project` | `order.#`, `visit.#`, `finance.#` | worker | KPI projections |
| `q.media.process` | `media.#` | worker | Image resize/thumbnails |
| `q.sync.fanout` | `sync.#` | worker | WS invalidate / device hints |
| `q.gps.process` | `gps.#` | worker | Geofence, compression |
| `q.outbox.relay` | internal poll or `outbox.publish` | worker | Publish outbox |
| `q.dlq` | via DLX | manual/ops | Failed messages |

### Core domain event names

```
identity.user.created
identity.user.roles_changed
tenant.branding.updated
catalog.product.updated
catalog.price.updated
crm.customer.updated
field.visit.checked_in
field.visit.checked_out
field.visit.photo_uploaded
field.gps.batch_received
order.submitted
order.status_changed
order.cancelled
return.created
return.approved
finance.payment.recorded
finance.credit_limit.exceeded
document.uploaded
notification.requested
sync.conflict.created
```

### Message envelope

```json
{
  "event_id": "uuid",
  "event_type": "order.submitted",
  "tenant_id": "uuid",
  "occurred_at": "ISO-8601",
  "aggregate_type": "order",
  "aggregate_id": "uuid",
  "payload": {},
  "correlation_id": "uuid",
  "version": 1
}
```

Retries: 3–5 with backoff → DLQ. Idempotent consumers via `event_id`.

---

## 16. Redis Cache Map

| Key pattern | TTL | Purpose |
|-------------|-----|---------|
| `session:refresh:{jti}` | until refresh exp | Refresh token validity |
| `auth:blacklist:{jti}` | access remaining TTL | Revoked access tokens |
| `auth:login_attempts:{tenant}:{login}` | 15m | Brute-force counter |
| `rbac:perms:{userId}:{hash}` | 15–60m | Permission set cache |
| `tenant:by_host:{host}` | 1h | Host → tenant_id |
| `tenant:branding:{tenantId}` | 1h | Branding JSON |
| `tenant:settings:{tenantId}` | 30m | Feature flags |
| `catalog:product:{tenantId}:{id}` | 10m | Hot product |
| `catalog:pricelist:{tenantId}:{id}` | 10m | Price list blob |
| `crm:customer:{tenantId}:{id}` | 10m | Hot customer |
| `dashboard:summary:{tenantId}:{userId}:{date}` | 1–5m | Dashboard widgets |
| `kpi:{tenantId}:{code}:{period}:{scope}` | 5m | KPI snapshot cache |
| `ratelimit:{tenantId}:{userId}:{route}` | sliding window | API rate limits |
| `idempotency:{tenantId}:{key}` | 24h | Idempotent responses |
| `ws:pubsub` channel | n/a | Redis pub/sub for WS scale |
| `gps:live:{tenantId}:{agentId}` | 5m | Last GPS point |
| `sync:lock:{deviceId}` | 30s | Prevent concurrent push |
| `outbox:leader` | 30s | Optional relay leader lock |

**Invalidation:** domain events → delete/update related keys (worker or in-process).

---

## 17. Background Jobs (Scheduler + Workers)

### Scheduler (cron)

| Job | Schedule | Description |
|-----|----------|-------------|
| `outbox_sweep` | every 10s | Publish pending outbox |
| `token_cleanup` | hourly | Purge expired refresh/reset tokens |
| `login_attempt_cleanup` | hourly | Purge old attempts |
| `kpi_recompute_daily` | 02:00 tenant TZ | Daily KPI snapshots |
| `kpi_recompute_hourly` | hourly | Near-real-time KPIs |
| `dashboard_warmup` | every 5m | Prefetch top dashboards |
| `gps_stale_check` | every 5m | Flag offline agents |
| `route_ Reminder` | daily morning | Push route summary to agents |
| `receivable_overdue_mark` | daily | Mark overdue AR |
| `notification_retry` | every minute | Retry failed deliveries |
| `media_orphan_cleanup` | daily | Delete unconfirmed uploads |
| `audit_partition_maintain` | monthly | Create partitions |
| `sync_changelog_compact` | daily | Compact old change log |
| `backup_verify_ping` | daily | Check backup freshness metric |

### Worker job handlers

| Job | Trigger | Work |
|-----|---------|------|
| SendPush | queue | FCM/APNs via tenant provider |
| SendEmail | queue | SMTP |
| SendSMS | queue | SMS gateway |
| ProcessImage | queue | Thumbnails for photos/logos |
| ProjectAnalytics | queue | Update read models |
| WriteAudit | queue | Ensure audit durability/export |
| GeofenceEvaluate | queue | Visit distance compliance |
| CreditAlert | event | Notify finance on limit breach |
| OrderStatusSideEffects | event | Stock reserve/release |
| BrandingCacheInvalidate | event | Clear Redis branding |

### Mobile background sync

| Job | Platform | Description |
|-----|----------|-------------|
| `BackgroundSyncTask` (`sfa.background.sync`) | Android WorkManager (~15m) / iOS BGTask + background fetch | Push outbox + pull deltas + upload/GPS flush |
| `GpsBatchFlush` | part of background + foreground SyncWorker | Flush buffered GPS |
| Device GPS | Flutter `geolocator` | Agent Routes GPS button + visit check-in coords |
| `PushTokenRefresh` | on token change | Register device |
