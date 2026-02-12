## Why

Sentinel needs a concrete v0 baseline so implementation can begin with useful tailnet monitoring instead of ad hoc experiments. Defining this now creates a clear contract for detecting peer presence changes while preserving an extension path for broader netmap change detection.

## What Changes

- Introduce a v0 netmap observation flow that snapshots netmap state, diffs successive snapshots, and emits typed events for peer online/offline transitions.
- Introduce a notifier pipeline with sink routing, dry-run behavior, rate limiting, batching, and suppression to reduce noisy alerts.
- Introduce persistent state requirements for last-seen snapshots and idempotency keys so restarts do not duplicate notifications.
- Introduce Sentinel CLI and config requirements using Cobra and Viper, including YAML/JSON config, env overrides, and command coverage for run/status/diff/testing flows.
- Introduce output and logging requirements for pretty human output (Lipgloss) and optional structured JSON logs (Zap), with deterministic no-color behavior.
- Introduce extensibility requirements so new detectors (routes, exit nodes, tags, endpoints, hostinfo, key indicators) can be added without redesign.

## Capabilities

### New Capabilities

- `netmap-presence-diffing`: Observe tailnet netmap snapshots and emit versioned typed events for peer online/offline changes.
- `notification-delivery-pipeline`: Route, suppress, batch, and deliver events to notification sinks with idempotency guarantees and dry-run support.
- `sentinel-cli-config`: Define Sentinel CLI commands, global flags, config parsing (YAML/JSON), and environment variable override behavior.
- `sentinel-output-and-logging`: Define pretty human output conventions, JSON logging mode, and safe-by-default redaction behavior.
- `extensible-netmap-detectors`: Define the plugin-style detector contract for adding new netmap change types beyond v0 presence events.

### Modified Capabilities

- None.

## Impact

- Affected code: new Sentinel command surface and core packages for netmap observation, snapshot/state storage, diff detection, event schema, policy/rules, and notifier sinks.
- APIs/contracts: canonical event schema and notifier interface (including idempotency key handling).
- Dependencies: `tailscale.com/tsnet`, `spf13/cobra`, `spf13/viper`, `uber-go/zap`, and `charmbracelet/lipgloss`.
- Systems/operations: daemon runtime behavior, persisted local state, log/notification outputs, and safe handling of tailnet-derived metadata.
