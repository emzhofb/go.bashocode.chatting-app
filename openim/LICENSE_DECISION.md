# OpenIM license decision

Status: **pending**

This repository is evaluating OpenIM for a proprietary, closed-source product. No external release or production use is authorized until the company has a written license decision.

## Verified facts

- OpenIM Server is published separately from the SDK Core and each client SDK must be checked independently.
- OpenIM SDK Core and the JavaScript WASM SDK publish an AGPL-3.0-or-later **or commercial license** model.
- OpenIM's commercial pricing is not published on the public documentation pages.
- Commercial licensing enquiries are directed to `contact@openim.io`.

Official references:

- [OpenIM SDK Core](https://github.com/openimsdk/openim-sdk-core)
- [OpenIM JavaScript WASM SDK](https://github.com/openimsdk/openim-sdk-js-wasm)
- [OpenIM Server](https://github.com/openimsdk/open-im-server)
- [OpenIM licensing](https://openim.io/en/licensing)
- [OpenIM Enterprise](https://openim.io/enterprise/)

## Decision checklist

- [ ] Confirm the product is proprietary/closed-source.
- [ ] Obtain legal review for OpenIM Server, SDK Core, WASM SDK, platform SDKs, demo clients, and plugins used.
- [ ] Obtain a written commercial quotation for 5,000–8,000 registered accounts.
- [ ] Confirm whether pricing is per company, application, deployment, domain, server, MAU, or registered user.
- [ ] Confirm SDK Core can be licensed separately from Business client UI.
- [ ] Confirm staging, disaster recovery, and multiple environments are covered.
- [ ] Confirm update, security patch, and support terms.
- [ ] Confirm modification and redistribution rights for SDK Core.
- [ ] Record the final state as `approved-commercial` or `approved-AGPL` only after written approval.

## Quotation request

Use the following context with OpenIM and store the response/contract outside this repository:

```text
We are evaluating OpenIM SDK Core for a proprietary, closed-source,
self-hosted messaging product with approximately 5,000–8,000 registered users.

Please provide:
1. Commercial license price and billing model.
2. Whether SDK Core can be licensed separately from Business client UI.
3. Limits by users, MAU, app, domain, deployment, or server count.
4. Rights to modify and redistribute the SDK in our client applications.
5. Coverage for development, staging, production, and disaster recovery.
6. Included updates, security patches, and support period.
7. Separate pricing for voice/video, meetings, HA, and Kubernetes support.
8. Renewal, termination, and perpetual-use terms.
```

