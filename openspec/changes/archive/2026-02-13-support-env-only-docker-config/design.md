## Context

Sentinel currently supports partial `SENTINEL_` environment overrides through Viper, but complex configuration sections (notifier sinks/routes, detector maps, ordered lists) are still primarily file-oriented. This blocks container-first deployment models where operators prefer to inject all runtime configuration via environment variables only.

The goal is to make environment-based configuration complete and deterministic without breaking existing YAML/JSON workflows, existing `SENTINEL_` scalar overrides, or `${VAR}` placeholder expansion in sink URLs.

## Goals / Non-Goals

**Goals:**
- Support running Sentinel with no config file by providing configuration entirely via env vars.
- Support deterministic env encoding for structured fields (maps/lists/objects) used by notifier and detector config.
- Preserve backward compatibility for file-based config and existing env override behavior.
- Provide clear parse/validation errors for malformed structured env values.
- Document environment-only docker/platform deployment patterns comprehensively.

**Non-Goals:**
- Replacing Viper for scalar overrides.
- Introducing a new external config service.
- Changing eventing, diffing, onboarding, or sink delivery semantics unrelated to config ingestion.

## Decisions

### 1) Introduce explicit structured env keys for complex config sections
Use dedicated env keys for sections that are not reliably representable as scalar overrides:
- `SENTINEL_DETECTORS` (JSON object)
- `SENTINEL_DETECTOR_ORDER` (JSON array)
- `SENTINEL_NOTIFIER_SINKS` (JSON array)
- `SENTINEL_NOTIFIER_ROUTES` (JSON array)

These keys are parsed after Viper unmarshal and before final validation.

Alternatives considered:
- Per-item indexed keys (`SENTINEL_NOTIFIER_SINKS_0_NAME`): rejected due to high complexity and poor operator ergonomics.
- Single full-config env blob only: rejected because it reduces composability with existing scalar overrides.

### 2) Preserve scalar override precedence while layering structured env parsing
Keep existing precedence model:
1. defaults
2. config file (if present)
3. scalar `SENTINEL_` overrides via Viper
4. structured env keys for complex sections

Structured env keys become authoritative for their sections when present, allowing file-less configuration while retaining selective override flexibility.

Alternatives considered:
- Structured keys before scalar overrides: rejected because current override expectations would change.

### 3) Add explicit env parse error surface with key attribution
Structured env parsing failures return actionable errors that include the env key name and parse/validation reason. This is required for platform debugging where users often only see startup logs.

Alternatives considered:
- Silent fallback to defaults on parse failure: rejected because it can mask misconfiguration in production.

### 4) Keep file-based behavior fully compatible
YAML/JSON loading remains unchanged. Existing `${VAR}` interpolation in sink URLs still works after merge. Environment-only mode is additive, not a replacement.

Alternatives considered:
- Forcing env-only mode with a flag: rejected to avoid introducing a breaking deployment mode split.

### 5) Expand docs around env-only deployment contract
Document:
- supported structured env keys and JSON shapes
- scalar key mapping examples
- `docker run -e` and platform env examples
- troubleshooting for parse/validation errors

## Risks / Trade-offs

- **[Risk] JSON-in-env quoting errors are common** -> **Mitigation:** include copy/paste-safe examples and explicit troubleshooting guidance.
- **[Risk] Ambiguous precedence between scalar and structured env values** -> **Mitigation:** define and document strict precedence order and apply it consistently in tests.
- **[Risk] Validation logic may reject legacy edge-case configs when env sections are present** -> **Mitigation:** add regression tests for current file-based behavior and mixed-mode configs.

## Migration Plan

1. Add structured env parsing helpers in `internal/config` and integrate into `Load` path.
2. Add/adjust validation and test coverage for env-only, mixed mode, and malformed structured values.
3. Update docs for environment-only container deployment and troubleshooting.
4. Validate using `validate-config`, unit tests, and docker examples.

Rollback strategy:
- Revert structured env parsing integration while retaining existing scalar overrides and file-based config behavior.
- Keep docs aligned with whichever behavior is active after rollback.

## Open Questions

- Should we also support an optional single `SENTINEL_CONFIG_JSON` full-document override for advanced users?
- Should structured env keys accept YAML in addition to JSON, or JSON only for deterministic parsing?
- Do we want a `sentinel print-effective-config` command in a follow-up change for easier runtime diagnostics?
