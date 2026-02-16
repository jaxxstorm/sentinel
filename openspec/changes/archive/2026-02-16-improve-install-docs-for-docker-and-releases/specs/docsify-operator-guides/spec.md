## MODIFIED Requirements

### Requirement: Sentinel SHALL document configuration format comprehensively
Sentinel documentation SHALL describe the configuration schema, defaults, sink and route configuration, and environment-variable interpolation behavior used at runtime, including `stdout/debug`, webhook, and Discord sink examples, SHALL include complete environment-only configuration guidance for container deployments, SHALL include a Docker Compose plus Railway template environment variable matrix that identifies required and optional variables, and SHALL present operator-oriented command examples using installed binary and Docker workflows while keeping source-based `go run` examples scoped to development guidance.

#### Scenario: Operator can configure sinks from docs alone
- **WHEN** an operator follows the configuration reference in `docs/`
- **THEN** the operator can configure `stdout/debug`, webhook, and Discord sinks including environment-backed webhook URLs

#### Scenario: Operator can run Sentinel with env vars only
- **WHEN** an operator follows docker/container configuration docs
- **THEN** they can deploy Sentinel without mounting a config file by using documented `SENTINEL_` environment variables, including complex structured config examples

#### Scenario: Operator can configure compose template variables correctly
- **WHEN** an operator follows compose and Railway template documentation
- **THEN** the operator can identify required vs optional variables and configure the template without guesswork

#### Scenario: Operator-facing examples avoid source-only prerequisites
- **WHEN** an operator reads runtime configuration and execution examples in user-facing docs
- **THEN** examples use `sentinel` binary or Docker commands rather than requiring `go run`

### Requirement: Sentinel SHALL provide release artifact workflow documentation
Sentinel documentation SHALL describe how release binaries and Docker images are produced, including GoReleaser usage, GitHub Actions workflow responsibilities, GHCR publish behavior, how the Docker `latest` tag relates to compose template defaults, and how operators install and run Sentinel from GitHub Release binaries.

#### Scenario: Operator can follow release documentation to publish artifacts
- **WHEN** an operator follows the release documentation in `docs/`
- **THEN** the operator can identify the tag trigger, locate binary artifacts in GitHub Releases, and locate Docker images in GHCR

#### Scenario: Operator understands compose image tag defaults
- **WHEN** an operator reads release and compose deployment docs
- **THEN** the operator understands when `latest` is used and how to pin a versioned image tag for deterministic rollouts

#### Scenario: Operator can install Sentinel from GitHub Releases
- **WHEN** an operator follows installation instructions in docs
- **THEN** the operator can download the correct release asset for their platform, extract it, and run `sentinel version`

### Requirement: Sentinel SHALL maintain a concise repository README
Sentinel SHALL provide a root `README.md` that gives a concise overview, installation and quick-start pathways for release binary and Docker usage, development-specific source-run guidance, and links to detailed Docsify documentation in plain technical style.

#### Scenario: README links to docs and quick start
- **WHEN** a new contributor opens the repository root
- **THEN** they can find quick-start instructions and links into the `docs/` site without reading long-form narrative text

#### Scenario: README presents multiple run paths clearly
- **WHEN** an operator reads the README quick-start section
- **THEN** they can choose between GitHub Release binary, Docker, and source development workflows without ambiguity
