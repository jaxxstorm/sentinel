## MODIFIED Requirements

### Requirement: Sentinel SHALL document configuration format comprehensively
Sentinel documentation SHALL describe the configuration schema, defaults, sink and route configuration, and environment-variable interpolation behavior used at runtime, including `stdout/debug`, webhook, and Discord sink examples, SHALL include complete environment-only configuration guidance for container deployments, and SHALL include a Docker Compose plus Railway template environment variable matrix that identifies required and optional variables.

#### Scenario: Operator can configure sinks from docs alone
- **WHEN** an operator follows the configuration reference in `docs/`
- **THEN** the operator can configure `stdout/debug`, webhook, and Discord sinks including environment-backed webhook URLs

#### Scenario: Operator can run Sentinel with env vars only
- **WHEN** an operator follows docker/container configuration docs
- **THEN** they can deploy Sentinel without mounting a config file by using documented `SENTINEL_` environment variables, including complex structured config examples

#### Scenario: Operator can configure compose template variables correctly
- **WHEN** an operator follows compose and Railway template documentation
- **THEN** the operator can identify required vs optional variables and configure the template without guesswork

### Requirement: Sentinel SHALL provide release artifact workflow documentation
Sentinel documentation SHALL describe how release binaries and Docker images are produced, including GoReleaser usage, GitHub Actions workflow responsibilities, GHCR publish behavior, and how the Docker `latest` tag relates to compose template defaults.

#### Scenario: Operator can follow release documentation to publish artifacts
- **WHEN** an operator follows the release documentation in `docs/`
- **THEN** the operator can identify the tag trigger, locate binary artifacts in GitHub Releases, and locate Docker images in GHCR

#### Scenario: Operator understands compose image tag defaults
- **WHEN** an operator reads release and compose deployment docs
- **THEN** the operator understands when `latest` is used and how to pin a versioned image tag for deterministic rollouts
