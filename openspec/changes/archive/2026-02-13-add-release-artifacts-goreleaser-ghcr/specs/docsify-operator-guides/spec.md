## MODIFIED Requirements

### Requirement: Sentinel SHALL document configuration format comprehensively
Sentinel documentation SHALL describe the configuration schema, defaults, sink and route configuration, and environment-variable interpolation behavior used at runtime, including `stdout/debug`, webhook, and Discord sink examples.

#### Scenario: Operator can configure sinks from docs alone
- **WHEN** an operator follows the configuration reference in `docs/`
- **THEN** the operator can configure `stdout/debug`, webhook, and Discord sinks including environment-backed webhook URLs

## ADDED Requirements

### Requirement: Sentinel SHALL provide release artifact workflow documentation
Sentinel documentation SHALL describe how release binaries and Docker images are produced, including GoReleaser usage, GitHub Actions workflow responsibilities, and GHCR publish behavior.

#### Scenario: Operator can follow release documentation to publish artifacts
- **WHEN** an operator follows the release documentation in `docs/`
- **THEN** the operator can identify the tag trigger, locate binary artifacts in GitHub Releases, and locate Docker images in GHCR
