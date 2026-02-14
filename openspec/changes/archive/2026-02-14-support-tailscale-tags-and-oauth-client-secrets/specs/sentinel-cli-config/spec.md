## MODIFIED Requirements

### Requirement: Config loading supports YAML, JSON, and environment overrides
Sentinel SHALL load configuration from YAML or JSON files, MUST apply environment variable overrides using the `SENTINEL_` prefix, SHALL support environment-only configuration for full runtime operation (including structured fields), SHALL validate expanded event type values including wildcard `*` for route event matching, SHALL validate sink type-specific configuration for supported sink types including `discord`, SHALL validate Tailscale onboarding fields including advertise tags and OAuth credential combinations, and operator documentation SHALL include working configuration examples that reflect this behavior.

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

#### Scenario: Discord sink configuration is valid
- **WHEN** an operator configures a sink with `type: discord` and a non-empty Discord webhook URL
- **THEN** Sentinel accepts the sink configuration and includes it in notifier wiring

#### Scenario: Invalid Discord sink configuration is rejected
- **WHEN** an operator configures a sink with `type: discord` but omits the webhook URL
- **THEN** Sentinel fails validation with a sink-specific configuration error

#### Scenario: Invalid structured environment value is rejected
- **WHEN** an operator sets malformed structured config content in a supported `SENTINEL_` env key
- **THEN** Sentinel fails config loading with a parse/validation error that identifies the offending env key

#### Scenario: Advertise tags are accepted from config and env
- **WHEN** an operator configures `tsnet.advertise_tags` in file or with the mapped `SENTINEL_` environment key
- **THEN** Sentinel validates and loads those tags into runtime tsnet configuration

#### Scenario: Invalid advertise tag format is rejected
- **WHEN** an operator configures an advertise tag that does not match required Tailscale tag format
- **THEN** Sentinel fails validation with a tag-specific configuration error

#### Scenario: OAuth client secret credential inputs are validated
- **WHEN** an operator configures OAuth credential fields for tsnet onboarding
- **THEN** Sentinel accepts complete valid combinations and rejects incomplete combinations with actionable validation errors

#### Scenario: Documentation includes env interpolation examples
- **WHEN** an operator reads the configuration documentation
- **THEN** the docs include explicit examples for `${VAR_NAME}` interpolation in sink URLs and `SENTINEL_` prefixed overrides, including environment-only docker usage
