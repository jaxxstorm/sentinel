## MODIFIED Requirements

### Requirement: Sentinel SHALL document configuration format comprehensively
Sentinel documentation SHALL describe the configuration schema, defaults, sink and route configuration, and environment-variable interpolation behavior used at runtime, including `stdout/debug`, webhook, and Discord sink examples, and SHALL include complete environment-only configuration guidance for container deployments.

#### Scenario: Operator can configure sinks from docs alone
- **WHEN** an operator follows the configuration reference in `docs/`
- **THEN** the operator can configure `stdout/debug`, webhook, and Discord sinks including environment-backed webhook URLs

#### Scenario: Operator can run Sentinel with env vars only
- **WHEN** an operator follows docker/container configuration docs
- **THEN** they can deploy Sentinel without mounting a config file by using documented `SENTINEL_` environment variables, including complex structured config examples

### Requirement: Sentinel SHALL provide operator troubleshooting guidance
Sentinel documentation SHALL include practical troubleshooting guidance for common runtime issues including missing webhook deliveries, idempotency suppression, sink connectivity failures, and malformed environment-sourced config values.

#### Scenario: Env config troubleshooting flow is documented
- **WHEN** an operator provides invalid structured env configuration
- **THEN** docs provide concrete checks for env key naming, expected structure, parse failures, and validation errors
