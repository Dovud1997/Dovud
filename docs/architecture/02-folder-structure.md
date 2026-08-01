# 2. Folder Structure (Monorepo)

```
sfa/
├── README.md
├── LICENSE
├── .gitignore
├── .editorconfig
├── docker-compose.yml                 # Local: Postgres, Redis, RabbitMQ, MinIO, API, Workers
├── Makefile
├── docs/
│   ├── architecture/                  # This design pack
│   ├── adr/                           # Architecture Decision Records
│   └── api/                           # Exported OpenAPI (generated)
│
├── deploy/
│   ├── docker/
│   │   ├── Dockerfile.api
│   │   ├── Dockerfile.worker
│   │   └── Dockerfile.scheduler
│   ├── helm/
│   │   └── sfa/
│   │       ├── Chart.yaml
│   │       ├── values.yaml
│   │       ├── values-staging.yaml
│   │       ├── values-prod.yaml
│   │       └── templates/
│   ├── k8s/
│   │   ├── namespaces/
│   │   ├── ingress/
│   │   └── network-policies/
│   └── terraform/                     # Optional infra (VPC, buckets, DNS)
│
├── scripts/
│   ├── migrate.sh
│   ├── seed.sh
│   └── gen-swagger.sh
│
├── backend/
│   ├── cmd/
│   │   ├── api/main.go                # HTTP + WS entrypoint
│   │   ├── worker/main.go             # RabbitMQ consumers
│   │   ├── scheduler/main.go          # Cron / periodic jobs
│   │   └── migrate/main.go
│   │
│   ├── api/
│   │   └── openapi/                   # Spec source / generated
│   │
│   ├── internal/
│   │   ├── platform/                  # Shared kernel (no business rules)
│   │   │   ├── config/
│   │   │   ├── logger/
│   │   │   ├── database/
│   │   │   ├── redis/
│   │   │   ├── rabbitmq/
│   │   │   ├── minio/
│   │   │   ├── httpx/                 # Fiber helpers, middleware
│   │   │   ├── auth/                  # JWT helpers
│   │   │   ├── i18n/
│   │   │   ├── errors/
│   │   │   ├── pagination/
│   │   │   ├── outbox/
│   │   │   └── telemetry/
│   │   │
│   │   ├── shared/                    # Shared kernel types (IDs, Money, Geo)
│   │   │   ├── domain/
│   │   │   └── dto/
│   │   │
│   │   ├── modules/                   # Bounded contexts
│   │   │   ├── identity/
│   │   │   │   ├── domain/
│   │   │   │   │   ├── entity/
│   │   │   │   │   ├── valueobject/
│   │   │   │   │   ├── repository/    # Interfaces
│   │   │   │   │   ├── service/       # Domain services
│   │   │   │   │   └── event/
│   │   │   │   ├── application/
│   │   │   │   │   ├── command/       # CQRS commands
│   │   │   │   │   ├── query/         # CQRS queries
│   │   │   │   │   └── dto/
│   │   │   │   ├── infrastructure/
│   │   │   │   │   ├── persistence/   # GORM repos
│   │   │   │   │   └── messaging/
│   │   │   │   └── interfaces/
│   │   │   │       ├── http/          # Handlers, routes
│   │   │   │       └── ws/
│   │   │   │
│   │   │   ├── tenant/                # Same layer pattern
│   │   │   ├── organization/
│   │   │   ├── catalog/
│   │   │   ├── crm/
│   │   │   ├── fieldforce/
│   │   │   ├── orders/
│   │   │   ├── finance/
│   │   │   ├── documents/
│   │   │   ├── notifications/
│   │   │   ├── analytics/
│   │   │   ├── sync/
│   │   │   └── audit/
│   │   │
│   │   ├── gateway/                   # Route aggregation, middleware stack
│   │   └── app/                       # DI composition root (fx)
│   │
│   ├── migrations/                    # SQL migrations
│   ├── seeds/
│   ├── configs/
│   │   ├── config.yaml
│   │   └── config.example.yaml
│   ├── go.mod
│   ├── go.sum
│   └── Makefile
│
├── frontend/
│   └── sfa_app/                       # Flutter monorepo-style app
│       ├── pubspec.yaml
│       ├── analysis_options.yaml
│       ├── l10n.yaml
│       ├── assets/
│       │   ├── fonts/
│       │   ├── icons/
│       │   ├── images/
│       │   └── branding/              # Fallback defaults
│       ├── lib/
│       │   ├── main.dart
│       │   ├── main_admin.dart        # Optional entry flavors
│       │   ├── bootstrap.dart
│       │   ├── app.dart
│       │   │
│       │   ├── core/                  # Shared kernel
│       │   │   ├── config/
│       │   │   ├── di/
│       │   │   ├── error/
│       │   │   ├── network/
│       │   │   ├── storage/           # Secure storage, Hive/Isar
│       │   │   ├── sync/
│       │   │   ├── theme/
│       │   │   ├── l10n/
│       │   │   ├── router/
│       │   │   ├── widgets/
│       │   │   └── utils/
│       │   │
│       │   ├── features/              # Feature modules (Clean Arch)
│       │   │   ├── auth/
│       │   │   │   ├── data/
│       │   │   │   ├── domain/
│       │   │   │   └── presentation/
│       │   │   ├── branding/
│       │   │   ├── dashboard/
│       │   │   ├── organization/
│       │   │   ├── catalog/
│       │   │   ├── crm/
│       │   │   ├── routes/
│       │   │   ├── visits/
│       │   │   ├── orders/
│       │   │   ├── returns/
│       │   │   ├── receivables/
│       │   │   ├── gps/
│       │   │   ├── documents/
│       │   │   ├── notifications/
│       │   │   ├── kpi/
│       │   │   ├── analytics/
│       │   │   ├── settings/
│       │   │   └── sync/
│       │   │
│       │   └── shared/
│       │       ├── models/
│       │       └── widgets/
│       │
│       ├── test/
│       ├── integration_test/
│       └── web/
│
└── packages/                          # Optional shared Dart packages later
    └── sfa_design_system/
```

---

## Module Internal Pattern (Backend)

Each module follows identical layering:

```
module/
├── domain/           # Entities, VOs, repository ports, domain events
├── application/      # Use cases (commands/queries), DTOs, mappers
├── infrastructure/   # GORM, Redis adapters, publishers
└── interfaces/       # HTTP/WS controllers, request validators
```

**Dependency rule:** `interfaces → application → domain ← infrastructure`

---

## Module Internal Pattern (Flutter Feature)

```
feature/
├── data/
│   ├── datasources/   # remote + local
│   ├── models/
│   └── repositories/  # implementations
├── domain/
│   ├── entities/
│   ├── repositories/  # contracts
│   └── usecases/
└── presentation/
    ├── bloc/ or cubit/ or riverpod
    ├── pages/
    └── widgets/
```
