# OpenIM upstream versions

All images use explicit tags. Immutable digests must be captured during the first successful pull and added before any production deployment.

| Component | Selected version | Source | Status |
| --- | --- | --- | --- |
| OpenIM Docker distribution | `v3.8` (`fb96638`) | [openim-docker release](https://github.com/openimsdk/openim-docker/releases/tag/v3.8) | selected for M0/M1 |
| OpenIM Server image | `v3.8.3-patch.12` | [openim-docker v3.8 env](https://github.com/openimsdk/openim-docker/blob/v3.8/.env) | selected for M1; digest pending |
| OpenIM Chat image | `v1.8.4-patch.2` | [openim-docker v3.8 env](https://github.com/openimsdk/openim-docker/blob/v3.8/.env) | selected for M1; digest pending |
| MongoDB | `7.0` | [MongoDB image](https://hub.docker.com/_/mongo) | selected for local M1 |
| Redis | `7.0.0` | [Redis image](https://hub.docker.com/_/redis) | selected for local M1 |
| etcd | `3.5.13` | [etcd image](https://hub.docker.com/r/bitnamilegacy/etcd) | selected for local M1 |
| Kafka | `3.5.1` | [Kafka image](https://hub.docker.com/r/bitnamilegacy/kafka) | selected for local M1 |
| MinIO | `RELEASE.2024-01-11T07-46-16Z` | [MinIO image](https://hub.docker.com/r/minio/minio) | selected for local M1 |
| OpenIM SDK Core | `v3.8.3-patch.14` | [SDK Core release](https://github.com/openimsdk/openim-sdk-core/releases/tag/v3.8.3-patch.14) | candidate; compatibility/license pending |
| OpenIM Web SDK | `@openim/wasm-client-sdk` `3.8.3-patch.15.1` | [npm package](https://www.npmjs.com/package/@openim/wasm-client-sdk) | candidate; validate against server before M3 |
| OpenIM Electron/Web demo | commit `62d7ca7b12e91144b315f36c8ebd1d9e0457a352` | [openim-electron-demo](https://github.com/openimsdk/openim-electron-demo/commit/62d7ca7b12e91144b315f36c8ebd1d9e0457a352) | imported into `openim/web`; AGPL-3.0 with additional upstream terms |

## Compatibility notes

The server/Chat image pair is copied from the official `openim-docker` v3.8 environment file. The SDK versions are recorded as candidates only; do not add them to a client until the M3 browser contract test confirms API/WebSocket compatibility and License Gate review is complete.
