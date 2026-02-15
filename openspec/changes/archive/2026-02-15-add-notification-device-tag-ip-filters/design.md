## Context

Sentinel notifier routes currently match only `event_types` and optional severity, so any route that subscribes to broad event sets receives notifications for every device in the tailnet. This makes wildcard routing noisy and forces operators to narrow to a small event list instead of scoping by target devices.

The current data model also does not carry full normalized device identity fields end-to-end for targeting. Event payloads consistently include `name`, but route matching cannot reliably filter on tags, owners, or IPs without deterministic identity inputs.

## Goals / Non-Goals

**Goals:**
- Add optional route-level device selectors for `name`, `tag`, `owner`, and `ip`.
- Preserve backward compatibility: routes without selectors keep current behavior.
- Make selector behavior deterministic and testable for mixed selector dimensions.
- Ensure device identity inputs include stable owner and IP values in both poll and realtime modes.
- Keep device-scoped event payloads aligned with selector fields so delivery and sink output share a consistent identity view.

**Non-Goals:**
- Adding regex/glob matching semantics for names or tags.
- Building a new non-device subject selector system in this change.
- Changing policy/idempotency semantics or sink transport behavior.
- Introducing control-plane API lookups at notify time.
- Resolving external owner identity directories beyond values available from Tailscale status/netmap data.

## Decisions

### 1. Add an explicit route device selector block

Decision:
- Extend notifier routes with an optional device selector object:
  - `notifier.routes[].device.names[]`
  - `notifier.routes[].device.tags[]`
  - `notifier.routes[].device.owners[]`
  - `notifier.routes[].device.ips[]`

Rationale:
- Keeps the filter scope explicit and local to each route.
- Leaves room to add additional subject selectors later without changing top-level route semantics.

Alternatives considered:
- Top-level route fields (`names`, `tags`, `ips`) without a nested object.
  - Rejected to avoid future namespace collisions with route-level fields.
- Global device filter shared by all routes.
  - Rejected because operators need per-route targeting behavior.

### 2. Define deterministic selector matching semantics

Decision:
- Route evaluation order remains: `event_types` -> `severities` -> `device selector`.
- Within each selector list (`names`, `tags`, `owners`, `ips`), matching is OR.
- Across selector dimensions, matching is AND (for configured dimensions only).
- If `device` selector is omitted, route behavior is unchanged.
- If a route has a `device` selector, non-device events do not match that route.

Rationale:
- Prevents ambiguous delivery when operators combine constraints.
- Preserves existing routing behavior by default.

Alternatives considered:
- OR across selector dimensions.
  - Rejected because it broadens matches unexpectedly and is harder to reason about.

### 3. Normalize device identity with canonical owner and IP sets

Decision:
- Extend source/snapshot peer models with normalized device identity data used for targeting.
- Populate canonical `owners` values from stable owner identifiers available from status/netmap data, and canonical `ips` from available peer address fields, with deterministic ordering.
- Carry canonical identity fields (`name`, `tags`, `owners`, `ips`) in device-scoped event payloads.

Rationale:
- Notifier matching needs stable device identity inputs without external lookups.
- Canonical payload fields improve sink observability and troubleshooting.

Alternatives considered:
- Resolve device identity dynamically from state store during notify.
  - Rejected to avoid extra coupling and stale lookup failure modes.

### 4. Validate selector inputs at config load time

Decision:
- Config validation rejects empty selector values and invalid IP literals.
- Structured env decoding for `SENTINEL_NOTIFIER_ROUTES` supports the new nested `device` object.
- Docs/examples include wildcard event routes combined with device selectors.

Rationale:
- Failing fast at config load avoids silently ineffective routes.
- Maintains parity between file-based and env-only config workflows.

## Risks / Trade-offs

- [Selector misconfiguration can silently reduce notification coverage] -> Mitigation: strict validation, explicit docs examples, and test coverage for match/no-match scenarios.
- [IP identity extraction may differ between poll and realtime source payload shapes] -> Mitigation: shared normalization helpers and cross-mode tests.
- [Adding identity fields to device-scoped event payloads can increase event size] -> Mitigation: include compact normalized values only (name/tags/owners/ips) and avoid volatile metadata.
- [Routes with device selectors dropping non-device events may surprise operators] -> Mitigation: document behavior and include validate-config messaging/examples.

## Migration Plan

1. Add config model and validation support for `notifier.routes[].device.{names,tags,owners,ips}`.
2. Extend source/snapshot normalization to include canonical owner/IP identity fields across poll and realtime modes.
3. Update device-scoped event emitters to include canonical selector identity fields in payloads.
4. Extend notifier route matching to apply device selector semantics.
5. Add tests for config parsing/validation, route matching, and source-mode parity.
6. Update operator docs and `config.example.yaml`.
7. Rollback strategy: remove device selector evaluation in notifier and ignore selector config fields while retaining existing route behavior.

## Open Questions

- Should `device.names` support canonical FQDN aliases in addition to current computed name, or remain exact-string matching only?
- Should `device.owners` match only stable user IDs in v1, or include optional login/email aliases when available?
- Should `device.ips` support CIDR/prefix selectors in this change, or strict IP literal matching only?
