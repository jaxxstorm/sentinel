## Why

Sentinel logs currently mix multiple formats (Zap structured logs, raw tsnet/tailscaled log lines, and sink JSON), which makes operational parsing noisy and inconsistent. Shutdown behavior on `Ctrl+C` is also unreliable, which can leave the runtime in an unclear termination state.

## What Changes

- Standardize runtime logging so Sentinel-managed records and embedded Tailscale runtime records follow a single structured logging contract.
- Add a stable `log_source` field (for example `sentinel`, `tailscale`, `sink`) so operators can distinguish origin without relying on message text.
- Remove empty `error_code` and `error_class` fields from onboarding/enrollment log records unless those fields have meaningful non-empty values.
- Normalize how event payload lines are emitted so sink output and runtime logs remain machine-readable without mixed formatting surprises.
- Fix signal handling for `Ctrl+C`/termination so `run` exits cleanly and predictably with proper context cancellation and shutdown logs.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `sentinel-output-and-logging`: require unified structured format and stable `log_source` attribution across runtime log emitters.
- `realtime-notification-and-logging`: require bus/event processing logs to include source attribution and consistent field conventions with other runtime logs.
- `tailscale-enrollment-status-reporting`: require omission of empty error classification fields and consistent structured output for onboarding status transitions.
- `sentinel-cli-config`: require reliable interrupt handling semantics for `run` so operator-initiated shutdown is graceful and deterministic.

## Impact

- Affected code: logging setup/factory, onboarding status logging, realtime source logging, notifier stdout/debug sink behavior, and CLI runtime signal handling.
- Behavioral impact: log records become more uniform and easier to ingest; termination via `Ctrl+C` becomes consistent for local and automated operation.
- APIs/dependencies: no new external dependencies expected; changes are in formatting, field population, and shutdown orchestration.
