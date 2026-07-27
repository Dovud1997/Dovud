#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "==> Backend deps"
cd "$ROOT/backend"
python3 -m pip install -q -r requirements.txt

if [[ ! -f .env ]]; then
  cp .env.example .env
  echo "Created backend/.env — add CONTROL_BOT_TOKEN locally (do not commit)."
fi

echo "==> Starting API on :8000"
export PATH="${HOME}/.local/bin:${PATH}"
uvicorn app.main:app --reload --host 127.0.0.1 --port 8000
