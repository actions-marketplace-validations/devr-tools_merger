# Extending MergeR

MergeR is designed as an open-source control plane core with first-party adapters and public extension contracts.

## Public SDK Surface

The public packages intended for external implementers are:

- `pkg/merger`
  - exported aliases for core domain types such as `ChangePacket`, `Mutation`, `RuntimeImpact`, and event envelopes
- `pkg/extensions`
  - provider interfaces for event buses, SCM adapters, persistence layers, mutation analyzers, and runtime graph sources

These packages allow downstream users to:

- plug in alternate SCM providers beyond GitHub
- replace NATS with another event backbone
- implement their own persistence layer
- ship organization-specific semantic analyzers
- ingest topology metadata from internal service catalogs or platform registries

## First-Party Adapters

This repository now includes first-party implementations for:

- GitHub App integration in `internal/github`
- NATS JetStream eventing in `internal/events`
- PostgreSQL persistence in `internal/store`

These are reference implementations, not hard requirements.

## Recommended Extension Strategy

1. Keep custom integrations in a separate module that imports `pkg/merger` and `pkg/extensions`.
2. Implement provider interfaces in that external module.
3. Add a thin bootstrap package or custom binary that wires your providers into the MergeR processor and services.

## Near-Term OSS Gaps

- provider registration and discovery is still constructor-based, not manifest-based
- runtime graph ingestion is source-pluggable but not yet remotely configurable
- mutation analyzers are compiled in-process; WASM or RPC analyzers are future work
