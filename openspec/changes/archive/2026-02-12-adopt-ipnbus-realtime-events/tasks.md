## 1. Realtime source foundation

- [x] 1.1 Add an IPNBus-backed source component that opens `WatchIPNBus` with initial state/netmap options
- [x] 1.2 Define source output translation from bus notifications into Sentinel `source.Netmap` updates
- [x] 1.3 Add tests for initial netmap bootstrap behavior from bus notifications

## 2. Watch lifecycle resilience

- [x] 2.1 Implement watch-stream reconnect loop with bounded exponential backoff
- [x] 2.2 Ensure reconnect loop exits cleanly on context cancellation/shutdown
- [x] 2.3 Add tests covering transient watch errors and successful resubscription

## 3. Runtime pipeline integration

- [x] 3.1 Integrate realtime source updates into existing snapshot normalization and diff flow
- [x] 3.2 Ensure no-op updates (unchanged normalized hash) skip detector/notification work
- [x] 3.3 Add tests for realtime peer online/offline transitions through diff/policy pipeline

## 4. Config compatibility and mode behavior

- [x] 4.1 Preserve existing config keys and defaults while enabling realtime observation as default source behavior
- [x] 4.2 Add/adjust source-mode validation and fallback behavior for polling compatibility paths
- [x] 4.3 Update config examples and validation tests to reflect realtime-first behavior

## 5. Logging and observability

- [x] 5.1 Add structured logs for bus-notification processing with stable fields
- [x] 5.2 Keep existing event JSON logging for derived Sentinel events in realtime mode
- [x] 5.3 Add tests/assertions for expected realtime log records and field stability

## 6. Notification delivery guarantees

- [x] 6.1 Verify bus-derived events route through existing notifier sink selection and policy/idempotency checks
- [x] 6.2 Ensure stdout/debug default sink behavior remains active for realtime events when external sinks are unavailable
- [x] 6.3 Add notifier/runtime tests for sink delivery and duplicate suppression with realtime-derived events

## 7. CLI/runtime command behavior

- [x] 7.1 Ensure `run` uses realtime watch semantics in steady state
- [x] 7.2 Ensure `run --once` has deterministic behavior under realtime source startup
- [x] 7.3 Ensure `dump-netmap` and status flows remain compatible with realtime source changes

## 8. Documentation and rollout

- [x] 8.1 Document realtime IPNBus operation model and reconnect behavior in project docs
- [x] 8.2 Document source-mode and compatibility expectations for operators
- [x] 8.3 Add migration/rollback notes for reverting to polling behavior if required

## 9. End-to-end verification

- [x] 9.1 Add integration test covering startup -> bus subscription -> event emission -> sink output
- [x] 9.2 Add integration test for watch interruption and recovery without process restart
- [x] 9.3 Run full test suite and verify no regressions in existing onboarding/presence flows
