# netmap-presence-diffing Specification

## Purpose
TBD - created by archiving change implement-basic-functionality. Update Purpose after archive.
## Requirements
### Requirement: Sentinel polls netmap and creates normalized snapshots
Sentinel SHALL poll the tailnet netmap at a configurable interval and persist a normalized snapshot representation for diffing.

#### Scenario: Poll cycle creates snapshot
- **WHEN** `sentinel run` starts with a valid tsnet session
- **THEN** Sentinel polls netmap on each configured interval and stores a normalized snapshot

### Requirement: Sentinel emits presence transition events only on state changes
Sentinel SHALL emit a presence event only when a peer transitions between online and offline states across consecutive normalized snapshots.

#### Scenario: Online transition produces event
- **WHEN** a peer is offline in snapshot N and online in snapshot N+1
- **THEN** Sentinel emits one `peer.online` event for that peer

#### Scenario: Unchanged state produces no event
- **WHEN** a peer remains online in both snapshot N and snapshot N+1
- **THEN** Sentinel emits no presence event for that peer

### Requirement: Presence events use a versioned canonical envelope
Sentinel SHALL include a versioned event envelope with stable identifiers for each emitted presence event, including `event_id`, `event_type`, `timestamp`, `subject_id`, `before_hash`, and `after_hash`.

#### Scenario: Event payload includes required envelope fields
- **WHEN** Sentinel emits a `peer.offline` event
- **THEN** the event contains all required envelope fields and valid before/after snapshot hashes

