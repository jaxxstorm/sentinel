## ADDED Requirements

### Requirement: Sentinel SHALL start with environment-only configuration
Sentinel SHALL support runtime startup with no YAML/JSON config file when sufficient `SENTINEL_` environment variables are provided.

#### Scenario: Environment-only startup without config file
- **WHEN** an operator runs Sentinel in a container with no mounted config file and provides required runtime values via environment variables
- **THEN** Sentinel loads configuration successfully and starts normal command execution

### Requirement: Sentinel SHALL support structured configuration values from environment variables
Sentinel SHALL accept structured configuration sections from environment variables for list/object fields that cannot be represented reliably as scalar overrides.

#### Scenario: Structured notifier config from environment
- **WHEN** an operator provides notifier sink and route structures via documented `SENTINEL_` environment variables
- **THEN** Sentinel parses and validates the structured values and wires notifier behavior equivalently to file-based config

### Requirement: Environment-driven configuration parsing SHALL be deterministic and validated
Sentinel SHALL provide deterministic parsing and explicit validation errors for malformed environment-sourced structured configuration values.

#### Scenario: Invalid structured environment value fails with actionable error
- **WHEN** an operator supplies malformed structured config content in an environment variable
- **THEN** Sentinel fails config loading with an error that identifies the environment key and validation/parsing reason

### Requirement: Environment-driven config SHALL preserve compatibility with existing file-based configuration
Environment-only configuration support SHALL NOT break existing YAML/JSON config loading, file-based defaults, or `${VAR}` placeholder expansion in sink URLs.

#### Scenario: Existing file-based config behavior remains valid
- **WHEN** an operator continues using YAML/JSON config with optional env overrides
- **THEN** Sentinel behavior remains backward compatible with existing configuration semantics
