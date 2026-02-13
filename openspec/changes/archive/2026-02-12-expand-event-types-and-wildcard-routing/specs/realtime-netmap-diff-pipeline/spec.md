## ADDED Requirements

### Requirement: Sentinel SHALL track additional normalized change dimensions in realtime mode
Sentinel SHALL normalize and track additional field groups required for expanded event emission, including peer routes, peer tags, selected peer identity attributes, and selected tailnet metadata.

#### Scenario: Route change affects normalized tracked dimensions
- **WHEN** a realtime update changes only a tracked route or tag field
- **THEN** Sentinel treats the update as meaningful for diff/event processing even when presence is unchanged

## MODIFIED Requirements

### Requirement: Sentinel SHALL execute diff and policy flow per meaningful realtime update
Sentinel SHALL run detector and policy evaluation when realtime updates produce meaningful normalized changes, including but not limited to presence transitions.

#### Scenario: Realtime peer transition triggers diff processing
- **WHEN** a bus-driven snapshot reflects a peer online or offline transition
- **THEN** Sentinel emits corresponding typed events through the existing diff and policy pipeline

#### Scenario: Realtime non-presence transition triggers diff processing
- **WHEN** a bus-driven update changes tracked non-presence fields such as peer routes, tags, or key-expiry indicators
- **THEN** Sentinel emits corresponding typed events through the existing diff and policy pipeline
