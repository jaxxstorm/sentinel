## MODIFIED Requirements

### Requirement: Sentinel polls netmap and creates normalized snapshots
Sentinel SHALL poll the tailnet netmap at a configurable interval and persist a normalized snapshot representation for diffing, including stable device identity fields required for routing and notification context (`name`, `tags`, `owners`, and device IP identity values).

#### Scenario: Poll cycle creates snapshot
- **WHEN** `sentinel run` starts with a valid tsnet session
- **THEN** Sentinel polls netmap on each configured interval and stores a normalized snapshot

#### Scenario: Poll snapshot includes device owner and IP identity values
- **WHEN** Sentinel normalizes poll-mode netmap peer data
- **THEN** each peer snapshot includes deterministic owner and peer IP identity values usable for notifier device selector matching
