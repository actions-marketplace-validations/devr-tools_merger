# API Surface

Phase 1 keeps API contracts intentionally thin:

- `proto/merger/v1/controlplane.proto` defines the initial gRPC surface placeholder.
- `cmd/merger-ingest` exposes HTTP webhook endpoints for GitHub.
- `cmd/merger-controlplane` exposes health endpoints and will later host control APIs and event consumers.

Future phases should add:

- Change Packet query APIs
- Evidence orchestration APIs
- Runtime graph ingest/query APIs
- Reviewer routing APIs
