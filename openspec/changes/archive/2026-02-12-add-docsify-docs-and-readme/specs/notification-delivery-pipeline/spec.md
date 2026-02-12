## MODIFIED Requirements

### Requirement: Sentinel routes events to configured sinks
Sentinel SHALL route emitted events to one or more configured notification sinks based on routing rules that match event type and severity, and operator documentation SHALL describe sink configuration and delivery visibility semantics.

#### Scenario: Matching route delivers to sink
- **WHEN** a `peer.online` event matches a route targeting `webhook-primary`
- **THEN** Sentinel enqueues the event for delivery to `webhook-primary`

#### Scenario: Sink behavior is documented for operators
- **WHEN** an operator reads sink documentation
- **THEN** docs explain `stdout/debug` event output, webhook retry behavior, and delivery success/failure log records

### Requirement: Dry-run mode prevents external delivery
Sentinel SHALL support dry-run mode that evaluates routing and policy decisions but MUST NOT send outbound sink requests, and docs SHALL include a dry-run validation workflow.

#### Scenario: Dry-run reports intended notification
- **WHEN** Sentinel runs with `--dry-run` and an event reaches delivery stage
- **THEN** Sentinel logs or prints the intended sink action without executing network delivery

#### Scenario: Docs include dry-run verification steps
- **WHEN** an operator follows notification testing docs
- **THEN** they can validate routing behavior using `test-notify` and `--dry-run` without external webhook side effects
