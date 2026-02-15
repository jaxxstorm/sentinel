## MODIFIED Requirements

### Requirement: Sentinel routes events to configured sinks
Sentinel SHALL route emitted events to one or more configured notification sinks based on routing rules that match event type and severity, SHALL support wildcard event-type matching via `*`, SHALL support optional device target selectors (`names`, `tags`, `owners`, `ips`) on routes for device-scoped events, SHALL support delivery to `stdout`, `webhook`, and `discord` sink types, and operator documentation SHALL describe sink configuration and delivery visibility semantics.

#### Scenario: Matching route delivers to sink
- **WHEN** a `peer.online` event matches a route targeting `webhook-primary`
- **THEN** Sentinel enqueues the event for delivery to `webhook-primary`

#### Scenario: Wildcard event route delivers expanded event families
- **WHEN** a route includes `event_types: ["*"]` and Sentinel emits an expanded non-presence event such as `peer.routes.changed`
- **THEN** Sentinel enqueues that event for delivery to the route's configured sinks

#### Scenario: Device name selector route matches targeted device event
- **WHEN** a route includes `device.names: ["nas-01"]` and Sentinel emits `peer.online` for `nas-01`
- **THEN** Sentinel enqueues the event for delivery to the route's configured sinks

#### Scenario: Device tag selector route excludes non-matching device event
- **WHEN** a route includes `device.tags: ["tag:prod"]` and Sentinel emits a device-scoped event for a device without `tag:prod`
- **THEN** Sentinel does not enqueue that event for the route

#### Scenario: Device owner selector route matches targeted device event
- **WHEN** a route includes `device.owners: ["123"]` and Sentinel emits a device-scoped event whose normalized owner identity includes `123`
- **THEN** Sentinel enqueues the event for delivery to the route's configured sinks

#### Scenario: Device IP selector route matches targeted device event
- **WHEN** a route includes `device.ips: ["100.64.0.10"]` and Sentinel emits a device-scoped event whose device identity includes `100.64.0.10`
- **THEN** Sentinel enqueues the event for delivery to the route's configured sinks

#### Scenario: Device selector route does not match non-device events
- **WHEN** a route includes `device` selectors and Sentinel emits a non-device event such as `daemon.state.changed`
- **THEN** Sentinel does not enqueue that event for the route

#### Scenario: Discord sink receives routed event
- **WHEN** a route targets a `discord` sink and a matching event is emitted
- **THEN** Sentinel sends the event to the configured Discord webhook endpoint

#### Scenario: Sink behavior is documented for operators
- **WHEN** an operator reads sink documentation
- **THEN** docs explain `stdout/debug` event output, webhook retry behavior, Discord delivery behavior, and delivery success/failure log records
