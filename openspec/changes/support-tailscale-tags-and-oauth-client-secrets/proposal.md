## Why

Sentinel can currently enroll with auth keys, but operators that rely on Tailscale ACL tags and OAuth-based credential flows cannot fully configure onboarding from Sentinel. Adding first-class support for advertise tags and OAuth client credentials removes manual post-enrollment steps and enables policy-compliant automation.

## What Changes

- Add onboarding configuration support for Tailscale node `advertise_tags` so Sentinel can request ACL tags during authentication.
- Add onboarding credential support for OAuth client-secret based auth flows, including config/env wiring and resolution behavior.
- Define deterministic precedence and validation rules when both auth keys and OAuth credentials are present.
- Extend runtime status and docs so operators can see which onboarding credential mode/source is active.

## Capabilities

### New Capabilities
- `tailscale-oauth-credential-authentication`: Define OAuth client-secret credential inputs, validation, source precedence, and onboarding behavior.

### Modified Capabilities
- `tailscale-node-onboarding`: Extend onboarding behavior to apply configured advertise tags during enrollment.
- `tailscale-auth-key-management`: Clarify interaction and precedence boundaries between auth-key and OAuth credential onboarding sources.
- `sentinel-cli-config`: Add and validate new config/env keys for advertise tags and OAuth credential fields, including env-only container usage.

## Impact

- Affected code:
  - `internal/config` (new tsnet config fields and validation)
  - `internal/cli` (flag/env wiring and status output)
  - `internal/onboarding` (credential selection and provider setup)
  - `docs/` and `config.example.yaml` (operator-facing configuration guidance)
- External behavior:
  - Expanded onboarding configuration surface for auth and tag assignment.
  - Potentially new env vars for OAuth client-secret credential paths.
- Dependencies/systems:
  - Uses existing `tsnet.Server` credential and tag fields; no new external service dependency expected.
