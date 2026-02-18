## ADDED Requirements

### Requirement: Notifier routes SHALL support global notification filters
Sentinel SHALL support an optional `notifier.routes[].filters` object with `include` and `exclude` filter groups, and each group SHALL support `device_names`, `tags`, `ips`, and `events` fields for route-level filtering.

#### Scenario: Include filters target events by tag
- **WHEN** a route defines `filters.include.tags: ["tag:prod"]` and Sentinel emits a device-scoped event for a device with `tag:prod`
- **THEN** Sentinel treats the route as matched for the include-tag dimension

#### Scenario: Include filters target events by IP
- **WHEN** a route defines `filters.include.ips: ["100.64.0.10"]` and Sentinel emits a device-scoped event whose identity includes `100.64.0.10`
- **THEN** Sentinel treats the route as matched for the include-IP dimension

#### Scenario: Exclude filters suppress noisy Mullvad nodes
- **WHEN** a route defines `filters.exclude.device_names: ["*.mullvad.ts.net"]` and Sentinel emits an event for `us-slc-wg-306.mullvad.ts.net`
- **THEN** Sentinel excludes that event from route matching

#### Scenario: Include filters target explicit event families
- **WHEN** a route defines `filters.include.events: ["peer.online", "peer.offline"]` and Sentinel emits `peer.online`
- **THEN** Sentinel treats the route as matched for the include-events dimension

#### Scenario: Exclude filters suppress specific event families
- **WHEN** a route defines `filters.exclude.events: ["peer.routes.changed"]` and Sentinel emits `peer.routes.changed`
- **THEN** Sentinel excludes that event from route matching

### Requirement: Filter semantics SHALL be deterministic and backward-compatible
Sentinel SHALL evaluate list values with OR semantics within each field, SHALL evaluate configured include fields with AND semantics across fields, SHALL apply exclude filters after include evaluation with exclusion precedence, and SHALL preserve existing route behavior when no `filters` object is configured.

#### Scenario: Exclude precedence over include
- **WHEN** a route has an include match and an exclude match for the same event
- **THEN** Sentinel does not match the route

#### Scenario: Route without filters behaves as before
- **WHEN** a route omits `filters` and omits legacy device selectors
- **THEN** route matching behavior remains equivalent to prior releases

#### Scenario: Legacy device selector remains supported
- **WHEN** a route defines legacy `device.names`, `device.tags`, or `device.ips` selectors
- **THEN** Sentinel maps legacy selectors to equivalent include filter behavior and preserves existing semantics
