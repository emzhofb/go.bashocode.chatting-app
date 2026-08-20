#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"
COMPOSE=(docker compose --env-file .env -f infra/compose/compose.dev.yml)

if [[ ! -f .env ]]; then
  echo "Missing .env. Run: cp .env.example .env" >&2
  exit 1
fi

if [[ "${1:-}" == "--reset" ]]; then
  read -r -p "Type RESET to delete all local OpenIM volumes: " confirmation
  if [[ "$confirmation" != "RESET" ]]; then
    echo "Reset cancelled"
    exit 1
  fi
  "${COMPOSE[@]}" down --volumes
else
  "${COMPOSE[@]}" down
fi

