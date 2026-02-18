## MODIFIED Requirements

### Requirement: Config loading supports YAML, JSON, and environment overrides
Sentinel SHALL load configuration from YAML or JSON files, MUST apply environment variable overrides using the `SENTINEL_` prefix, SHALL support environment-only configuration for full runtime operation (including structured fields), SHALL validate expanded event type values including wildcard `*` for route event matching, SHALL validate sink type-specific configuration for supported sink types including `discord`, SHALL validate notifier device-target selector fields (`notifier.routes[].device.names`, `notifier.routes[].device.tags`, `notifier.routes[].device.owners`, `notifier.routes[].device.ips`), SHALL validate Tailscale onboarding fields including advertise tags and OAuth credential combinations, SHALL define canonical shorthand route env keys using the `SENTINEL_NOTIFIER_ROUTE_*` namespace, SHALL treat legacy shorthand aliases (`SENTINEL_NOTIFIER_ROUTE_EVENT_TYPE` and `SENTINEL_NOTIFIER_SINK`) as deprecated compatibility inputs, SHALL prefer canonical shorthand keys when both canonical and legacy aliases are present, and operator documentation SHALL include working configuration examples and migration notes that reflect this behavior.

#### Scenario: Environment overrides file value
- **WHEN** `poll_interval` is set in config file and `SENTINEL_POLL_INTERVAL` is also set
- **THEN** Sentinel uses the environment-provided value at runtime

#### Scenario: Environment-only config without file is accepted
- **WHEN** no config file is supplied and required runtime settings are provided via `SENTINEL_` environment variables
- **THEN** Sentinel loads, validates, and runs with environment-provided configuration only

#### Scenario: Structured config can be supplied via environment
- **WHEN** list/object config sections (for example notifier sinks/routes) are provided via documented structured `SENTINEL_` environment variables
- **THEN** Sentinel parses the structured values and applies equivalent runtime behavior to file-based config

#### Scenario: Wildcard event type route is valid
- **WHEN** an operator configures `notifier.routes[].event_types` as `['*']`
- **THEN** Sentinel accepts the config as valid and enables wildcard event matching at runtime

#### Scenario: Device selector route is valid
- **WHEN** an operator configures `notifier.routes[].device` with non-empty `names`, `tags`, `owners`, and/or `ips` values
- **THEN** Sentinel accepts the config as valid and includes selector fields in runtime route wiring

#### Scenario: Invalid device selector IP is rejected
- **WHEN** an operator configures `notifier.routes[].device.ips` with an invalid IP literal
- **THEN** Sentinel fails validation with a route selector-specific configuration error

#### Scenario: Discord sink configuration is valid
- **WHEN** an operator configures a sink with `type: discord` and a non-empty Discord webhook URL
- **THEN** Sentinel accepts the sink configuration and includes it in notifier wiring

#### Scenario: Invalid Discord sink configuration is rejected
- **WHEN** an operator configures a sink with `type: discord` but omits the webhook URL
- **THEN** Sentinel fails validation with a sink-specific configuration error

#### Scenario: Invalid structured environment value is rejected
- **WHEN** an operator sets malformed structured config content in a supported `SENTINEL_` env key
- **THEN** Sentinel fails config loading with a parse/validation error that identifies the offending env key

#### Scenario: Canonical route sink shorthand key is honored
- **WHEN** an operator sets `SENTINEL_NOTIFIER_ROUTE_SINKS` for shorthand route configuration
- **THEN** Sentinel uses that value as the route sink list for the appended shorthand route

#### Scenario: Legacy route sink shorthand alias remains compatible
- **WHEN** an operator sets `SENTINEL_NOTIFIER_SINK` and does not set `SENTINEL_NOTIFIER_ROUTE_SINKS`
- **THEN** Sentinel accepts the alias and uses it as shorthand route sink input

#### Scenario: Canonical shorthand key wins over legacy alias
- **WHEN** an operator sets both `SENTINEL_NOTIFIER_ROUTE_SINKS` and `SENTINEL_NOTIFIER_SINK`
- **THEN** Sentinel resolves shorthand route sinks from `SENTINEL_NOTIFIER_ROUTE_SINKS` and ignores the legacy alias

#### Scenario: Advertise tags are accepted from config and env
- **WHEN** an operator configures `tsnet.advertise_tags` in file or with the mapped `SENTINEL_` environment key
- **THEN** Sentinel validates and loads those tags into runtime tsnet configuration

#### Scenario: Invalid advertise tag format is rejected
- **WHEN** an operator configures an advertise tag that does not match required Tailscale tag format
- **THEN** Sentinel fails validation with a tag-specific configuration error

#### Scenario: OAuth client secret credential inputs are validated
- **WHEN** an operator configures OAuth credential fields for tsnet onboarding
- **THEN** Sentinel accepts complete valid combinations and rejects incomplete combinations with actionable validation errors

#### Scenario: Documentation includes canonical shorthand notifier examples
- **WHEN** an operator reads the configuration documentation
- **THEN** the docs use canonical shorthand route env keys in examples and include migration guidance for deprecated legacy aliases
