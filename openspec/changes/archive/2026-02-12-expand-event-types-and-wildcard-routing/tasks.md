## 1. Event catalog foundations

- [x] 1.1 Add event type constants and helper constructors for new peer, daemon, prefs, and tailnet event families in `internal/event`
- [x] 1.2 Define payload conventions per event family and add unit tests validating stable event IDs/idempotency keys
- [x] 1.3 Update docs-facing event taxonomy references used by config/docs examples

## 2. Realtime state extraction and normalization

- [x] 2.1 Extend realtime source parsing to capture additional tracked fields needed for non-presence diffs
- [x] 2.2 Update snapshot normalization/hash inputs for newly tracked stable fields while preserving volatile-field exclusions
- [x] 2.3 Add tests proving equivalent notify frames remain no-op and changed tracked fields alter snapshot hash

## 3. Expanded diff detectors

- [x] 3.1 Implement detector logic for new peer change families (membership, routes, tags, machine authorization, key lifecycle, hostinfo-derived signals)
- [x] 3.2 Implement detector logic for daemon/prefs/tailnet metadata transitions emitted from realtime updates
- [x] 3.3 Add detector and engine tests covering presence + non-presence transitions and unchanged-state suppression

## 4. Notifier wildcard route matching

- [x] 4.1 Update notifier route matcher to treat `event_types: [\"*\"]` as match-all
- [x] 4.2 Define deterministic behavior for mixed wildcard + literal event type lists and document it in tests
- [x] 4.3 Preserve explicit-literal matching behavior for routes that do not include `*`

## 5. Config validation and operator examples

- [x] 5.1 Update config validation to accept wildcard event type routes and reject invalid route event type definitions
- [x] 5.2 Update default/example configuration to demonstrate explicit event routing and wildcard routing usage
- [x] 5.3 Add config tests for wildcard acceptance and backward compatibility with existing explicit routes

## 6. Runtime and notifier integration verification

- [x] 6.1 Add integration tests verifying expanded events flow end-to-end through policy and notifier sinks
- [x] 6.2 Add integration tests verifying wildcard routes deliver expanded events without requiring explicit enumeration
- [x] 6.3 Verify existing `peer.online`/`peer.offline` behavior and sink delivery remain unchanged

## 7. Documentation updates

- [x] 7.1 Update docs for supported event types and route filtering semantics (including wildcard `*`)
- [x] 7.2 Update troubleshooting guidance with examples for validating expanded events and wildcard routes
- [x] 7.3 Document migration guidance for operators moving from explicit presence-only routing to broader catalogs
