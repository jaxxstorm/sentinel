## Why

Sentinel currently relies on polling and can miss or delay tailnet change detection when status snapshots are sparse or stale. Moving to IPNBus-backed realtime updates makes event detection immediate and reliable while preserving the existing diff/policy/notification pipeline.

## What Changes

- Add an IPNBus-driven observation mode that consumes `WatchIPNBus` notifications (initial state + netmap updates) from the active tsnet server.
- Use bus updates as the primary trigger for netmap normalization, diffing, policy evaluation, and notifier delivery instead of fixed-interval polling.
- Preserve the current configuration model and simplify where needed by keeping existing config keys valid while introducing clear source-mode behavior for realtime operation.
- Emit structured logs whenever a bus event is processed and whenever Sentinel emits a derived change event.
- Keep notifier behavior consistent: detected events continue to route through configured sinks, with stdout/debug as a safe default.
- Add reconnect and backoff handling for IPNBus disconnects so runtime operation remains resilient.

## Capabilities

### New Capabilities
- `ipnbus-realtime-observation`: Subscribe to and process IPNBus notifications from the embedded tsnet node as the primary runtime event source.
- `realtime-netmap-diff-pipeline`: Convert relevant bus updates into normalized snapshots and run the existing diff/policy pipeline on each meaningful update.
- `realtime-notification-and-logging`: Guarantee event-time structured logging and notifier sink dispatch for events produced from bus-driven updates.

### Modified Capabilities
- None.

## Impact

- Affected code: runtime/source integration (`internal/source`, `internal/app`, `internal/cli`), onboarding/runtime lifecycle coordination, and related tests.
- APIs/behavior: runtime execution shifts from poll-first to realtime-bus-first while preserving CLI/config compatibility and sink routing semantics.
- Dependencies: no new external dependencies expected; uses existing `tailscale.com/client/local` IPNBus APIs already in the module.
- Operational impact: lower event latency, fewer missed transitions, and clearer runtime diagnostics when source connectivity changes.
