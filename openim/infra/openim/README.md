# OpenIM upstream deployment notes

The Compose file in `../compose/compose.dev.yml` is a small, local-only wrapper around the official `openim-docker` v3.8 deployment shape. It keeps OpenIM Server and Chat Server unmodified and wires the required MongoDB, Redis, etcd, Kafka, and MinIO services on a private Docker network.

The images are pinned to explicit release tags in `.env.example`. Do not replace them with `latest`. Before a production deployment, capture immutable digests for every image and record them in `openim/UPSTREAM_VERSIONS.md`.

Only loopback ports are published in development. A production deployment must put HTTPS/WSS and access control in front of the OpenIM API and message gateway, and must not expose MongoDB, Redis, Kafka, or etcd publicly.

OpenIM administrative APIs and administrator credentials belong in the future `app-api`; they are not called by this infrastructure task and must never be sent to a browser.

