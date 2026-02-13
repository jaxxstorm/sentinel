## 1. Discord Sink Implementation

- [x] 1.1 Add a new `DiscordSink` implementation in `internal/notify` that satisfies the `Sink` interface.
- [x] 1.2 Implement Discord request payload rendering with stable event context fields and payload summary text.
- [x] 1.3 Implement bounded retry/backoff and structured success/failure logging for Discord sends.

## 2. Notifier Wiring and Routing

- [x] 2.1 Wire `discord` sink construction into CLI notifier setup alongside existing sink types.
- [x] 2.2 Ensure notifier routing can target Discord sinks using existing route matching behavior.
- [x] 2.3 Ensure idempotency behavior remains unchanged for Discord deliveries.

## 3. Configuration and Validation

- [x] 3.1 Extend sink type validation to accept `discord` and reject unsupported sink types.
- [x] 3.2 Add sink-type-specific config validation for Discord webhook URL requirements.
- [x] 3.3 Update `config.example.yaml` with a Discord sink example and route configuration.

## 4. Tests

- [x] 4.1 Add unit tests for Discord sink payload shape and successful delivery handling.
- [x] 4.2 Add unit tests for Discord sink retry/error paths and logging-visible outcomes.
- [x] 4.3 Add config validation tests for valid/invalid Discord sink entries.
- [x] 4.4 Add notifier integration tests confirming routed events are delivered to Discord sinks.

## 5. Documentation

- [x] 5.1 Update sink documentation to include Discord configuration and behavior.
- [x] 5.2 Update troubleshooting docs with Discord delivery validation and failure diagnosis steps.
