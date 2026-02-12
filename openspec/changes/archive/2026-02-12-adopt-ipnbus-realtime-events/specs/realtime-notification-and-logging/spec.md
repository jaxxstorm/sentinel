## ADDED Requirements

### Requirement: Sentinel SHALL log bus-driven processing events with structured fields
Sentinel SHALL emit structured logs for realtime bus notification handling and derived netmap processing outcomes.

#### Scenario: Bus update processing emits structured runtime log
- **WHEN** Sentinel processes an IPNBus notification in realtime mode
- **THEN** Sentinel logs an event including stable fields identifying the bus/update context and processing result

### Requirement: Sentinel SHALL dispatch derived events to notifier sinks in realtime mode
Sentinel SHALL route bus-derived Sentinel events through the existing notifier configuration and sink routing behavior.

#### Scenario: Realtime presence event is delivered to configured sinks
- **WHEN** a bus-driven diff produces a notifiable event
- **THEN** Sentinel sends that event to each configured route sink that matches event type and severity

### Requirement: Sentinel SHALL preserve default sink behavior for realtime events
Sentinel SHALL ensure realtime-derived events are still emitted to the default stdout/debug sink when no external sink is available.

#### Scenario: Realtime event still surfaces without webhook configuration
- **WHEN** no valid webhook sink is configured
- **THEN** Sentinel emits realtime-derived notification output through stdout/debug sink behavior

### Requirement: Sentinel SHALL apply existing policy and idempotency controls to realtime notifications
Sentinel SHALL apply suppression, batching, rate limiting, and idempotency safeguards before sending bus-derived notifications.

#### Scenario: Duplicate realtime event is suppressed
- **WHEN** a bus-derived event resolves to an idempotency key that is already recorded
- **THEN** Sentinel suppresses duplicate outbound sink delivery for that event
