# Sales Force Automation (SFA)

White Label SaaS platform for sales force automation — multi-tenant, offline-capable, enterprise-ready.

**Status:** Architecture design complete — awaiting confirmation before implementation.

## Architecture Documentation

Full system design (no application code yet):

👉 **[docs/architecture/README.md](docs/architecture/README.md)**

| Topic | Doc |
|-------|-----|
| System architecture | [01](docs/architecture/01-system-architecture.md) |
| Folder structure | [02](docs/architecture/02-folder-structure.md) |
| Backend | [03](docs/architecture/03-backend-architecture.md) |
| Frontend (Flutter) | [04](docs/architecture/04-frontend-architecture.md) |
| Database & indexes | [05](docs/architecture/05-database-design.md) |
| ER / UML / Sequence | [06](docs/architecture/06-diagrams.md) |
| Entities & relations | [07](docs/architecture/07-entities-and-relations.md) |
| API map | [08](docs/architecture/08-api-specification.md) |
| Services, RabbitMQ, Redis, Jobs | [09](docs/architecture/09-infrastructure-map.md) |
| Screens & flows | [10](docs/architecture/10-screens-and-flows.md) |
| Offline sync | [11](docs/architecture/11-offline-sync.md) |
| Security, scale, backup, deploy | [12](docs/architecture/12-security-scale-ops.md) |

## Planned Stack

- **Backend:** Go, Fiber, GORM, PostgreSQL, Redis, RabbitMQ, MinIO
- **Frontend:** Flutter (Android, iOS, Web Admin)
- **Architecture:** Clean Architecture, DDD, CQRS (where needed), RBAC, JWT + Refresh
- **Delivery:** Docker, Kubernetes-ready, White Label SaaS

## Locales & Themes

- Languages: Russian, Uzbek, English
- Themes: Light, Dark, Brand Color (per tenant)
