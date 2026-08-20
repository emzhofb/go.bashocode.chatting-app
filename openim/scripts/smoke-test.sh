#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"
COMPOSE=(docker compose --env-file .env -f infra/compose/compose.dev.yml)

if [[ ! -f .env ]]; then
  echo "Missing .env. Run: cp .env.example .env" >&2
  exit 1
fi

env_value() {
  local key="$1"
  awk -F= -v key="$key" '$1 == key {sub(/^[^=]*=/, ""); print; exit}' .env
}

OPENIM_API_PORT="$(env_value OPENIM_API_PORT)"
OPENIM_CHAT_API_PORT="$(env_value OPENIM_CHAT_API_PORT)"
MINIO_PORT="$(env_value MINIO_PORT)"
APP_HTTP_PORT="$(env_value APP_HTTP_PORT)"
OPENIM_API_PORT="${OPENIM_API_PORT:-10002}"
OPENIM_CHAT_API_PORT="${OPENIM_CHAT_API_PORT:-10008}"
MINIO_PORT="${MINIO_PORT:-10005}"
APP_HTTP_PORT="${APP_HTTP_PORT:-8080}"

required_services=(mongo redis etcd kafka minio postgres openim-server openim-chat app-api)
running_services="$(${COMPOSE[@]} ps --status running --services)"
for service in "${required_services[@]}"; do
  if ! grep -qx "$service" <<<"$running_services"; then
    echo "$service is not running" >&2
    "${COMPOSE[@]}" ps
    exit 1
  fi
done

check_http() {
  local name="$1" url="$2" status
  status="$(curl --silent --show-error --max-time 10 --output /dev/null --write-out '%{http_code}' "$url")"
  [[ "$status" != "000" ]] || { echo "$name is unreachable" >&2; exit 1; }
}

check_http "OpenIM API" "http://127.0.0.1:${OPENIM_API_PORT}/"
check_http "OpenIM Chat API" "http://127.0.0.1:${OPENIM_CHAT_API_PORT}/"
check_http "MinIO" "http://127.0.0.1:${MINIO_PORT}/minio/health/live"
check_http "app-api" "http://127.0.0.1:${APP_HTTP_PORT}/health/live"

echo "OpenIM local infrastructure smoke test passed"
"${COMPOSE[@]}" ps
