# 4. Frontend Architecture (Flutter)

## 4.1 Clients

| Client | Platform | Primary users |
|--------|----------|---------------|
| SFA Mobile | Android, iOS | Sales agents, managers (field) |
| SFA Admin | Flutter Web | Tenant admins, managers, finance, warehouse |
| Customer Portal | Flutter Web (restricted routes) | Customer contacts (orders, AR, docs) |

Single codebase with **flavors / role-based navigation shells**.

---

## 4.2 Architectural Style

**Feature-first Clean Architecture** + presentation state management.

Recommended stack:

| Concern | Choice |
|---------|--------|
| State | Riverpod 2.x (or Bloc if team prefers) |
| Routing | go_router |
| DI | Riverpod providers / get_it |
| Local DB | Isar or Drift (SQLite) |
| Secure storage | flutter_secure_storage |
| Network | Dio + interceptors |
| WS | web_socket_channel |
| i18n | flutter_localizations + ARB (`ru`, `uz`, `en`) |
| Theming | Dynamic ThemeData from branding API |
| Maps / GPS | geolocator + google_maps_flutter / mapbox |
| Push | firebase_messaging (+ APNs) |
| Images | cached_network_image + image_picker + camera |
| Charts | fl_chart |
| Offline queue | Local outbox table + Background Fetch / Workmanager |

---

## 4.3 Presentation Shells

```
AppBootstrap
 ├── BrandingLoader (public API / cache)
 ├── AuthGate
 └── RoleShell
      ├── AdminShell (Web-first navigation rail)
      ├── ManagerShell
      ├── AgentShell (mobile-first bottom nav)
      └── CustomerShell
```

UI must feel premium: dense where needed (admin tables), spacious and glanceable for agent field UI. Avoid generic purple SaaS look; brand color is tenant-driven.

### Design system tokens

```
--color-primary          ← tenant brand
--color-primary-contrast
--color-surface
--color-surface-elevated
--color-bg
--color-text
--color-text-muted
--color-success | warning | danger | info
--radius-sm|md|lg
--space-1..8
--font-display           ← expressive (e.g. Manrope / Plus Jakarta / Geologica)
--font-body
--shadow-elev-1
```

Themes: **Light**, **Dark**, **Brand** (primary overridden; surfaces adapt).

Motion (2–3 intentional):
1. Brand splash → home fade/slide
2. Order status transitions
3. Sync indicator pulse / success check

---

## 4.4 Data Flow

```
UI → Notifier/Bloc → UseCase → Repository
                           ├── RemoteDataSource (REST/WS)
                           └── LocalDataSource (Isar)
                                    ▲
                              SyncEngine
```

### Repository policy

- **Online:** write-through local + remote (or remote-first for admin web)
- **Offline (agent):** write local outbox → background sync
- **Reads:** local-first for catalog/customers/routes; invalidate on sync pull

---

## 4.5 Navigation Map (logical)

See [10-screens-and-flows.md](./10-screens-and-flows.md) for full screen inventory.

Admin Web: persistent side nav + top bar (tenant logo, search, notifications, locale, theme).  
Agent Mobile: Home · Customers · Orders · Route · More.

---

## 4.6 i18n

- ARB files: `app_en.arb`, `app_ru.arb`, `app_uz.arb`
- Server error keys mapped to localized strings
- RTL not required (current languages LTR)
- Date/number formats via `intl`

---

## 4.7 White Label Client Behavior

1. Cold start: load cached branding
2. Fetch `/api/v1/public/branding`
3. Apply ThemeData + assets (logo, icons)
4. Persist branding version; refetch when `branding_version` changes (WS/push/header)

Custom domain (web): host → tenant branding automatically.

---

## 4.8 Quality Bar vs SAP / Odoo / Dynamics

| Dimension | Approach |
|-----------|----------|
| Visual hierarchy | One job per screen; no cluttered KPI soup in agent home |
| Speed | Offline-first critical path < 100ms local |
| Density control | Admin: compact tables; Agent: large tap targets |
| Empty states | Actionable, branded |
| Feedback | Inline validation, toast sparingly, clear sync states |
| Accessibility | Contrast WCAG AA, scalable type |

---

## 4.9 Testing

- Unit: use cases, sync merge
- Widget: critical forms (login, order)
- Golden: theme light/dark/brand
- Integration: offline order → sync
