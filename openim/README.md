# OpenIM Chat Product

This directory contains the self-hosted OpenIM proof-of-concept described by [OPENIM_IMPLEMENTATION_PLAYBOOK.md](OPENIM_IMPLEMENTATION_PLAYBOOK.md). The current implementation covers the M0/M1 foundation, the M2 app API foundation, and the first M3 web chat vertical slice.

## Start locally

Requirements: Docker Engine/Desktop with Compose v2 and `curl`.

```sh
cd openim
cp .env.example .env
./scripts/dev-up.sh
./scripts/smoke-test.sh
```

The smoke test checks that all required containers are running/healthy and that the loopback OpenIM and product API endpoints respond. It does not send messages; that requires the M3 client work.

Run the web client separately:

```sh
cd openim/web
npm install
npm run dev
```

Open `http://127.0.0.1:3000` after the stack is healthy.

Useful development endpoints:

- OpenIM API: `http://127.0.0.1:10002`
- OpenIM message gateway: `ws://127.0.0.1:10001`
- OpenIM Chat API: `http://127.0.0.1:10008`
- Product API: `http://127.0.0.1:8080`
- MinIO API/console: `http://127.0.0.1:10005` / `http://127.0.0.1:10004`

Stop the stack without deleting data:

```sh
./scripts/dev-down.sh
```

To intentionally reset all local OpenIM data, use the separately guarded destructive operation:

```sh
./scripts/dev-down.sh --reset
```

## Scope and safety

- The values in `.env.example` are local development placeholders only.
- Never commit `.env`, real credentials, tokens, or private keys.
- Do not read or write OpenIM's MongoDB directly from application code; use its official Server API/SDK contracts.
- The product API keeps identity/session data in the separate PostgreSQL `app_chat` database.
- `OPENIM_ADMIN_SECRET` is consumed only by `app-api`; it must never be sent to a browser.
- Do not claim end-to-end encryption. TLS and encryption at rest are not E2EE.
