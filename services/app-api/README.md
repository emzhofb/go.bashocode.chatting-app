# `app-api`

Go business API for product accounts, sessions, moderation, audit, and Tinode REST authentication. Chat messages and topics remain in Tinode.

## M2 local run

The local Compose file currently starts PostgreSQL and Tinode. To run `app-api`, set these values in `.env` or export them in the shell:

```sh
export APP_ENV=development
export APP_HTTP_ADDR=:8080
export APP_INTERNAL_HTTP_ADDR=:8081
export APP_DATABASE_URL='postgresql://app_chat:development-only@localhost:5432/app_chat?sslmode=disable'
go run ./cmd/api
```

Health endpoints:

```sh
curl -i http://localhost:8080/health/live
curl -i http://localhost:8080/health/ready
```

The internal listener is intentionally not published by the Compose file. It must be reachable only on the private container network once `app-api` is containerized.

