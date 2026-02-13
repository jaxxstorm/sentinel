## MODIFIED Requirements

### Requirement: Sentinel routes events to configured sinks
Sentinel SHALL route emitted events to one or more configured notification sinks based on routing rules that match event type and severity, SHALL support wildcard event-type matching via `*`, and operator documentation SHALL describe sink configuration and delivery visibility semantics.

#### Scenario: Matching route delivers to sink
- **WHEN** a `peer.online` event matches a route targeting `webhook-primary`
- **THEN** Sentinel enqueues the event for delivery to `webhook-primary`

#### Scenario: Wildcard event route delivers expanded event families
- **WHEN** a route includes `event_types: ["*"]` and Sentinel emits an expanded non-presence event such as `peer.routes.changed`
- **THEN** Sentinel enqueues that event for delivery to the route's configured sinks

#### Scenario: Sink behavior is documented for operators
- **WHEN** an operator reads sink documentation
- **THEN** docs explain `stdout/debug` event output, webhook retry behavior, and delivery success/failure log records
