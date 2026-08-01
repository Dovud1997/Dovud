# 11. Full API Structure

Base URL: `https://{host}/api/v1`  
WebSocket: `wss://{host}/ws/v1`  
Auth: `Authorization: Bearer <access_token>`  
Tenant: from JWT / `X-Tenant-ID` / Host  
Idempotency: `Idempotency-Key` on mutating POSTs  
Locale: `Accept-Language: ru|uz|en`  
API versioning: URI prefix `/api/v1`; breaking changes → `/api/v2`

Common query params: `page`, `per_page`, `sort`, `q`, `updated_since`, `include_deleted`

---

## 11.1 Public

| Method | Path | Description |
|--------|------|-------------|
| GET | `/public/health` | Liveness |
| GET | `/public/ready` | Readiness |
| GET | `/public/branding` | White-label branding by host/code |
| GET | `/public/locales` | Supported locales |

---

## 11.2 Auth

| Method | Path | Description |
|--------|------|-------------|
| POST | `/auth/login` | Login → access + refresh |
| POST | `/auth/refresh` | Rotate tokens |
| POST | `/auth/logout` | Revoke session |
| POST | `/auth/forgot-password` | Start reset |
| POST | `/auth/reset-password` | Complete reset |
| GET | `/auth/me` | Current user profile |
| PATCH | `/auth/me` | Update profile/locale/theme |
| POST | `/auth/change-password` | Change password |
| POST | `/auth/devices` | Register device/push token |
| DELETE | `/auth/devices/{id}` | Remove device |

---

## 11.3 RBAC

| Method | Path | Description |
|--------|------|-------------|
| GET | `/permissions` | Permission catalog |
| GET | `/roles` | List roles |
| POST | `/roles` | Create role |
| GET | `/roles/{id}` | Get role |
| PUT | `/roles/{id}` | Update role |
| DELETE | `/roles/{id}` | Soft delete role |
| PUT | `/roles/{id}/permissions` | Replace permissions |
| GET | `/users` | List users |
| POST | `/users` | Invite/create user |
| GET | `/users/{id}` | Get user |
| PUT | `/users/{id}` | Update user |
| DELETE | `/users/{id}` | Deactivate |
| PUT | `/users/{id}/roles` | Assign roles |
| POST | `/users/{id}/reset-password` | Admin reset |

---

## 11.4 Tenant / White Label

| Method | Path | Description |
|--------|------|-------------|
| GET | `/tenant` | Current tenant |
| PUT | `/tenant` | Update tenant profile |
| GET | `/tenant/branding` | Get branding |
| PUT | `/tenant/branding` | Update branding |
| POST | `/tenant/branding/assets` | Upload logo/icon (presign flow) |
| GET | `/tenant/domains` | List domains |
| POST | `/tenant/domains` | Add domain |
| DELETE | `/tenant/domains/{id}` | Remove domain |
| GET | `/tenant/providers` | List providers (secrets masked) |
| PUT | `/tenant/providers/{type}` | Upsert SMTP/SMS/Push |
| POST | `/tenant/providers/{type}/test` | Send test email/sms/push |
| GET | `/tenant/settings` | Settings |
| PUT | `/tenant/settings` | Update settings |

---

## 11.5 Organization

### Companies
| Method | Path |
|--------|------|
| GET/POST | `/companies` |
| GET/PUT/DELETE | `/companies/{id}` |

### Branches
| Method | Path |
|--------|------|
| GET/POST | `/branches` |
| GET/PUT/DELETE | `/branches/{id}` |

### Warehouses
| Method | Path |
|--------|------|
| GET/POST | `/warehouses` |
| GET/PUT/DELETE | `/warehouses/{id}` |
| GET | `/warehouses/{id}/stocks` |
| PUT | `/warehouses/{id}/stocks/{productId}` |

---

## 11.6 Catalog

### Manufacturers / Categories / Products
| Method | Path |
|--------|------|
| GET/POST | `/manufacturers` |
| GET/PUT/DELETE | `/manufacturers/{id}` |
| GET/POST | `/categories` |
| GET/PUT/DELETE | `/categories/{id}` |
| GET | `/categories/tree` |
| GET/POST | `/products` |
| GET/PUT/DELETE | `/products/{id}` |
| POST | `/products/{id}/images` |
| DELETE | `/products/{id}/images/{imageId}` |
| POST | `/products/import` |
| GET | `/products/export` |

### Prices / Promotions
| Method | Path |
|--------|------|
| GET/POST | `/price-lists` |
| GET/PUT/DELETE | `/price-lists/{id}` |
| GET/POST | `/price-lists/{id}/prices` |
| PUT | `/price-lists/{id}/prices/bulk` |
| GET/POST | `/promotions` |
| GET/PUT/DELETE | `/promotions/{id}` |
| PUT | `/promotions/{id}/items` |
| POST | `/pricing/quote` | Quote cart with promos |

---

## 11.7 CRM

| Method | Path |
|--------|------|
| GET/POST | `/customers` |
| GET/PUT/DELETE | `/customers/{id}` |
| GET/POST | `/customers/{id}/contacts` |
| PUT/DELETE | `/customers/{id}/contacts/{contactId}` |
| GET/POST | `/customers/{id}/addresses` |
| PUT/DELETE | `/customers/{id}/addresses/{addressId}` |
| GET | `/customers/{id}/timeline` | Visits/orders/AR summary |
| GET/POST | `/customer-categories` |
| GET/PUT/DELETE | `/customer-categories/{id}` |

---

## 11.8 Field Force

### Agents / Routes
| Method | Path |
|--------|------|
| GET/POST | `/agents` |
| GET/PUT/DELETE | `/agents/{id}` |
| GET/PUT | `/agents/{id}/territories` |
| GET/POST | `/routes` |
| GET/PUT/DELETE | `/routes/{id}` |
| PUT | `/routes/{id}/stops` | Replace ordered stops |
| POST | `/routes/generate` | Auto-generate (optional) |

### Visits / GPS / Photos / Comments
| Method | Path |
|--------|------|
| GET/POST | `/visits` |
| GET | `/visits/{id}` |
| POST | `/visits/check-in` |
| POST | `/visits/{id}/check-out` |
| POST | `/visits/{id}/photos` |
| DELETE | `/visits/{id}/photos/{photoId}` |
| GET/POST | `/visits/{id}/comments` |
| POST | `/gps/points` | Batch GPS upload |
| GET | `/gps/agents/{id}/live` | Latest position |
| GET | `/gps/agents/{id}/track` | History |

---

## 11.9 Orders & Returns

| Method | Path |
|--------|------|
| GET/POST | `/orders` |
| GET | `/orders/{id}` |
| PUT | `/orders/{id}` | Edit draft |
| POST | `/orders/{id}/submit` |
| POST | `/orders/{id}/confirm` |
| POST | `/orders/{id}/cancel` |
| POST | `/orders/{id}/status` | Transition |
| GET | `/orders/{id}/history` |
| GET/POST | `/returns` |
| GET | `/returns/{id}` |
| POST | `/returns/{id}/approve` |
| POST | `/returns/{id}/reject` |

---

## 11.10 Finance (Receivables)

| Method | Path |
|--------|------|
| GET | `/receivables` |
| GET | `/receivables/{id}` |
| POST | `/receivables/{id}/payments` |
| GET | `/customers/{id}/balance` |
| GET/PUT | `/customers/{id}/credit-limit` |
| GET | `/finance/aging` | Aging report |

---

## 11.11 Documents / Files / Comments

| Method | Path |
|--------|------|
| POST | `/files/presign` | Upload URL |
| POST | `/files/complete` | Confirm upload |
| GET | `/files/{id}` | Meta + download URL |
| DELETE | `/files/{id}` |
| GET/POST | `/documents` |
| GET/PUT/DELETE | `/documents/{id}` |
| POST | `/documents/{id}/files` |
| GET/POST | `/comments` | `entity_type`, `entity_id` |
| DELETE | `/comments/{id}` |

---

## 11.12 Notifications

| Method | Path |
|--------|------|
| GET | `/notifications` |
| POST | `/notifications/{id}/read` |
| POST | `/notifications/read-all` |
| GET | `/notifications/unread-count` |
| POST | `/notifications/test` | Admin test |

---

## 11.13 KPI / Analytics / Dashboard

| Method | Path |
|--------|------|
| GET | `/dashboard/summary` | Role-aware widgets |
| GET | `/kpi` | Definitions |
| GET | `/kpi/snapshots` | Values by period/agent/branch |
| GET | `/analytics/sales` | Sales series |
| GET | `/analytics/visits` | Visit compliance |
| GET | `/analytics/agents/{id}` | Agent performance |
| GET | `/analytics/products` | Product ranking |
| GET | `/analytics/receivables` | AR analytics |
| GET/PUT | `/dashboard/layout` | Personal/role layout |

---

## 11.14 Offline Sync

| Method | Path |
|--------|------|
| POST | `/sync/bootstrap` | Initial snapshot meta |
| GET | `/sync/pull` | Delta since cursor |
| POST | `/sync/push` | Apply client ops |
| GET | `/sync/conflicts` | List conflicts |
| POST | `/sync/conflicts/{id}/resolve` | Resolve |
| GET | `/sync/status` | Device sync status |

Pull query: `cursor`, `types[]`, `limit`  
Push body: `{ device_id, ops: [{ op_id, entity_type, entity_id, op, base_version, payload, client_ts }] }`

---

## 11.15 Audit

| Method | Path |
|--------|------|
| GET | `/audit-logs` | Filter by actor/entity/date |
| GET | `/audit-logs/{id}` | Detail |

---

## 11.16 WebSocket Events

Connect: `/ws/v1?token=...`

| Event | Direction | Payload |
|-------|-----------|---------|
| `notification.created` | S→C | notification |
| `order.updated` | S→C | order id/status |
| `visit.updated` | S→C | visit summary |
| `sync.invalidate` | S→C | entity types |
| `gps.agent.updated` | S→C | agent lat/lng |
| `kpi.tick` | S→C | optional dashboard |
| `ping/pong` | bi | keepalive |

---

## 11.17 Platform Super-Admin (optional namespace)

| Method | Path |
|--------|------|
| GET/POST | `/platform/tenants` |
| GET/PUT | `/platform/tenants/{id}` |
| POST | `/platform/tenants/{id}/suspend` |

---

## 11.18 Error Codes (sample)

| Code | HTTP | Meaning |
|------|------|---------|
| `AUTH_INVALID` | 401 | Bad credentials/token |
| `AUTH_FORBIDDEN` | 403 | RBAC deny |
| `TENANT_NOT_FOUND` | 404 | Unknown tenant |
| `VALIDATION_FAILED` | 422 | Input errors |
| `CONFLICT_VERSION` | 409 | Optimistic lock / sync |
| `CREDIT_LIMIT_EXCEEDED` | 409 | Finance rule |
| `IDEMPOTENCY_REPLAY` | 200/409 | Replay semantics |
| `RATE_LIMITED` | 429 | Too many requests |
