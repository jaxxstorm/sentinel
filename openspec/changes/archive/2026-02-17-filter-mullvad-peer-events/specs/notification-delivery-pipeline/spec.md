## MODIFIED Requirements

### Requirement: Sentinel routes events to configured sinks
Sentinel SHALL route emitted events to one or more configured notification sinks based on routing rules that match event type and severity, SHALL support wildcard event-type matching via `*`, SHALL support optional route-level notification filters for device names, tags, IPs, and event types with include/exclude semantics, SHALL preserve compatibility with legacy `device` selector fields by mapping them to include filters, SHALL support delivery to `stdout`, `webhook`, and `discord` sink types, and operator documentation SHALL describe sink configuration and filter-driven delivery visibility semantics.

#### Scenario: Matching route delivers to sink
- **WHEN** a `peer.online` event matches a route targeting `webhook-primary`
- **THEN** Sentinel enqueues the event for delivery to `webhook-primary`

#### Scenario: Wildcard event route delivers expanded event families
- **WHEN** a route includes `event_types: ["*"]` and Sentinel emits an expanded non-presence event such as `peer.routes.changed`
- **THEN** Sentinel enqueues that event for delivery to the route's configured sinks

#### Scenario: Include tag filter matches targeted device event
- **WHEN** a route includes `filters.include.tags: ["tag:prod"]` and Sentinel emits a device-scoped event for a device carrying `tag:prod`
- **THEN** Sentinel enqueues the event for delivery to the route's configured sinks

#### Scenario: Include IP filter matches targeted device event
- **WHEN** a route includes `filters.include.ips: ["100.64.0.10"]` and Sentinel emits a device-scoped event whose identity includes `100.64.0.10`
- **THEN** Sentinel enqueues the event for delivery to the route's configured sinks

#### Scenario: Exclude filter suppresses noisy device class
- **WHEN** a route includes `filters.exclude.device_names: ["*.mullvad.ts.net"]` and Sentinel emits a matching peer-scoped event
- **THEN** Sentinel does not enqueue that event for the route

#### Scenario: Exclude filter takes precedence over include match
- **WHEN** a route has both include and exclude filters that match the same event
- **THEN** Sentinel does not enqueue the event for that route

#### Scenario: Include events filter scopes wildcard route
- **WHEN** a route includes `event_types: ["*"]` and `filters.include.events: ["daemon.state.changed"]`
- **THEN** Sentinel enqueues only matching daemon-state events for that route

#### Scenario: Exclude events filter suppresses selected event family
- **WHEN** a route includes `event_types: ["*"]` and `filters.exclude.events: ["peer.tags.changed"]`
- **THEN** Sentinel does not enqueue `peer.tags.changed` events for that route

#### Scenario: Legacy device selectors still route correctly
- **WHEN** a route uses legacy `device.names`, `device.tags`, or `device.ips` fields and Sentinel emits a matching event
- **THEN** Sentinel evaluates and routes the event with equivalent include-filter behavior

#### Scenario: Sink behavior is documented for operators
- **WHEN** an operator reads sink documentation
- **THEN** docs explain `stdout/debug` event output, webhook retry behavior, Discord delivery behavior, and delivery success/failure log records
