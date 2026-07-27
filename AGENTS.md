# AGENTS.md

## Cursor Cloud specific instructions

### Where the code lives
The default base branch (`main`) contains only `README.md`. The actual product code
(`backend/`, `admin/`, `docs/`, `scripts/`) lives on feature branches — the superset is
`cursor/instagram-media-pipeline-edf0`. Environment work here is based on that branch.

### Services
- **Backend API** — FastAPI (`backend/`), served with uvicorn on port `8000`.
- **Admin panel** — Vite + React (`admin/`), dev server on port `5173`.
- `mobile/` is an empty placeholder (no code).

The app is designed to run with **zero external services**: it uses a SQLite file DB
(auto-created at `backend/data/platform.db`) and an in-process task queue by default.
Postgres and Redis (see `docker-compose.yml`) are optional and only needed for the
`arq` queue backend or a production-like DB; ARQ silently falls back to in-process if
Redis is unavailable.

### Running (standard commands are in `README.md`)
- Backend: `cd backend && source .venv/bin/activate && uvicorn app.main:app --reload --port 8000`.
  The DB schema, a seeded admin user (`admin@example.com` / `admin123`) and the default
  org are created automatically on startup.
- Admin: `cd admin && npm run dev`.

### Non-obvious caveats
- **Vite binds to `localhost` (IPv6 `::1`), not `127.0.0.1`.** `curl http://127.0.0.1:5173`
  returns nothing (exit 7); use `http://localhost:5173` or `http://[::1]:5173`. The
  browser on the Desktop works fine either way.
- The admin talks to the API at `http://127.0.0.1:8000/api` by default
  (override with `VITE_API_BASE`). CORS is preconfigured for ports 5173.
- Plugin credentials accept the literal token `demo:local` to exercise the full
  create/activate/publish flow offline (no real Telegram/Instagram/Meta tokens needed).
- `SECRET_ENCRYPTION_KEY` may be left empty — a deterministic dev key is derived from
  `SECRET_KEY`, so agent secrets survive restarts.

### Test / lint
- Backend tests: `cd backend && pytest -q`.
- Admin lint: `cd admin && npm run lint` (oxlint). There is no configured Python linter.
