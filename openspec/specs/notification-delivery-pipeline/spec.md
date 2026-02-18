# notification-delivery-pipeline Specification

## Purpose
TBD - created by archiving change implement-basic-functionality. Update Purpose after archive.
## Requirements
### Requirement: Sentinel routes events to configured sinks
Sentinel SHALL route emitted events to one or more configured notification sinks based on routing rules that match event type and severity, SHALL support wildcard event-type matching via `*`, SHALL support optional route-level notification filters for device names, tags, IPs, and event types with include/exclude semantics, SHALL preserve compatibility with legacy `device` selector fields by mapping `device.names`/`device.tags`/`device.ips` to include filters while retaining `device.owners` behavior, SHALL support delivery to `stdout`, `webhook`, and `discord` sink types, and operator documentation SHALL describe sink configuration and filter-driven delivery visibility semantics.

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

#### Scenario: Device owner selector route matches targeted device event
- **WHEN** a route includes `device.owners: ["123"]` and Sentinel emits a device-scoped event whose normalized owner identity includes `123`
- **THEN** Sentinel enqueues the event for delivery to the route's configured sinks

#### Scenario: Device selector route does not match non-device events
- **WHEN** a route includes identity-based filters (`filters.include.device_names` or `filters.include.tags` or `filters.include.ips`) and Sentinel emits a non-device event such as `daemon.state.changed`
- **THEN** Sentinel does not enqueue that event for the route

#### Scenario: Legacy device selectors still route correctly
- **WHEN** a route uses legacy `device.names`, `device.tags`, or `device.ips` fields and Sentinel emits a matching event
- **THEN** Sentinel evaluates and routes the event with equivalent include-filter behavior

#### Scenario: Discord sink receives routed event
- **WHEN** a route targets a `discord` sink and a matching event is emitted
- **THEN** Sentinel sends the event to the configured Discord webhook endpoint

#### Scenario: Sink behavior is documented for operators
- **WHEN** an operator reads sink documentation
- **THEN** docs explain `stdout/debug` event output, webhook retry behavior, Discord delivery behavior, and delivery success/failure log records

### Requirement: Sentinel enforces notification idempotency
Sentinel SHALL compute an idempotency key per notification attempt and suppress duplicate deliveries when the same key was already sent within the configured retention window.

#### Scenario: Duplicate event is suppressed
- **WHEN** an event with an idempotency key already recorded in state is processed again
- **THEN** Sentinel does not deliver it to sinks and records it as suppressed

### Requirement: Sentinel applies noise-control policies before delivery
Sentinel SHALL apply debounce, suppression windows, rate limiting, and batching policies before sink delivery.

#### Scenario: Debounce suppresses flapping peer events
- **WHEN** the same peer toggles online/offline repeatedly within the debounce window
- **THEN** Sentinel suppresses intermediate events and emits only the debounced result

### Requirement: Dry-run mode prevents external delivery
Sentinel SHALL support dry-run mode that evaluates routing and policy decisions but MUST NOT send outbound sink requests, and docs SHALL include a dry-run validation workflow.

#### Scenario: Dry-run reports intended notification
- **WHEN** Sentinel runs with `--dry-run` and an event reaches delivery stage
- **THEN** Sentinel logs or prints the intended sink action without executing network delivery

#### Scenario: Docs include dry-run verification steps
- **WHEN** an operator follows notification testing docs
- **THEN** they can validate routing behavior using `test-notify` and `--dry-run` without external webhook side effects
