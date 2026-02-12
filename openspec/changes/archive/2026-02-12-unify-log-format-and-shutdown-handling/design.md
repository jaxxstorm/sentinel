## Context

Sentinel currently emits logs from multiple origins with different formats:

- Sentinel runtime logs use Zap pretty/JSON formatting.
- Embedded Tailscale/tsnet logs still appear as raw stdlib log lines (for example `2026/02/12 ...`), without structured fields.
- Notifier stdout/debug output emits raw event JSON lines that do not include runtime log metadata.

This makes ingestion and troubleshooting harder, especially when correlating onboarding transitions, bus updates, and sink deliveries.

The current `run` command also uses `context.Background()` without signal-bound cancellation, so `Ctrl+C` can end with abrupt process termination (`signal: interrupt`) instead of a deterministic graceful shutdown path.

## Goals / Non-Goals

**Goals:**

- Standardize runtime log formatting across Sentinel and embedded Tailscale log emitters.
- Add stable origin attribution via `log_source` for all runtime records.
- Omit empty enrollment error fields (`error_code`, `error_class`) unless values are present.
- Preserve machine-readable event output while making source attribution explicit.
- Ensure `Ctrl+C` and termination signals trigger graceful shutdown (context cancellation, clean exit path, no noisy fatal output).

**Non-Goals:**

- Changing Sentinel event schema (`event.Event`) or notification routing semantics.
- Replacing Zap with a different logger.
- Implementing log shipping/aggregation features beyond local process output.
- Redesigning detector/policy behavior.

## Decisions

### 1. Route tsnet logs through Sentinel logging adapters

Decision:

- Configure `tsnet.Server.UserLogf` and `tsnet.Server.Logf` in runtime wiring.
- Implement small adapters that forward tsnet log lines into Zap with `log_source: "tailscale"`.
- Use level mapping:
  - `UserLogf` at `INFO` (operator-relevant)
  - `Logf` at `DEBUG` (backend/noisy diagnostics)

Rationale:

- tsnet exposes native hooks, so we avoid global `log.SetOutput` side effects and keep source-local control.
- Guarantees formatting consistency with existing `pretty`/`json` modes.

Alternatives considered:

- Capturing global stdlib logger output via `log.SetOutput`: too broad and can affect unrelated packages.
- Keeping tsnet logs raw: fails the unified format requirement.

### 2. Introduce a consistent `log_source` field contract

Decision:

- Add `log_source` field to runtime log records:
  - `sentinel` for core runtime/pipeline logs
  - `tailscale` for tsnet/local backend logs
  - `sink` for notifier sink emission records

Rationale:

- Operators can filter by origin in both pretty and JSON modes without parsing message text.
- Keeps field semantics stable as logging grows.

Alternatives considered:

- Encoding source in message prefixes only: harder to parse and less stable.

### 3. Suppress empty enrollment error fields

Decision:

- Build enrollment logging fields conditionally so `error_code` and `error_class` are included only when non-empty.
- Apply this behavior to onboarding status and enrollment completion/failure logs.

Rationale:

- Removes noisy empty-string fields while preserving useful diagnostics when errors exist.

Alternatives considered:

- Keep empty fields for strict key presence: rejected because signal-to-noise is poor and user explicitly requested omission.

### 4. Normalize sink stdout/debug output with source attribution

Decision:

- Keep notification payload machine-readable JSON, but wrap/annotate with `log_source: "sink"` and sink identity fields.
- Ensure behavior remains compatible when no external webhook sink is configured.

Rationale:

- Preserves the “always visible event output” property while making stream origin explicit.

Alternatives considered:

- Leave sink output as bare event JSON: preserves backward behavior but keeps mixed-format ambiguity.

### 5. Make signal handling explicit and graceful in CLI run path

Decision:

- Use `signal.NotifyContext` in `run` command (and other long-lived commands if needed) for `os.Interrupt` and `SIGTERM`.
- Pass that context to `Runner.Run`/`RunOnce`.
- Treat context cancellation due interrupt as graceful exit (no error stack/no `signal: interrupt` tail output).

Rationale:

- Aligns user expectation: `Ctrl+C` should cancel work and exit cleanly.
- Prevents noisy “poll cycle failed” logs during shutdown.

Alternatives considered:

- Rely on default process interrupt handling: non-deterministic and currently noisy.

### 6. Keep compatibility with existing log-format controls

Decision:

- No config key changes for format selection (`pretty` vs `json` remains).
- `log_source` becomes additive; existing fields and message names remain stable where possible.

Rationale:

- Minimizes migration cost and avoids breaking downstream parsing pipelines unnecessarily.

Alternatives considered:

- Introduce a new logging schema/version flag: unnecessary for this scoped cleanup.

## Risks / Trade-offs

- [Risk: Increased log volume from `tailscale` debug stream] -> Mitigation: map backend `Logf` to debug level and honor current log-level filtering.
- [Risk: Existing consumers of bare sink JSON may break if wrapping changes] -> Mitigation: preserve event payload shape and document transition; add compatibility tests.
- [Risk: Conditional field emission could impact dashboards expecting always-present keys] -> Mitigation: document that empty keys are intentionally omitted and keep key names unchanged when values exist.
- [Risk: Signal-context cancellation could surface context errors in nested loops] -> Mitigation: treat interrupt-driven context cancellation as normal shutdown path in CLI and runner.

## Migration Plan

1. Add logging adapters and wire `tsnet.Server.UserLogf` / `Logf` in runtime construction.
2. Add `log_source` fields across Sentinel runtime records and sink emission logs.
3. Update onboarding status logging to conditionally include error fields only when populated.
4. Implement signal-bound contexts in `run` command and normalize cancellation handling.
5. Add/update tests for:
   - formatted tailscale-source logs
   - `log_source` field presence and stability
   - omission of empty `error_code`/`error_class`
   - graceful `Ctrl+C`/interrupt exit behavior
6. Update operator docs/example output to reflect unified format and source attribution.

Rollback:

- Revert to previous logger wiring (remove tsnet adapters and source annotations).
- Restore previous unconditional enrollment field behavior if required.
- Revert signal-context changes in CLI run path.

## Open Questions

- Should sink output remain a standalone JSON event line plus a separate log line, or become a single wrapped record only?
- Should `tailscale` backend debug logs be fully disabled in pretty mode unless `--log-level=debug` is set (strict opt-in)?
- Do we want to include a per-record correlation key (for example cycle id) in this change, or defer to a follow-up?
