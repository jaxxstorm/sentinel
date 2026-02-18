## MODIFIED Requirements

### Requirement: Environment-driven configuration parsing SHALL be deterministic and validated
Sentinel SHALL provide deterministic parsing and explicit validation errors for malformed environment-sourced structured configuration values, SHALL define canonical shorthand route env keys for route-scoped values, SHALL continue to accept deprecated shorthand aliases for backward compatibility, and SHALL use canonical shorthand values when both canonical and alias keys are set for the same field.

#### Scenario: Invalid structured environment value fails with actionable error
- **WHEN** an operator supplies malformed structured config content in an environment variable
- **THEN** Sentinel fails config loading with an error that identifies the environment key and validation/parsing reason

#### Scenario: Canonical shorthand route key is used when both canonical and alias are set
- **WHEN** an operator sets both `SENTINEL_NOTIFIER_ROUTE_EVENT_TYPES` and deprecated `SENTINEL_NOTIFIER_ROUTE_EVENT_TYPE`
- **THEN** Sentinel resolves shorthand route event types from `SENTINEL_NOTIFIER_ROUTE_EVENT_TYPES`

#### Scenario: Legacy shorthand route key is used when canonical is absent
- **WHEN** an operator sets deprecated `SENTINEL_NOTIFIER_SINK` and does not set `SENTINEL_NOTIFIER_ROUTE_SINKS`
- **THEN** Sentinel accepts the alias and resolves shorthand route sinks from that value
