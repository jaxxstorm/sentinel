## MODIFIED Requirements

### Requirement: Config loading supports YAML, JSON, and environment overrides
Sentinel SHALL load configuration from YAML or JSON files and MUST apply environment variable overrides using the `SENTINEL_` prefix, SHALL validate expanded event type values including wildcard `*` for route event matching, and operator documentation SHALL include working configuration examples that reflect this behavior.

#### Scenario: Environment overrides file value
- **WHEN** `poll_interval` is set in config file and `SENTINEL_POLL_INTERVAL` is also set
- **THEN** Sentinel uses the environment-provided value at runtime

#### Scenario: Wildcard event type route is valid
- **WHEN** an operator configures `notifier.routes[].event_types` as `['*']`
- **THEN** Sentinel accepts the config as valid and enables wildcard event matching at runtime

#### Scenario: Documentation includes env interpolation examples
- **WHEN** an operator reads the configuration documentation
- **THEN** the docs include explicit examples for `${VAR_NAME}` interpolation in sink URLs and `SENTINEL_` prefixed overrides
