## MODIFIED Requirements

### Requirement: Sentinel SHALL provide a stable event taxonomy and payload baseline
Sentinel SHALL define a stable catalog of supported event types and SHALL attach consistent payload keys per event family so sink integrations can process events without source-specific branching, and device-scoped events SHALL include stable selector identity fields (`name`, `tags`, `owners`, `ips`) for route filtering and sink context.

#### Scenario: Event type catalog includes non-presence families
- **WHEN** operators configure routing for emitted events
- **THEN** documented event types include peer lifecycle, route/tag/identity, daemon lifecycle, prefs, and tailnet metadata change families

#### Scenario: Payload includes normalized subject and diff context
- **WHEN** Sentinel emits an event in the expanded catalog
- **THEN** the event contains `event_type`, `subject_id`, `subject_type`, and payload fields sufficient to identify changed attributes

#### Scenario: Device event payload includes selector identity baseline
- **WHEN** Sentinel emits a device-scoped event
- **THEN** event payload includes stable `name`, `tags`, `owners`, and `ips` fields aligned with notifier device selector matching
