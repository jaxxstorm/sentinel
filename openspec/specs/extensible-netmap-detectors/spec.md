# extensible-netmap-detectors Specification

## Purpose
TBD - created by archiving change implement-basic-functionality. Update Purpose after archive.
## Requirements
### Requirement: Sentinel provides a detector plugin contract
Sentinel SHALL define a detector interface that accepts normalized before/after snapshots and returns typed events for a single change family.

#### Scenario: Presence detector implements contract
- **WHEN** Sentinel initializes the v0 presence detector
- **THEN** the detector is invoked through the shared detector interface and returns only presence events

### Requirement: Detector execution order is deterministic and configurable
Sentinel SHALL execute enabled detectors in a deterministic order defined by configuration.

#### Scenario: Configured detector order is respected
- **WHEN** detectors are configured as `presence`, `routes`, `tags`
- **THEN** Sentinel runs detectors in that exact sequence for each diff cycle

### Requirement: Detector enablement is configuration-driven
Sentinel SHALL allow each detector to be enabled or disabled through configuration without code changes.

#### Scenario: Disabled detector does not run
- **WHEN** the `routes` detector is configured as disabled
- **THEN** Sentinel skips routes detection and emits no route-change events

### Requirement: New detectors can be added without redesigning core flow
Sentinel SHALL support adding new detectors for additional netmap changes without changing the core observer, notifier, or sink contracts.

#### Scenario: New detector emits new event type
- **WHEN** a new endpoint-change detector is registered
- **THEN** Sentinel includes its events in the same policy and notification pipeline as existing detectors

