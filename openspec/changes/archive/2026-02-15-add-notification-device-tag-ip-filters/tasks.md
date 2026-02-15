## 1. Route selector config model and validation

- [x] 1.1 Extend notifier route config structs with optional `device.names`, `device.tags`, `device.owners`, and `device.ips` fields.
- [x] 1.2 Update config/env decoding for `SENTINEL_NOTIFIER_ROUTES` so nested device selector fields are loaded in env-only mode.
- [x] 1.3 Add validation rules for selector fields (non-empty values, valid IP literals) with route-indexed error messages.

## 2. Device identity normalization

- [x] 2.1 Extend source peer decoding (status and netmap JSON paths) to extract canonical device owner and peer IP identity values.
- [x] 2.2 Add snapshot peer fields and normalization helpers for deterministic ordering/deduplication of owner and identity IP values.
- [x] 2.3 Add unit tests proving poll and realtime source modes produce equivalent normalized device identity behavior.

## 3. Device-scoped event payload baseline updates

- [x] 3.1 Update device-scoped event emitters to include stable selector identity fields (`name`, `tags`, `owners`, `ips`) in payloads.
- [x] 3.2 Add/adjust event and detector tests to verify selector identity payload fields are present and stable.

## 4. Notifier route matching with device selectors

- [x] 4.1 Extend notifier route matching to evaluate device selector dimensions after existing event type/severity checks.
- [x] 4.2 Implement deterministic selector semantics (OR within dimension, AND across configured dimensions, non-device events do not match device-filtered routes).
- [x] 4.3 Add notifier tests for name/tag/owner/IP selector matching, mismatch cases, and backward compatibility for routes without selectors.

## 5. Operator examples and documentation

- [x] 5.1 Update `config.example.yaml` with device selector route examples for name-, tag-, owner-, and IP-based targeting.
- [x] 5.2 Update configuration docs with selector schema, matching semantics, and non-device event behavior for device-filtered routes.

## 6. End-to-end verification

- [x] 6.1 Add integration coverage showing filtered routes reduce notifications to selected devices while wildcard event routing remains enabled.
- [x] 6.2 Run focused test suites for `internal/config`, `internal/source`, `internal/snapshot`, `internal/diff`, and `internal/notify`.
