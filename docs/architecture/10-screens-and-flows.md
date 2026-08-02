# 18–22. Flutter Screen Map & User Flows

## 18. Full Screen Map

### A. Shared / Auth / Shell

| ID | Screen | Platforms |
|----|--------|-----------|
| S01 | Splash / Branding bootstrap | All |
| S02 | Tenant select (optional multi-tenant app) | Mobile |
| S03 | Login | All |
| S04 | Forgot password | All |
| S05 | Reset password | All |
| S06 | Force change password | All |
| S07 | Main shell (role-based) | All |
| S08 | Notifications list | All |
| S09 | Notification detail | All |
| S10 | Profile & preferences (locale, theme) | All |
| S11 | Security (sessions/devices) | All |
| S12 | About / App info | All |
| S13 | Sync center (status, conflicts) | Mobile |
| S14 | Conflict resolve | Mobile |

### B. Admin / Manager (Web-first, also tablet)

| ID | Screen |
|----|--------|
| A01 | Dashboard |
| A02 | Analytics hub |
| A03 | KPI browser |
| A04 | Users list |
| A05 | User create/edit |
| A06 | Roles & permissions |
| A07 | Companies |
| A08 | Branches |
| A09 | Warehouses |
| A10 | Warehouse stock |
| A11 | Manufacturers |
| A12 | Categories tree |
| A13 | Products list |
| A14 | Product detail/edit |
| A15 | Price lists |
| A16 | Price list editor |
| A17 | Promotions list |
| A18 | Promotion editor |
| A19 | Customers list |
| A20 | Customer 360 (contacts, AR, visits, orders) |
| A21 | Contacts |
| A22 | Agents list |
| A23 | Agent detail (KPI, GPS, routes) |
| A24 | Routes planner |
| A25 | Route detail |
| A26 | Visits monitor |
| A27 | Visit detail (photos, comments, map) |
| A28 | Live GPS map |
| A29 | GPS track history |
| A30 | Orders list |
| A31 | Order detail / approve |
| A32 | Returns list |
| A33 | Return detail |
| A34 | Receivables list |
| A35 | Receivable detail / payment |
| A36 | Aging report |
| A37 | Documents |
| A38 | Document detail |
| A39 | Files browser |
| A40 | Audit logs |
| A41 | Tenant settings |
| A42 | White Label branding studio |
| A43 | Domains |
| A44 | Providers (Email/SMS/Push) |
| A45 | Feature modules / settings |
| A46 | Import/Export jobs |

### C. Agent Mobile

| ID | Screen |
|----|--------|
| G01 | Agent home (today route + sync status) |
| G02 | Today route map/list |
| G03 | Stop / Customer quick view |
| G04 | Visit check-in |
| G05 | Active visit workspace |
| G06 | Photo report capture |
| G07 | Visit comment |
| G08 | Visit check-out |
| G09 | Customers nearby / search |
| G10 | Customer detail (offline) |
| G11 | Create/edit order |
| G12 | Product picker |
| G13 | Cart / promo quote |
| G14 | Order success |
| G15 | Orders history |
| G16 | Order detail |
| G17 | Create return |
| G18 | Receivable / collect payment |
| G19 | Documents for customer |
| G20 | My KPI |
| G21 | GPS permission / accuracy helper |

### D. Customer Portal

| ID | Screen |
|----|--------|
| C01 | Customer login |
| C02 | My orders |
| C03 | Order detail |
| C04 | Create order request (optional) |
| C05 | Invoices / receivables |
| C06 | Documents |
| C07 | Profile / contacts |
| C08 | Notifications |

---

## 19. User Flow (Generic Authenticated)

```
Splash → Load branding → Has tokens?
  ├─ No → Login → Validate → Shell by role
  └─ Yes → Refresh if needed → Bootstrap sync/profile → Shell by role

In Shell:
  Dashboard/Home → Module → Detail → Action → Feedback
  Notifications → Deep link entity
  Profile → Locale / Theme / Password
  Logout → Revoke → Login
```

---

## 20. Admin Flow

```
Login (tenant_admin / sales_manager / finance / warehouse)
  → Dashboard (KPIs, alerts)
  → Setup path (first-time):
       Branding → Branches → Warehouses → Roles/Users → Catalog → Customers → Agents
  → Daily ops:
       Routes planner → Assign agents
       Orders approval queue → Confirm/Cancel
       Visits monitor + Live GPS
       Receivables / overdue follow-up
       Audit review (as needed)
  → White Label:
       Branding studio → Upload logo → Colors → Preview → Save
       Domains → Verify DNS
       Providers → Test email/SMS/push
```

**Happy path — create product & price**

```
Products → Create → Set category/manufacturer → Save
→ Price lists → Add price → Effective dates → Save
→ (event) agents sync pull gets product/price
```

**Happy path — approve order**

```
Orders (Submitted) → Open → Review lines/credit → Confirm
→ Status history + push to agent + analytics project
```

---

## 21. Agent Flow

```
Login → Background sync bootstrap
→ Home: today's route + sync badge
→ Open stop → Check-in (GPS)
→ Visit workspace:
     ├─ Capture photos
     ├─ Add comment
     ├─ Create order (catalog offline)
     ├─ View AR / take payment note
     └─ Check-out
→ Continue next stop
→ End of day: Sync Center ensures outbox empty
```

**Offline path**

```
No network → All critical writes to local DB + outbox
→ Queue: visits, orders, photos (local files), GPS, comments
→ Network up → BackgroundSync push/pull
→ Conflicts → Sync Center → Resolve
```

**Order path**

```
Customer → New order → Pick products → See promo quote
→ Submit locally → Sync → Server ack → Status updates via pull/WS
```

---

## 22. Customer Flow

```
Portal login (customer_portal role)
→ Home summary (open orders, balance)
→ Orders list → Detail / PDF document
→ Receivables → Payment instructions / history
→ Documents download
→ Optional: order request → Pending agent/manager processing
→ Notifications of shipment/status
```

---

## Navigation IA (Agent Mobile)

```
Tab1 Home      → G01
Tab2 Customers → G09/G10
Tab3 Orders    → G15/G11
Tab4 Route     → G02
Tab5 More      → Profile, KPI, Sync, Notifications, Settings
```

## Navigation IA (Admin Web)

```
Dashboard
Catalog ▸ Products, Categories, Manufacturers, Prices, Promotions
CRM ▸ Customers, Contacts
Field ▸ Agents, Routes, Visits, GPS
Sales ▸ Orders, Returns
Finance ▸ Receivables, Aging
Content ▸ Documents, Files
Insights ▸ Analytics, KPI
Admin ▸ Users, Roles, Org, Branding, Domains, Providers, Audit, Settings
```
