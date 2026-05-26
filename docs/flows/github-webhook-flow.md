# GitHub Webhook Flow

## Phase 1 Execution Path

1. GitHub emits a `pull_request` webhook for `opened`, `reopened`, or `synchronize`.
2. `cmd/merger-ingest` accepts the request at `/webhooks/github`.
3. The handler extracts:
   - GitHub event type
   - delivery ID as correlation ID
   - repository and PR identity
4. The ingest processor emits `PROpened`.
5. The GitHub adapter fetches pull request metadata and unified diff text.
6. `pkg/diff` converts unified diff into normalized changed files.
7. The processor creates a `ChangePacket` shell and emits `ChangePacketCreated`.
8. The mutation engine:
   - derives signals from patch content
   - matches semantic mutation rules
   - emits `MutationDetected`
9. The runtime graph resolver estimates:
   - services and ownership touched
   - blast radius
   - criticality placeholder
10. The risk engine computes:
    - risk entries
    - aggregate risk score
    - emits `RiskAssigned`
11. The policy engine resolves:
    - reviewers
    - evidence requirements
    - deployment requirements
    - blocking decisions
12. The merge lane assigner selects `GREEN`, `YELLOW`, `RED`, or `BLACK`.
13. The processor emits `MergeLaneAssigned`.
14. If policies block or violate requirements, `PolicyViolationDetected` is emitted.
15. The GitHub Check Run publisher posts the MergeR summary back to GitHub.

## Future Expansion Points

- Replace in-memory eventing with NATS/Kafka.
- Fan out Change Packet processing across specialized services.
- Persist packet state and evidence state transitions.
- Feed deployment outcomes back into risk and policy calibration.
