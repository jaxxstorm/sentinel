## ADDED Requirements

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
The `run` command SHALL support `--dry-run` and `--once` execution modes.

#### Scenario: One-shot run exits after one poll cycle
- **WHEN** an operator runs `sentinel run --once`
- **THEN** Sentinel performs one observe/diff cycle and exits with a final status code

### Requirement: Config loading supports YAML, JSON, and environment overrides
Sentinel SHALL load configuration from YAML or JSON files and MUST apply environment variable overrides using the `SENTINEL_` prefix.

#### Scenario: Environment overrides file value
- **WHEN** `poll_interval` is set in config file and `SENTINEL_POLL_INTERVAL` is also set
- **THEN** Sentinel uses the environment-provided value at runtime

