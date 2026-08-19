# Upstream versions

Pin every upstream image/repository before deployment. Do not use `latest` outside a deliberate upgrade task.

| Component | Version/tag | Source | Status |
| --- | --- | --- | --- |
| Tinode server image | `tinode/tinode:0.25.3` | <https://hub.docker.com/r/tinode/tinode> | selected for M0/M1 |
| Tinode server release | `v0.25.3` | <https://github.com/tinode/chat/releases/tag/v0.25.3> | selected for M0/M1 |
| PostgreSQL image | `postgres:16-alpine` | <https://hub.docker.com/_/postgres> | selected for local development |
| Tinode JS SDK | not added yet | <https://github.com/tinode/tinode-js> | planned M5 |
| Tinode web client | not added yet | <https://github.com/tinode/webapp> | planned M5 |
| coturn | not added yet | <https://github.com/coturn/coturn> | planned M10 |

When a dependency is pulled, record the immutable image digest or commit SHA in the table and in the change log. The current environment cannot resolve GitHub from the shell, so the release tag is the authoritative M0 pin and the digest must be captured during the first successful image pull.

