# sentinel-cli-config Specification

## Purpose
TBD - created by archiving change implement-basic-functionality. Update Purpose after archive.

## Requirements
### Requirement: Sentinel provides a standard command surface
Sentinel SHALL expose the commands `run`, `status`, `diff`, `dump-netmap`, `test-notify`, `validate-config`, and `version`.

#### Scenario: CLI shows expected commands
- **WHEN** an operator runs `sentinel --help`
- **THEN** the command list includes all required Sentinel commands

#### Scenario: Version command is available as a first-class command
- **WHEN** an operator runs `sentinel version --help`
- **THEN** Sentinel returns command help for `version` without requiring runtime config or tsnet startup

### Requirement: All commands accept shared config and logging flags
Every Sentinel command SHALL accept `--config`, and global flags SHALL include `--log-format`, `--log-level`, and `--no-color`.

#### Scenario: Global flags are available on subcommands
- **WHEN** an operator runs `sentinel status --help`
- **THEN** help output includes `--config`, `--log-format`, `--log-level`, and `--no-color`

### Requirement: Run command supports dry-run and one-shot execution
The `run` command SHALL support `--dry-run` and `--once` execution modes, and SHALL handle interrupt and termination signals with graceful context cancellation semantics.

#### Scenario: One-shot run exits after one poll cycle
- **WHEN** an operator runs `sentinel run --once`
- **THEN** Sentinel performs one observe/diff cycle and exits with a final status code

#### Scenario: Interrupt triggers graceful shutdown
- **WHEN** an operator sends `Ctrl+C` (`SIGINT`) while `sentinel run` is active
- **THEN** Sentinel cancels runtime context, performs graceful shutdown handling, and exits without an unstructured `signal: interrupt` tail error line

### Requirement: Config loading supports YAML, JSON, and environment overrides
Sentinel SHALL load configuration from YAML or JSON files, MUST apply environment variable overrides using the `SENTINEL_` prefix, SHALL support environment-only configuration for full runtime operation (including structured fields), SHALL validate expanded event type values including wildcard `*` for route event matching, SHALL validate sink type-specific configuration for supported sink types including `discord`, and operator documentation SHALL include working configuration examples that reflect this behavior.

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

#### Scenario: Documentation includes env interpolation examples
- **WHEN** an operator reads the configuration documentation
- **THEN** the docs include explicit examples for `${VAR_NAME}` interpolation in sink URLs and `SENTINEL_` prefixed overrides, including environment-only docker usage

### Requirement: Version command SHALL report standardized build metadata
The `sentinel version` command SHALL print version, commit, and build timestamp fields derived from the runtime version metadata model.

#### Scenario: Version command reports release metadata
- **WHEN** an operator runs `sentinel version` on a release binary
- **THEN** output includes non-empty version, commit hash, and build timestamp values consistent with build metadata

#### Scenario: Version command reports fallback metadata for local builds
- **WHEN** an operator runs `sentinel version` on a local untagged build
- **THEN** output includes documented fallback values rather than missing fields
