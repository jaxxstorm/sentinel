## Why

Operators currently have to hand-assemble local container runtime settings and Railway deployment settings, which makes onboarding fragile and inconsistent. A first-party Docker Compose template should provide a safe default for local testing and a Railway-importable model with explicit required vs optional environment inputs.

## What Changes

- Add a repository-managed Docker Compose template that is valid for Railway import and structured for local use.
- Define a local overlay/build workflow so local development builds the image from source while preserving a deployable Railway-oriented compose model.
- Define required vs optional environment variable handling in the template, including a required auth key path for default auth-key onboarding.
- Define secret handling guidance and template conventions so secret values are not committed in plaintext.
- Update release image publishing behavior so Railway can track an intentionally maintained `latest` image tag alongside versioned tags.
- Update operator docs with end-to-end usage for local compose and Railway template import.

## Capabilities

### New Capabilities
- `docker-compose-deployment-template`: Standardize Compose-based local and Railway deployment templates, environment variable coverage, required/optional semantics, and secret-safe usage patterns.

### Modified Capabilities
- `docsify-operator-guides`: Add operator documentation for local compose usage, Railway template import, environment variable matrix, and secret management conventions.
- `ghcr-docker-publish-workflow`: Extend Docker publish requirements to include maintaining a `latest` tag suitable for Railway template defaults.

## Impact

- Affected code/config: `docker-compose.yml` (and local override file if used), optional `.env.example`, docs under `docs/`, and GitHub Actions workflow(s) responsible for Docker image publication.
- Affected systems: local Docker Compose runtime and Railway template import/deploy flow.
- Dependencies/APIs: no new runtime dependencies; relies on existing Docker/Buildx/GHCR and Railway compose support.
