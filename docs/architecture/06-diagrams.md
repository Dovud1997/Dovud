# 6–8. Diagrams (ER / UML / Sequence)

## 6. ER Diagram (Logical)

```mermaid
erDiagram
  TENANTS ||--o{ USERS : has
  TENANTS ||--|| TENANT_BRANDING : has
  TENANTS ||--o{ TENANT_DOMAINS : has
  TENANTS ||--o{ TENANT_PROVIDERS : has
  TENANTS ||--o{ BRANCHES : has
  TENANTS ||--o{ WAREHOUSES : has
  TENANTS ||--o{ PRODUCTS : has
  TENANTS ||--o{ CUSTOMERS : has
  TENANTS ||--o{ ORDERS : has

  USERS ||--o{ USER_ROLES : has
  ROLES ||--o{ USER_ROLES : grants
  ROLES ||--o{ ROLE_PERMISSIONS : has
  PERMISSIONS ||--o{ ROLE_PERMISSIONS : included

  COMPANIES ||--o{ BRANCHES : contains
  BRANCHES ||--o{ WAREHOUSES : has
  BRANCHES ||--o{ CUSTOMERS : serves
  BRANCHES ||--o{ SALES_AGENTS : employs

  MANUFACTURERS ||--o{ PRODUCTS : makes
  CATEGORIES ||--o{ CATEGORIES : parent
  CATEGORIES ||--o{ PRODUCTS : classifies
  PRICE_LISTS ||--o{ PRODUCT_PRICES : contains
  PRODUCTS ||--o{ PRODUCT_PRICES : priced
  PROMOTIONS ||--o{ PROMOTION_ITEMS : includes

  CUSTOMERS ||--o{ CUSTOMER_CONTACTS : has
  CUSTOMERS ||--o{ CUSTOMER_ADDRESSES : has
  CUSTOMERS ||--o{ ORDERS : places
  CUSTOMERS ||--o{ RECEIVABLES : owes
  CUSTOMERS ||--o{ VISITS : receives

  USERS ||--o| SALES_AGENTS : profile
  SALES_AGENTS ||--o{ ROUTES : runs
  ROUTES ||--o{ ROUTE_STOPS : contains
  CUSTOMERS ||--o{ ROUTE_STOPS : stop
  SALES_AGENTS ||--o{ VISITS : performs
  ROUTE_STOPS ||--o| VISITS : results_in
  VISITS ||--o{ VISIT_PHOTOS : has
  VISITS ||--o{ VISIT_COMMENTS : has
  VISITS ||--o{ GPS_TRACKS : tracks
  VISITS ||--o{ ORDERS : may_create

  ORDERS ||--o{ ORDER_LINES : contains
  PRODUCTS ||--o{ ORDER_LINES : item
  ORDERS ||--o{ RETURNS : may_return
  RETURNS ||--o{ RETURN_LINES : contains
  WAREHOUSES ||--o{ WAREHOUSE_STOCKS : holds
  PRODUCTS ||--o{ WAREHOUSE_STOCKS : stocked

  RECEIVABLES ||--o{ RECEIVABLE_PAYMENTS : paid_by
  DOCUMENTS ||--o{ DOCUMENT_FILES : has
  FILES ||--o{ DOCUMENT_FILES : linked
  USERS ||--o{ NOTIFICATIONS : receives
  USERS ||--o{ AUDIT_LOGS : acts
```

---

## 7. UML — Component / Package

```mermaid
flowchart TB
  subgraph Clients
    Mobile[Flutter Mobile]
    Admin[Flutter Web Admin]
    Portal[Customer Portal]
  end

  subgraph Gateway
    API[Fiber API Gateway]
    WS[WebSocket Hub]
  end

  subgraph Modules
    ID[Identity]
    TN[Tenant/WhiteLabel]
    ORG[Organization]
    CAT[Catalog]
    CRM[CRM]
    FF[FieldForce]
    ORD[Orders]
    FIN[Finance]
    DOC[Documents]
    NTF[Notifications]
    AN[Analytics]
    SYN[Sync]
    AUD[Audit]
  end

  subgraph Infra
    PG[(PostgreSQL)]
    RD[(Redis)]
    RQ[[RabbitMQ]]
    S3[(MinIO)]
    WRK[Workers]
  end

  Mobile --> API
  Admin --> API
  Portal --> API
  Mobile --> WS
  Admin --> WS
  API --> ID & TN & ORG & CAT & CRM & FF & ORD & FIN & DOC & NTF & AN & SYN & AUD
  WS --> SYN & NTF
  Modules --> PG & RD
  Modules --> RQ
  DOC --> S3
  RQ --> WRK
  WRK --> PG & RD & S3 & NTF
```

### UML — Class (Order Aggregate excerpt)

```mermaid
classDiagram
  class Order {
    +UUID id
    +UUID tenantId
    +String number
    +OrderStatus status
    +Money grandTotal
    +long version
    +addLine()
    +applyPromotion()
    +submit()
    +cancel()
  }
  class OrderLine {
    +UUID productId
    +Decimal qty
    +Money unitPrice
    +Money lineTotal
  }
  class OrderStatus {
    <<enumeration>>
    DRAFT
    SUBMITTED
    CONFIRMED
    PICKING
    SHIPPED
    DELIVERED
    CANCELLED
  }
  Order "1" *-- "1..*" OrderLine
  Order --> OrderStatus
```

### UML — Use Case (Agent)

```mermaid
flowchart LR
  Agent((Sales Agent))
  Agent --> Login
  Agent --> ViewRoute
  Agent --> CheckInVisit
  Agent --> CreateOrder
  Agent --> CapturePhoto
  Agent --> CollectPayment
  Agent --> SyncOffline
  Manager((Manager))
  Manager --> AssignRoute
  Manager --> ApproveOrder
  Manager --> ViewKPI
  Admin((Tenant Admin))
  Admin --> ManageUsers
  Admin --> ManageCatalog
  Admin --> ConfigureBranding
```

---

## 8. Sequence Diagrams

### 8.1 Login + Token Refresh

```mermaid
sequenceDiagram
  participant C as Client
  participant API as API Gateway
  participant ID as Identity
  participant RD as Redis
  participant PG as PostgreSQL

  C->>API: POST /api/v1/auth/login
  API->>ID: Authenticate(email, password, device)
  ID->>PG: Load user + roles + permissions
  ID->>RD: Store refresh token hash / session
  ID-->>API: access_jwt + refresh_token
  API-->>C: 200 tokens + user profile

  C->>API: POST /api/v1/auth/refresh
  API->>ID: RotateRefresh(token)
  ID->>RD: Validate + rotate + blacklist old
  ID-->>API: new tokens
  API-->>C: 200
```

### 8.2 Create Order (Online)

```mermaid
sequenceDiagram
  participant C as Agent App
  participant API as Orders API
  participant CAT as Catalog
  participant FIN as Finance
  participant PG as PostgreSQL
  participant OB as Outbox
  participant RQ as RabbitMQ
  participant NTF as Notify Worker

  C->>API: POST /orders (Idempotency-Key)
  API->>CAT: Validate products/prices/promos
  API->>FIN: Check credit limit
  API->>PG: TX create order + lines + outbox
  PG-->>API: OK
  API-->>C: 201 Order
  OB->>RQ: order.submitted
  RQ->>NTF: Send push to manager
```

### 8.3 Offline Order + Sync Push

```mermaid
sequenceDiagram
  participant C as Agent App (Offline)
  participant L as Local DB
  participant S as Sync API
  participant ORD as Orders Module
  participant PG as PostgreSQL
  participant WS as WebSocket

  C->>L: Create local order (client_request_id)
  C->>L: Enqueue outbox op
  Note over C,L: Network restored
  C->>S: POST /sync/push [ops...]
  S->>ORD: Apply op (idempotent)
  alt version conflict
    ORD-->>S: Conflict
    S-->>C: conflict payload
    C->>L: Mark conflict / merge UI
  else success
    ORD->>PG: Persist + bump version
    S-->>C: ack + server entities
    C->>L: Update local + cursor
    S->>WS: sync hint to managers
  end
```

### 8.4 Visit Check-in with GPS + Photo

```mermaid
sequenceDiagram
  participant C as Agent App
  participant API as FieldForce API
  participant DOC as Documents
  participant M as MinIO
  participant PG as PostgreSQL
  participant RQ as RabbitMQ

  C->>API: POST /visits/check-in {lat,lng,customer_id}
  API->>PG: Create visit + gps_event
  API-->>C: visit_id
  C->>DOC: POST /files/presign
  DOC-->>C: upload URL
  C->>M: PUT photo
  C->>API: POST /visits/{id}/photos
  API->>PG: Save visit_photo
  API->>RQ: visit.photo_uploaded
```

### 8.5 White Label Branding Resolve

```mermaid
sequenceDiagram
  participant C as Client
  participant API as Tenant API
  participant RD as Redis
  participant PG as PostgreSQL

  C->>API: GET /public/branding (Host or tenant code)
  API->>RD: GET branding:{tenant}
  alt cache hit
    RD-->>API: branding JSON
  else miss
    API->>PG: Load tenant_branding + domains
    API->>RD: SET cache TTL
  end
  API-->>C: branding + version
  C->>C: Apply theme/assets
```
