# Sales Force Automation (SFA) — Architecture Documentation

**Product:** White Label SaaS for Sales Force Automation  
**Status:** Design Phase — awaiting confirmation before implementation  
**Version:** 1.0.0-design  
**Languages:** Russian (ru), Uzbek (uz), English (en)  
**Clients:** Flutter (Android / iOS / Web Admin)

---

## Documents

| # | Document | Description |
|---|----------|-------------|
| 01 | [System Architecture](./01-system-architecture.md) | High-level architecture, principles, bounded contexts |
| 02 | [Folder Structure](./02-folder-structure.md) | Monorepo layout (backend + frontend) |
| 03 | [Backend Architecture](./03-backend-architecture.md) | Clean Architecture, DDD, CQRS, DI, modules |
| 04 | [Frontend Architecture](./04-frontend-architecture.md) | Flutter Clean Architecture, theming, i18n |
| 05 | [Database Design](./05-database-design.md) | Tables, indexes, tenancy, soft-delete |
| 06 | [Diagrams](./06-diagrams.md) | ER, UML, Sequence diagrams (Mermaid) |
| 07 | [Entities & Relations](./07-entities-and-relations.md) | All domain entities and relationships |
| 08 | [API Specification](./08-api-specification.md) | Full REST + WebSocket map, versioning |
| 09 | [Infrastructure Map](./09-infrastructure-map.md) | Services, RabbitMQ, Redis, Jobs |
| 10 | [Flutter Screens & Flows](./10-screens-and-flows.md) | Screen map, User/Admin/Agent/Customer flows |
| 11 | [Offline Sync](./11-offline-sync.md) | Sync protocol, conflict resolution |
| 12 | [Security, Scale, Ops](./12-security-scale-ops.md) | Security, scaling, backup, deployment |

---

## Design Decisions (Summary)

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Deployment shape | Modular Monolith → Microservices-ready | Faster MVP, clear module boundaries for future split |
| Multi-tenancy | Shared DB + `tenant_id` + RLS-ready | Cost-efficient SaaS; row-level isolation |
| White Label | Tenant branding + custom domain + provider configs | No code change per customer |
| Auth | JWT access + Refresh Token (Redis blacklist/rotation) | Mobile-friendly, revocable |
| Storage | MinIO (S3-compatible) | Self-hosted, K8s-friendly |
| Messaging | RabbitMQ | Reliable async, retries, DLQ |
| Cache | Redis | Sessions, tokens, hot reads, rate limits |
| ORM | GORM | Productivity + migrations; repositories abstract it |
| API | Fiber REST v1 + WebSocket | Performance, Go ecosystem |
| Mobile/Web | Flutter single codebase | Android, iOS, Admin Web |

---

## Confirmation Gate

**No application code will be written until this architecture is explicitly approved.**

Please review all documents and confirm or request changes.
