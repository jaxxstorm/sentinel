## MODIFIED Requirements

### Requirement: Sentinel provides a standard command surface
Sentinel SHALL expose the commands `run`, `status`, `diff`, `dump-netmap`, `test-notify`, `validate-config`, and `version`.

#### Scenario: CLI shows expected commands
- **WHEN** an operator runs `sentinel --help`
- **THEN** the command list includes all required Sentinel commands

#### Scenario: Version command is available as a first-class command
- **WHEN** an operator runs `sentinel version --help`
- **THEN** Sentinel returns command help for `version` without requiring runtime config or tsnet startup

## ADDED Requirements

### Requirement: Version command SHALL report standardized build metadata
The `sentinel version` command SHALL print version, commit, and build timestamp fields derived from the runtime version metadata model.

#### Scenario: Version command reports release metadata
- **WHEN** an operator runs `sentinel version` on a release binary
- **THEN** output includes non-empty version, commit hash, and build timestamp values consistent with build metadata

#### Scenario: Version command reports fallback metadata for local builds
- **WHEN** an operator runs `sentinel version` on a local untagged build
- **THEN** output includes documented fallback values rather than missing fields
