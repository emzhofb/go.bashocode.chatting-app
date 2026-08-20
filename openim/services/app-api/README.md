# OpenIM app-api

`app-api` owns product identity and sessions. It stores only application data in PostgreSQL and uses the official OpenIM Platform API for user provisioning and user-token issuance.

Endpoints in this M2 foundation:

- `POST /v1/auth/register`
- `POST /v1/auth/login`
- `POST /v1/auth/refresh`
- `POST /v1/auth/logout` (Bearer app access token)
- `GET /v1/me` (Bearer app access token)
- `GET /v1/openim/session` (Bearer app access token; returns a user token for the client SDK)
- `GET /v1/users/search?q=...` (Bearer app access token; returns active users without email/password fields)

The OpenIM administrator secret and admin token are backend-only. Access and refresh tokens are opaque random values; only SHA-256 hashes are stored in PostgreSQL. Structured logs never include credentials or message content.

Logout revokes the application access token and asks OpenIM to force-logout the configured platform. If the OpenIM revocation call fails, the application token is still revoked and the endpoint returns an upstream error so the client can retry.

The provisioning path follows the official Platform API contracts: read the user, call `user_register` when absent, then issue a user token with `auth/get_user_token`. OpenIM data is never read from MongoDB directly.
