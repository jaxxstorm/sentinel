# tailscale-auth-key-management Specification

## Purpose
TBD - created by archiving change add-tailscale-node-auth-onboarding. Update Purpose after archive.
## Requirements
### Requirement: Sentinel SHALL accept auth key from flag, env, or config
Sentinel SHALL accept a Tailscale auth key from CLI flag, environment variable, or configuration file.

#### Scenario: Auth key is resolved from configured sources
- **WHEN** an operator provides an auth key in configuration
- **THEN** Sentinel uses that key for enrollment if a higher-precedence source is not present

### Requirement: Sentinel SHALL enforce auth key precedence
Sentinel SHALL resolve auth key precedence as CLI flag over environment variable over configuration file.

#### Scenario: CLI flag overrides environment and config
- **WHEN** auth key values are present in flag, env, and config
- **THEN** Sentinel uses the flag value and ignores lower-precedence sources

### Requirement: Sentinel SHALL validate and redact auth key material
Sentinel SHALL validate auth key presence/format before enrollment and MUST redact raw auth key values from logs, status, and error messages.

#### Scenario: Validation failure does not leak key
- **WHEN** Sentinel rejects an auth key as malformed
- **THEN** emitted logs and errors contain no raw auth key content

### Requirement: Sentinel SHALL support configurable fallback policy
Sentinel SHALL support a configuration switch controlling whether interactive login fallback is permitted after auth key failure.

#### Scenario: Interactive fallback is disabled
- **WHEN** auth key enrollment fails and fallback is disabled
- **THEN** Sentinel reports auth failure and does not initiate interactive login

