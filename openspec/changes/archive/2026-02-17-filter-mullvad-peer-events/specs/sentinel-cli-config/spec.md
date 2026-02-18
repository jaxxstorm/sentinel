## ADDED Requirements

### Requirement: Config loading SHALL validate notification route filter fields
Sentinel SHALL accept optional `notifier.routes[].filters.include` and `notifier.routes[].filters.exclude` field groups with `device_names`, `tags`, `ips`, and `events` values, SHALL validate each configured filter entry, SHALL support legacy `notifier.routes[].device.*` fields as backward-compatible aliases, and SHALL expose documented examples showing global filtering and Mullvad suppression.

#### Scenario: Include/exclude filter route is valid
- **WHEN** an operator configures `notifier.routes[].filters.include.tags` and `notifier.routes[].filters.exclude.device_names` with non-empty values
- **THEN** Sentinel accepts the configuration and includes those filters in runtime route wiring

#### Scenario: Invalid filter value is rejected
- **WHEN** an operator configures `notifier.routes[].filters.include.ips` with an invalid IP/CIDR value
- **THEN** Sentinel fails configuration validation with a route filter specific error

#### Scenario: Unknown filter event value is rejected
- **WHEN** an operator configures `notifier.routes[].filters.include.events` with an unknown event type value
- **THEN** Sentinel fails configuration validation with a route filter event-specific error

#### Scenario: Legacy device selector config remains valid
- **WHEN** an operator configures legacy `notifier.routes[].device.names`, `device.tags`, or `device.ips` fields
- **THEN** Sentinel accepts the configuration and maps those fields to equivalent route include filters

#### Scenario: Documentation includes global filter and Mullvad examples
- **WHEN** an operator reads configuration documentation for notifier routes
- **THEN** documentation includes examples for device/tag/IP/event filtering, include/exclude usage, suppressing `*.mullvad.ts.net` events, and a full event type reference
