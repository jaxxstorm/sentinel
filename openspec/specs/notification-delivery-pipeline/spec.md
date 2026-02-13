# notification-delivery-pipeline Specification

## Purpose
TBD - created by archiving change implement-basic-functionality. Update Purpose after archive.
## Requirements
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
