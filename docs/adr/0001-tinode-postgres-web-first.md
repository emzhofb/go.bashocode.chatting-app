# ADR 0001: Tinode + PostgreSQL + web-first

- Status: accepted
- Date: 2026-08-19

## Context

The product needs WhatsApp/Telegram-like messaging with no per-user license fee, while keeping product accounts, moderation, and branding under our control. Building message routing, topics, receipts, presence, sync, attachments, and call signaling from scratch would make the first release much larger and harder to validate.

## Decision

Use the upstream Tinode server as a separately deployed messaging engine, PostgreSQL as its storage backend, and a separate Go `app-api` for business-owned accounts, sessions, reports, moderation, and audit records. Start with a responsive web/PWA client using the official Tinode JavaScript SDK/reference client. Keep Tinode server source unmodified and pin an explicit release/image tag.

Use separate PostgreSQL databases and roles for Tinode and the application. Integrate application authentication through Tinode's REST authenticator; do not read or write Tinode's internal tables from `app-api`.

For the local PostgreSQL bootstrap, use the dedicated role `tinode_admin` and an empty fallback database with the same name. Tinode's upstream initializer then creates the target `tinode` database. The application role remains `app_chat` and has no permission on either Tinode database.

## Consequences

Positive:

- Messaging capabilities arrive earlier through a mature protocol/server.
- Product business logic remains separate and testable.
- Vertical scaling and Docker Compose are adequate for the initial target.

Costs and constraints:

- Tinode server is GPLv3; distribution and on-premise scenarios require legal review.
- Tinode does not provide stable end-to-end encryption for this plan; the product must not claim E2EE.
- Upstream upgrades and reference-client customization need explicit pinning and regression tests.
- Group calls, federation, and native mobile are outside the first implementation path.

## Rejected alternatives

- Building the entire messaging backend in-house: too much protocol, sync, storage, and abuse-prevention surface for the first release.
- Using a hosted chat API: conflicts with the self-hosted/no per-user licensing requirement and reduces infrastructure control.
- Forking Tinode server immediately: increases maintenance and licensing obligations before a product need is proven.
