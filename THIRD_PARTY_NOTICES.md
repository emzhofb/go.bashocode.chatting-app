# Third-party notices

This file is the starting inventory for the implementation. Update it whenever a dependency or upstream client is added.

## Tinode server

- Project: [tinode/chat](https://github.com/tinode/chat)
- Version/image: `v0.25.3` / `tinode/tinode:0.25.3`
- License: GPL-3.0
- Usage: unmodified container/process; not linked into `app-api`.
- The local Compose bootstrap follows Tinode's PostgreSQL fallback-database requirement; this does not grant the app role access to Tinode databases.
- Required action before distribution: perform a formal GPL compliance and deployment review.

## Tinode JavaScript SDK and web client

- SDK: [tinode/tinode-js](https://github.com/tinode/tinode-js)
- Package: `tinode-sdk@0.25.3`
- Reference client: [tinode/webapp](https://github.com/tinode/webapp) (used for local acceptance; M5 shell is a separate application)
- License: Apache-2.0 (verify the exact pinned commit and notice files when added).

## PostgreSQL

- Project: [postgres](https://www.postgresql.org/)
- Usage: local and production database service.
- License: PostgreSQL License.

## coturn (future M10)

- Project: [coturn/coturn](https://github.com/coturn/coturn)
- Usage: self-hosted TURN server for 1-to-1 WebRTC calls only.
- License: BSD-3-Clause.
