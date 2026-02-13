## Context

Sentinel currently supports `stdout` and `webhook` notifier sinks. The notification and routing pipeline already includes idempotency, retries, and structured logging. Discord usage today requires operators to use a generic webhook sink, which lacks sink-specific payload structure and validation guarantees. This change adds a native `discord` sink without changing existing routing/policy behavior.

## Goals / Non-Goals

**Goals:**
- Add a first-class `discord` sink type to `notifier.sinks[]`.
- Deliver routed events to Discord webhooks with readable message structure.
- Reuse existing notifier semantics: route matching, idempotency suppression, retry/backoff, and sink-level delivery logs.
- Keep configuration compatible with current sink model and env interpolation behavior.

**Non-Goals:**
- Building a bidirectional Discord bot or slash-command integration.
- Supporting advanced Discord features (threads, attachments, role mentions, embeds customization per route).
- Changing event generation, detector behavior, or policy semantics.

## Decisions

### 1) Implement Discord as a dedicated sink implementation in `internal/notify`
- Add `DiscordSink` that satisfies the existing `notify.Sink` interface.
- Reuse the same HTTP client timeout and retry profile as `WebhookSink` for consistent behavior.

Alternatives considered:
- Reusing `WebhookSink` directly with no new type: rejected because it cannot enforce Discord-specific payload shape or config validation.
- Provider abstraction shared by webhook/discord: deferred as unnecessary for initial scope.

### 2) Use Discord webhook JSON payload with stable, compact text content
- Send webhook payload with top-level `content` containing a concise event summary.
- Include event identity, subject, and key payload fields in a deterministic text format.
- Keep the original Sentinel notification JSON available in stdout/debug logs; Discord message focuses on operator readability.

Alternatives considered:
- Full embed-based rich payload: deferred to keep initial implementation simple and robust.
- Raw Sentinel JSON payload as message body: rejected due to poor readability and Discord message constraints.

### 3) Extend sink config validation with sink-type-specific rules
- Accept `type: discord` as valid sink type.
- Require non-empty URL for discord sink entries.
- Keep URL env interpolation path unchanged (via existing config loading behavior).

Alternatives considered:
- Soft-skip invalid discord sinks like placeholder webhook URLs: rejected for missing URL because this is a deterministic config error.

### 4) Keep routing behavior unchanged
- Existing `notifier.routes[].event_types` and `severities` filtering behavior remains unchanged.
- `discord` is simply another sink target in route `sinks` arrays.

Alternatives considered:
- Introducing discord-only route keys: rejected as unnecessary complexity.

## Risks / Trade-offs

- **[Risk] Discord API response semantics differ from generic webhooks** -> **Mitigation:** treat any non-2xx as failure with retry; log sink name and status code.
- **[Risk] Message length overflow for large payload fields** -> **Mitigation:** truncate payload summary in Discord content while retaining full structured event in normal sink/debug logs.
- **[Risk] Validation strictness can break existing configs if users switch sink type incorrectly** -> **Mitigation:** provide explicit validation errors and config examples.

## Migration Plan

1. Add config schema/validation support for `discord` sink type.
2. Implement `DiscordSink` and wire it in CLI sink construction.
3. Add unit tests for discord payload generation, retry behavior, and notifier routing.
4. Update docs and `config.example.yaml` with a Discord sink example.
5. Rollout with no breaking default behavior; existing `stdout`/`webhook` configs continue to work unchanged.

Rollback:
- Remove Discord sink entries from config to disable usage immediately.
- Revert code paths for `discord` sink wiring if needed; existing sink types are independent.

## Open Questions

- Should we support Discord embeds in v1, or keep plain `content` only?
- Should Discord sink support optional username/avatar overrides in config?
- Should sink-level payload truncation limits be configurable?
