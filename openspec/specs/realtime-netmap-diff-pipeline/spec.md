# realtime-netmap-diff-pipeline Specification

## Purpose
TBD - created by archiving change adopt-ipnbus-realtime-events. Update Purpose after archive.
## Requirements
### Requirement: Sentinel SHALL derive normalized snapshots from realtime bus netmap updates
Sentinel SHALL convert relevant IPNBus netmap notifications into Sentinel snapshot inputs using the existing normalization rules before diff execution.

#### Scenario: Netmap update becomes normalized snapshot
- **WHEN** a watch notification contains updated netmap peer state
- **THEN** Sentinel normalizes that state into a new snapshot and computes a deterministic snapshot hash

### Requirement: Sentinel SHALL track additional normalized change dimensions in realtime mode
Sentinel SHALL normalize and track additional field groups required for expanded event emission, including peer routes, peer tags, selected peer identity attributes, and selected tailnet metadata.

#### Scenario: Route change affects normalized tracked dimensions
- **WHEN** a realtime update changes only a tracked route or tag field
- **THEN** Sentinel treats the update as meaningful for diff/event processing even when presence is unchanged

### Requirement: Sentinel SHALL execute diff and policy flow per meaningful realtime update
Sentinel SHALL run detector and policy evaluation when realtime updates produce meaningful normalized changes, including but not limited to presence transitions.

#### Scenario: Realtime peer transition triggers diff processing
- **WHEN** a bus-driven snapshot reflects a peer online or offline transition
- **THEN** Sentinel emits corresponding typed events through the existing diff and policy pipeline

#### Scenario: Realtime non-presence transition triggers diff processing
- **WHEN** a bus-driven update changes tracked non-presence fields such as peer routes, tags, or key-expiry indicators
- **THEN** Sentinel emits corresponding typed events through the existing diff and policy pipeline

### Requirement: Sentinel SHALL suppress no-op updates in realtime mode
Sentinel SHALL avoid unnecessary detector and notification work when consecutive bus updates do not change normalized snapshot content.

#### Scenario: Identical normalized snapshots produce no emitted events
- **WHEN** two consecutive realtime updates normalize to the same snapshot hash
- **THEN** Sentinel records no netmap-change events for that update

### Requirement: Sentinel SHALL preserve detector extensibility in realtime mode
Sentinel SHALL continue to use the configured detector ordering and enablement model for realtime-triggered diff evaluations.

#### Scenario: Disabled detector remains inactive under realtime updates
- **WHEN** a detector is disabled in configuration
- **THEN** Sentinel does not execute that detector for bus-triggered snapshot updates
