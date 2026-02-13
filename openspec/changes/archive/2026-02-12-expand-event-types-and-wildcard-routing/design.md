## Context

Sentinel currently operates with realtime IPNBus intake but emits only presence events (`peer.online`, `peer.offline`). The source already receives richer `ipn.Notify` frames (state, prefs, netmap, engine, health, auth flow signals), but the downstream snapshot and detector path tracks a limited field set. Routing currently expects explicit event type lists, which makes broad rollout of additional event families operationally tedious.

This change expands observable event coverage and introduces wildcard route matching (`*`) so operators can opt into all events without enumerating every event type.

## Goals / Non-Goals

**Goals:**
- Define and emit additional typed events from realtime IPNBus/NetMap updates.
- Preserve event schema stability (`event_type`, subject fields, hashes, payload) while expanding event families.
- Support `notifier.routes[].event_types: [\"*\"]` as match-all semantics.
- Keep explicit event-type routing behavior backward-compatible.
- Keep suppression/no-op semantics so added signal does not create duplicate spam.

**Non-Goals:**
- Replacing the notifier architecture or sink contracts.
- Introducing a separate event transport or message bus.
- Breaking existing configs that use explicit `event_types`.
- Expanding into non-IPNBus external data sources.

## Decisions

### 1. Add an explicit event catalog layer

Decision:
- Introduce a centralized event-type catalog in `internal/event` that declares supported event names, subject scope, and payload expectations.

Rationale:
- Avoid ad-hoc string literals spread across detectors and notifier routing.
- Provide one place to evolve docs, validation, and route matching behavior.

Alternatives considered:
- Keep free-form event type strings only in detectors. Rejected due to drift risk and validation complexity.

### 2. Expand normalized tracked state in source/snapshot

Decision:
- Extend normalized tracked fields for peers and selected tailnet-level values so non-presence transitions can be diffed deterministically.
- Keep volatile fields excluded to preserve no-op suppression behavior.

Rationale:
- Additional events require deterministic before/after comparisons on stable fields.
- Existing hash/no-op logic remains useful if field selection is disciplined.

Alternatives considered:
- Emit directly from raw `ipn.Notify` without normalization. Rejected because it bypasses current deterministic diff/idempotency model.

### 3. Add focused detectors for new event families

Decision:
- Keep the detector model and add focused detectors (or extend existing detectors) for route/tag/identity and daemon/prefs transitions.
- Preserve detector ordering and enablement behavior via config.

Rationale:
- Fits current architecture and keeps implementation testable by capability.
- Minimizes risk to policy and notifier stages.

Alternatives considered:
- Single monolithic detector for all changes. Rejected for maintainability and test isolation.

### 4. Implement wildcard route matching in notifier routing

Decision:
- Treat `*` in `routes[].event_types` as match-all.
- If a route contains `*` plus explicit literals, route still behaves as match-all.
- Retain explicit literal matching unchanged when `*` is absent.

Rationale:
- Clear operator semantics and low cognitive overhead.
- Backward-compatible behavior for existing route definitions.

Alternatives considered:
- Introduce prefix/glob matching beyond `*`. Deferred to keep semantics simple and low-risk.

### 5. Validate and document new routing/event semantics

Decision:
- Update config validation to accept wildcard event types and reject invalid/empty route type definitions.
- Update docs/example config to show both explicit routing and wildcard routing patterns.

Rationale:
- Prevent runtime surprises and preserve current operator workflow.

Alternatives considered:
- Runtime-only permissive handling with no validation changes. Rejected due to poor operator feedback.

## Risks / Trade-offs

- [Event volume growth from expanded catalog] -> Mitigation: preserve no-op suppression, debounce/suppression/rate-limit policy, and encourage route scoping.
- [Field instability causing noisy diffs] -> Mitigation: keep volatile fields excluded from normalized tracked state and add tests for stable hashing behavior.
- [Config ambiguity around `*`] -> Mitigation: deterministic semantics (`*` => match-all), explicit docs, and validation tests.
- [Behavioral regressions in existing presence flows] -> Mitigation: preserve current presence detector behavior and add regression coverage for `peer.online/offline`.

## Migration Plan

1. Add event catalog constants/types and tests for supported event names.
2. Extend normalized state inputs to include additional diffable fields.
3. Implement/extend detectors for new event families and add unit tests per family.
4. Update notifier route matcher for wildcard semantics and add routing tests.
5. Update config validation/defaults and docs examples.
6. Run integration tests for presence + non-presence events end-to-end through notifier.
7. Rollback strategy: disable new detectors (or route only explicit presence events) and keep wildcard unused in config.

## Open Questions

- Should wildcard matching be limited to only `*` as the sole element, or remain permissive when combined with literals?
- Which expanded events should be enabled by default versus gated by detector enablement to control noise?
- Do we need a per-event-family severity map in config now, or keep current fixed severity model for this iteration?
