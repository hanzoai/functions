# functions — Hanzo FaaS plane

Functions-as-a-service. **The live API is `/v1/functions` in `hanzoai/cloud`**
(`apps/functions`) — registry, triggers, metrics and invoke, with invoke
delegating to the sandboxed executor. This repo is the standalone control
plane, and it is NOT deployed: the cluster install was removed on 2026-08-04
(`universe` 798fac65) after serving zero invocations in 36 days, because the
cloud plugin had been answering the same API the whole time. Two
implementations of one product; this was the copy nobody called.

Upstream attribution lives in `NOTICE` and nowhere else.

`CLAUDE.md` and `AGENTS.md` are symlinks to this file — edit `LLM.md`, never
the symlinks, and never commit them.

## Request path

```
client ──HTTPS──> api.hanzo.ai (hanzoai/gateway)
                    │  validates IAM JWT, strips client identity headers,
                    │  injects X-Org-Id / X-User-Id / X-Request-Id
                    ▼
          /v1/functions/*  ──> router.fission.svc.cluster.local:80
                    │            (Fission router)
                    ▼
          function pod (python|nodejs|rust env, scale-to-zero)
```

A function whose HTTP trigger is `/{name}` is reachable at
`https://api.hanzo.ai/v1/functions/{name}`.

## Gateway route (owned by `gateway-config`)

One route; the router is the single data-plane entrypoint.

| Field | Value |
|-------|-------|
| Match | `PathPrefix: /v1/functions` (all methods) |
| Upstream | `http://router.fission.svc.cluster.local:80` |
| Path rewrite | strip `/v1/functions`, forward the remainder as the router path |
| Auth | IAM JWT required, as with every `/v1/*` |
| Inject | `X-Org-Id` (JWT `owner`), `X-User-Id` (`sub`), `X-User-Email`, `X-Request-Id`; **strip** any client-supplied identity headers |
| Timeout | ≥ 60s — a cold start specializes an env pod on first call |

**Multi-network.** One control plane watches two namespaces: `fission` =
prod/mainnet, `fission-sandbox` = sandbox/testnet+devnet. The router resolves
triggers across both, so sandbox can route by host or prefix. Network selection
(mainnet/testnet/devnet RPC) is injected per-function as env from the project's
network — not a gateway concern.

## Billing / metering

None here. An invocation is paid for where it is served: `/v1/functions` in cloud
(`apps/functions`), through cloud's one metering client, which refuses a call it
cannot weigh. This plane's router meter (`pkg/usagemeter`, over commerce's copy of
that client) was deleted with the copy; this plane serves no traffic.

## `/v1` surface

Two planes. **Invoke** goes gateway→router. **Management** is
k8s-CRD + storagesvc work belonging in a `functions` module in cloud (already the
`/v1` backend), org-scoped by `X-Org-Id`.

| Method & path | Action | Maps to |
|---------------|--------|---------|
| `GET /v1/functions/runtimes` | list runtimes | Environment CRs |
| `POST /v1/functions` | create | code → Package + Function CR + HTTPTrigger `/{name}` |
| `GET /v1/functions` | list (org-scoped) | Function CRs where `hanzo.ai/org={org}` |
| `GET /v1/functions/{name}` | detail | Function + Trigger + status |
| `DELETE /v1/functions/{name}` | delete | cascade Function+Package+Trigger |
| `POST /v1/functions/{name}` | **invoke** | data-plane via gateway→router |
| `GET /v1/functions/{name}/logs` | logs | stream env-pod logs |
| `GET /v1/functions/{name}/metrics` | usage | invocations / GPU-seconds from commerce |

Create body:

```json
{ "name": "thumb", "runtime": "python", "network": "mainnet",
  "code": "<source>", "entrypoint": "main", "env": {"K":"V"}, "gpu": false }
```

**Multi-tenancy.** v1 uses one namespace per network and isolates orgs by name
prefix + label (Function `"{org}-{name}"`, label `hanzo.ai/org={org}`, trigger
`/{org}/{name}`); list filters on the label. That is a demo-grade boundary, not a
security boundary. Hardening is one namespace per org per network (`fn-{org}` /
`fn-{org}-sandbox`) so RBAC, resource quota and network policy isolate tenants at
the k8s boundary — the control plane already watches N namespaces via
`additionalFissionNamespaces`.

## Ownership boundaries

| Concern | Owner |
|---------|-------|
| Fission control plane, runtimes, env images | this repo |
| `/v1/functions` gateway route + identity injection | `gateway-config` |
| `/v1/functions` management backend + console UI | cloud |
