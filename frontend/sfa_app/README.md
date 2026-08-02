# SFA Flutter Client (P0)

White-label shell for Android / iOS / Web Admin.

## Features in P0

- Branding bootstrap from `/public/branding`
- Login / session restore / logout
- Dynamic brand theme (Light / Dark / Brand colors)
- Locale placeholders: `ru`, `uz`, `en` (ARB files ready)
- Role-aware dashboard shell
- Admin: Users / Roles / Branding studio pages

## Run

```bash
# Start API first (SQLite local or docker compose)
cd ../../backend && SFA_DATABASE_DSN='sqlite:file:./sfa_dev.db?cache=shared&mode=rwc' go run ./cmd/api

cd ../frontend/sfa_app
flutter pub get
flutter run -d chrome --dart-define=API_BASE_URL=http://localhost:8080/api/v1
```

Demo credentials:

- Tenant: `demo`
- Admin: `admin@demo.local` / `Admin123!`
- Agent: `agent@demo.local` / `Agent123!`
