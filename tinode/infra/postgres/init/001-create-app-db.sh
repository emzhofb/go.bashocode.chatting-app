#!/usr/bin/env bash
set -euo pipefail

: "${POSTGRES_TINODE_PASSWORD:?POSTGRES_TINODE_PASSWORD is required}"
: "${POSTGRES_TINODE_USER:?POSTGRES_TINODE_USER is required}"
: "${POSTGRES_APP_PASSWORD:?POSTGRES_APP_PASSWORD is required}"

psql --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" \
  --set=ON_ERROR_STOP=1 \
  --set=tinode_user="$POSTGRES_TINODE_USER" \
  --set=tinode_password="$POSTGRES_TINODE_PASSWORD" \
  --set=app_password="$POSTGRES_APP_PASSWORD" <<'SQL'
SELECT format('CREATE ROLE %I LOGIN CREATEDB PASSWORD %L', :'tinode_user', :'tinode_password')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'tinode_user')\gexec

ALTER ROLE :"tinode_user" CREATEDB;

SELECT format('CREATE ROLE app_chat LOGIN PASSWORD %L', :'app_password')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_chat')\gexec

SELECT format('CREATE DATABASE %I OWNER app_chat', 'app_chat')
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'app_chat')\gexec

SELECT format('CREATE DATABASE %I OWNER %I', :'tinode_user', :'tinode_user')
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = :'tinode_user')\gexec
SQL
