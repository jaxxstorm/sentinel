## Why

Sentinel currently supports stdout and generic webhooks, but not a first-class Discord sink. Operators that use Discord for incident visibility need a direct, supported sink type that can be configured and routed like other sinks.

## What Changes

- Add a new notification sink type `discord` that posts Sentinel events to a configured Discord webhook.
- Define Discord sink payload formatting so key event fields (event type, subject, severity, timestamp, and payload summary) are easy to read in a channel.
- Add configuration support and validation for Discord sink fields in `notifier.sinks[]`.
- Ensure Discord sink participates in existing routing, idempotency, retry/backoff, and logging behavior.
- Update docs and config examples to show Discord sink setup and routing.

## Capabilities

### New Capabilities
- `discord-notification-sink`: Native Discord sink behavior, payload formatting, delivery semantics, and observability.

### Modified Capabilities
- `notification-delivery-pipeline`: Extend sink support and routing expectations to include Discord delivery.
- `sentinel-cli-config`: Extend sink configuration schema/validation and examples for Discord sink configuration.

## Impact

- Affected code: `internal/notify` sink implementations and wiring, `internal/config` validation/defaults, CLI wiring and docs.
- External integration: outbound HTTPS requests to Discord webhook endpoints.
- Testing: add unit tests for sink send behavior and notifier routing, plus config validation tests for discord sink entries.
