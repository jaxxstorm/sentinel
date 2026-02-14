## Context

Sentinel onboarding currently resolves one credential type (auth key) and applies it to `tsnet.Server.AuthKey` before enrollment. This leaves two operator gaps:
1. No first-class way to request ACL tags (`AdvertiseTags`) during enrollment.
2. No first-class way to supply OAuth client-secret credentials for tsnet-managed auth-key generation flows.

Because onboarding is a startup gate for the entire poll/diff pipeline, these changes affect config loading, runtime wiring, enrollment mode selection, and status observability.

## Goals / Non-Goals

**Goals:**
- Support configuring Tailscale advertise tags for enrollment and apply them before onboarding starts.
- Support OAuth client-secret credential inputs (client secret and related identifiers) for tsnet onboarding.
- Define deterministic precedence between auth-key and OAuth credentials, with explicit validation errors.
- Expose onboarding credential source/mode in status/logging for operator debugging.
- Preserve backward compatibility for existing auth-key-only users.

**Non-Goals:**
- Implementing broader OAuth token exchange logic outside fields supported directly by `tsnet.Server`.
- Managing long-term secret rotation or external secret-manager integrations.
- Changing notification, diff, or detector pipelines.

## Decisions

### 1. Add explicit tsnet credential and tag config fields
- Add `tsnet.advertise_tags` (array of strings).
- Add OAuth credential fields mapped to tsnet server capabilities:
  - `tsnet.client_secret`
  - `tsnet.client_id` (optional, where required)
  - `tsnet.id_token` / `tsnet.audience` (optional advanced fields)
- Add documented `SENTINEL_` env mappings for these fields.

Rationale:
- Keeps behavior explicit and config-driven.
- Aligns directly with `tsnet.Server` fields to avoid custom auth abstractions.

Alternatives considered:
- Parse a provider-specific OAuth credentials file shape only.
  - Rejected: too provider-specific and ambiguous across environments.
- Single opaque JSON blob only.
  - Rejected: harder validation and discoverability.

### 2. Introduce credential resolution precedence for onboarding
- Retain existing auth key precedence: flag > `SENTINEL_TAILSCALE_AUTH_KEY` > config auth key.
- Add OAuth credential resolution after auth-key sources in `auto` mode.
- Extend onboarding mode semantics to include OAuth-first execution when explicitly configured (either via a new `oauth` mode or explicit credential-available path in `auto`).

Rationale:
- Auth-key behavior must remain stable for existing users.
- OAuth credentials should be first-class without silently overriding explicit auth keys.

Alternatives considered:
- Prefer OAuth over auth key in `auto`.
  - Rejected: would be a surprising behavior change for current deployments.

### 3. Apply advertise tags at tsnet server construction time
- Set `tsnet.Server.AdvertiseTags` during runtime wiring before any `LocalClient()`/status calls.
- Validate tags use Tailscale tag format (`tag:<name>`) to fail fast with actionable errors.

Rationale:
- `tsnet` consumes these fields at startup/enrollment time.
- Early validation prevents ambiguous control-plane rejections.

Alternatives considered:
- Apply tags after startup through API mutation.
  - Rejected: not supported by the current onboarding abstraction and risks racey behavior.

### 4. Expand status and logging for credential clarity
- Extend status output to include credential mode/source (auth key vs OAuth) and whether advertise tags are configured.
- Keep secrets redacted and never print raw client secret/auth key values.

Rationale:
- Operators need clear diagnostics when onboarding does not behave as expected.

## Risks / Trade-offs

- [Risk] Added config surface increases validation complexity.
  - Mitigation: strict field-level validation and targeted config tests for each credential source.
- [Risk] OAuth credential fields may be provided incompletely, causing confusing startup failures.
  - Mitigation: explicit validation errors that name missing companion fields.
- [Risk] Tag assignment may still be rejected by tailnet policy despite valid syntax.
  - Mitigation: preserve onboarding status classification and remediation guidance indicating policy-side rejection.
- [Risk] Startup behavior drift in `auto` mode.
  - Mitigation: codify precedence/order in specs and tests (existing state > auth key > OAuth > interactive).

## Migration Plan

1. Add new tsnet config fields, env wiring, and validation behind backward-compatible defaults.
2. Update onboarding resolver/wiring to apply tags and OAuth credentials before first probe.
3. Add/adjust tests for config loading, precedence, onboarding mode selection, and redaction safety.
4. Update docs (`configuration`, `docker-image`, and examples) with new env-only and file-based patterns.
5. Rollback strategy: revert to auth-key-only resolution path and ignore new OAuth/tag fields while keeping existing keys intact.

## Open Questions

- Should explicit `tsnet.login_mode=oauth` be introduced, or should OAuth remain an implicit branch under `auto` only?
- Do we require `client_id` whenever `client_secret` is set, or allow secret-only for environments where tsnet can resolve ID externally?
- Should we support `*_FILE` env conventions for secrets in this change, or defer to a follow-up?
