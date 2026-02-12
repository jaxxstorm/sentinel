# realtime-netmap-diff-pipeline Specification

## Purpose
TBD - created by archiving change adopt-ipnbus-realtime-events. Update Purpose after archive.
## Requirements
### Requirement: Sentinel SHALL derive normalized snapshots from realtime bus netmap updates
Sentinel SHALL convert relevant IPNBus netmap notifications into Sentinel snapshot inputs using the existing normalization rules before diff execution.

#### Scenario: Netmap update becomes normalized snapshot
- **WHEN** a watch notification contains updated netmap peer state
- **THEN** Sentinel normalizes that state into a new snapshot and computes a deterministic snapshot hash

### Requirement: Sentinel SHALL execute diff and policy flow per meaningful realtime update
Sentinel SHALL run detector and policy evaluation when realtime updates produce meaningful snapshot changes.

#### Scenario: Realtime peer transition triggers diff processing
- **WHEN** a bus-driven snapshot reflects a peer online or offline transition
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

