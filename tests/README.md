# Tests

Tests live under `tests/` to keep source and verification artifacts separated.

Recommended split:

- `tests/lanes`: merge lane decision tests
- `tests/mutations`: semantic mutation classification tests
- `tests/policy`: policy evaluation tests

Future phases should add:

- contract tests for GitHub webhook and check integrations
- runtime graph fixture tests
- evidence orchestration workflow tests
- benchmark suites for high-volume PR ingest
