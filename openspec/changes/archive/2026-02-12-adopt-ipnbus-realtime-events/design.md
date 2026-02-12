## Context

Sentinel currently operates as a poll-driven pipeline: fetch current netmap-like state, normalize snapshots, run detectors, apply policy, and notify sinks. In practice, polling introduces latency and can miss short-lived transitions when intermediate state changes are not visible via point-in-time status reads. Recent debugging showed cases where `Status().Peer` was empty while realtime netmap updates were visible on IPNBus, creating gaps between actual tailnet changes and Sentinel events.

The project already has the required runtime pieces (diff engine, policy layer, notifier sinks, onboarding, state persistence). The design goal is to change the source trigger model to realtime bus updates without breaking existing config or notification behavior.

## Goals / Non-Goals

**Goals:**
- Make IPNBus notifications from the local tsnet server the primary runtime event source.
- Preserve current detector/policy/notifier architecture and event schemas.
- Keep existing configuration valid while simplifying source behavior toward realtime-by-default.
- Emit structured logs for bus activity and derived Sentinel events.
- Ensure reconnect/backoff behavior when the bus stream disconnects.

**Non-Goals:**
- Replacing Sentinel’s event schema or policy semantics.
- Building a TUI/watch UI in Sentinel runtime.
- Adding new non-presence detectors as part of this change.
- Removing existing sinks or changing sink contracts.

## Decisions

### 1. Use IPNBus watch stream as the primary source

Decision:
- Introduce an IPNBus-backed source implementation that holds a long-lived `WatchIPNBus` subscription and emits normalized `source.Netmap` updates when meaningful notifications arrive (`NetMap`, and relevant state/prefs changes where applicable).

Rationale:
- The stream reflects tailnet state transitions in near realtime and avoids polling blind spots.
- Matches the user’s requirement for realtime events directly from the tsnet runtime context.

Alternatives considered:
- Keep polling and only reduce interval: lower latency but still lossy and noisier.
- Fully event-driven rewrite that bypasses snapshot/diff: faster but breaks existing architecture and increases migration risk.

### 2. Keep diff/policy/notifier pipeline unchanged downstream

Decision:
- Reuse current snapshot normalization, detector execution, policy suppression/rate limiting, and notifier dispatch unchanged.

Rationale:
- Minimizes regression risk and preserves stable behavior for sinks, metrics, and events.
- Keeps changes focused on source and runtime orchestration.

Alternatives considered:
- Emit final notifications directly from raw IPNBus events: simpler path but duplicates logic and bypasses existing controls.

### 3. Maintain config compatibility with a simplified source-mode model

Decision:
- Keep existing config keys valid.
- Add/clarify a source mode concept (defaulting to realtime) while retaining poll interval settings for compatibility and optional fallback paths.

Rationale:
- Users should not need to rewrite configs for this migration.
- Simplifies operator mental model: Sentinel is realtime first.

Alternatives considered:
- Hard breaking switch to event-only config and remove polling keys immediately.

### 4. Add resilient watch lifecycle with reconnect/backoff

Decision:
- On watch termination or transient localapi failures, reconnect with exponential backoff bounded by existing backoff settings.
- Log reconnect attempts and successful resubscriptions.

Rationale:
- Long-lived streams fail in practice; runtime must self-heal.

Alternatives considered:
- Fail-fast on watch error: operationally brittle.

### 5. Log both bus-level and Sentinel-level events

Decision:
- Add structured debug/info logs for bus notifications processed (type, peer count/hash deltas where available).
- Keep existing Sentinel event logs and sink delivery logs.

Rationale:
- Required to diagnose source behavior and prove end-to-end delivery.
- Supports both human and JSON log modes.

Alternatives considered:
- Only log final Sentinel events: insufficient for diagnosing source issues.

## Risks / Trade-offs

- [Watch stream churn under unstable localapi] -> Mitigation: bounded exponential backoff, reconnect telemetry, and startup health visibility.
- [Higher event volume compared with polling] -> Mitigation: keep existing policy debounce/suppression/rate-limit and only process meaningful netmap notifications.
- [Behavioral differences between status snapshots and bus netmap payloads] -> Mitigation: normalize both to a single internal `source.Netmap` shape and cover with source-level tests.
- [Config confusion during migration] -> Mitigation: maintain backward compatibility and document source-mode defaults clearly in config examples.

## Migration Plan

1. Add IPNBus source component and unit tests for notification-to-netmap translation.
2. Integrate source into runtime loop as realtime trigger while preserving existing pipeline.
3. Keep polling path as compatibility/fallback behavior during transition.
4. Update CLI/config docs and examples to reflect realtime-first operation.
5. Add integration tests for:
   - realtime peer online/offline detection
   - reconnect/backoff behavior
   - sink delivery from bus-derived events
6. Rollout plan:
   - default realtime enabled
   - fallback to polling only on explicit mode or unrecoverable watch constraints
7. Rollback strategy:
   - switch source mode back to polling-only behavior using config/flag guard
   - retain existing state store and notifier semantics unchanged.

## Open Questions

- Should Sentinel process every `Notify` frame or only those containing `NetMap` updates?
- Do we want a configurable coalescing window for bursty netmap updates before running diff?
- Should source mode be exposed as a top-level config key or nested under `tsnet`?
- What minimum log fields are required at info level versus debug level for bus notifications?
