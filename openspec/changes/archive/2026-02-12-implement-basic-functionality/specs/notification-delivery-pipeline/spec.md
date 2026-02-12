## ADDED Requirements

### Requirement: Sentinel routes events to configured sinks
Sentinel SHALL route emitted events to one or more configured notification sinks based on routing rules that match event type and severity.

#### Scenario: Matching route delivers to sink
- **WHEN** a `peer.online` event matches a route targeting `webhook-primary`
- **THEN** Sentinel enqueues the event for delivery to `webhook-primary`

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
Sentinel SHALL support dry-run mode that evaluates routing and policy decisions but MUST NOT send outbound sink requests.

#### Scenario: Dry-run reports intended notification
- **WHEN** Sentinel runs with `--dry-run` and an event reaches delivery stage
- **THEN** Sentinel logs or prints the intended sink action without executing network delivery

