# Tinode Chat Product

Self-hosted chat product built around the upstream Tinode messaging server.

The implementation source of truth is [TINODE_IMPLEMENTATION_PLAYBOOK.md](TINODE_IMPLEMENTATION_PLAYBOOK.md). Work through one checkbox at a time and do not modify Tinode server source.

## Current status

Milestones M0–M4 are implemented; the local M1 Tinode environment and reference-client personal-chat smoke test are healthy. Group chat, attachments, reconnect, and the custom branded web client remain later milestones.

## Start the local Tinode environment

Requirements:

- Docker Desktop or Docker Engine with Compose v2;
- at least 2 GB free memory for the local containers;
- network access for the first image pull.

Run:

```sh
cp .env.example .env
docker compose --env-file .env -f infra/compose/compose.dev.yml up -d
./scripts/smoke-test.sh
```

Open <http://localhost:6060/> to use the Tinode reference web client served by the Tinode image.

The development Compose stack also starts `app-api` on <http://127.0.0.1:8080/>. Its internal REST authenticator listener is reachable only inside the Compose network at `app-api:8081`.

For local M2 `app-api` development, PostgreSQL is bound to `127.0.0.1:5432` only. It is not exposed to the LAN; production Compose must keep the database private.

Stop without deleting data:

```sh
docker compose --env-file .env -f infra/compose/compose.dev.yml down
```

`docker compose down -v` deletes the local PostgreSQL and media volumes. Use it only when intentionally resetting the development environment.

## Test checklist

Run backend tests from the repository root:

```sh
(cd services/app-api && go test ./...)
./scripts/smoke-test.sh
```

Manual personal-chat smoke test:

1. Open `http://localhost:6060/` in two browser tabs.
2. Create or sign in as two development users, for example `alice01` and `bob01`.
3. In Alice's tab, choose the chat/new-chat icon, then **by id**.
4. Enter Bob's Tinode public ID (`usr...`) and choose **Subscribe**.
5. Send a message from Alice and reply from Bob.
6. Confirm both tabs show the messages and the delivery/read indicators (`done`/`done_all`).
7. Reload both tabs and confirm the conversation history remains.

The current manual acceptance covers account provisioning, personal chat, and basic receipts. Group, attachment, and offline/reconnect scenarios are intentionally still unchecked in the playbook.

## Repository rules

- Tinode server stays upstream and is pinned to an explicit image tag.
- Application data and Tinode data use separate PostgreSQL databases and roles. The dedicated Tinode role is named `tinode_admin` in development because Tinode's upstream PostgreSQL initializer requires a same-named fallback database when the role is not `postgres`.
- Secrets belong in the local `.env` or deployment secret manager, never Git.
- Chat messages/topics stay in Tinode; the future `app-api` owns product accounts, sessions, moderation, and audit data.

# go.bashocode.chatting-app
