# 9–10. Entities & Relationships

All business entities are **independent aggregates** with their own identity and lifecycle. Cross-entity links use IDs; cascading business rules live in application services / domain events.

---

## 9. Entity Catalog

### Tenant & Platform

| Entity | Description | Key attributes |
|--------|-------------|----------------|
| Tenant | SaaS customer | code, name, status, timezone, currency |
| TenantBranding | White-label look | app_name, colors, logo, icons, version |
| TenantDomain | Host mapping | host, is_primary, ssl_status |
| TenantProvider | Channel credentials | type(smtp/sms/push), config_enc |
| TenantSetting | Feature toggles | key, value_json |

### Identity

| Entity | Description |
|--------|-------------|
| User | Person account |
| Role | Named permission set |
| Permission | `resource:action` |
| UserRole | Assignment (+ optional branch_id scope) |
| RefreshToken | Long-lived session |
| UserDevice | Device + push token |
| LoginAttempt | Security telemetry |

### Organization

| Entity | Description |
|--------|-------------|
| Company | Legal entity inside tenant |
| Branch | Sales/ops unit |
| Warehouse | Stock location |
| WarehouseStock | Qty on hand / reserved |

### Catalog

| Entity | Description |
|--------|-------------|
| Manufacturer | Brand/maker |
| Category | Tree category |
| Product | Sellable SKU |
| ProductImage | Media link |
| PriceList | Named price book |
| ProductPrice | Price row |
| Promotion | Campaign |
| PromotionItem | Scoped products/categories |
| PromotionRule | Discount logic |

### CRM

| Entity | Description |
|--------|-------------|
| Customer | Outlet / buyer |
| CustomerContact | Person |
| CustomerAddress | Address + geo |
| CustomerCategory | Segment tag |

### Field Force

| Entity | Description |
|--------|-------------|
| SalesAgent | Field rep profile |
| AgentTerritory | Coverage |
| Route | Planned journey |
| RouteStop | Customer stop |
| Visit | Actual visit |
| VisitPhoto | Photo report |
| VisitComment | Comment |
| GpsTrack | Breadcrumb point |
| GpsEvent | Check-in/out, geofence |

### Commerce

| Entity | Description |
|--------|-------------|
| Order | Sales order |
| OrderLine | Line |
| OrderStatusHistory | Status trail |
| Return | Return header |
| ReturnLine | Return line |

### Finance

| Entity | Description |
|--------|-------------|
| Receivable | AR open item |
| ReceivablePayment | Payment allocation |
| CreditLimit | Limit policy |

### Content & Comms

| Entity | Description |
|--------|-------------|
| Document | Business document |
| File | Object storage meta |
| Comment | Polymorphic comment |
| Notification | User notification |
| NotificationDelivery | Channel attempt |

### Analytics & Sync

| Entity | Description |
|--------|-------------|
| KpiDefinition | Metric definition |
| KpiSnapshot | Computed value |
| DashboardLayout | Widgets layout |
| SyncDevice | Device cursor |
| SyncConflict | Conflict record |
| AuditLog | Immutable audit |
| OutboxEvent | Integration event |

---

## 10. Relationships Matrix

Legend: `1` one, `*` many, `0..1` optional one.

| From | To | Cardinality | Nature |
|------|----|-------------|--------|
| Tenant | User | 1:* | Ownership |
| Tenant | Branch | 1:* | Ownership |
| Tenant | Product | 1:* | Ownership |
| Tenant | Customer | 1:* | Ownership |
| Tenant | TenantBranding | 1:1 | Ownership |
| Tenant | TenantDomain | 1:* | Ownership |
| Tenant | TenantProvider | 1:* | Ownership |
| User | Role | *:* | via UserRole |
| Role | Permission | *:* | via RolePermission |
| User | UserDevice | 1:* | Ownership |
| User | SalesAgent | 1:0..1 | Profile |
| Company | Branch | 1:* | Structure |
| Branch | Warehouse | 1:* | Structure |
| Branch | SalesAgent | 1:* | Assignment |
| Branch | Customer | 1:* | Service area (default) |
| Category | Category | 1:* | Tree parent |
| Manufacturer | Product | 1:* | Classification |
| Category | Product | 1:* | Classification |
| PriceList | ProductPrice | 1:* | Composition |
| Product | ProductPrice | 1:* | Pricing |
| Product | ProductImage | 1:* | Media |
| Promotion | PromotionItem | 1:* | Composition |
| Promotion | PromotionRule | 1:* | Composition |
| Customer | CustomerContact | 1:* | Composition |
| Customer | CustomerAddress | 1:* | Composition |
| Customer | Order | 1:* | Reference |
| Customer | Visit | 1:* | Reference |
| Customer | Receivable | 1:* | Reference |
| Customer | RouteStop | 1:* | Reference |
| SalesAgent | Route | 1:* | Ownership |
| Route | RouteStop | 1:* | Composition |
| RouteStop | Visit | 1:0..1 | Outcome |
| SalesAgent | Visit | 1:* | Performance |
| Visit | VisitPhoto | 1:* | Composition |
| Visit | VisitComment | 1:* | Composition |
| Visit | GpsTrack | 1:* | Telemetry |
| Visit | Order | 1:* | Optional link |
| SalesAgent | Order | 1:* | Created by |
| Warehouse | Order | 1:* | Fulfill from |
| Order | OrderLine | 1:* | Composition |
| Product | OrderLine | 1:* | Reference |
| Order | Return | 1:* | Optional |
| Return | ReturnLine | 1:* | Composition |
| Warehouse | WarehouseStock | 1:* | Composition |
| Product | WarehouseStock | 1:* | Reference |
| Receivable | ReceivablePayment | 1:* | Composition |
| Document | File | *:* | via DocumentFile |
| Entity (poly) | Comment | 1:* | Polymorphic |
| Entity (poly) | File | 1:* | via EntityFile |
| User | Notification | 1:* | Recipient |
| User | AuditLog | 1:* | Actor |
| User | SyncDevice | 1:* | Devices |

### Polymorphic targets

`Comment` / `EntityFile` may attach to: `Visit`, `Order`, `Return`, `Customer`, `Document`, `Product` (configurable allow-list).

### Independence rules

1. Deleting (soft) a Product does not cascade-delete historical OrderLines.
2. Deactivating an Agent does not delete Routes/Visits/Orders.
3. Customer merge is an explicit domain operation (rewrites references + audit).
4. Branding/Provider changes never require redeploy.
5. Sync versions are per-entity; no global lock across aggregates.
