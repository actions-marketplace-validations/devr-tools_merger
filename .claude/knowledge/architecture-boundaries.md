# Architecture & System Boundaries

- merger is a **control plane made of two services**, not a CLI: `cmd/merger-ingest` (GitHub webhook ingest) and `cmd/merger-controlplane` (analysis + gRPC/HTTP). Platform deps are PostgreSQL, Redis, and NATS JetStream.
- The Go module path is `github.com/mergerhq/merger`, but the GitHub home is `github.com/devr-tools/merger` — it is part of the devr-tools tool family (siblings: `codeguard`, `cleanr`). The module/remote mismatch is intentional legacy, not an error.
- Core pipeline lives under `internal/`: `ingest → mutations → runtimegraph → risk → policy → lanes`. A PR becomes a Change Packet, mutations are classified, blast radius is estimated, risk is scored, policy is applied, and a merge lane (`GREEN`/`YELLOW`/`RED`/`BLACK`) is assigned.
- The `merger` CLI (`cmd/merger`, logic in `internal/cli`) is the installable local face. It does NOT run the services; it drives `internal/scan`, an offline pipeline that reuses the same engines (mutations → runtimegraph → risk → policy → lanes) against a raw unified diff, with a filesystem-backed content loader instead of the GitHub content API. Keep `internal/scan` in sync with `internal/ingest`'s stage order if the pipeline changes.
- Public extension seams live in `pkg/extensions` (SCM, topology, event, analyzer, persistence adapters). First-party impls (GitHub, NATS, PostgreSQL) are the reference implementations; the seams exist so other orgs can swap them.
- `pkg/merger` is public type aliases only; `pkg/diff` is reusable unified-diff parsing; `pkg/identity` is shared identity types.
- Policies are YAML (`config/policies`) and composable: `when` mutation conditions → `require` reviewers/evidence/deployment → `action` minimum lane.
