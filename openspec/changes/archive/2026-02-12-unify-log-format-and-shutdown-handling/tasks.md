## 1. Unified Runtime Log Sources

- [x] 1.1 Add tsnet log adapters in runtime wiring using `tsnet.Server.UserLogf` and `tsnet.Server.Logf` to route embedded logs through Sentinel logger output.
- [x] 1.2 Add stable `log_source` fields to Sentinel runtime log records and ensure source values are consistent (`sentinel`, `tailscale`, `sink`).
- [x] 1.3 Update realtime bus-processing log calls to include source attribution while preserving existing structured fields used for diagnostics.

## 2. Enrollment Log Field Hygiene

- [x] 2.1 Refactor onboarding/enrollment log field construction to omit `error_code` when value is empty.
- [x] 2.2 Refactor onboarding/enrollment log field construction to omit `error_class` when value is empty.
- [x] 2.3 Add or update tests to assert empty enrollment error fields are not emitted while non-empty values are preserved.

## 3. Sink Output Normalization

- [x] 3.1 Define and implement sink output attribution strategy so stdout/debug sink-visible records can be identified with `log_source=sink`.
- [x] 3.2 Ensure sink output remains machine-readable JSON and does not regress default stdout/debug behavior when webhook sinks are unavailable.
- [x] 3.3 Add notifier/runtime tests that validate source-attributed sink output shape and compatibility with existing idempotency flow.

## 4. Graceful Signal Shutdown

- [x] 4.1 Update `run` command context lifecycle to use signal-aware cancellation (`SIGINT`/`SIGTERM`) instead of background context only.
- [x] 4.2 Ensure runner shutdown path treats interrupt-triggered context cancellation as graceful exit (no trailing unstructured interrupt error output).
- [x] 4.3 Add command/runtime tests for `Ctrl+C` behavior and verify deterministic termination semantics.

## 5. Docs and Regression Coverage

- [x] 5.1 Update operator-facing docs/config examples to describe unified log formatting, `log_source`, and empty error-field omission behavior.
- [x] 5.2 Add/update tests for mixed log-source formatting consistency across pretty and JSON modes.
- [x] 5.3 Run full test suite and verify no regressions in onboarding, realtime observation, notifier routing, and CLI command behavior.
