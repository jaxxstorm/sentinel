# sentinel-cli-config Specification

## Purpose
TBD - created by archiving change implement-basic-functionality. Update Purpose after archive.

## Requirements
### Requirement: Sentinel provides a standard command surface
Sentinel SHALL expose the commands `run`, `status`, `diff`, `dump-netmap`, `test-notify`, and `validate-config`.

#### Scenario: CLI shows expected commands
- **WHEN** an operator runs `sentinel --help`
- **THEN** the command list includes all required Sentinel commands

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
Sentinel SHALL load configuration from YAML or JSON files and MUST apply environment variable overrides using the `SENTINEL_` prefix, and operator documentation SHALL include working configuration examples that reflect this behavior.

#### Scenario: Environment overrides file value
- **WHEN** `poll_interval` is set in config file and `SENTINEL_POLL_INTERVAL` is also set
- **THEN** Sentinel uses the environment-provided value at runtime

#### Scenario: Documentation includes env interpolation examples
- **WHEN** an operator reads the configuration documentation
- **THEN** the docs include explicit examples for `${VAR_NAME}` interpolation in sink URLs and `SENTINEL_` prefixed overrides
