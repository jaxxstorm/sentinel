## Context

Sentinel is a new daemon/CLI that embeds a Tailscale node through `tsnet` and observes netmap changes for alerting. The initial proposal commits to a v0 that detects peer online/offline transitions, while preserving a stable extension point for additional netmap-derived change types. The design must balance operator-friendly output, reliable deduplicated notifications, and clear boundaries between observation, diffing, policy, and delivery.

## Goals / Non-Goals

**Goals:**

- Deliver a production-usable v0 for presence change detection on top of periodic netmap snapshots.
- Establish a modular architecture: netmap source, snapshot normalization, diff detectors, rules/policy, notifier pipeline, and sinks.
- Guarantee restart-safe deduplication via persisted snapshot and idempotency state.
- Provide consistent CLI and config behavior with Cobra + Viper, including YAML/JSON and `SENTINEL_` env overrides.
- Support human-friendly pretty output and machine-friendly JSON logging without changing runtime behavior.

**Non-Goals:**

- Implement all future detectors (routes, endpoints, tags, exit-node, hostinfo, key indicators) in v0.
- Build a full interactive UI; output remains formatted terminal text only.
- Introduce distributed coordination or shared remote state in v0.
- Guarantee exactly-once delivery to external sinks; v0 targets at-least-once with idempotency keys.

## Decisions

### 1. Layered runtime pipeline with explicit interfaces

Sentinel will run a single pipeline with clear interfaces:
`NetmapSource -> SnapshotStore -> DiffEngine -> PolicyEngine -> Notifier -> Sink(s)`.

- `NetmapSource` is mockable and hides `tsnet`/local-client specifics.
- `DiffEngine` emits typed versioned events, independent of notification transport.
- `PolicyEngine` owns debounce, suppression windows, rate limiting, and batching.
- `Notifier` handles sink fanout, retries, and idempotency metadata.

Why this over a monolithic loop:
- Easier unit testing and isolated failure handling.
- Future detectors and sinks can be added without redesign.

### 2. Poll-based netmap observation with jitter and backoff

v0 uses periodic polling of the netmap from the embedded Tailscale node.

- Configurable `poll_interval`.
- Small randomized jitter prevents synchronized polling bursts across many instances.
- Backoff on source errors to reduce failure amplification.

Alternative considered:
- Push/event-driven netmap updates. Rejected for v0 due to higher complexity and weaker control over retry behavior.

### 3. Snapshot normalization before detector execution

Raw netmap objects include volatile fields that can cause noisy diffs. Sentinel will normalize snapshots into stable internal records before hashing and detection.

- v0 presence detector uses stable peer identity fields and online status.
- Volatile network/path details are excluded from v0 presence normalization.
- Normalized snapshots are hashed for before/after references in events.

Alternative considered:
- Diff raw structs directly. Rejected due to high false-positive risk and unstable event identity.

### 4. Detector plugin contract with deterministic ordering

`DiffEngine` executes a configured ordered list of detectors.

- Interface shape: `Detect(before, after) -> []Event`.
- Each detector owns one change family (v0: presence).
- Detector order is explicit in config for deterministic outputs.

Alternative considered:
- One detector handling all change types. Rejected because it couples unrelated logic and slows independent evolution.

### 5. Versioned event envelope and derived idempotency key

All emitted events use a shared envelope (`schema_version`, `event_id`, `event_type`, `severity`, `timestamp`, `subject`, `before_hash`, `after_hash`, payload). Notification idempotency key is deterministic from event type + subject + before/after fingerprints.

Alternative considered:
- Random UUID-only event IDs. Rejected because restart/retry dedupe would be unreliable.

### 6. Local persistent state abstraction with file-backed default

State is abstracted behind a store interface with a local file-backed implementation in v0.

- Persists last normalized snapshot.
- Persists recently sent idempotency keys (with TTL) for dedupe across restarts.
- Writes are atomic (temp file + rename) to avoid partial-state corruption.

Alternative considered:
- In-memory only state. Rejected because duplicates after restarts violate reliability goals.

### 7. Policy-first notification flow

Policies execute before sink delivery:

- Debounce/hysteresis per detector/type.
- Suppression windows.
- Rate limiting and batching.
- Dry-run mode logs/prints intended notifications but does not call sinks.

Alternative considered:
- Implement policy independently per sink. Rejected due to inconsistent behavior and duplicated logic.

### 8. Unified CLI/config and logging/output modes

- Cobra command tree: `run`, `status`, `diff`, `dump-netmap`, `test-notify`, `validate-config`.
- Viper loads YAML or JSON config, then applies env overrides with `SENTINEL_` prefix.
- Zap logger supports `pretty` (console encoder) and `json` formats.
- Lipgloss styles user-facing terminal output; `NO_COLOR` disables ANSI consistently.

Alternative considered:
- Separate logging and output frameworks per command. Rejected to keep behavior consistent and testable.

## Risks / Trade-offs

- [Polling misses short-lived transitions between intervals] -> Mitigate with configurable short intervals and documented trade-offs for high-churn tailnets.
- [Normalization may exclude fields needed by future detectors] -> Keep raw snapshot access available to detectors and version normalization rules.
- [File-backed state may contend on slow disks] -> Use compact state shape, atomic writes, and bounded idempotency key retention.
- [Suppression/rate-limit rules can hide important events] -> Add explicit suppressed counters/log fields and per-rule observability.
- [At-least-once sink delivery can still duplicate externally] -> Require idempotency keys in notifier contract and include keys in sink payloads where possible.

## Migration Plan

1. Add foundational packages and interfaces (`source`, `snapshot`, `diff`, `policy`, `notify`, `state`) behind internal contracts.
2. Implement v0 presence detector and event envelope; validate with unit tests over normalized snapshots.
3. Implement file-backed state store and notifier idempotency checks; verify restart behavior with integration tests.
4. Implement CLI/config surface and logging/output mode switches; add config validation command.
5. Roll out with `--dry-run` in real environments first, tune debounce/suppression, then enable active sinks.
6. Rollback strategy: disable sinks (or enable dry-run), keep polling and diffing active for diagnostics, and revert to previous binary/config if needed.

## Open Questions

- What should the default `poll_interval` be for v0 (`5s`, `10s`, or `30s`) given expected tailnet sizes?
- Should v0 severity be fixed for presence events or configurable per sink routing rule?
- Which sink should be first-class in v0 (webhook vs Slack) for strongest end-to-end validation?
- How long should idempotency keys be retained by default to balance dedupe and state size?
