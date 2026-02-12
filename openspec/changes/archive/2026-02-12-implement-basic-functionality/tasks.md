## 1. Project scaffolding and core contracts

- [x] 1.1 Create internal package structure for `source`, `snapshot`, `diff`, `policy`, `notify`, `state`, and shared `event` types
- [x] 1.2 Define core interfaces (`NetmapSource`, `Detector`, `PolicyEngine`, `Notifier`, `StateStore`) and wire constructors
- [x] 1.3 Add build-time dependency wiring for `tsnet`, `cobra`, `viper`, `zap`, and `lipgloss`

## 2. Configuration and runtime settings

- [x] 2.1 Define Sentinel config schema for poll interval, detector toggles/order, sink routes, debounce/suppression, rate limiting, batching, and state store
- [x] 2.2 Implement config loading for YAML and JSON files through Viper
- [x] 2.3 Implement `SENTINEL_` environment override mapping and precedence tests
- [x] 2.4 Add config validation logic shared by startup and `validate-config` command

## 3. Netmap observation and snapshot persistence

- [x] 3.1 Implement `tsnet`-backed netmap source adapter behind `NetmapSource`
- [x] 3.2 Implement poll loop with configurable interval, jitter, and error backoff
- [x] 3.3 Implement snapshot normalization for v0 presence detection fields
- [x] 3.4 Implement snapshot hashing and persistence of last normalized snapshot in the state store

## 4. Diff engine and detector framework

- [x] 4.1 Implement diff engine that executes enabled detectors in deterministic configured order
- [x] 4.2 Implement v0 `PresenceDetector` for online/offline transitions only
- [x] 4.3 Implement detector enable/disable behavior from configuration
- [x] 4.4 Add detector registration mechanism to support future detector plugins without core pipeline changes

## 5. Event model and idempotency

- [x] 5.1 Implement versioned canonical event envelope fields (`event_id`, `event_type`, `timestamp`, `subject_id`, `before_hash`, `after_hash`)
- [x] 5.2 Implement deterministic idempotency key derivation from event identity and snapshot fingerprints
- [x] 5.3 Implement persisted idempotency-key retention with TTL and duplicate suppression checks

## 6. Policy engine and notification delivery

- [x] 6.1 Implement policy evaluation stage for debounce/hysteresis, suppression windows, rate limiting, and batching
- [x] 6.2 Implement notifier routing engine that maps events to configured sinks by event type/severity
- [x] 6.3 Implement dry-run behavior that reports intended deliveries and performs no outbound sink calls
- [x] 6.4 Implement first sink adapter (webhook) with retry/backoff and idempotency key propagation

## 7. CLI commands and operator UX

- [x] 7.1 Implement Cobra root command with global flags `--config`, `--log-format`, `--log-level`, and `--no-color`
- [x] 7.2 Implement `run` command with `--dry-run` and `--once` modes
- [x] 7.3 Implement `status`, `diff`, `dump-netmap`, `test-notify`, and `validate-config` commands
- [x] 7.4 Implement pretty human output formatting via Lipgloss with deterministic styling rules
- [x] 7.5 Implement `NO_COLOR` and `--no-color` handling to strip ANSI sequences

## 8. Logging, redaction, and observability

- [x] 8.1 Implement Zap logger factory for `pretty` and `json` modes with stable structured fields
- [x] 8.2 Implement default redaction rules for sensitive peer metadata in both output modes
- [x] 8.3 Add counters/timing metrics for polling, diffs, emitted events, notifications, suppressions, and state store errors

## 9. State store reliability

- [x] 9.1 Implement file-backed state store with atomic write semantics (temp file + rename)
- [x] 9.2 Add corruption/recovery handling for partially written or invalid state files
- [x] 9.3 Add restart behavior tests proving snapshot continuity and notification dedupe across process restarts

## 10. Verification and acceptance testing

- [x] 10.1 Add unit tests for presence diff scenarios (transition emits event, unchanged state emits none)
- [x] 10.2 Add unit tests for detector ordering and enable/disable configuration behavior
- [x] 10.3 Add unit tests for policy controls (debounce, suppression, rate limit, batching)
- [x] 10.4 Add integration tests for CLI config precedence (file + env) and command flag availability
- [x] 10.5 Add end-to-end dry-run test for `sentinel run --once --dry-run` showing observe -> diff -> route pipeline without sink delivery
