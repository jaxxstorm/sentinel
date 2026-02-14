## ADDED Requirements

### Requirement: Sentinel SHALL provide a Docker Compose template for local and Railway deployment
Sentinel SHALL provide a repository-managed Docker Compose template that can be imported by Railway and SHALL provide a local compose overlay for source-build development workflows.

#### Scenario: Railway can import the compose template
- **WHEN** an operator imports the repository compose template into Railway
- **THEN** Railway can parse the compose configuration and deploy Sentinel using the documented image and environment variable inputs

#### Scenario: Local compose workflow builds from source
- **WHEN** an operator runs the documented local compose command with the local overlay
- **THEN** Docker Compose builds the Sentinel image from the local repository and starts the service with compatible runtime settings

### Requirement: Compose template SHALL expose complete runtime environment variable coverage
The compose template SHALL expose all documented runtime `SENTINEL_` environment variables with explicit required/optional semantics and defaults aligned to Sentinel runtime behavior.

#### Scenario: Required variables are identified for default auth-key onboarding
- **WHEN** an operator reviews the compose template and accompanying env example
- **THEN** required variables for default auth-key onboarding are clearly identified and optional values are separated

#### Scenario: Optional variables can remain unset
- **WHEN** an operator omits optional compose environment values
- **THEN** compose configuration remains valid and Sentinel starts with default behavior or mode-specific validation errors only where expected

### Requirement: Compose template SHALL keep secret values externalized
The compose template SHALL NOT embed live secret literals and MUST guide operators to source secrets from `.env`/environment injection locally and platform-managed variables in Railway.

#### Scenario: Repository templates do not contain secret literals
- **WHEN** an operator inspects committed compose and env example files
- **THEN** files contain placeholders/documentation only and no usable credential material

#### Scenario: Railway secret injection path is documented
- **WHEN** an operator follows Railway template setup guidance
- **THEN** secret values are configured through Railway variables/secrets rather than committed compose YAML values
