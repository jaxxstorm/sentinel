# docsify-operator-guides Specification

## Purpose
Defines operator documentation requirements for Sentinel using a Docsify-renderable docs site and a concise root README.

## Requirements
### Requirement: Sentinel SHALL provide Docsify-renderable operator documentation
Sentinel SHALL provide a documentation site under `docs/` that can be rendered by Docsify without additional conversion steps.

#### Scenario: Docsify entrypoint is present
- **WHEN** an operator serves the repository with Docsify tooling
- **THEN** the `docs/` content renders a navigable documentation site

### Requirement: Sentinel SHALL document configuration format comprehensively
Sentinel documentation SHALL describe the configuration schema, defaults, sink and route configuration, and environment-variable interpolation behavior used at runtime, including `stdout/debug`, webhook, and Discord sink examples, and SHALL include complete environment-only configuration guidance for container deployments.

#### Scenario: Operator can configure sinks from docs alone
- **WHEN** an operator follows the configuration reference in `docs/`
- **THEN** the operator can configure `stdout/debug`, webhook, and Discord sinks including environment-backed webhook URLs

#### Scenario: Operator can run Sentinel with env vars only
- **WHEN** an operator follows docker/container configuration docs
- **THEN** they can deploy Sentinel without mounting a config file by using documented `SENTINEL_` environment variables, including complex structured config examples

### Requirement: Sentinel SHALL provide release artifact workflow documentation
Sentinel documentation SHALL describe how release binaries and Docker images are produced, including GoReleaser usage, GitHub Actions workflow responsibilities, and GHCR publish behavior.

#### Scenario: Operator can follow release documentation to publish artifacts
- **WHEN** an operator follows the release documentation in `docs/`
- **THEN** the operator can identify the tag trigger, locate binary artifacts in GitHub Releases, and locate Docker images in GHCR

### Requirement: Sentinel SHALL provide operator troubleshooting guidance
Sentinel documentation SHALL include practical troubleshooting guidance for common runtime issues including missing webhook deliveries, idempotency suppression, sink connectivity failures, and malformed environment-sourced config values.

#### Scenario: Webhook troubleshooting flow is documented
- **WHEN** an operator observes sink output without webhook delivery
- **THEN** docs provide concrete checks for endpoint health, runtime logs, retries, and idempotency state

#### Scenario: Env config troubleshooting flow is documented
- **WHEN** an operator provides invalid structured env configuration
- **THEN** docs provide concrete checks for env key naming, expected structure, parse failures, and validation errors

### Requirement: Sentinel SHALL maintain a concise repository README
Sentinel SHALL provide a root `README.md` that gives a concise overview, quick-start commands, and links to detailed Docsify documentation in plain technical style.

#### Scenario: README links to docs and quick start
- **WHEN** a new contributor opens the repository root
- **THEN** they can find quick-start instructions and links into the `docs/` site without reading long-form narrative text
