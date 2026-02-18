## Why

Notifier shorthand environment variables are currently inconsistent and ambiguous (for example `SENTINEL_NOTIFIER_SINK` vs `SENTINEL_NOTIFIER_ROUTE_SINKS`), which makes route wiring error-prone and hard to debug. We need a single, predictable naming model so operators can configure notifier sinks and routes reliably in env-only deployments.

## What Changes

- Define canonical shorthand notifier env var names with consistent namespace and plurality rules.
- Keep legacy shorthand aliases temporarily for backward compatibility, but document and treat them as deprecated.
- Define deterministic precedence and conflict behavior when canonical and legacy aliases are both set.
- Update config validation and operator-facing documentation to use only canonical names in examples.

## Capabilities

### New Capabilities
- *(none)*

### Modified Capabilities
- `sentinel-cli-config`: Standardize shorthand notifier env var naming, parsing precedence, and compatibility behavior.
- `env-only-config-ingestion`: Ensure env-only notifier parsing remains deterministic when canonical and legacy alias keys coexist.

## Impact

- Affected code: `internal/config/config.go` notifier shorthand env parsing and validation paths.
- Affected tests: notifier/env parsing tests in `internal/config/config_test.go` and CLI wiring coverage.
- Affected docs: `docs/configuration.md`, docker-related configuration docs, and `.env` examples.
- Runtime behavior: clearer route/sink env mapping with explicit deprecation path for ambiguous legacy keys.
