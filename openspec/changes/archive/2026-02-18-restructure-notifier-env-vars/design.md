## Context

Sentinel supports both structured notifier env vars (`SENTINEL_NOTIFIER_SINKS`, `SENTINEL_NOTIFIER_ROUTES`) and shorthand composite env vars that append a route/sink. The shorthand surface has drifted into inconsistent naming, especially for route sinks (`SENTINEL_NOTIFIER_ROUTE_SINKS` vs `SENTINEL_NOTIFIER_SINK`) and route event types (`SENTINEL_NOTIFIER_ROUTE_EVENT_TYPES` vs `SENTINEL_NOTIFIER_ROUTE_EVENT_TYPE`), which creates operator confusion and misconfiguration risk.

This change standardizes shorthand naming while preserving current deployments that still use legacy aliases.

## Goals / Non-Goals

**Goals:**
- Define and document canonical shorthand notifier env var names.
- Keep legacy shorthand aliases working during migration.
- Enforce deterministic precedence when canonical and legacy aliases are both set.
- Keep existing structured-env behavior and file-based config compatibility intact.

**Non-Goals:**
- Removing legacy shorthand aliases in this change.
- Changing notifier routing semantics beyond env key resolution.
- Replacing structured JSON env overrides with a new encoding format.

## Decisions

### 1) Canonical naming model for shorthand route fields
Route-scoped shorthand keys are canonical only when prefixed with `SENTINEL_NOTIFIER_ROUTE_`.

Canonical examples:
- `SENTINEL_NOTIFIER_ROUTE_EVENT_TYPES`
- `SENTINEL_NOTIFIER_ROUTE_SINKS`
- `SENTINEL_NOTIFIER_ROUTE_SEVERITIES`
- `SENTINEL_NOTIFIER_ROUTE_FILTER_INCLUDE_*`
- `SENTINEL_NOTIFIER_ROUTE_FILTER_EXCLUDE_*`

Sink-definition shorthand keys remain:
- `SENTINEL_NOTIFIER_SINK_NAME`
- `SENTINEL_NOTIFIER_SINK_TYPE`
- `SENTINEL_NOTIFIER_SINK_URL`

Alternatives considered:
- Keep dual naming indefinitely with no canonical set: rejected because it preserves ambiguity.
- Rename every notifier env var in one breaking cut: rejected due to migration cost.

### 2) Legacy shorthand aliases remain supported but secondary
Legacy aliases continue to parse for compatibility:
- `SENTINEL_NOTIFIER_ROUTE_EVENT_TYPE` (alias for `...EVENT_TYPES`)
- `SENTINEL_NOTIFIER_SINK` (alias for `...ROUTE_SINKS`)

Resolver behavior is deterministic:
- If canonical is present, use canonical.
- Else if alias is present, use alias.
- Else leave unset and apply existing defaults.

Alternatives considered:
- Hard-fail when alias is used: rejected as unnecessarily breaking.

### 3) Keep current precedence layering and append behavior
The existing config layering is preserved:
1. defaults
2. file config
3. scalar env overrides
4. structured env overrides
5. shorthand composite append

Shorthand alias cleanup does not change append semantics; it only changes which key names are canonical.

Alternatives considered:
- Move shorthand earlier in precedence: rejected because it would alter existing effective config behavior.

### 4) Documentation follows canonical names only
Operator docs and examples switch to canonical shorthand keys. Legacy alias names remain documented only in migration notes/troubleshooting context.

Alternatives considered:
- Continue showing both names in main examples: rejected because it keeps ambiguity visible in first-run paths.

## Risks / Trade-offs

- **[Risk] Existing deployments may rely on alias names that become less visible in docs** -> **Mitigation:** keep aliases functional and provide explicit migration notes from alias to canonical keys.
- **[Risk] Mixed canonical+alias environments could produce surprising outcomes** -> **Mitigation:** enforce deterministic canonical-first precedence and cover with targeted tests.
- **[Risk] Users may still misunderstand structured-vs-shorthand behavior** -> **Mitigation:** keep precedence table explicit and include side-by-side examples in docs.

## Migration Plan

1. Introduce canonical/alias mapping constants in config parsing for shorthand route fields.
2. Implement canonical-first resolution for the alias pairs.
3. Add/adjust unit tests for canonical-only, alias-only, and mixed-key inputs.
4. Update configuration docs/examples to canonical names and add migration notes for aliases.
5. Validate with `sentinel validate-config` and config package tests.

Rollback strategy:
- Revert parser and docs changes together to restore pre-change alias behavior and examples.

## Open Questions

- Should alias usage emit a deprecation warning in a follow-up change once warning plumbing is standardized across config load paths?
- Do we want a future `sentinel config explain-env` command to print canonical key mappings and precedence at runtime?
