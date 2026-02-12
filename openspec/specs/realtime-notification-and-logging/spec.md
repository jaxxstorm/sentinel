# realtime-notification-and-logging Specification

## Purpose
TBD - created by archiving change adopt-ipnbus-realtime-events. Update Purpose after archive.

## Requirements
### Requirement: Sentinel SHALL log bus-driven processing events with structured fields
Sentinel SHALL emit structured logs for realtime bus notification handling and derived netmap processing outcomes, and SHALL include stable source attribution fields to distinguish Sentinel-generated and Tailscale-originated runtime records.

#### Scenario: Bus update processing emits structured runtime log
- **WHEN** Sentinel processes an IPNBus notification in realtime mode
- **THEN** Sentinel logs an event including stable fields identifying the bus/update context and processing result

#### Scenario: Realtime processing record includes source attribution
- **WHEN** Sentinel emits a realtime bus-processing log record
- **THEN** the record includes `log_source=sentinel` or equivalent stable source attribution for runtime filtering

### Requirement: Sentinel SHALL dispatch derived events to notifier sinks in realtime mode
Sentinel SHALL route bus-derived Sentinel events through the existing notifier configuration and sink routing behavior.

#### Scenario: Realtime presence event is delivered to configured sinks
- **WHEN** a bus-driven diff produces a notifiable event
- **THEN** Sentinel sends that event to each configured route sink that matches event type and severity

### Requirement: Sentinel SHALL preserve default sink behavior for realtime events
Sentinel SHALL ensure realtime-derived events are still emitted to the default stdout/debug sink when no external sink is available, and sink-visible runtime output SHALL preserve machine-readable event JSON with explicit source attribution.

#### Scenario: Realtime event still surfaces without webhook configuration
- **WHEN** no valid webhook sink is configured
- **THEN** Sentinel emits realtime-derived notification output through stdout/debug sink behavior

#### Scenario: Sink output includes source attribution
- **WHEN** a realtime-derived event is emitted via stdout/debug sink
- **THEN** the emitted output includes stable attribution that the record originated from the sink path

### Requirement: Sentinel SHALL apply existing policy and idempotency controls to realtime notifications
Sentinel SHALL apply suppression, batching, rate limiting, and idempotency safeguards before sending bus-derived notifications.

#### Scenario: Duplicate realtime event is suppressed
- **WHEN** a bus-derived event resolves to an idempotency key that is already recorded
- **THEN** Sentinel suppresses duplicate outbound sink delivery for that event
