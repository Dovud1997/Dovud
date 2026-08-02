# 23. Offline Sync Mechanism

## Goals

- Agents work full critical path without network: customers, catalog (cached), prices, routes, visits, orders, returns, photos, comments, GPS
- Guaranteed eventual consistency
- Idempotent, ordered-enough delivery
- Clear conflict UX
- Battery-friendly background sync

---

## Syncable Entity Set

| Entity | Pull | Push | Conflict strategy |
|--------|------|------|-------------------|
| Products / Categories / Manufacturers | ✓ | ✗ (admin online) | Server wins |
| PriceLists / ProductPrices / Promotions | ✓ | ✗ | Server wins |
| Customers / Contacts / Addresses | ✓ | ✓ (limited fields) | Field-level merge / server wins configurable |
| Routes / RouteStops | ✓ | status updates | Server plan wins; status merge |
| Visits | ✓ | ✓ | Last-write-wins on notes; append photos/comments |
| VisitPhotos / Comments | ✓ | ✓ | Append-only |
| Orders / OrderLines | ✓ | ✓ create/update draft | `client_request_id` idempotency; version on edit |
| Returns | ✓ | ✓ | Same as orders |
| Receivables | ✓ | payments (optional) | Server balance authoritative |
| GPS points | ✓ (own) | ✓ | Append-only |
| Notifications | ✓ | read receipts | Server wins |
| Documents meta | ✓ | ✓ upload | Append-only files |

---

## Client Local Store

- DB: Isar/Drift
- File cache: photos awaiting upload
- Tables: `local_entities`, `outbox_ops`, `sync_cursors`, `conflict_records`, `file_uploads`

### Outbox operation

```json
{
  "op_id": "uuid",
  "entity_type": "order",
  "entity_id": "uuid",
  "op": "create|update|delete|custom",
  "base_version": 3,
  "payload": {},
  "client_ts": "ISO-8601",
  "status": "pending|sending|acked|failed|conflict"
}
```

---

## Protocol

### 1) Bootstrap

`POST /sync/bootstrap`

Returns: schema/version, branding version, initial cursors, feature flags, user scope (branch/agent).

### 2) Pull (delta)

`GET /sync/pull?cursor=...&types=product,customer,order&limit=200`

Response:

```json
{
  "next_cursor": "...",
  "has_more": true,
  "changes": [
    {
      "entity_type": "product",
      "entity_id": "...",
      "version": 12,
      "deleted": false,
      "updated_at": "...",
      "data": {}
    }
  ],
  "branding_version": 4
}
```

Server change feed sourced from `updated_at+id` indexes and/or `sync_change_log`.

### 3) Push

`POST /sync/push`

```json
{
  "device_id": "...",
  "ops": [ /* outbox ops */ ]
}
```

Per-op results: `acked | conflict | rejected` with server entity snapshot.

### 4) File upload alignment

1. Presign (or offline defer)
2. Upload to MinIO when online
3. Push metadata op referencing `file_id` / checksum
4. VisitPhoto links to file

Photos can sync after order/visit metadata (server accepts pending media state).

---

## Conflict Resolution

### Detection

- Client `base_version` < server `version` on update
- Unique business key clash without matching `client_request_id`

### Strategies

| Case | Strategy |
|------|----------|
| Order create retry | Idempotent via `client_request_id` |
| Order draft concurrent edit | 409 conflict → UI merge or take server |
| Visit notes | Prefer latest `client_ts` if within policy; else manual |
| Append-only (photos, GPS, comments) | No conflict; unique `op_id` |
| Master data | Server wins silently on pull |

Conflicts stored in `sync_conflicts` + client Sync Center (S13/S14).

---

## Ordering & Locks

- Device push serialized via Redis lock `sync:lock:{deviceId}`
- Ops in a batch applied in client order
- Pull after push recommended (client loop: push → pull)

---

## Background Sync

```
App foreground: sync every N seconds + on connectivity regain
App background: OS task every ~15m (best effort)
Triggers: order saved, visit checkout, photo captured, manual "Sync now"
```

Indicators: synced / syncing / offline pending(count) / conflict(count).

---

## Security on Sync

- JWT required; scope limited to agent’s branch/customers/routes
- Payload size limits; batch max ops
- Virus/mime checks on uploads
- Audit: sync sessions and conflict resolutions

---

## Failure Handling

| Failure | Behavior |
|---------|----------|
| Network drop mid-batch | Retry same `op_id`s; server idempotent |
| 401 | Refresh token → retry; else re-login |
| 409 conflict | Mark op conflict; continue others |
| 422 validation | Mark failed; user fix |
| 429/5xx | Exponential backoff |

---

## Versioning

- Entity `version` increments on each server mutation
- API header `X-Min-App-Version` for forced upgrades when sync protocol changes
- Sync protocol version in bootstrap (`sync_protocol: 1`)
