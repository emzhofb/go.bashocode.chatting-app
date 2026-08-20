#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

COMPOSE=(docker compose --env-file .env -f infra/compose/compose.dev.yml)

if [[ ! -f .env ]]; then
  echo "Missing .env. Run: cp .env.example .env" >&2
  exit 1
fi

if ! "${COMPOSE[@]}" ps --status running --services | grep -qx postgres; then
  echo "postgres service is not running" >&2
  "${COMPOSE[@]}" ps
  exit 1
fi

if ! "${COMPOSE[@]}" ps --status running --services | grep -qx tinode; then
  echo "tinode service is not running" >&2
  "${COMPOSE[@]}" ps
  exit 1
fi

if ! "${COMPOSE[@]}" ps --status running --services | grep -qx app-api; then
  echo "app-api service is not running" >&2
  "${COMPOSE[@]}" ps
  exit 1
fi

curl --fail --silent --show-error --max-time 10 http://localhost:6060/ >/dev/null
curl --fail --silent --show-error --max-time 10 http://127.0.0.1:8080/health/ready >/dev/null

echo "Tinode + app-api HTTP smoke test passed"
"${COMPOSE[@]}" ps
