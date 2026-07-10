# Changelog

## Unreleased

### Features

* add the user-facing `merger` CLI (`cmd/merger`) with `scan`, `validate`,
  `init`, and `version` commands; `scan` runs the analysis pipeline offline
  against a diff and assigns a merge lane, with `-format text|json` and a
  `-fail-on-lane` CI gate. Config is auto-discovered from `.merger/`.
* add `internal/scan`, an offline pipeline that reuses the mutations, runtime
  graph, risk, policy, and lane engines without the ingest/service dependencies
* scaffold the Phase 1 control-plane slice: GitHub webhook ingest, PR diff
  parsing, Change Packet generation, rule-based semantic mutation detection,
  risk scoring, policy evaluation, and merge-lane assignment
  (`GREEN`/`YELLOW`/`RED`/`BLACK`)
* publish public extension seams (`pkg/extensions`) for SCM, topology, event,
  analyzer, and persistence adapters
* add first-party GitHub, NATS JetStream, and PostgreSQL implementations

### Ecosystem

* adopt the devr-tools tool-family conventions: Apache-2.0 `LICENSE`,
  `internal/version` package, `.golangci.yml` lint config, `SECURITY.md`,
  `CONTRIBUTING.md`, and this changelog
