## Why

Sentinel currently emits only peer presence events, which leaves a lot of useful IPNBus and NetMap signal unavailable to operators. We need richer event coverage and simpler routing controls now that realtime IPNBus ingestion is in place.

## What Changes

- Expand event generation beyond `peer.online` and `peer.offline` to include additional peer, netmap, prefs, and daemon lifecycle changes sourced from `ipn.Notify`.
- Define a stable event type catalog and payload conventions for newly supported events so sink integrations can rely on predictable structure.
- Update notifier route matching so `notifier.routes[].event_types` supports both explicit lists and `*` as a wildcard for all event types.
- Preserve current behavior for existing configs that list explicit event types, while allowing operators to opt into wildcard routing without breaking existing route definitions.

## Capabilities

### New Capabilities
- `ipnbus-event-catalog`: Define and emit an expanded set of observable event types from IPNBus/NetMap updates.
- `notifier-event-type-wildcard-routing`: Support `*` route matching for event types in notifier routing.

### Modified Capabilities
- `realtime-netmap-diff-pipeline`: Extend diffing from presence-only output to broader event extraction and normalization.
- `notification-delivery-pipeline`: Update route filtering semantics to include wildcard event type matching.
- `sentinel-cli-config`: Expand config validation and defaults to recognize new event types and wildcard routing.

## Impact

- Affected code: `internal/source`, `internal/diff`, `internal/event`, `internal/notify`, `internal/config`, and related tests.
- Operator-facing behavior: more event types can be emitted and routed; `event_types: ["*"]` becomes valid.
- Compatibility: existing explicit event routes remain supported; wildcard support is additive.
