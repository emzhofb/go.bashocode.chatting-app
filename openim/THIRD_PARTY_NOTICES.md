# OpenIM third-party notices

This is an inventory for the M0 due-diligence pass. It is not a substitute for the license files shipped by each upstream component.

| Component | Version | License/source note |
| --- | --- | --- |
| OpenIM Server | v3.8.3-patch.12 | See [upstream repository](https://github.com/openimsdk/open-im-server) and its license files. |
| OpenIM Chat | v1.8.4-patch.2 | See the image/repository metadata from the official OpenIM distribution. |
| OpenIM SDK Core | v3.8.3-patch.14 | AGPL-3.0-or-later or commercial; legal review required. |
| OpenIM WASM SDK | 3.8.3-patch.15.1 | AGPL-3.0-or-later or commercial; legal review required. |
| MongoDB | 7.0 | Container image distribution; review upstream notices before redistribution. |
| Redis | 7.0.0 | Container image distribution; review upstream notices before redistribution. |
| etcd | 3.5.13 | Container image distribution; review upstream notices before redistribution. |
| Kafka | 3.5.1 | Container image distribution; review upstream notices before redistribution. |
| MinIO | RELEASE.2024-01-11T07-46-16Z | Container image distribution; review upstream notices before redistribution. |

Before release, run a dependency/license scanner over the actual images and client/backend dependencies, retain upstream LICENSE/NOTICE files, and ensure no AGPL component is linked into proprietary code without legal approval.

