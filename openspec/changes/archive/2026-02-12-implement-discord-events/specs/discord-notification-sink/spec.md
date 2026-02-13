## ADDED Requirements

### Requirement: Sentinel SHALL provide a native Discord notification sink
Sentinel SHALL support a `discord` sink type that posts notification events to a configured Discord webhook endpoint.

#### Scenario: Discord sink delivers routed event
- **WHEN** a routed event targets a configured `discord` sink with a valid webhook URL
- **THEN** Sentinel sends the event to the Discord webhook and records the delivery as successful on 2xx responses

### Requirement: Sentinel SHALL format Discord deliveries for operator readability
Sentinel SHALL map notification events to a Discord-friendly payload structure that includes event identity and essential change context.

#### Scenario: Discord message includes core event context
- **WHEN** Sentinel emits any routed event to a Discord sink
- **THEN** the delivered Discord payload includes event type, subject id/type, severity, timestamp, and a concise payload summary

### Requirement: Sentinel SHALL retain delivery reliability semantics for Discord sink sends
Discord sink delivery SHALL use bounded retries and failure reporting semantics consistent with existing outbound sinks.

#### Scenario: Discord sink retries transient delivery failure
- **WHEN** a Discord sink request fails due to transport error or non-2xx response
- **THEN** Sentinel retries using bounded backoff and emits structured success/failure logs that identify the sink
