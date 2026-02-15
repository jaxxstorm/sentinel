# notifier-device-target-filtering Specification

## Purpose
TBD - created by archiving change add-notification-device-tag-ip-filters. Update Purpose after archive.

## Requirements
### Requirement: Notifier routes SHALL support device target selectors
Sentinel SHALL support an optional `notifier.routes[].device` selector object with `names`, `tags`, `owners`, and `ips` fields so routes can target subsets of device-scoped events.

#### Scenario: Route matches device event by name selector
- **WHEN** a route defines `device.names: ["nas-01"]` and Sentinel emits a device-scoped event for `nas-01`
- **THEN** Sentinel treats the route as matched (subject to other route filters)

#### Scenario: Route matches device event by tag selector
- **WHEN** a route defines `device.tags: ["tag:prod"]` and Sentinel emits a device-scoped event for a device carrying `tag:prod`
- **THEN** Sentinel treats the route as matched (subject to other route filters)

#### Scenario: Route matches device event by owner selector
- **WHEN** a route defines `device.owners: ["123"]` and Sentinel emits a device-scoped event for a device with normalized owner identity value `123`
- **THEN** Sentinel treats the route as matched (subject to other route filters)

#### Scenario: Route matches device event by IP selector
- **WHEN** a route defines `device.ips: ["100.64.0.10"]` and Sentinel emits a device-scoped event for a device whose normalized identity IP set contains `100.64.0.10`
- **THEN** Sentinel treats the route as matched (subject to other route filters)

### Requirement: Device selector semantics SHALL be deterministic and backward-compatible
Sentinel SHALL evaluate device selector lists with OR semantics inside each field and AND semantics across configured fields, and routes without a device selector SHALL preserve existing behavior.

#### Scenario: Route without device selector behaves as before
- **WHEN** a route defines only `event_types`/`severities` and omits `device`
- **THEN** route matching behavior remains equivalent to prior releases

#### Scenario: Multi-field selector requires all configured dimensions
- **WHEN** a route defines `device.names` and `device.tags` and an event matches name but not tag
- **THEN** Sentinel does not match the route

#### Scenario: Device selector does not match non-device subject events
- **WHEN** a route defines `device` selectors and Sentinel emits a non-device event such as `daemon.state.changed`
- **THEN** Sentinel does not match the route
