# ipnbus-event-catalog Specification

## Purpose
TBD - created by archiving change expand-event-types-and-wildcard-routing. Update Purpose after archive.

## Requirements
### Requirement: Sentinel SHALL emit expanded typed events from IPNBus and NetMap changes
Sentinel SHALL emit typed events beyond presence-only signals by deriving peer, daemon, preference, and tailnet-change events from `ipn.Notify` and `NetMap` updates.

#### Scenario: Peer route update emits typed event
- **WHEN** a peer's `PrimaryRoutes` value changes between consecutive normalized updates
- **THEN** Sentinel emits a `peer.routes.changed` event with peer identity and before/after route summaries

#### Scenario: Daemon state transition emits typed event
- **WHEN** an IPNBus notification includes a new daemon state value different from the previous observed state
- **THEN** Sentinel emits a `daemon.state.changed` event describing previous and current states

### Requirement: Sentinel SHALL provide a stable event taxonomy and payload baseline
Sentinel SHALL define a stable catalog of supported event types and SHALL attach consistent payload keys per event family so sink integrations can process events without source-specific branching.

#### Scenario: Event type catalog includes non-presence families
- **WHEN** operators configure routing for emitted events
- **THEN** documented event types include peer lifecycle, route/tag/identity, daemon lifecycle, prefs, and tailnet metadata change families

#### Scenario: Payload includes normalized subject and diff context
- **WHEN** Sentinel emits an event in the expanded catalog
- **THEN** the event contains `event_type`, `subject_id`, `subject_type`, and payload fields sufficient to identify changed attributes

### Requirement: Sentinel SHALL suppress duplicate emits for unchanged semantic state
Sentinel SHALL avoid emitting expanded events when a notification does not produce semantic change in normalized tracked fields.

#### Scenario: Repeated equivalent notify frame produces no event
- **WHEN** consecutive notifications carry identical effective values for tracked event fields
- **THEN** Sentinel emits no additional expanded event for that field set
