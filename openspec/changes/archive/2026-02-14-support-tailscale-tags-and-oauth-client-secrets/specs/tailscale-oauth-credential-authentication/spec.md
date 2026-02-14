## ADDED Requirements

### Requirement: Sentinel SHALL accept OAuth client credentials for tsnet onboarding
Sentinel SHALL accept OAuth credential fields needed by tsnet onboarding from configuration and `SENTINEL_` environment variables, including client secret and companion identity fields.

#### Scenario: OAuth client credentials are resolved from environment
- **WHEN** an operator provides valid OAuth credential environment variables and does not provide an auth key
- **THEN** Sentinel config loading succeeds and onboarding wiring uses OAuth credential inputs for tsnet enrollment

### Requirement: Sentinel SHALL validate and redact OAuth secret material
Sentinel SHALL validate required OAuth credential combinations before enrollment and MUST redact raw OAuth secret values from logs, status, and errors.

#### Scenario: Missing required OAuth companion field fails validation
- **WHEN** an operator provides a client secret without required companion fields
- **THEN** Sentinel fails configuration or onboarding initialization with an actionable validation error and no secret leakage
