## Context

Sentinel currently supports container execution but does not ship a canonical Docker Compose deployment template that operators can reuse for both local development and Railway onboarding. As a result, operators manually assemble environment variables, volume/state mappings, and notifier configuration, which creates drift between local and hosted setups.

The requested outcome is a Compose-first deployment model that:
- works locally with source builds,
- can be drag-and-dropped into Railway as a template,
- covers all supported `SENTINEL_` runtime environment variables with clear required/optional expectations, and
- keeps secrets out of committed files.

## Goals / Non-Goals

**Goals:**
- Provide a Railway-importable Compose template as the canonical deployment model.
- Provide a local Compose workflow that builds Sentinel from source.
- Define a complete environment variable matrix in the template/docs, with clear required vs optional semantics.
- Ensure secret values are sourced from external env/secret stores instead of inline YAML literals.
- Ensure release Docker publishing keeps a predictable `latest` tag suitable for Railway defaults.

**Non-Goals:**
- Re-architect Sentinel runtime config parsing or login-mode validation behavior.
- Add a new secret manager integration inside Sentinel.
- Replace existing release/versioned image tagging strategy.

## Decisions

### 1. Use a two-file Compose strategy
- `docker-compose.yml` is the canonical, Railway-importable template and defaults to GHCR image usage.
- `docker-compose.local.yml` overlays local behavior (source build + local tags/volumes) for developer runs.

Rationale:
- Railway template import is most reliable with a single, straightforward base file.
- Local source builds and Railway image pulls have different needs; separating them avoids conditional complexity.

Alternatives considered:
- Single compose file containing both `build` and `image`.
  - Rejected due ambiguous platform behavior and reduced clarity for template users.

### 2. Represent env vars explicitly with required/optional semantics
- Compose template includes all currently supported runtime `SENTINEL_` variables as explicit entries.
- Variables are marked/documented as required vs optional; default mode expects auth-key onboarding and therefore requires auth credentials for that default path.
- Optional variables default to empty or safe defaults, preserving startup behavior where possible.

Rationale:
- Operators can audit the full config surface in one place.
- Railway variable setup becomes deterministic.

Alternatives considered:
- Minimal env set only.
  - Rejected because it hides supported configuration and makes template extension error-prone.

### 3. Keep secrets out of committed template values
- No secret literal is committed in compose YAML.
- Local workflow uses `.env` (ignored by git) and ships `.env.example` placeholders.
- Railway guidance uses platform-managed variables/secrets.

Rationale:
- Prevents accidental credential leakage while keeping setup ergonomic.

Alternatives considered:
- Inline placeholder secret values directly in compose file.
  - Rejected due copy/paste leakage risk.

### 4. Maintain a `latest` image in GHCR for template defaults
- Release/publish workflow requirements are extended so template users can point to a maintained `latest` tag while still supporting explicit version pinning.

Rationale:
- Railway templates benefit from a sensible default image tag.
- Advanced users can pin immutable version tags.

Alternatives considered:
- Version-only tags.
  - Rejected because template setup becomes more complex for first-time users.

## Risks / Trade-offs

- [Risk] `latest` tag mutability can surprise operators.
  - Mitigation: document pinning by version tag for deterministic rollouts.
- [Risk] Environment variable surface drifts as new config fields are added.
  - Mitigation: keep template/env matrix in sync with config docs and add checklist tasks for updates.
- [Risk] Operators may misread mode-specific required credentials.
  - Mitigation: document required variables per login mode with explicit examples.

## Migration Plan

1. Add compose template and local overlay files plus `.env.example`.
2. Update docs for local compose usage, Railway template import, and env/secret setup.
3. Update GHCR workflow behavior/documentation for `latest` support.
4. Validate local compose path (`docker compose` config/build/run) and release workflow validation.
5. Rollback: remove compose artifacts and revert workflow/doc updates if unexpected deployment issues appear.

## Open Questions

- Should Railway template defaults use `latest` or a pinned semver tag with manual updates?
- Should the template include optional Discord sink wiring by default, or keep notifier defaults to stdout-only for safest bootstrapping?
- Do we want CI coverage that validates compose template env variable completeness against config documentation?
