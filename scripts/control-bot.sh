#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT/backend"

if [[ -z "${CONTROL_BOT_TOKEN:-}" ]]; then
  if [[ -f .env ]]; then
    # shellcheck disable=SC1091
    set -a
    source .env
    set +a
  fi
fi

if [[ -z "${CONTROL_BOT_TOKEN:-}" ]]; then
  echo "Set CONTROL_BOT_TOKEN in backend/.env or environment (never commit the token)."
  exit 1
fi

export PLATFORM_API_BASE="${PLATFORM_API_BASE:-http://127.0.0.1:8000/api}"
export PATH="${HOME}/.local/bin:${PATH}"
python3 -m bot.control_bot
